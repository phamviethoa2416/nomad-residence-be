package public

import (
	"log/slog"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type RoomHandler struct {
	roomUsecase    port.RoomUsecase
	pricingUsecase port.PricingUsecase
	logger         *slog.Logger
}

func NewRoomHandler(roomUsecase port.RoomUsecase, pricingUsecase port.PricingUsecase, logger *slog.Logger) *RoomHandler {
	return &RoomHandler{roomUsecase: roomUsecase, pricingUsecase: pricingUsecase, logger: logger}
}

func (h *RoomHandler) ListRooms(c *gin.Context) {
	var req request.ListRoomsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var minPrice *decimal.Decimal
	if req.MinPrice > 0 {
		v := decimal.NewFromInt(int64(req.MinPrice))
		minPrice = &v
	}
	var maxPrice *decimal.Decimal
	if req.MaxPrice > 0 {
		v := decimal.NewFromInt(int64(req.MaxPrice))
		maxPrice = &v
	}

	f := filter.RoomFilter{
		RoomType:  entity.RoomType(req.RoomType),
		MinGuests: req.Guests,
		MinPrice:  minPrice,
		MaxPrice:  maxPrice,
		Page:      req.Page,
		Limit:     req.Limit,
		Status:    entity.RoomStatusActive,
	}

	var checkin, checkout *time.Time
	if req.Checkin != "" && req.Checkout != "" {
		ci, err1 := utils.ParseDate(req.Checkin)
		co, err2 := utils.ParseDate(req.Checkout)
		if err1 == nil && err2 == nil {
			checkin = &ci
			checkout = &co
		}
	}

	rooms, total, err := h.roomUsecase.ListAvailableRooms(c.Request.Context(), f, checkin, checkout)
	if err != nil {
		h.logger.Error("failed to list available rooms", slog.Any("error", err))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 12
	}

	c.Header("Cache-Control", "public, max-age=60, stale-while-revalidate=30")
	c.JSON(200, gin.H{
		"rooms": rooms,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *RoomHandler) GetRoomDetail(c *gin.Context) {
	slug := c.Param("slug")

	var req request.GetRoomDetailQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	room, err := h.roomUsecase.GetRoomBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error("failed to get room detail", slog.Any("error", err))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var breakdown *entity.PriceBreakdown
	if req.Checkin != "" && req.Checkout != "" {
		ci, err1 := utils.ParseDate(req.Checkin)
		co, err2 := utils.ParseDate(req.Checkout)
		if err1 == nil && err2 == nil {
			_, breakdown, err = h.roomUsecase.GetRoomDetailWithPrice(c.Request.Context(), room.ID, &ci, &co)
			if err != nil {
				h.logger.Warn("price calculation failed", slog.Any("error", err))
				breakdown = nil
			}
		}
	}

	c.Header("Cache-Control", "public, max-age=60, stale-while-revalidate=30")
	c.JSON(200, gin.H{
		"room":          room,
		"price_details": breakdown,
	})
}

func (h *RoomHandler) GetRoomCalendar(c *gin.Context) {
	slug := c.Param("slug")

	var req request.GetRoomCalendarQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	room, err := h.roomUsecase.GetRoomBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error("failed to get room by slug", slog.Any("error", err))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	from := utils.TruncateToDate(time.Now())
	to := from.AddDate(0, 0, 90)

	if req.From != "" {
		if t, err := utils.ParseDate(req.From); err == nil {
			from = t
		}
	}
	if req.To != "" {
		if t, err := utils.ParseDate(req.To); err == nil {
			to = t
		}
	}

	calendar, err := h.roomUsecase.GetRoomCalendar(c.Request.Context(), room.ID, from, to)
	if err != nil {
		h.logger.Error("failed to get room calendar", slog.Any("error", err))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"room_id": room.ID,
		"slug":    room.Slug,
		"dates":   calendar,
	})
}

func (h *RoomHandler) GetRoomAvailability(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || parsedID == 0 {
		c.JSON(404, gin.H{"error": apperrors.ErrRoomNotFound.Message})
		return
	}
	roomID := uint(parsedID)

	var req request.GetRoomDetailQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Checkin == "" || req.Checkout == "" {
		c.JSON(400, gin.H{"error": "Vui lòng cung cấp checkin và checkout"})
		return
	}

	ci, err1 := utils.ParseDate(req.Checkin)
	co, err2 := utils.ParseDate(req.Checkout)
	if err1 != nil || err2 != nil {
		c.JSON(400, gin.H{"error": apperrors.ErrInvalidDates.Message})
		return
	}

	available, err := h.roomUsecase.CheckAvailability(c.Request.Context(), roomID, ci, co)
	if err != nil {
		h.logger.Error("failed to check room availability", slog.Any("error", err))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"available": available})
}
