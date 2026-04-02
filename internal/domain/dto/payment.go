package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreatePaymentRequest struct {
	BookingID uint            `json:"booking_id" binding:"required"`
	Amount    decimal.Decimal `json:"amount"     binding:"required"`
	Method    string          `json:"method"     binding:"required,oneof=vietqr vnpay cash"`
}

type QRCallbackRequest struct {
	TransactionID string          `json:"transaction_id" binding:"required"`
	Amount        decimal.Decimal `json:"amount"         binding:"required"`
	Content       string          `json:"content"`
	Time          string          `json:"time"`
}

type UpdatePaymentRequest struct {
	Status    string  `json:"status"     binding:"required,oneof=paid failed refunded"`
	AdminNote *string `json:"admin_note"`
}

type PaymentResponse struct {
	ID              uint            `json:"id"`
	BookingID       uint            `json:"booking_id"`
	Amount          decimal.Decimal `json:"amount"`
	Method          string          `json:"method"`
	QRTransactionID *string         `json:"qr_transaction_id,omitempty"`
	Status          string          `json:"status"`
	PaidAt          *time.Time      `json:"paid_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

type QRPaymentResponse struct {
	PaymentID   uint            `json:"payment_id"`
	BookingCode string          `json:"booking_code"`
	Amount      decimal.Decimal `json:"amount"`
	QRContent   string          `json:"qr_content"`
	BankAccount string          `json:"bank_account"`
	BankName    string          `json:"bank_name"`
	ExpiresAt   time.Time       `json:"expires_at"`
}
