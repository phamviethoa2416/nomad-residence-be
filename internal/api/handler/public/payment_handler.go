package public

import (
	"errors"
	"log/slog"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PaymentHandler struct {
	paymentUsecase port.PaymentUsecase
	bookingUsecase port.BookingUsecase
	vietqrSvc      port.VietQRService
	logger         *slog.Logger
}

func NewPaymentHandler(
	paymentUsecase port.PaymentUsecase,
	bookingUsecase port.BookingUsecase,
	vietqrSvc port.VietQRService,
	logger *slog.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		paymentUsecase: paymentUsecase,
		bookingUsecase: bookingUsecase,
		vietqrSvc:      vietqrSvc,
		logger:         logger,
	}
}

func (h *PaymentHandler) GetQRPayment(c *gin.Context) {
	var req request.PaymentBookingCodeQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.Normalize()

	booking, err := h.bookingUsecase.GetBookingByCode(c.Request.Context(), req.BookingCode)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
			return
		}
		h.logger.Error("failed to get booking by code", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	if h.vietqrSvc == nil {
		c.JSON(500, gin.H{"error": "VietQR service chưa được cấu hình"})
		return
	}

	qrInfo, err := h.vietqrSvc.GenerateQRInfo(booking.BookingCode, booking.TotalAmount)
	if err != nil {
		h.logger.Error("failed to generate vietqr", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Không thể tạo mã QR"})
		return
	}

	c.JSON(200, gin.H{
		"booking_code":  booking.BookingCode,
		"room_name":     booking.Room.Name,
		"checkin_date":  booking.CheckinDate,
		"checkout_date": booking.CheckoutDate,
		"num_nights":    booking.NumNights,
		"amount":        booking.TotalAmount,
		"expires_at":    booking.ExpiresAt,
		"qr": gin.H{
			"data_url":         qrInfo.QRDataURL,
			"raw_code":         qrInfo.QRCode,
			"transfer_content": qrInfo.TransferContent,
			"account_no":       qrInfo.AccountNo,
			"account_name":     qrInfo.AccountName,
			"bank_name":        qrInfo.BankName,
		},
		"instructions": []string{
			"1. Mở app ngân hàng hoặc ví điện tử",
			"2. Chọn Quét mã QR hoặc Chuyển khoản",
			"3. Quét mã QR hoặc nhập thông tin tài khoản",
			"4. Nhập đúng nội dung: " + qrInfo.TransferContent,
			"5. Xác nhận và hoàn tất thanh toán",
		},
	})
}

// VietQRWebhook POST /api/v1/payments/vietqr/webhook
// Response format {error, errorReason, toastMessage, object} theo yêu cầu của VietQR gateway.
func (h *PaymentHandler) VietQRWebhook(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(200, gin.H{
			"error":        true,
			"errorReason":  "MISSING_TOKEN",
			"toastMessage": "Thiếu Authorization header",
			"object":       nil,
		})
		return
	}

	if h.vietqrSvc != nil && !h.vietqrSvc.VerifyWebhookToken(authHeader) {
		c.JSON(200, gin.H{
			"error":        true,
			"errorReason":  "INVALID_TOKEN",
			"toastMessage": "Token xác thực không hợp lệ",
			"object":       nil,
		})
		return
	}

	var body struct {
		TransactionID string  `json:"transactionid"`
		Amount        float64 `json:"amount"`
		Content       string  `json:"content"`
		BankAccount   string  `json:"bankaccount"`
		OrderID       string  `json:"orderId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.TransactionID == "" {
		c.JSON(200, gin.H{
			"error":        true,
			"errorReason":  "INVALID_PAYLOAD",
			"toastMessage": "Webhook payload không hợp lệ",
			"object":       nil,
		})
		return
	}

	// Ưu tiên content, fallback orderId để extract booking code
	content := body.Content
	if content == "" {
		content = body.OrderID
	}

	ref := gin.H{"reftransactionid": body.TransactionID}

	err := h.paymentUsecase.HandleVietQRWebhook(
		c.Request.Context(),
		body.TransactionID,
		decimal.NewFromFloat(body.Amount),
		content,
	)
	if err != nil {
		var appErr *apperrors.AppError
		ok := errors.As(err, &appErr)
		if ok && appErr.Code == "ALREADY_PROCESSED" {
			c.JSON(200, gin.H{
				"error":        false,
				"errorReason":  nil,
				"toastMessage": "Đơn hàng đã được xác nhận trước đó",
				"object":       ref,
			})
			return
		}
		h.logger.Error("VietQR webhook failed",
			slog.String("transaction_id", body.TransactionID),
			slog.Any("error", err),
		)
		if ok {
			c.JSON(200, gin.H{
				"error":        true,
				"errorReason":  appErr.Code,
				"toastMessage": appErr.Message,
				"object":       nil,
			})
			return
		}
		c.JSON(200, gin.H{
			"error":        true,
			"errorReason":  "INTERNAL_SERVER_ERROR",
			"toastMessage": "Lỗi server",
			"object":       nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"error":        false,
		"errorReason":  nil,
		"toastMessage": "Xác nhận thanh toán thành công",
		"object":       ref,
	})
}

func (h *PaymentHandler) CheckPaymentStatus(c *gin.Context) {
	var req request.PaymentBookingCodeQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.Normalize()

	booking, err := h.bookingUsecase.GetBookingByCode(c.Request.Context(), req.BookingCode)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
			return
		}
		h.logger.Error("failed to get booking by code", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	isExpired := string(booking.Status) == "canceled"
	if !isExpired && booking.ExpiresAt != nil {
		isExpired = time.Now().After(*booking.ExpiresAt)
	}

	c.JSON(200, gin.H{
		"booking_code": booking.BookingCode,
		"status":       booking.Status,
		"is_paid":      string(booking.Status) == "confirmed",
		"is_expired":   isExpired,
		"confirmed_at": booking.ConfirmedAt,
		"total_amount": booking.TotalAmount,
	})
}
