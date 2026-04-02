package entity

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
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
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	BookingCode string `gorm:"type:varchar(20);uniqueIndex;not null" json:"booking_code"`

	RoomID uint `gorm:"not null;index:idx_room_date_range" json:"room_id"`

	GuestName  string  `gorm:"type:varchar(255);not null" json:"guest_name"`
	GuestPhone string  `gorm:"type:varchar(20);not null;index" json:"guest_phone"`
	GuestEmail *string `gorm:"type:varchar(255)" json:"guest_email,omitempty"`
	GuestNote  *string `gorm:"type:text" json:"guest_note,omitempty"`

	CheckinDate  time.Time `gorm:"type:date;not null;index:idx_room_date_range" json:"checkin_date"`
	CheckoutDate time.Time `gorm:"type:date;not null;index:idx_room_date_range" json:"checkout_date"`

	NumGuests int `gorm:"default:1;check:num_guests > 0" json:"num_guests"`
	NumNights int `gorm:"not null;check:num_nights > 0" json:"num_nights"`

	BaseTotal   decimal.Decimal `gorm:"type:decimal(12,0);not null;check:base_total >= 0" json:"base_total"`
	CleaningFee decimal.Decimal `gorm:"type:decimal(12,0);default:0;check:cleaning_fee >= 0" json:"cleaning_fee"`
	Discount    decimal.Decimal `gorm:"type:decimal(12,0);default:0;check:discount >= 0" json:"discount"`

	TotalAmount decimal.Decimal `gorm:"type:decimal(12,0);not null;check:total_amount >= 0" json:"total_amount"`
	Currency    string          `gorm:"type:varchar(10);default:'VND'" json:"currency"`

	PriceBreakdown *datatypes.JSON `gorm:"type:jsonb" json:"price_breakdown,omitempty"`

	Status BookingStatus `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	Source BookingSource `gorm:"type:varchar(20);default:'website'" json:"source"`

	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	CanceledAt  *time.Time `json:"canceled_at,omitempty"`

	CancelReason *string `gorm:"type:varchar(255)" json:"cancel_reason,omitempty"`

	ExpiresAt *time.Time `gorm:"index" json:"expires_at,omitempty"`

	AdminNote *string `gorm:"type:text" json:"admin_note,omitempty"`

	Room     Room      `gorm:"foreignKey:RoomID" json:"-"`
	Payments []Payment `gorm:"foreignKey:BookingID;constraint:OnDelete:CASCADE" json:"payments,omitempty"`

	BaseModel
}

func (Booking) TableName() string {
	return "bookings"
}
