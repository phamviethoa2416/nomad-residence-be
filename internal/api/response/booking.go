package response

import (
	"time"

	"github.com/shopspring/decimal"
)

type BookingPaymentResponse struct {
	ID     uint            `json:"id"`
	Amount decimal.Decimal `json:"amount"`
	Method string          `json:"method"`
	Status string          `json:"status"`
	PaidAt *time.Time      `json:"paid_at,omitempty"`
}

type BookingResponse struct {
	ID             uint                     `json:"id"`
	BookingCode    string                   `json:"booking_code"`
	RoomID         uint                     `json:"room_id"`
	Room           *RoomSummaryResponse     `json:"room,omitempty"`
	GuestName      string                   `json:"guest_name"`
	GuestPhone     string                   `json:"guest_phone"`
	GuestEmail     *string                  `json:"guest_email,omitempty"`
	GuestNote      *string                  `json:"guest_note,omitempty"`
	CheckinDate    time.Time                `json:"checkin_date"`
	CheckoutDate   time.Time                `json:"checkout_date"`
	NumGuests      int                      `json:"num_guests"`
	NumNights      int                      `json:"num_nights"`
	BaseTotal      decimal.Decimal          `json:"base_total"`
	CleaningFee    decimal.Decimal          `json:"cleaning_fee"`
	Discount       decimal.Decimal          `json:"discount"`
	TotalAmount    decimal.Decimal          `json:"total_amount"`
	PriceBreakdown interface{}              `json:"price_breakdown,omitempty"`
	Status         string                   `json:"status"`
	Source         string                   `json:"source"`
	AdminNote      *string                  `json:"admin_note,omitempty"`
	ConfirmedAt    *time.Time               `json:"confirmed_at,omitempty"`
	CanceledAt     *time.Time               `json:"canceled_at,omitempty"`
	CancelReason   *string                  `json:"cancel_reason,omitempty"`
	ExpiresAt      *time.Time               `json:"expires_at,omitempty"`
	Payments       []BookingPaymentResponse `json:"payments,omitempty"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
}

type BookingSummaryResponse struct {
	ID           uint            `json:"id"`
	BookingCode  string          `json:"booking_code"`
	GuestName    string          `json:"guest_name"`
	GuestPhone   string          `json:"guest_phone"`
	CheckinDate  time.Time       `json:"checkin_date"`
	CheckoutDate time.Time       `json:"checkout_date"`
	NumNights    int             `json:"num_nights"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
}
