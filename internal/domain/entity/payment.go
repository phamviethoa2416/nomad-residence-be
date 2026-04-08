package entity

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
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
	ID        uint `json:"id"`
	BookingID uint `json:"booking_id"`

	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`

	Method PaymentMethod `json:"method"`
	Status PaymentStatus `json:"status"`

	IdempotencyKey string `json:"-"`

	ExternalTransactionID *string `json:"external_transaction_id,omitempty"`

	PaidAt      *time.Time       `json:"paid_at,omitempty"`
	RawResponse *json.RawMessage `json:"raw_response,omitempty"`
	AdminNote   *string          `json:"admin_note,omitempty"`

	BaseModel

	Booking Booking `json:"-"`
}
