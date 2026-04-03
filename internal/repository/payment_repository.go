package repository

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
)

type PaymentRepository interface {
	FindByID(ctx context.Context, id uint) (*entity.Payment, error)
	FindByBookingID(ctx context.Context, bookingID uint) ([]entity.Payment, error)
	FindByQRTransactionID(ctx context.Context, txID string) (*entity.Payment, error)
	Create(ctx context.Context, payment *entity.Payment) error
	Update(ctx context.Context, payment *entity.Payment) error
}
