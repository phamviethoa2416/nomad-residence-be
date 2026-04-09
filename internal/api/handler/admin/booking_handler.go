package admin

import (
	"errors"
	"log/slog"
	"nomad-residence-be/internal/api/middlewares"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingUsecase port.BookingUsecase
	logger         *slog.Logger
}

func NewBookingHandler(bookingUsecase port.BookingUsecase, logger *slog.Logger) *BookingHandler {
	return &BookingHandler{bookingUsecase: bookingUsecase, logger: logger}
}

func (h *BookingHandler) ListBookings(c *gin.Context) {
	var req request.ListBookingsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	req.ApplyDefaults()
	if field, msg := req.Validate(); field != "" {
		c.JSON(400, gin.H{"error": msg, "field": field})
		return
	}

	f := filter.BookingFilter{
		Status:      entity.BookingStatus(req.Status),
		RoomID:      req.RoomID,
		GuestPhone:  req.GuestPhone,
		CheckinFrom: req.DateFrom,
		CheckinTo:   req.DateTo,
		Page:        req.Page,
		Limit:       req.Limit,
	}

	bookings, total, err := h.bookingUsecase.ListBookings(c.Request.Context(), f)
	if err != nil {
		h.logger.Error("list_bookings failed", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, gin.H{
		"bookings": bookings,
		"pagination": gin.H{
			"page":        req.Page,
			"limit":       req.Limit,
			"total":       total,
			"total_pages": (total + int64(req.Limit) - 1) / int64(req.Limit),
		},
	})
}

func (h *BookingHandler) GetBooking(c *gin.Context) {
	var params request.BookingIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	booking, err := h.bookingUsecase.GetBookingByID(c.Request.Context(), params.ID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("get_booking failed", slog.Uint64("id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, booking)
}

func (h *BookingHandler) ConfirmBooking(c *gin.Context) {
	var params request.BookingIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.ConfirmBookingRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	booking, err := h.bookingUsecase.ConfirmBooking(c.Request.Context(), params.ID, body.AdminNote)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("confirm_booking failed", slog.Uint64("id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	adminCtx := middlewares.GetAdminContext(c)
	if adminCtx == nil {
		c.JSON(401, gin.H{"error": "Vui lòng đăng nhập", "code": "UNAUTHORIZED"})
		return
	}

	h.logger.Info("booking_confirm_request",
		slog.Uint64("booking_id", uint64(booking.ID)),
		slog.String("booking_code", booking.BookingCode),
		slog.Uint64("admin_id", uint64(adminCtx.ID)),
	)

	c.JSON(200, gin.H{"data": booking, "message": "Đã xác nhận đơn"})
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	var params request.BookingIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.CancelBookingRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	booking, err := h.bookingUsecase.CancelBooking(c.Request.Context(), params.ID, body.Reason)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("cancel_booking failed", slog.Uint64("id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	adminCtx := middlewares.GetAdminContext(c)
	if adminCtx == nil {
		c.JSON(401, gin.H{"error": "Vui lòng đăng nhập", "code": "UNAUTHORIZED"})
		return
	}

	h.logger.Info("booking_cancel_request",
		slog.Uint64("booking_id", uint64(booking.ID)),
		slog.String("booking_code", booking.BookingCode),
		slog.Uint64("admin_id", uint64(adminCtx.ID)),
		slog.String("reason", body.Reason),
	)

	c.JSON(200, gin.H{"data": booking, "message": "Đã hủy đơn"})
}

func (h *BookingHandler) CreateManualBooking(c *gin.Context) {
	var body request.CreateManualBookingRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	body.ApplyDefaults()
	if field, msg := body.Validate(); field != "" {
		c.JSON(400, gin.H{"error": msg, "field": field})
		return
	}

	var guestEmail, guestNote *string
	if body.GuestEmail != "" {
		guestEmail = &body.GuestEmail
	}
	if body.GuestNote != "" {
		guestNote = &body.GuestNote
	}

	booking, err := h.bookingUsecase.CreateManualBooking(
		c.Request.Context(),
		body.RoomID,
		body.CheckinDate,
		body.CheckoutDate,
		body.GuestName,
		body.GuestPhone,
		guestEmail,
		guestNote,
		body.NumGuests,
		body.Source,
		body.AdminNote,
	)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("create_manual_booking failed", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(201, booking)
}
