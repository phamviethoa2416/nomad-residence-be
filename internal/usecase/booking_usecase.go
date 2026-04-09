package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"nomad-residence-be/internal/domain/port"
	"nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"time"

	"github.com/shopspring/decimal"
)

const (
	bookingExpiryMinutes = 30
)

type BookingUsecase struct {
	tx                  port.TransactionManager
	bookingRepo         port.BookingRepository
	roomRepo            port.RoomRepository
	blockedDateRepo     port.BlockedDateRepository
	paymentRepo         port.PaymentRepository
	pricingUsecase      port.PricingUsecase
	notificationService port.NotificationService
	logger              *slog.Logger
}

func NewBookingUsecase(
	tx port.TransactionManager,
	bookingRepo port.BookingRepository,
	roomRepo port.RoomRepository,
	blockedDateRepo port.BlockedDateRepository,
	paymentRepo port.PaymentRepository,
	pricingUsecase port.PricingUsecase,
	notificationService port.NotificationService,
	logger *slog.Logger,
) *BookingUsecase {
	return &BookingUsecase{
		tx:                  tx,
		bookingRepo:         bookingRepo,
		roomRepo:            roomRepo,
		blockedDateRepo:     blockedDateRepo,
		paymentRepo:         paymentRepo,
		pricingUsecase:      pricingUsecase,
		notificationService: notificationService,
		logger:              logger,
	}
}

// --- Guest-facing ---

func (uc *BookingUsecase) CreateBooking(
	ctx context.Context,
	roomID uint,
	checkinStr, checkoutStr string,
	numGuests int,
	guestName, guestPhone string,
	guestEmail, guestNote *string,
) (*entity.Booking, error) {
	checkin, err := utils.ParseDate(checkinStr)
	if err != nil {
		return nil, errors.ErrInvalidDates
	}

	checkout, err := utils.ParseDate(checkoutStr)
	if err != nil {
		return nil, errors.ErrInvalidDates
	}

	if !checkout.After(checkin) {
		return nil, errors.ErrInvalidDates
	}

	today := utils.TruncateToDate(time.Now())
	if checkin.Before(today) {
		return nil, errors.ErrPastCheckin
	}

	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil || room.Status != entity.RoomStatusActive {
		return nil, errors.ErrRoomNotFound
	}

	if numGuests > room.MaxGuests {
		return nil, errors.ErrGuestsExceeded
	}

	numNights := utils.NightsBetween(checkin, checkout)
	if numNights < room.MinNights {
		return nil, errors.ErrMinNights
	}
	if numNights > room.MaxNights {
		return nil, errors.ErrMaxNights
	}

	breakdown, err := uc.pricingUsecase.CalculatePriceForRoom(ctx, room, checkin, checkout)
	if err != nil {
		return nil, err
	}

	priceJSON, _ := json.Marshal(breakdown)

	var booking *entity.Booking

	err = uc.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := uc.bookingRepo.LockRoom(txCtx, roomID); err != nil {
			return err
		}

		available, err := uc.bookingRepo.IsAvailable(txCtx, roomID, checkin, checkout, nil)
		if err != nil {
			return err
		}
		if !available {
			return errors.ErrRoomNotAvailable
		}

		now := time.Now()
		expiresAt := utils.AddMinutes(now, bookingExpiryMinutes)

		booking = &entity.Booking{
			BookingCode:  generateBookingCode(now),
			RoomID:       roomID,
			GuestName:    guestName,
			GuestPhone:   guestPhone,
			GuestEmail:   guestEmail,
			GuestNote:    guestNote,
			CheckinDate:  checkin,
			CheckoutDate: checkout,
			NumGuests:    numGuests,
			NumNights:    numNights,
			BaseTotal:    breakdown.BaseTotal,
			CleaningFee:  breakdown.CleaningFee,
			Discount:     breakdown.Discount,
			TotalAmount:  breakdown.Total,
			Currency:     "VND",
			Status:       entity.BookingPending,
			Source:       entity.BookingSourceWebsite,
			ExpiresAt:    &expiresAt,
			RefundStatus: entity.RefundNone,
		}

		breakdownJSON := json.RawMessage(priceJSON)
		booking.PriceBreakdown = &breakdownJSON

		return uc.bookingRepo.Create(txCtx, booking)
	})

	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (uc *BookingUsecase) LookupBooking(ctx context.Context, code, phone string) (*entity.Booking, error) {
	booking, err := uc.bookingRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if booking == nil || booking.GuestPhone != phone {
		return nil, errors.ErrBookingNotFound
	}

	uc.normalizeBooking(booking)

	return booking, nil
}

