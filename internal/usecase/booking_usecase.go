package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"nomad-residence-be/internal/infrastructure/notification"
	"nomad-residence-be/internal/repository"
	apperrors "nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	bookingExpiryMinutes = 30
)

type BookingUsecase struct {
	db              *gorm.DB
	bookingRepo     repository.BookingRepository
	roomRepo        repository.RoomRepository
	blockedDateRepo repository.BlockedDateRepository
	paymentRepo     repository.PaymentRepository
	pricingUsecase  *PricingUsecase
	notifService    *notification.Service
	logger          *slog.Logger
}

func NewBookingUsecase(
	db *gorm.DB,
	bookingRepo repository.BookingRepository,
	roomRepo repository.RoomRepository,
	blockedDateRepo repository.BlockedDateRepository,
	paymentRepo repository.PaymentRepository,
	pricingUsecase *PricingUsecase,
	notifService *notification.Service,
	logger *slog.Logger,
) *BookingUsecase {
	return &BookingUsecase{
		db:              db,
		bookingRepo:     bookingRepo,
		roomRepo:        roomRepo,
		blockedDateRepo: blockedDateRepo,
		paymentRepo:     paymentRepo,
		pricingUsecase:  pricingUsecase,
		notifService:    notifService,
		logger:          logger,
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
		return nil, apperrors.ErrInvalidDates
	}
	checkout, err := utils.ParseDate(checkoutStr)
	if err != nil {
		return nil, apperrors.ErrInvalidDates
	}

	if !checkout.After(checkin) {
		return nil, apperrors.ErrInvalidDates
	}

	today := utils.TruncateToDate(time.Now())
	if checkin.Before(today) {
		return nil, apperrors.ErrPastCheckin
	}

	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil || room.Status != entity.RoomStatusActive {
		return nil, apperrors.ErrRoomNotFound
	}

	if numGuests > room.MaxGuests {
		return nil, apperrors.ErrGuestsExceeded
	}

	numNights := utils.NightsBetween(checkin, checkout)
	if numNights < room.MinNights {
		return nil, apperrors.ErrMinNights
	}
	if numNights > room.MaxNights {
		return nil, apperrors.ErrMaxNights
	}

	breakdown, err := uc.pricingUsecase.CalculatePriceForRoom(ctx, room, checkin, checkout)
	if err != nil {
		return nil, err
	}

	priceJSON, _ := json.Marshal(breakdown)

	var booking *entity.Booking

	err = repository.RunInTx(ctx, uc.db, func(txCtx context.Context) error {
		if err := uc.bookingRepo.LockRoom(txCtx, roomID); err != nil {
			return err
		}

		available, err := uc.bookingRepo.IsAvailable(txCtx, roomID, checkin, checkout, nil)
		if err != nil {
			return err
		}
		if !available {
			return apperrors.ErrRoomNotAvailable
		}

		now := time.Now()
		expiresAt := utils.AddMinutes(now, bookingExpiryMinutes)

		booking = &entity.Booking{
			BookingCode:  generateBookingCode(),
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
		}

		breakdownJSON := json.RawMessage(priceJSON)
		dtJSON := datatypesJSON(breakdownJSON)
		booking.PriceBreakdown = &dtJSON

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
	if booking == nil {
		return nil, apperrors.ErrBookingNotFound
	}
	if booking.GuestPhone != phone {
		return nil, apperrors.ErrBookingNotFound
	}
	return booking, nil
}

// --- Admin operations ---

func (uc *BookingUsecase) ListBookings(ctx context.Context, f filter.BookingFilter) ([]entity.Booking, int64, error) {
	return uc.bookingRepo.FindAll(ctx, f)
}

func (uc *BookingUsecase) GetBookingByID(ctx context.Context, id uint) (*entity.Booking, error) {
	booking, err := uc.bookingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, apperrors.ErrBookingNotFound
	}
	return booking, nil
}

func (uc *BookingUsecase) ConfirmBooking(ctx context.Context, id uint, adminNote string) (*entity.Booking, error) {
	var result *entity.Booking

	err := repository.RunInTx(ctx, uc.db, func(txCtx context.Context) error {
		booking, err := uc.bookingRepo.FindByID(txCtx, id)
		if err != nil {
			return err
		}
		if booking == nil {
			return apperrors.ErrBookingNotFound
		}

		if booking.Status != entity.BookingPending {
			return apperrors.ErrInvalidStatus
		}

		if booking.ExpiresAt != nil && booking.ExpiresAt.Before(time.Now()) {
			return apperrors.ErrBookingExpired
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
			return apperrors.ErrRoomNotAvailable
		}

		now := time.Now()
		booking.Status = entity.BookingConfirmed
		booking.ConfirmedAt = &now
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

	uc.sendConfirmationNotifications(ctx, result)

	return result, nil
}

func (uc *BookingUsecase) CancelBooking(ctx context.Context, id uint, reason string) (*entity.Booking, error) {
	var result *entity.Booking

	err := repository.RunInTx(ctx, uc.db, func(txCtx context.Context) error {
		booking, err := uc.bookingRepo.FindByID(txCtx, id)
		if err != nil {
			return err
		}
		if booking == nil {
			return apperrors.ErrBookingNotFound
		}

		if booking.Status != entity.BookingPending && booking.Status != entity.BookingConfirmed {
			return apperrors.ErrInvalidStatus
		}

		now := time.Now()
		booking.Status = entity.BookingCanceled
		booking.CanceledAt = &now
		booking.CancelReason = &reason

		if err := uc.bookingRepo.Update(txCtx, booking); err != nil {
			return err
		}

		if err := uc.removeBlockedDatesForBooking(txCtx, booking); err != nil {
			uc.logger.Error("failed to remove blocked dates for canceled booking",
				slog.Uint64("booking_id", uint64(booking.ID)),
				slog.Any("error", err),
			)
		}

		result = booking
		return nil
	})
	if err != nil {
		return nil, err
	}

	uc.sendCancellationNotifications(ctx, result)

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
		return nil, apperrors.ErrInvalidStatus
	}
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
		return nil, apperrors.ErrInvalidDates
	}
	checkout, err := utils.ParseDate(checkoutStr)
	if err != nil {
		return nil, apperrors.ErrInvalidDates
	}

	if !checkout.After(checkin) {
		return nil, apperrors.ErrInvalidDates
	}

	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, apperrors.ErrRoomNotFound
	}

	numNights := utils.NightsBetween(checkin, checkout)

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
		BookingCode:  generateBookingCode(),
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
	}
	if adminNote != "" {
		booking.AdminNote = &adminNote
	}

	priceJSON, _ := json.Marshal(breakdown)
	breakdownJSON := json.RawMessage(priceJSON)
	dtJSON := datatypesJSON(breakdownJSON)
	booking.PriceBreakdown = &dtJSON

	err = repository.RunInTx(ctx, uc.db, func(txCtx context.Context) error {
		if err := uc.bookingRepo.Create(txCtx, booking); err != nil {
			return err
		}
		return uc.createBlockedDatesForBooking(txCtx, booking)
	})
	if err != nil {
		return nil, err
	}

	uc.sendConfirmationNotifications(ctx, booking)

	return booking, nil
}

