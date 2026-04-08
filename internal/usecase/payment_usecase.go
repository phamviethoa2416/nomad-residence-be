package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/repository"
	apperrors "nomad-residence-be/pkg/errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PaymentUsecase struct {
	db              *gorm.DB
	paymentRepo     repository.PaymentRepository
	bookingRepo     repository.BookingRepository
	blockedDateRepo repository.BlockedDateRepository
	bookingUsecase  *BookingUsecase
	logger          *slog.Logger
}

func NewPaymentUsecase(
	db *gorm.DB,
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	blockedDateRepo repository.BlockedDateRepository,
	bookingUsecase *BookingUsecase,
	logger *slog.Logger,
) *PaymentUsecase {
	return &PaymentUsecase{
		db:              db,
		paymentRepo:     paymentRepo,
		bookingRepo:     bookingRepo,
		blockedDateRepo: blockedDateRepo,
		bookingUsecase:  bookingUsecase,
		logger:          logger,
	}
}

func (uc *PaymentUsecase) CreatePayment(
	ctx context.Context,
	bookingID uint,
	amount decimal.Decimal,
	method string,
) (*entity.Payment, error) {
	booking, err := uc.bookingRepo.FindByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, apperrors.ErrBookingNotFound
	}

	if booking.Status != entity.BookingPending && booking.Status != entity.BookingConfirmed {
		return nil, apperrors.ErrInvalidStatus
	}

	if booking.ExpiresAt != nil && booking.ExpiresAt.Before(time.Now()) && booking.Status == entity.BookingPending {
		return nil, apperrors.ErrBookingExpired
	}

	payment := &entity.Payment{
		BookingID:      bookingID,
		Amount:         amount,
		Currency:       "VND",
		Method:         entity.PaymentMethod(method),
		Status:         entity.PaymentPending,
		IdempotencyKey: generateIdempotencyKey(bookingID, method),
	}

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUsecase) GetPaymentByID(ctx context.Context, id uint) (*entity.Payment, error) {
	payment, err := uc.paymentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, apperrors.ErrPaymentNotFound
	}
	return payment, nil
}

func (uc *PaymentUsecase) GetPaymentsByBookingID(ctx context.Context, bookingID uint) ([]entity.Payment, error) {
	return uc.paymentRepo.FindByBookingID(ctx, bookingID)
}

func (uc *PaymentUsecase) GetPaymentsByBookingCode(ctx context.Context, bookingCode string) ([]entity.Payment, error) {
	booking, err := uc.bookingRepo.FindByCode(ctx, bookingCode)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, apperrors.ErrBookingNotFound
	}
	return uc.paymentRepo.FindByBookingID(ctx, booking.ID)
}

// HandleVietQRWebhook processes incoming VietQR payment webhooks.
// It extracts the booking code from the payment content, verifies the amount,
// and confirms the booking if payment matches.
func (uc *PaymentUsecase) HandleVietQRWebhook(
	ctx context.Context,
	transactionID string,
	amount decimal.Decimal,
	content string,
) error {
	existing, err := uc.paymentRepo.FindByQRTransactionID(ctx, transactionID)
	if err != nil {
		return err
	}
	if existing != nil {
		return apperrors.ErrAlreadyProcessed
	}

	bookingCode := extractBookingCode(content)
	if bookingCode == "" {
		uc.logger.Warn("VietQR webhook: cannot extract booking code",
			slog.String("content", content),
			slog.String("transaction_id", transactionID),
		)
		return apperrors.ErrBookingNotFound
	}

	booking, err := uc.bookingRepo.FindByCode(ctx, bookingCode)
	if err != nil {
		return err
	}
	if booking == nil {
		return apperrors.ErrBookingNotFound
	}

	if booking.Status != entity.BookingPending {
		if booking.Status == entity.BookingConfirmed {
			return apperrors.ErrAlreadyProcessed
		}
		return apperrors.ErrInvalidStatus
	}

	if !amount.Equal(booking.TotalAmount) {
		uc.logger.Warn("VietQR webhook: amount mismatch",
			slog.String("booking_code", bookingCode),
			slog.String("expected", booking.TotalAmount.String()),
			slog.String("received", amount.String()),
		)
		return apperrors.ErrInvalidAmount
	}

	return uc.confirmPaymentAndBooking(ctx, booking, transactionID, amount)
}

