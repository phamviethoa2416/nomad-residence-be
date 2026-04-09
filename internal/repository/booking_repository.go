package repository

import (
	"context"
	"errors"
	"fmt"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"nomad-residence-be/internal/repository/model"
	"nomad-residence-be/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *bookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) FindAll(ctx context.Context, f filter.BookingFilter) ([]entity.Booking, int64, error) {
	db := DB(ctx, r.db).WithContext(ctx).Model(&model.Booking{})

	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	if f.RoomID != nil {
		db = db.Where("room_id = ?", *f.RoomID)
	}
	if f.GuestPhone != "" {
		db = db.Where("guest_phone = ?", f.GuestPhone)
	}
	if f.CheckinFrom != "" {
		db = db.Where("checkin_date >= ?", f.CheckinFrom)
	}
	if f.CheckinTo != "" {
		db = db.Where("checkin_date <= ?", f.CheckinTo)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, limit := utils.NormalizePage(f.Page, f.Limit)
	offset := (page - 1) * limit

	var bookings []model.Booking
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
	if err != nil {
		return nil, 0, err
	}

	return model.BookingsToDomain(bookings), total, nil
}

func (r *bookingRepository) FindByID(ctx context.Context, id uint) (*entity.Booking, error) {
	var m model.Booking
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Room").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *bookingRepository) FindByCode(ctx context.Context, code string) (*entity.Booking, error) {
	var m model.Booking
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Room").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("booking_code = ?", code).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *bookingRepository) FindByGuestPhone(ctx context.Context, phone string) ([]entity.Booking, error) {
	var bookings []model.Booking
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Room", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		Where("guest_phone = ?", phone).
		Order("created_at DESC").
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}
	return model.BookingsToDomain(bookings), nil
}

func (r *bookingRepository) FindConfirmedOverlapping(ctx context.Context, roomID uint, minDate, maxDate time.Time) ([]entity.Booking, error) {
	rangeEndExclusive := maxDate.AddDate(0, 0, 1)
	var bookings []model.Booking
	err := DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ?", roomID).
		Where("status = ?", entity.BookingConfirmed).
		Where("checkin_date < ? AND checkout_date > ?", rangeEndExclusive, minDate).
		Order("checkin_date ASC").
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}
	return model.BookingsToDomain(bookings), nil
}

func (r *bookingRepository) Create(ctx context.Context, booking *entity.Booking) error {
	m := model.BookingFromDomain(booking)
	if err := DB(ctx, r.db).WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*booking = *m.ToDomain()
	return nil
}

func (r *bookingRepository) Update(ctx context.Context, booking *entity.Booking) error {
	m := model.BookingFromDomain(booking)
	if err := DB(ctx, r.db).WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	*booking = *m.ToDomain()
	return nil
}

func (r *bookingRepository) IsAvailable(
	ctx context.Context,
	roomID uint,
	checkin, checkout time.Time,
	excludeBookingID *uint,
) (bool, error) {
	db := DB(ctx, r.db).WithContext(ctx)
	now := time.Now()

	confirmedQ := db.Model(&model.Booking{}).
		Where("room_id = ?", roomID).
		Where("status = ?", entity.BookingConfirmed).
		Where("checkin_date < ? AND checkout_date > ?", checkout, checkin)

	pendingQ := db.Model(&model.Booking{}).
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
	blockedQ := db.Model(&model.BlockedDate{}).
		Where("room_id = ?", roomID).
		Where("date >= ? AND date < ?", checkin, checkout)
	if excludeBookingID != nil {
		bookingRef := fmt.Sprintf("booking:%d", *excludeBookingID)
		blockedQ = blockedQ.Where("(source_ref IS NULL OR source_ref <> ?)", bookingRef)
	}
	err := blockedQ.Count(&blockedCount).Error
	if err != nil {
		return false, err
	}

	return blockedCount == 0, nil
}

func (r *bookingRepository) LockRoom(ctx context.Context, roomID uint) error {
	return DB(ctx, r.db).WithContext(ctx).
		Exec(`SELECT id FROM rooms WHERE id = ? FOR UPDATE`, roomID).Error
}

func (r *bookingRepository) CancelExpiredPending(ctx context.Context) (int64, error) {
	now := time.Now()
	result := DB(ctx, r.db).WithContext(ctx).
		Model(&model.Booking{}).
		Where("status = ? AND expires_at < ?", entity.BookingPending, now).
		Updates(map[string]interface{}{
			"status": entity.BookingExpired,
		})
	return result.RowsAffected, result.Error
}

func (r *bookingRepository) MarkCompletedPastCheckout(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	result := DB(ctx, r.db).WithContext(ctx).
		Model(&model.Booking{}).
		Where("status = ? AND checkout_date < ?", entity.BookingConfirmed, startOfToday).
		Update("status", entity.BookingCompleted)
	return result.RowsAffected, result.Error
}
