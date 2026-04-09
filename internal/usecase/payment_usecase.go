package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/port"
	"nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PaymentUsecase struct {
	db              *gorm.DB
	tx              port.TransactionManager
	paymentRepo     port.PaymentRepository
	bookingRepo     port.BookingRepository
	blockedDateRepo port.BlockedDateRepository
	bookingUsecase  *BookingUsecase
	logger          *slog.Logger
}

func NewPaymentUsecase(
	db *gorm.DB,
	tx port.TransactionManager,
	paymentRepo port.PaymentRepository,
	bookingRepo port.BookingRepository,
	blockedDateRepo port.BlockedDateRepository,
	bookingUsecase *BookingUsecase,
	logger *slog.Logger,
) *PaymentUsecase {
	return &PaymentUsecase{
		db:              db,
		tx:              tx,
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
		return nil, errors.ErrBookingNotFound
	}

	if booking.Status != entity.BookingPending && booking.Status != entity.BookingConfirmed {
		return nil, errors.ErrInvalidStatus
	}

	if booking.ExpiresAt != nil && booking.ExpiresAt.Before(time.Now()) && booking.Status == entity.BookingPending {
		return nil, errors.ErrBookingExpired
	}

	idempotencyKey := fmt.Sprintf("%d-%s-%s", bookingID, method, uuid.New().String()[:8])

	payment := &entity.Payment{
		BookingID:      bookingID,
		Amount:         amount,
		Currency:       "VND",
		Method:         entity.PaymentMethod(method),
		Status:         entity.PaymentPending,
		IdempotencyKey: idempotencyKey,
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
		return nil, errors.ErrPaymentNotFound
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
		return nil, errors.ErrBookingNotFound
	}
	return uc.paymentRepo.FindByBookingID(ctx, booking.ID)
}

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
		return errors.ErrAlreadyProcessed
	}

	bookingCode := extractBookingCode(content)
	if bookingCode == "" {
		uc.logger.Warn("VietQR webhook: cannot extract booking code",
			slog.String("content", content),
			slog.String("transaction_id", transactionID),
		)
		return errors.ErrBookingNotFound
	}

	booking, err := uc.bookingRepo.FindByCode(ctx, bookingCode)
	if err != nil {
		return err
	}
	if booking == nil {
		return errors.ErrBookingNotFound
	}

	return uc.tx.RunInTx(ctx, func(txCtx context.Context) error {
		alreadyProcessed, err := uc.paymentRepo.FindByQRTransactionID(txCtx, transactionID)
		if err != nil {
			return err
		}
		if alreadyProcessed != nil {
			return errors.ErrAlreadyProcessed
		}

		if err := uc.bookingRepo.LockRoom(txCtx, booking.RoomID); err != nil {
			return err
		}

		lockedBooking, err := uc.bookingRepo.FindByID(txCtx, booking.ID)
		if err != nil {
			return err
		}
		if lockedBooking == nil {
			return errors.ErrBookingNotFound
		}

		if lockedBooking.Status != entity.BookingPending {
			if lockedBooking.Status == entity.BookingConfirmed {
				return errors.ErrAlreadyProcessed
			}
			return errors.ErrInvalidStatus
		}
		if lockedBooking.ExpiresAt != nil && lockedBooking.ExpiresAt.Before(time.Now()) {
			return errors.ErrBookingExpired
		}

		if !amount.Equal(lockedBooking.TotalAmount) {
			uc.logger.Warn("VietQR webhook: amount mismatch",
				slog.String("booking_code", bookingCode),
				slog.String("expected", lockedBooking.TotalAmount.String()),
				slog.String("received", amount.String()),
				slog.String("transaction_id", transactionID),
			)

			note := fmt.Sprintf("Discrepancy: expected=%s received=%s", lockedBooking.TotalAmount.String(), amount.String())
			failed := &entity.Payment{
				BookingID:             lockedBooking.ID,
				Amount:                amount,
				Currency:              "VND",
				Method:                entity.PaymentVietQR,
				Status:                entity.PaymentFailed,
				IdempotencyKey:        uuid.New().String(),
				ExternalTransactionID: &transactionID,
				AdminNote:             &note,
			}

			if err := uc.paymentRepo.Create(txCtx, failed); err != nil {
				return err
			}
			return errors.ErrInvalidAmount
		}

		pendingPayment, err := uc.paymentRepo.FindPendingByBookingID(txCtx, lockedBooking.ID)
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
				BookingID:             lockedBooking.ID,
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

		if err := uc.confirmLockedBooking(txCtx, lockedBooking, "Tự động xác nhận qua VietQR"); err != nil {
			return err
		}

		return nil
	})
}

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
		return nil, errors.ErrPaymentNotFound
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

func (uc *PaymentUsecase) ConfirmAllPendingPayments(ctx context.Context, bookingID uint, adminNote string) error {
	return uc.paymentRepo.UpdateManyPendingToSuccess(ctx, bookingID, adminNote)
}

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

func (uc *PaymentUsecase) confirmLockedBooking(ctx context.Context, booking *entity.Booking, adminNote string) error {
	now := time.Now()
	booking.Status = entity.BookingConfirmed
	booking.ConfirmedAt = &now
	booking.ExpiresAt = nil
	if adminNote != "" {
		booking.AdminNote = &adminNote
	}

	if err := uc.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}

	dates := utils.GetDateRange(booking.CheckinDate, booking.CheckoutDate)
	blocked := make([]entity.BlockedDate, 0, len(dates))
	ref := fmt.Sprintf("booking:%d", booking.ID)
	for _, d := range dates {
		blocked = append(blocked, entity.BlockedDate{
			RoomID:    booking.RoomID,
			Date:      d,
			Source:    entity.BlockBooking,
			SourceRef: &ref,
		})
	}

	return uc.blockedDateRepo.BulkCreate(ctx, blocked)
}