// --- Admin operations ---

func (uc *BookingUsecase) ListBookings(ctx context.Context, f filter.BookingFilter) ([]entity.Booking, int64, error) {
	bookings, total, err := uc.bookingRepo.FindAll(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	for i := range bookings {
		uc.normalizeBooking(&bookings[i])
	}
	return bookings, total, nil
}

func (uc *BookingUsecase) GetBookingByID(ctx context.Context, id uint) (*entity.Booking, error) {
	booking, err := uc.bookingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, errors.ErrBookingNotFound
	}

	uc.normalizeBooking(booking)
	return booking, nil
}

func (uc *BookingUsecase) GetBookingByCode(ctx context.Context, code string) (*entity.Booking, error) {
	booking, err := uc.bookingRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, errors.ErrBookingNotFound
	}

	uc.normalizeBooking(booking)
	return booking, nil
}

func (uc *BookingUsecase) ConfirmBooking(ctx context.Context, id uint, adminNote string) (*entity.Booking, error) {
	var result *entity.Booking

	err := uc.tx.RunInTx(ctx, func(txCtx context.Context) error {
		booking, err := uc.bookingRepo.FindByID(txCtx, id)
		if err != nil {
			return err
		}
		if booking == nil {
			return errors.ErrBookingNotFound
		}

		uc.normalizeBooking(booking)

		if booking.Status != entity.BookingPending {
			if booking.Status == entity.BookingExpired {
				return errors.ErrBookingExpired
			}
			return errors.ErrInvalidStatus
		}

		if err := uc.bookingRepo.LockRoom(txCtx, booking.RoomID); err != nil {
			return err
		}

		available, err := uc.bookingRepo.IsAvailable(
			txCtx, booking.RoomID,
			booking.CheckinDate, booking.CheckoutDate,
			&booking.ID,
		)

		if err != nil {
			return err
		}
		if !available {
			return errors.ErrRoomNotAvailable
		}

		now := time.Now()
		booking.Status = entity.BookingConfirmed
		booking.ConfirmedAt = &now
		booking.ExpiresAt = nil

		if adminNote != "" {
			booking.AdminNote = &adminNote
		}

		if err := uc.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}

		if err := uc.createBlockedDatesForBooking(txCtx, booking); err != nil {
			return err
		}

		result = booking
		return nil
	})
	if err != nil {
		return nil, err
	}

	if uc.notificationService == nil {
		return result, nil
	}

	room, err := uc.roomRepo.FindByID(ctx, result.RoomID)
	if err == nil && room != nil {
		result.Room = *room
	}

	if err := uc.notificationService.SendBookingConfirmationEmail(ctx, result); err != nil {
		uc.logger.Error("failed to send booking confirmation email",
			slog.Uint64("booking_id", uint64(result.ID)),
			slog.Any("error", err),
		)
	}

	if err := uc.notificationService.NotifyAdminBookingConfirmed(ctx, result); err != nil {
		uc.logger.Error("failed to send admin telegram notification",
			slog.Uint64("booking_id", uint64(result.ID)),
			slog.Any("error", err),
		)
	}

	return result, nil
}

