package vietqr

import (
	"crypto/subtle"
	"fmt"
	"net/url"
	"nomad-residence-be/config"
	"nomad-residence-be/internal/domain/port"
	"strings"

	"github.com/shopspring/decimal"
)

type Service struct {
	cfg config.VietQRConfig
}

func NewService(cfg config.VietQRConfig) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) GenerateQRInfo(bookingCode string, amount decimal.Decimal) (*port.VietQRInfo, error) {
	if s.cfg.BankBin == "" || s.cfg.AccountNo == "" || s.cfg.AccountName == "" {
		return nil, fmt.Errorf("vietqr config is incomplete")
	}

	template := s.cfg.Template
	if template == "" {
		template = "compact2"
	}

	transferContent := strings.ToUpper(strings.TrimSpace(bookingCode))
	if transferContent == "" {
		return nil, fmt.Errorf("booking code is required")
	}

	amountInt := amount.IntPart()
	if amountInt <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	basePath := fmt.Sprintf("https://img.vietqr.io/image/%s-%s-%s.png", s.cfg.BankBin, s.cfg.AccountNo, template)
	q := url.Values{}
	q.Set("amount", fmt.Sprintf("%d", amountInt))
	q.Set("addInfo", transferContent)
	q.Set("accountName", s.cfg.AccountName)

	dataURL := basePath + "?" + q.Encode()

	bankName := s.cfg.BankName
	if bankName == "" {
		bankName = s.cfg.BankBin
	}

	return &port.VietQRInfo{
		QRDataURL:       dataURL,
		QRCode:          dataURL,
		TransferContent: transferContent,
		AccountNo:       s.cfg.AccountNo,
		AccountName:     s.cfg.AccountName,
		BankName:        bankName,
	}, nil
}

func (s *Service) VerifyWebhookToken(authHeader string) bool {
	expected := strings.TrimSpace(s.cfg.WebhookToken)
	if expected == "" {
		return true
	}

	token := strings.TrimSpace(authHeader)
	const bearer = "Bearer "
	if strings.HasPrefix(token, bearer) {
		token = strings.TrimSpace(token[len(bearer):])
	}

	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}