// --- Helpers ---

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
		return nil, apperrors.ErrBookingNotFound
	}
	if booking.Status != entity.BookingConfirmed {
		return nil, apperrors.ErrInvalidStatus
	}

	booking.Status = entity.BookingCompleted
	if err := uc.bookingRepo.Update(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

func (uc *BookingUsecase) sendConfirmationNotifications(ctx context.Context, booking *entity.Booking) {
	if uc.notifService == nil {
		return
	}

	room, err := uc.roomRepo.FindByID(ctx, booking.RoomID)
	if err == nil && room != nil {
		booking.Room = *room
	}

	if err := uc.notifService.SendBookingConfirmationEmail(ctx, booking); err != nil {
		uc.logger.Error("failed to send booking confirmation email",
			slog.Uint64("booking_id", uint64(booking.ID)),
			slog.Any("error", err),
		)
	}

	if err := uc.notifService.NotifyAdminBookingConfirmed(ctx, booking); err != nil {
		uc.logger.Error("failed to send admin telegram notification",
			slog.Uint64("booking_id", uint64(booking.ID)),
			slog.Any("error", err),
		)
	}
}

func (uc *BookingUsecase) sendCancellationNotifications(ctx context.Context, booking *entity.Booking) {
	if uc.notifService == nil {
		return
	}

	room, err := uc.roomRepo.FindByID(ctx, booking.RoomID)
	if err == nil && room != nil {
		booking.Room = *room
	}

	if err := uc.notifService.SendBookingCancellationEmail(ctx, booking); err != nil {
		uc.logger.Error("failed to send booking cancellation email",
			slog.Uint64("booking_id", uint64(booking.ID)),
			slog.Any("error", err),
		)
	}
}

func generateBookingCode() string {
	now := time.Now()
	return fmt.Sprintf("NR%s%04d",
		now.Format("060102"),
		now.UnixNano()%10000,
	)
}

// datatypesJSON converts json.RawMessage into gorm datatypes.JSON.
func datatypesJSON(raw json.RawMessage) datatypes.JSON {
	return datatypes.JSON(raw)
}
