package model

import (
	"nomad-residence-be/internal/domain/entity"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type Booking struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	BookingCode string `gorm:"type:varchar(20);uniqueIndex;not null"`

	RoomID uint `gorm:"not null;index:idx_room_date_range"`

	GuestName  string  `gorm:"type:varchar(255);not null"`
	GuestPhone string  `gorm:"type:varchar(20);not null;index"`
	GuestEmail *string `gorm:"type:varchar(255)"`
	GuestNote  *string `gorm:"type:text"`

	CheckinDate  time.Time `gorm:"type:date;not null;index:idx_room_date_range"`
	CheckoutDate time.Time `gorm:"type:date;not null;index:idx_room_date_range"`

	NumGuests int `gorm:"default:1;check:num_guests > 0"`
	NumNights int `gorm:"not null;check:num_nights > 0"`

	BaseTotal   decimal.Decimal `gorm:"type:decimal(12,0);not null;check:base_total >= 0"`
	CleaningFee decimal.Decimal `gorm:"type:decimal(12,0);default:0;check:cleaning_fee >= 0"`
	Discount    decimal.Decimal `gorm:"type:decimal(12,0);default:0;check:discount >= 0"`

	TotalAmount decimal.Decimal `gorm:"type:decimal(12,0);not null;check:total_amount >= 0"`
	Currency    string          `gorm:"type:varchar(10);default:'VND'"`

	PriceBreakdown *datatypes.JSON `gorm:"type:jsonb"`

	Status entity.BookingStatus `gorm:"type:varchar(20);default:'pending';index"`
	Source entity.BookingSource `gorm:"type:varchar(20);default:'website'"`

	ConfirmedAt *time.Time
	CanceledAt  *time.Time

	CancelReason *string `gorm:"type:varchar(255)"`

	ExpiresAt *time.Time `gorm:"index"`

	RequiresRefund   bool                `gorm:"not null;default:false"`
	RefundableAmount decimal.Decimal     `gorm:"type:decimal(12,0);not null;default:0"`
	RefundStatus     entity.RefundStatus `gorm:"type:varchar(20);not null;default:'none';index"`

	AdminNote *string `gorm:"type:text"`

	Room     Room      `gorm:"foreignKey:RoomID"`
	Payments []Payment `gorm:"foreignKey:BookingID;constraint:OnDelete:CASCADE"`

	BaseModel
}

func (Booking) TableName() string { return "bookings" }

func BookingFromDomain(e *entity.Booking) *Booking {
	if e == nil {
		return nil
	}
	return &Booking{
		ID: e.ID, BookingCode: e.BookingCode, RoomID: e.RoomID,
		GuestName: e.GuestName, GuestPhone: e.GuestPhone,
		GuestEmail: e.GuestEmail, GuestNote: e.GuestNote,
		CheckinDate: e.CheckinDate, CheckoutDate: e.CheckoutDate,
		NumGuests: e.NumGuests, NumNights: e.NumNights,
		BaseTotal: e.BaseTotal, CleaningFee: e.CleaningFee, Discount: e.Discount,
		TotalAmount: e.TotalAmount, Currency: e.Currency,
		PriceBreakdown: toJSONPtr(e.PriceBreakdown),
		Status:         e.Status, Source: e.Source,
		ConfirmedAt: e.ConfirmedAt, CanceledAt: e.CanceledAt,
		CancelReason: e.CancelReason, ExpiresAt: e.ExpiresAt,
		RequiresRefund: e.RequiresRefund, RefundableAmount: e.RefundableAmount, RefundStatus: e.RefundStatus,
		AdminNote: e.AdminNote,
		BaseModel: baseFromDomain(e.BaseModel),
	}
}

func (m *Booking) ToDomain() *entity.Booking {
	if m == nil {
		return nil
	}
	b := &entity.Booking{
		ID: m.ID, BookingCode: m.BookingCode, RoomID: m.RoomID,
		GuestName: m.GuestName, GuestPhone: m.GuestPhone,
		GuestEmail: m.GuestEmail, GuestNote: m.GuestNote,
		CheckinDate: m.CheckinDate, CheckoutDate: m.CheckoutDate,
		NumGuests: m.NumGuests, NumNights: m.NumNights,
		BaseTotal: m.BaseTotal, CleaningFee: m.CleaningFee, Discount: m.Discount,
		TotalAmount: m.TotalAmount, Currency: m.Currency,
		PriceBreakdown: fromJSONPtr(m.PriceBreakdown),
		Status:         m.Status, Source: m.Source,
		ConfirmedAt: m.ConfirmedAt, CanceledAt: m.CanceledAt,
		CancelReason: m.CancelReason, ExpiresAt: m.ExpiresAt,
		RequiresRefund: m.RequiresRefund, RefundableAmount: m.RefundableAmount, RefundStatus: m.RefundStatus,
		AdminNote: m.AdminNote,
		BaseModel: m.BaseModel.toDomainBase(),
	}
	if m.Room.ID != 0 {
		b.Room = *m.Room.ToDomain()
	}
	for i := range m.Payments {
		b.Payments = append(b.Payments, *m.Payments[i].ToDomain())
	}
	return b
}

func BookingsToDomain(models []Booking) []entity.Booking {
	result := make([]entity.Booking, len(models))
	for i := range models {
		result[i] = *models[i].ToDomain()
	}
	return result
}
