package public

import (
	"errors"
	"log/slog"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingUsecase port.BookingUsecase
	logger         *slog.Logger
}

func NewBookingHandler(bookingUsecase port.BookingUsecase, logger *slog.Logger) *BookingHandler {
	return &BookingHandler{bookingUsecase: bookingUsecase, logger: logger}
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req request.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.Normalize()
	if field, msg := req.Validate(); field != "" {
		c.JSON(400, gin.H{"error": msg})
		return
	}

	var guestEmail, guestNote *string
	if req.GuestEmail != "" {
		guestEmail = &req.GuestEmail
	}
	if req.GuestNote != "" {
		guestNote = &req.GuestNote
	}

	booking, err := h.bookingUsecase.CreateBooking(
		c.Request.Context(),
		req.RoomID,
		req.CheckinDate,
		req.CheckoutDate,
		req.NumGuests,
		req.GuestName,
		req.GuestPhone,
		guestEmail,
		guestNote,
	)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
			return
		}
		h.logger.Error("unhandled booking error", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
	}

	c.JSON(201, gin.H{
		"booking_code":    booking.BookingCode,
		"room_name":       booking.Room.Name,
		"checkin_date":    booking.CheckinDate,
		"checkout_date":   booking.CheckoutDate,
		"num_nights":      booking.NumNights,
		"total_amount":    booking.TotalAmount,
		"price_breakdown": booking.PriceBreakdown,
		"status":          booking.Status,
		"expires_at":      booking.ExpiresAt,
		"payment_url":     "/api/v1/payments/vietqr?booking_code=" + booking.BookingCode,
	})
}

func (h *BookingHandler) LookupBooking(c *gin.Context) {
	var req request.LookupBookingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.Normalize()

	booking, err := h.bookingUsecase.LookupBooking(c.Request.Context(), req.Code, req.Phone)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
			return
		}
		h.logger.Error("unhandled booking error", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
	}

	c.JSON(200, gin.H{
		"booking_code":  booking.BookingCode,
		"room_name":     booking.Room.Name,
		"checkin_date":  utils.FormatDate(booking.CheckinDate),
		"checkout_date": utils.FormatDate(booking.CheckoutDate),
		"num_nights":    booking.NumNights,
		"num_guests":    booking.NumGuests,
		"total_amount":  booking.TotalAmount,
		"status":        booking.Status,
		"checkin_time":  booking.Room.CheckinTime,
		"checkout_time": booking.Room.CheckoutTime,
		"address":       booking.Room.Address,
	})
}
