package request

import (
	"nomad-residence-be/pkg/validator"
	"strings"
)

type CreateBookingRequest struct {
	RoomID       uint   `json:"room_id"       binding:"required,gt=0"`
	CheckinDate  string `json:"checkin_date"  binding:"required,datetime=2006-01-02"`
	CheckoutDate string `json:"checkout_date" binding:"required,datetime=2006-01-02"`
	NumGuests    int    `json:"num_guests"    binding:"required,min=1"`
	GuestName    string `json:"guest_name"    binding:"required,min=1,max=255"`
	GuestPhone   string `json:"guest_phone"   binding:"required,min=8,max=20"`
	GuestEmail   string `json:"guest_email"   binding:"omitempty,email,max=255"`
	GuestNote    string `json:"guest_note"    binding:"omitempty,max=1000"`
}

func (r *CreateBookingRequest) Normalize() {
	r.GuestPhone = validator.NormalizePhone(r.GuestPhone)
	r.GuestEmail = strings.ToLower(strings.TrimSpace(r.GuestEmail))
	r.GuestName = strings.TrimSpace(r.GuestName)
}

func (r *CreateBookingRequest) Validate() (string, string) {
	return validator.ValidateDateRange(r.CheckinDate, r.CheckoutDate)
}

type LookupBookingRequest struct {
	Code  string `form:"code"  binding:"required,min=1"`
	Phone string `form:"phone" binding:"required,min=1"`
}

func (r *LookupBookingRequest) Normalize() {
	r.Phone = validator.NormalizePhone(r.Phone)
	r.Code = strings.TrimSpace(r.Code)
}

type BookingIDParam struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type ListBookingsQuery struct {
	Status     string `form:"status"`
	RoomID     *uint  `form:"room_id"    binding:"omitempty,gt=0"`
	GuestPhone string `form:"guest_phone"`
	DateFrom   string `form:"date_from"  binding:"omitempty,datetime=2006-01-02"`
	DateTo     string `form:"date_to"    binding:"omitempty,datetime=2006-01-02"`
	Page       int    `form:"page"       binding:"omitempty,min=1"`
	Limit      int    `form:"limit"      binding:"omitempty,min=1,max=100"`
}

func (q *ListBookingsQuery) ApplyDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
}

func (q *ListBookingsQuery) Validate() (string, string) {
	return validator.ValidateDateRangeFromTo(q.DateFrom, q.DateTo)
}

type ConfirmBookingRequest struct {
	AdminNote string `json:"admin_note" binding:"omitempty,max=1000"`
}

type CancelBookingRequest struct {
	Reason string `json:"reason" binding:"required,min=1,max=500"`
}

type CreateManualBookingRequest struct {
	RoomID        uint   `json:"room_id"        binding:"required,gt=0"`
	CheckinDate   string `json:"checkin_date"   binding:"required,datetime=2006-01-02"`
	CheckoutDate  string `json:"checkout_date"  binding:"required,datetime=2006-01-02"`
	GuestName     string `json:"guest_name"     binding:"required,min=1,max=255"`
	GuestPhone    string `json:"guest_phone"    binding:"required,min=1,max=20"`
	GuestEmail    string `json:"guest_email"    binding:"omitempty,email,max=255"`
	GuestNote     string `json:"guest_note"     binding:"omitempty,max=1000"`
	NumGuests     int    `json:"num_guests"     binding:"omitempty,min=1"`
	Source        string `json:"source"         binding:"omitempty,max=50"`
	PaymentMethod string `json:"payment_method" binding:"omitempty,max=30"`
	AdminNote     string `json:"admin_note"     binding:"omitempty,max=1000"`
}

func (r *CreateManualBookingRequest) ApplyDefaults() {
	if r.NumGuests <= 0 {
		r.NumGuests = 1
	}
	if r.Source == "" {
		r.Source = "admin_manual"
	}
	if r.PaymentMethod == "" {
		r.PaymentMethod = "cash"
	}
}

func (r *CreateManualBookingRequest) Validate() (string, string) {
	return validator.ValidateDateRange(r.CheckinDate, r.CheckoutDate)
}

type BookingFilterRequest struct {
	Status      string `form:"status"        binding:"omitempty"`
	RoomID      *uint  `form:"room_id"`
	GuestPhone  string `form:"guest_phone"`
	CheckinFrom string `form:"checkin_from"  binding:"omitempty,datetime=2006-01-02"`
	CheckinTo   string `form:"checkin_to"    binding:"omitempty,datetime=2006-01-02"`
	Page        int    `form:"page"          binding:"omitempty,min=1"`
	Limit       int    `form:"limit"         binding:"omitempty,min=1,max=100"`
}

type UpdateBookingStatusRequest struct {
	Status       string  `json:"status"        binding:"required,oneof=confirmed canceled completed"`
	AdminNote    *string `json:"admin_note"`
	CancelReason *string `json:"cancel_reason" binding:"required_if=Status canceled"`
}
