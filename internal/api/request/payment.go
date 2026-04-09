package request

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type PaymentBookingCodeQuery struct {
	BookingCode string `form:"booking_code" binding:"required,min=1"`
}

func (q *PaymentBookingCodeQuery) Normalize() {
	upper := make([]byte, len(q.BookingCode))
	for i := range q.BookingCode {
		c := q.BookingCode[i]
		if c >= 'a' && c <= 'z' {
			upper[i] = c - 32
		} else {
			upper[i] = c
		}
	}
	q.BookingCode = string(upper)
}

type VietQRWebhookBody struct {
	TransactionID string      `json:"transaction_id"`
	Amount        json.Number `json:"amount"`
	Content       string      `json:"content"`
	BankAccount   string      `json:"bank_account"`
	OrderID       string      `json:"orderId"`
}

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
