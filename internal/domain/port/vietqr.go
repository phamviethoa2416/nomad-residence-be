package port

import "github.com/shopspring/decimal"

type VietQRInfo struct {
	QRDataURL       string
	QRCode          string
	TransferContent string
	AccountNo       string
	AccountName     string
	BankName        string
}

type VietQRService interface {
	GenerateQRInfo(bookingCode string, amount decimal.Decimal) (*VietQRInfo, error)
	VerifyWebhookToken(authHeader string) bool
}
