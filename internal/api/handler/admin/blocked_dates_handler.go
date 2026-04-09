package admin

import (
	"fmt"
	"log/slog"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type BlockedDateHandler struct {
	roomUsecase port.RoomUsecase
	blockedRepo port.BlockedDateRepository
	bookingRepo port.BookingRepository
	logger      *slog.Logger
}

func NewBlockedDateHandler(
	roomUsecase port.RoomUsecase,
	blockedRepo port.BlockedDateRepository,
	bookingRepo port.BookingRepository,
	logger *slog.Logger,
) *BlockedDateHandler {
	return &BlockedDateHandler{
		roomUsecase: roomUsecase,
		blockedRepo: blockedRepo,
		bookingRepo: bookingRepo,
		logger:      logger,
	}
}

func (h *BlockedDateHandler) ListBlockedDates(c *gin.Context) {
	var params request.BlockedDateRoomParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var q request.ListBlockedDatesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if field, msg := q.Validate(); field != "" {
		c.JSON(400, gin.H{"error": msg, "field": field})
		return
	}

	from, err := utils.ParseDate(q.From)
	if err != nil {
		c.JSON(400, gin.H{"error": "Ngày from không hợp lệ"})
		return
	}

	to := from.AddDate(0, 3, 0)
	if q.To != "" {
		if t, err := utils.ParseDate(q.To); err == nil {
			to = t
		}
	}

	dates, err := h.roomUsecase.GetBlockedDates(c.Request.Context(), params.ID, from, to)
	if err != nil {
		h.logger.Error("list_blocked_dates failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, dates)
}

// BlockDates POST /api/v1/admin/rooms/:id/blocked-dates
// Kiểm tra xung đột với booking confirmed trước khi block.
func (h *BlockedDateHandler) BlockDates(c *gin.Context) {
	var params request.BlockedDateRoomParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.BlockDatesRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	dates := make([]time.Time, 0, len(body.Dates))
	for _, s := range body.Dates {
		t, err := utils.ParseDate(s)
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Ngày không hợp lệ: %s", s)})
			return
		}
		dates = append(dates, t)
	}

	// Kiểm tra xung đột với booking confirmed — port từ blockDates controller JS
	// Sort để lấy min/max cho query range
	minDate, maxDate := dates[0], dates[0]
	for _, d := range dates[1:] {
		if d.Before(minDate) {
			minDate = d
		}
		if d.After(maxDate) {
			maxDate = d
		}
	}

	conflicts, err := h.bookingRepo.FindConfirmedOverlapping(c.Request.Context(), params.ID, minDate, maxDate)
	if err != nil {
		h.logger.Error("block_dates: conflict check failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	if err := h.roomUsecase.BlockDates(c.Request.Context(), params.ID, dates, body.Reason); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("block_dates failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	msg := fmt.Sprintf("Đã block %d ngày", len(dates))
	if len(conflicts) > 0 {
		msg = fmt.Sprintf("Đã block %d ngày. Cảnh báo: %d booking bị xung đột!", len(dates), len(conflicts))
	}

	var conflictsData interface{}
	if len(conflicts) > 0 {
		conflictsData = conflicts
	}

	c.JSON(201, gin.H{
		"data": gin.H{
			"blocked":   len(dates),
			"conflicts": conflictsData,
		},
		"message": msg,
	})
}

// UnblockDate DELETE /api/v1/admin/blocked-dates/:id
// Không cho phép gỡ block của booking — phải hủy booking thay thế.
func (h *BlockedDateHandler) UnblockDate(c *gin.Context) {
	var params request.BlockedDateIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	date, err := h.blockedRepo.FindByID(c.Request.Context(), params.ID)
	if err != nil {
		h.logger.Error("unblock_date: get failed", slog.Uint64("id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}
	if date == nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy bản ghi", "code": "NOT_FOUND"})
		return
	}
	if date.Source == entity.BlockBooking {
		c.JSON(400, gin.H{
			"error": "Không thể gỡ block của đơn đặt phòng, hãy hủy đơn thay thế",
			"code":  "CANNOT_UNBLOCK_BOOKING",
		})
		return
	}

	if err := h.blockedRepo.DeleteByRoomAndDate(c.Request.Context(), date.RoomID, date.Date); err != nil {
		h.logger.Error("unblock_date: delete failed", slog.Uint64("id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, gin.H{"message": "Đã gỡ block ngày"})
}
