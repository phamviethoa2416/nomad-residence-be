package errors

import "fmt"

type AppError struct {
	HTTPStatus int
	Code       string
	Message    string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(httpStatus int, code, message string) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
}

var (
	ErrRoomNotFound             = New(404, "ROOM_NOT_FOUND", "Phòng không tồn tại hoặc không hoạt động")
	ErrRoomNotAvailable         = New(409, "ROOM_NOT_AVAILABLE", "Phòng đã được đặt trong khoảng thời gian này")
	ErrGuestsExceeded           = New(400, "GUESTS_EXCEEDED", "Số khách vượt quá giới hạn của phòng")
	ErrMinNights                = New(400, "MIN_NIGHTS_REQUIRED", "Không đủ số đêm tối thiểu")
	ErrMaxNights                = New(400, "MAX_NIGHTS_EXCEEDED", "Vượt quá số đêm tối đa")
	ErrInvalidDates             = New(400, "INVALID_DATES", "Ngày check-in/check-out không hợp lệ")
	ErrPastCheckin              = New(400, "INVALID_DATES", "Ngày check-in không thể là ngày trong quá khứ")
	ErrBookingNotFound          = New(404, "BOOKING_NOT_FOUND", "Không tìm thấy đơn đặt phòng")
	ErrBookingExpired           = New(400, "BOOKING_EXPIRED", "Đơn đặt phòng đã hết hạn, vui lòng đặt lại")
	ErrInvalidStatus            = New(400, "INVALID_STATUS", "Trạng thái đơn không hợp lệ cho thao tác này")
	ErrPaymentNotFound          = New(404, "PAYMENT_NOT_FOUND", "Không tìm thấy thông tin thanh toán")
	ErrInvalidPaymentTransition = New(400, "PAYMENT_INVALID_TRANSITION", "Trạng thái thanh toán không hợp lệ")
	ErrInvalidAmount            = New(400, "INVALID_AMOUNT", "Số tiền thanh toán không khớp")
	ErrAlreadyProcessed         = New(200, "ALREADY_PROCESSED", "Đơn hàng đã được xác nhận trước đó")
	ErrIcalLinkNotFound         = New(404, "ICAL_LINK_NOT_FOUND", "Không tìm thấy iCal link")
	ErrPricingRuleNotFound      = New(404, "PRICING_RULE_NOT_FOUND", "Không tìm thấy pricing rule")
)

func Wrapf(httpStatus int, code, format string, args ...any) *AppError {
	return New(httpStatus, code, fmt.Sprintf(format, args...))
}
