package entity

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type PaymentStatus string
type PaymentMethod string

const (
	PaymentPending  PaymentStatus = "pending"
	PaymentPaid     PaymentStatus = "paid"
	PaymentFailed   PaymentStatus = "failed"
	PaymentRefunded PaymentStatus = "refunded"

	PaymentVietQR PaymentMethod = "vietqr"
)

type Payment struct {
	ID        uint `gorm:"primaryKey;autoIncrement" json:"id"`
	BookingID uint `gorm:"not null;index" json:"booking_id"`

	Amount   decimal.Decimal `gorm:"type:decimal(12,0);not null;check:amount >= 0" json:"amount"`
	Currency string          `gorm:"type:varchar(10);default:'VND'" json:"currency"`

	Method PaymentMethod `gorm:"type:varchar(30);not null;index" json:"method"`
	Status PaymentStatus `gorm:"type:varchar(20);default:'pending';index" json:"status"`

	IdempotencyKey string `gorm:"type:varchar(100);uniqueIndex" json:"-"`

	ExternalTransactionID *string `gorm:"type:varchar(100);uniqueIndex" json:"external_transaction_id,omitempty"`

	PaidAt      *time.Time      `json:"paid_at,omitempty"`
	RawResponse *datatypes.JSON `gorm:"type:jsonb" json:"raw_response,omitempty"`
	AdminNote   *string         `gorm:"type:text" json:"admin_note,omitempty"`

	BaseModel

	Booking Booking `gorm:"foreignKey:BookingID" json:"-"`
}