func (uc *BookingUsecase) CancelBooking(ctx context.Context, id uint, reason string) (*entity.Booking, error) {
	var result *entity.Booking

	err := uc.tx.RunInTx(ctx, func(txCtx context.Context) error {
		booking, err := uc.bookingRepo.FindByID(txCtx, id)
		if err != nil {
			return err
		}

		if booking == nil {
			return errors.ErrBookingNotFound
		}
		if booking.RefundStatus == "" {
			booking.RefundStatus = entity.RefundNone
		}
		now := time.Now()
		uc.normalizeBooking(booking)

		if booking.Status != entity.BookingPending && booking.Status != entity.BookingConfirmed {
			return errors.ErrInvalidStatus
		}
		wasConfirmed := booking.Status == entity.BookingConfirmed

		if err := uc.bookingRepo.LockRoom(txCtx, booking.RoomID); err != nil {
			return err
		}

		booking.Status = entity.BookingCanceled
		booking.CanceledAt = &now
		booking.CancelReason = &reason
		booking.ExpiresAt = nil
		booking.RequiresRefund = false
		booking.RefundableAmount = decimal.Zero
		booking.RefundStatus = entity.RefundNone

		if wasConfirmed {
			paidAmount, err := uc.calculatePaidAmount(txCtx, booking.ID)
			if err != nil {
				return err
			}
			if paidAmount.GreaterThan(decimal.Zero) {
				booking.RequiresRefund = true
				booking.RefundableAmount = paidAmount
				booking.RefundStatus = entity.RefundPending

				refundNote := fmt.Sprintf("Refund required: %s VND (manual payout review)", paidAmount.String())
				if booking.AdminNote == nil || *booking.AdminNote == "" {
					booking.AdminNote = &refundNote
				} else {
					merged := *booking.AdminNote + " | " + refundNote
					booking.AdminNote = &merged
				}
			}
		}

		if err := uc.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}

		if err := uc.removeBlockedDatesForBooking(txCtx, booking); err != nil {
			return err
		}

		result = booking
		return nil
	})
	if err != nil {
		return nil, err
	}

	if uc.notificationService == nil {
		return result, nil
	}

	room, err := uc.roomRepo.FindByID(ctx, result.RoomID)
	if err == nil && room != nil {
		result.Room = *room
	}

	if err := uc.notificationService.SendBookingCancellationEmail(ctx, result); err != nil {
		uc.logger.Error("failed to send booking cancellation email",
			slog.Uint64("booking_id", uint64(result.ID)),
			slog.Any("error", err),
		)
	}

	return result, nil
}

func (uc *BookingUsecase) UpdateBookingStatus(
	ctx context.Context,
	id uint,
	status string,
	adminNote, cancelReason *string,
) (*entity.Booking, error) {
	switch entity.BookingStatus(status) {
	case entity.BookingConfirmed:
		note := ""
		if adminNote != nil {
			note = *adminNote
		}
		return uc.ConfirmBooking(ctx, id, note)
	case entity.BookingCanceled:
		reason := "Hủy bởi admin"
		if cancelReason != nil {
			reason = *cancelReason
		}
		return uc.CancelBooking(ctx, id, reason)
	case entity.BookingCompleted:
		return uc.markCompleted(ctx, id)
	default:
		return nil, errors.ErrInvalidStatus
	}
}

