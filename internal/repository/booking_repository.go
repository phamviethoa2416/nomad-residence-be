package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/dto"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type BookingRepository interface {
	FindAll(ctx context.Context, filter dto.BookingFilterRequest) ([]entity.Booking, int64, error)
	FindByID(ctx context.Context, id uint) (*entity.Booking, error)
	FindByCode(ctx context.Context, code string) (*entity.Booking, error)
	FindByGuestPhone(ctx context.Context, phone string) ([]entity.Booking, error)

	Create(ctx context.Context, booking *entity.Booking) error
	Update(ctx context.Context, booking *entity.Booking) error

	IsAvailable(ctx context.Context, roomID uint, checkin, checkout time.Time, excludeBookingID *uint) (bool, error)

	ExpirePendingBookings(ctx context.Context) (int64, error)
	LockRoom(ctx context.Context, roomID uint) error
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) FindAll(ctx context.Context, filter dto.BookingFilterRequest) ([]entity.Booking, int64, error) {
	db := DB(ctx, r.db).WithContext(ctx).Model(&entity.Booking{})

	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.RoomID != nil {
		db = db.Where("room_id = ?", *filter.RoomID)
	}
	if filter.GuestPhone != "" {
		db = db.Where("guest_phone = ?", filter.GuestPhone)
	}
	if filter.CheckinFrom != "" {
		db = db.Where("checkin_date >= ?", filter.CheckinFrom)
	}
	if filter.CheckinTo != "" {
		db = db.Where("checkin_date <= ?", filter.CheckinTo)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, limit := utils.NormalizePage(filter.Page, filter.Limit)
	offset := (page - 1) * limit

	var bookings []entity.Booking
	err := db.
		Preload("Room", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&bookings).Error

	return bookings, total, err
}

func (r *bookingRepository) FindByID(ctx context.Context, id uint) (*entity.Booking, error) {
	var booking entity.Booking
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Room").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&booking, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &booking, err
}

func (r *bookingRepository) FindByCode(ctx context.Context, code string) (*entity.Booking, error) {
	var booking entity.Booking
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Room").
		Where("booking_code = ?", code).
		First(&booking).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &booking, err
}

func (r *bookingRepository) FindByGuestPhone(ctx context.Context, phone string) ([]entity.Booking, error) {
	var bookings []entity.Booking
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Room", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		Where("guest_phone = ?", phone).
		Order("created_at DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) Create(ctx context.Context, booking *entity.Booking) error {
	return DB(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &bookingRepository{db: tx}

		if err := txRepo.LockRoom(ctx, booking.RoomID); err != nil {
			return err
		}

		available, err := txRepo.IsAvailable(ctx, booking.RoomID, booking.CheckinDate, booking.CheckoutDate, nil)
		if err != nil {
			return err
		}
		if !available {
			return errors.New("room is not available for the selected dates")
		}

		return tx.Create(booking).Error
	})
}

func (r *bookingRepository) Update(ctx context.Context, booking *entity.Booking) error {
	return DB(ctx, r.db).WithContext(ctx).Save(booking).Error
}

func (r *bookingRepository) IsAvailable(
	ctx context.Context,
	roomID uint,
	checkin, checkout time.Time,
	excludeBookingID *uint,
) (bool, error) {
	db := DB(ctx, r.db).WithContext(ctx)
	now := time.Now()

	confirmedQ := db.Model(&entity.Booking{}).
		Where("room_id = ?", roomID).
		Where("status = ?", entity.BookingConfirmed).
		Where("checkin_date < ? AND checkout_date > ?", checkout, checkin)

	pendingQ := db.Model(&entity.Booking{}).
		Where("room_id = ?", roomID).
		Where("status = ?", entity.BookingPending).
		Where("expires_at > ?", now).
		Where("checkin_date < ? AND checkout_date > ?", checkout, checkin)

	if excludeBookingID != nil {
		confirmedQ = confirmedQ.Where("id != ?", *excludeBookingID)
		pendingQ = pendingQ.Where("id != ?", *excludeBookingID)
	}

	var confirmedCount, pendingCount int64
	if err := confirmedQ.Count(&confirmedCount).Error; err != nil {
		return false, err
	}
	if confirmedCount > 0 {
		return false, nil
	}
	if err := pendingQ.Count(&pendingCount).Error; err != nil {
		return false, err
	}
	if pendingCount > 0 {
		return false, nil
	}

	var blockedCount int64
	err := db.Model(&entity.BlockedDate{}).
		Where("room_id = ?", roomID).
		Where("date >= ? AND date < ?", checkin, checkout).
		Count(&blockedCount).Error
	if err != nil {
		return false, err
	}

	return blockedCount == 0, nil
}

func (r *bookingRepository) ExpirePendingBookings(ctx context.Context) (int64, error) {
	now := time.Now()
	result := DB(ctx, r.db).WithContext(ctx).
		Model(&entity.Booking{}).
		Where("status = ? AND expires_at <= ?", entity.BookingPending, now).
		Updates(map[string]interface{}{
			"status":        entity.BookingExpired,
			"canceled_at":   now,
			"cancel_reason": "Hết hạn tự động",
		})
	return result.RowsAffected, result.Error
}

func (r *bookingRepository) LockRoom(ctx context.Context, roomID uint) error {
	return DB(ctx, r.db).WithContext(ctx).
		Exec(`SELECT id FROM rooms WHERE id = ? FOR UPDATE`, roomID).Error
}