// HandleQRCallback processes QR payment callbacks, updating the payment record
// and confirming the booking.
func (uc *PaymentUsecase) HandleQRCallback(
	ctx context.Context,
	transactionID string,
	amount decimal.Decimal,
	content string,
) error {
	return uc.HandleVietQRWebhook(ctx, transactionID, amount, content)
}

func (uc *PaymentUsecase) UpdatePaymentStatus(
	ctx context.Context,
	paymentID uint,
	status string,
	adminNote *string,
) (*entity.Payment, error) {
	payment, err := uc.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, apperrors.ErrPaymentNotFound
	}

	payment.Status = entity.PaymentStatus(status)
	if adminNote != nil {
		payment.AdminNote = adminNote
	}

	if status == string(entity.PaymentPaid) {
		now := time.Now()
		payment.PaidAt = &now
	}

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, err
	}

	if payment.Status == entity.PaymentPaid {
		booking, bErr := uc.bookingRepo.FindByID(ctx, payment.BookingID)
		if bErr == nil && booking != nil && booking.Status == entity.BookingPending {
			if _, err := uc.bookingUsecase.ConfirmBooking(ctx, booking.ID, "Tự động xác nhận qua thanh toán"); err != nil {
				uc.logger.Error("failed to auto-confirm booking after payment",
					slog.Uint64("booking_id", uint64(booking.ID)),
					slog.Any("error", err),
				)
			}
		}
	}

	return payment, nil
}

// ConfirmAllPendingPayments marks all pending payments for a booking as paid (admin action).
func (uc *PaymentUsecase) ConfirmAllPendingPayments(ctx context.Context, bookingID uint, adminNote string) error {
	return uc.paymentRepo.UpdateManyPendingToSuccess(ctx, bookingID, adminNote)
}

// --- Internal helpers ---

func (uc *PaymentUsecase) confirmPaymentAndBooking(
	ctx context.Context,
	booking *entity.Booking,
	transactionID string,
	amount decimal.Decimal,
) error {
	return repository.RunInTx(ctx, uc.db, func(txCtx context.Context) error {
		pendingPayment, err := uc.paymentRepo.FindPendingByBookingID(txCtx, booking.ID)
		if err != nil {
			return err
		}

		now := time.Now()
		if pendingPayment != nil {
			pendingPayment.Status = entity.PaymentPaid
			pendingPayment.ExternalTransactionID = &transactionID
			pendingPayment.PaidAt = &now
			if err := uc.paymentRepo.Update(txCtx, pendingPayment); err != nil {
				return err
			}
		} else {
			newPayment := &entity.Payment{
				BookingID:             booking.ID,
				Amount:                amount,
				Currency:              "VND",
				Method:                entity.PaymentVietQR,
				Status:                entity.PaymentPaid,
				IdempotencyKey:        uuid.New().String(),
				ExternalTransactionID: &transactionID,
				PaidAt:                &now,
			}
			if err := uc.paymentRepo.Create(txCtx, newPayment); err != nil {
				return err
			}
		}

		if _, err := uc.bookingUsecase.ConfirmBooking(txCtx, booking.ID, "Tự động xác nhận qua VietQR"); err != nil {
			return err
		}

		return nil
	})
}

func generateIdempotencyKey(bookingID uint, method string) string {
	return fmt.Sprintf("%d-%s-%s", bookingID, method, uuid.New().String()[:8])
}

// extractBookingCode attempts to extract a booking code (format "NRYYMMDDXXXX")
// from the payment content/description string.
func extractBookingCode(content string) string {
	upper := strings.ToUpper(content)
	idx := strings.Index(upper, "NR")
	if idx == -1 {
		return ""
	}

	candidate := upper[idx:]
	end := 0
	for end < len(candidate) && end < 14 {
		c := candidate[end]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			end++
		} else {
			break
		}
	}

	if end >= 10 {
		return candidate[:end]
	}

	return ""
}