func (uc *BookingUsecase) createBlockedDatesForBooking(ctx context.Context, booking *entity.Booking) error {
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

func (uc *BookingUsecase) removeBlockedDatesForBooking(ctx context.Context, booking *entity.Booking) error {
	ref := fmt.Sprintf("booking:%d", booking.ID)
	return uc.blockedDateRepo.DeleteByRoomAndSource(ctx, booking.RoomID, ref)
}

func (uc *BookingUsecase) markCompleted(ctx context.Context, id uint) (*entity.Booking, error) {
	booking, err := uc.bookingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, errors.ErrBookingNotFound
	}
	if booking.Status != entity.BookingConfirmed {
		return nil, errors.ErrInvalidStatus
	}

	booking.Status = entity.BookingCompleted
	if err := uc.bookingRepo.Update(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

func (uc *BookingUsecase) CreateManualBooking(
	ctx context.Context,
	roomID uint,
	checkinStr, checkoutStr string,
	guestName, guestPhone string,
	guestEmail, guestNote *string,
	numGuests int,
	source, adminNote string,
) (*entity.Booking, error) {
	checkin, err := utils.ParseDate(checkinStr)
	if err != nil {
		return nil, errors.ErrInvalidDates
	}
	checkout, err := utils.ParseDate(checkoutStr)
	if err != nil {
		return nil, errors.ErrInvalidDates
	}

	if !checkout.After(checkin) {
		return nil, errors.ErrInvalidDates
	}

	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.ErrRoomNotFound
	}
	// Manual booking allows admin override for inactive rooms.
	// We only enforce room existence here by design.

	if numGuests > room.MaxGuests {
		return nil, errors.ErrGuestsExceeded
	}

	numNights := utils.NightsBetween(checkin, checkout)
	if numNights < room.MinNights {
		return nil, errors.ErrMinNights
	}
	if numNights > room.MaxNights {
		return nil, errors.ErrMaxNights
	}

	breakdown, err := uc.pricingUsecase.CalculatePriceForRoom(ctx, room, checkin, checkout)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	bookingSource := entity.BookingSource(source)
	if bookingSource == "" {
		bookingSource = entity.BookingSourceAdmin
	}

	booking := &entity.Booking{
		BookingCode:  generateBookingCode(now),
		RoomID:       roomID,
		GuestName:    guestName,
		GuestPhone:   guestPhone,
		GuestEmail:   guestEmail,
		GuestNote:    guestNote,
		CheckinDate:  checkin,
		CheckoutDate: checkout,
		NumGuests:    numGuests,
		NumNights:    numNights,
		BaseTotal:    breakdown.BaseTotal,
		CleaningFee:  breakdown.CleaningFee,
		Discount:     breakdown.Discount,
		TotalAmount:  breakdown.Total,
		Currency:     "VND",
		Status:       entity.BookingConfirmed,
		Source:       bookingSource,
		ConfirmedAt:  &now,
		RefundStatus: entity.RefundNone,
	}
	if adminNote != "" {
		booking.AdminNote = &adminNote
	}

	priceJSON, _ := json.Marshal(breakdown)
	breakdownJSON := json.RawMessage(priceJSON)
	booking.PriceBreakdown = &breakdownJSON

	err = uc.tx.RunInTx(ctx, func(txCtx context.Context) error {
		if err := uc.bookingRepo.LockRoom(txCtx, roomID); err != nil {
			return err
		}
		available, err := uc.bookingRepo.IsAvailable(txCtx, roomID, checkin, checkout, nil)
		if err != nil {
			return err
		}
		if !available {
			return errors.ErrRoomNotAvailable
		}

		if err := uc.bookingRepo.Create(txCtx, booking); err != nil {
			return err
		}
		return uc.createBlockedDatesForBooking(txCtx, booking)
	})
	if err != nil {
		return nil, err
	}

	if uc.notificationService == nil {
		return booking, nil
	}

	room, err = uc.roomRepo.FindByID(ctx, booking.RoomID)
	if err == nil && room != nil {
		booking.Room = *room
	}

	if err := uc.notificationService.SendBookingConfirmationEmail(ctx, booking); err != nil {
		uc.logger.Error("failed to send booking confirmation email",
			slog.Uint64("booking_id", uint64(booking.ID)),
			slog.Any("error", err),
		)
	}

	if err := uc.notificationService.NotifyAdminBookingConfirmed(ctx, booking); err != nil {
		uc.logger.Error("failed to send admin telegram notification",
			slog.Uint64("booking_id", uint64(booking.ID)),
			slog.Any("error", err),
		)
	}

	return booking, nil
}

func (uc *BookingUsecase) normalizeBooking(booking *entity.Booking) {
	if booking == nil {
		return
	}
	if booking.RefundStatus == "" {
		booking.RefundStatus = entity.RefundNone
	}
	if booking.Status == entity.BookingPending && booking.ExpiresAt != nil && booking.ExpiresAt.Before(time.Now()) {
		booking.Status = entity.BookingExpired
	}
}

func (uc *BookingUsecase) calculatePaidAmount(ctx context.Context, bookingID uint) (decimal.Decimal, error) {
	payments, err := uc.paymentRepo.FindByBookingID(ctx, bookingID)
	if err != nil {
		return decimal.Zero, err
	}

	total := decimal.Zero
	for _, p := range payments {
		if p.Status == entity.PaymentPaid {
			total = total.Add(p.Amount)
		}
	}
	return total, nil
}

func generateBookingCode(now time.Time) string {
	randomPart := make([]byte, 4)
	_, err := rand.Read(randomPart)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("NR%s%s", now.Format("060102"), hex.EncodeToString(randomPart))
}
