package repository

import (
	"context"
	"nomad-residence-be/internal/domain/dto"
	"nomad-residence-be/internal/domain/entity"
	"time"
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
}
