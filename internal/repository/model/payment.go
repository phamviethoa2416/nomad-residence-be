package model

import (
	"nomad-residence-be/internal/domain/entity"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type Payment struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	BookingID uint `gorm:"not null;index"`

	Amount   decimal.Decimal      `gorm:"type:decimal(12,0);not null;check:amount >= 0"`
	Currency string               `gorm:"type:varchar(10);default:'VND'"`
	Method   entity.PaymentMethod `gorm:"type:varchar(30);not null;index"`
	Status   entity.PaymentStatus `gorm:"type:varchar(20);default:'pending';index"`

	IdempotencyKey string `gorm:"type:varchar(100);uniqueIndex"`

	ExternalTransactionID *string `gorm:"type:varchar(100);uniqueIndex"`

	PaidAt      *time.Time      `json:"paid_at,omitempty"`
	RawResponse *datatypes.JSON `gorm:"type:jsonb"`
	AdminNote   *string         `gorm:"type:text"`

	BaseModel

	Booking Booking `gorm:"foreignKey:BookingID"`
}

func PaymentFromDomain(e *entity.Payment) *Payment {
	if e == nil {
		return nil
	}
	return &Payment{
		ID: e.ID, BookingID: e.BookingID,
		Amount: e.Amount, Currency: e.Currency,
		Method: e.Method, Status: e.Status,
		IdempotencyKey:        e.IdempotencyKey,
		ExternalTransactionID: e.ExternalTransactionID,
		PaidAt:                e.PaidAt, RawResponse: toJSONPtr(e.RawResponse),
		AdminNote: e.AdminNote,
		BaseModel: baseFromDomain(e.BaseModel),
	}
}

func (m *Payment) ToDomain() *entity.Payment {
	if m == nil {
		return nil
	}
	return &entity.Payment{
		ID: m.ID, BookingID: m.BookingID,
		Amount: m.Amount, Currency: m.Currency,
		Method: m.Method, Status: m.Status,
		IdempotencyKey:        m.IdempotencyKey,
		ExternalTransactionID: m.ExternalTransactionID,
		PaidAt:                m.PaidAt, RawResponse: fromJSONPtr(m.RawResponse),
		AdminNote: m.AdminNote,
		BaseModel: m.BaseModel.toDomainBase(),
	}
}

func PaymentsToDomain(models []Payment) []entity.Payment {
	result := make([]entity.Payment, len(models))
	for i := range models {
		result[i] = *models[i].ToDomain()
	}
	return result
}
