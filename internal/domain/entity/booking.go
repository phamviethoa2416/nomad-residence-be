package entity

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type BookingStatus string
type BookingSource string

const (
	BookingPending   BookingStatus = "pending"
	BookingConfirmed BookingStatus = "confirmed"
	BookingCanceled  BookingStatus = "canceled"
	BookingCompleted BookingStatus = "completed"
	BookingExpired   BookingStatus = "expired"

	BookingSourceWebsite BookingSource = "website"
	BookingSourceAdmin   BookingSource = "admin"
	BookingSourceIcal    BookingSource = "ical"
)

type Booking struct {
	ID uint `json:"id"`

	BookingCode string `json:"booking_code"`

	RoomID uint `json:"room_id"`

	GuestName  string  `json:"guest_name"`
	GuestPhone string  `json:"guest_phone"`
	GuestEmail *string `json:"guest_email,omitempty"`
	GuestNote  *string `json:"guest_note,omitempty"`

	CheckinDate  time.Time `json:"checkin_date"`
	CheckoutDate time.Time `json:"checkout_date"`

	NumGuests int `json:"num_guests"`
	NumNights int `json:"num_nights"`

	BaseTotal   decimal.Decimal `json:"base_total"`
	CleaningFee decimal.Decimal `json:"cleaning_fee"`
	Discount    decimal.Decimal `json:"discount"`

	TotalAmount decimal.Decimal `json:"total_amount"`
	Currency    string          `json:"currency"`

	PriceBreakdown *json.RawMessage `json:"price_breakdown,omitempty"`

	Status BookingStatus `json:"status"`
	Source BookingSource `json:"source"`

	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	CanceledAt  *time.Time `json:"canceled_at,omitempty"`

	CancelReason *string `json:"cancel_reason,omitempty"`

	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	AdminNote *string `json:"admin_note,omitempty"`

	Room     Room      `json:"-"`
	Payments []Payment `json:"payments,omitempty"`

	BaseModel
}
