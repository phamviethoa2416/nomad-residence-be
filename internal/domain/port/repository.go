package port

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"time"
)

type AdminRepository interface {
	FindAll(ctx context.Context) ([]entity.Admin, error)
	FindByID(ctx context.Context, id uint) (*entity.Admin, error)
	FindByEmail(ctx context.Context, email string) (*entity.Admin, error)
	Create(ctx context.Context, admin *entity.Admin) error
	Update(ctx context.Context, admin *entity.Admin) error
	Delete(ctx context.Context, id uint) error
}

type RoomRepository interface {
	FindAll(ctx context.Context, filter filter.RoomFilter) ([]entity.Room, int64, error)
	FindByID(ctx context.Context, id uint) (*entity.Room, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Room, error)
	Create(ctx context.Context, room *entity.Room) error
	Update(ctx context.Context, room *entity.Room) error
	Delete(ctx context.Context, id uint) error

	AddImages(ctx context.Context, images []entity.RoomImage) error
	DeleteImage(ctx context.Context, imageID uint) error
	ResetPrimaryImages(ctx context.Context, roomID uint) error
	UpdateImageSortOrder(ctx context.Context, imageID uint, sortOrder int) error

	AddAmenities(ctx context.Context, amenities []entity.RoomAmenity) error
	DeleteAmenity(ctx context.Context, amenityID uint) error
	ReplaceAmenities(ctx context.Context, roomID uint, amenities []entity.Amenity) error
}

type BookingRepository interface {
	FindAll(ctx context.Context, filter filter.BookingFilter) ([]entity.Booking, int64, error)
	FindByID(ctx context.Context, id uint) (*entity.Booking, error)
	FindByCode(ctx context.Context, code string) (*entity.Booking, error)
	FindByGuestPhone(ctx context.Context, phone string) ([]entity.Booking, error)

	Create(ctx context.Context, booking *entity.Booking) error
	Update(ctx context.Context, booking *entity.Booking) error

	IsAvailable(ctx context.Context, roomID uint, checkin, checkout time.Time, excludeBookingID *uint) (bool, error)
	LockRoom(ctx context.Context, roomID uint) error

	CancelExpiredPending(ctx context.Context) (int64, error)
	MarkCompletedPastCheckout(ctx context.Context) (int64, error)
	FindConfirmedOverlapping(ctx context.Context, roomID uint, minDate, maxDate time.Time) ([]entity.Booking, error)
}

type BlockedDateRepository interface {
	FindByID(ctx context.Context, id uint) (*entity.BlockedDate, error)
	FindByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) ([]entity.BlockedDate, error)
	BulkCreate(ctx context.Context, dates []entity.BlockedDate) error
	DeleteByRoomAndDate(ctx context.Context, roomID uint, date time.Time) error
	DeleteByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) error
	DeleteByRoomAndSource(ctx context.Context, roomID uint, source string) error
	GetUnavailableRoomIDs(ctx context.Context, checkin, checkout time.Time) ([]uint, error)
}

type PaymentRepository interface {
	FindByID(ctx context.Context, id uint) (*entity.Payment, error)
	FindByBookingID(ctx context.Context, bookingID uint) ([]entity.Payment, error)
	FindByQRTransactionID(ctx context.Context, txID string) (*entity.Payment, error)
	FindPendingByBookingID(ctx context.Context, bookingID uint) (*entity.Payment, error)
	Create(ctx context.Context, payment *entity.Payment) error
	Update(ctx context.Context, payment *entity.Payment) error
	UpdateManyPendingToSuccess(ctx context.Context, bookingID uint, adminNote string) error
}

type PricingRuleRepository interface {
	FindByRoomID(ctx context.Context, roomID uint) ([]entity.PricingRule, error)
	FindByID(ctx context.Context, id uint) (*entity.PricingRule, error)
	FindActiveRulesForRange(ctx context.Context, roomID uint, from, to time.Time) ([]entity.PricingRule, error)

	Create(ctx context.Context, rule *entity.PricingRule) error
	Update(ctx context.Context, rule *entity.PricingRule) error
	Delete(ctx context.Context, id uint) error
}

type IcalLinkRepository interface {
	FindByRoomID(ctx context.Context, roomID uint) ([]entity.IcalLink, error)
	FindByID(ctx context.Context, id uint) (*entity.IcalLink, error)
	FindActiveImportLinks(ctx context.Context) ([]entity.IcalLink, error)

	Create(ctx context.Context, link *entity.IcalLink) error
	Update(ctx context.Context, link *entity.IcalLink) error
	Delete(ctx context.Context, id uint) error
}

type SettingRepository interface {
	FindAll(ctx context.Context) ([]entity.Setting, error)
	FindByKey(ctx context.Context, key string) (*entity.Setting, error)
	Upsert(ctx context.Context, setting *entity.Setting) error
	Delete(ctx context.Context, key string) error
}
