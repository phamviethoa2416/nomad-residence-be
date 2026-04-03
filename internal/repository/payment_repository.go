package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	FindByID(ctx context.Context, id uint) (*entity.Payment, error)
	FindByBookingID(ctx context.Context, bookingID uint) ([]entity.Payment, error)
	FindByQRTransactionID(ctx context.Context, txID string) (*entity.Payment, error)
	Create(ctx context.Context, payment *entity.Payment) error
	Update(ctx context.Context, payment *entity.Payment) error
	UpdateManyPendingToSuccess(ctx context.Context, bookingID uint, adminNote string) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) FindByID(ctx context.Context, id uint) (*entity.Payment, error) {
	var payment entity.Payment
	err := DB(ctx, r.db).WithContext(ctx).First(&payment, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &payment, err
}

func (r *paymentRepository) FindByBookingID(ctx context.Context, bookingID uint) ([]entity.Payment, error) {
	var payments []entity.Payment
	err := DB(ctx, r.db).WithContext(ctx).
		Where("booking_id = ?", bookingID).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) FindByQRTransactionID(ctx context.Context, txID string) (*entity.Payment, error) {
	var payment entity.Payment
	err := DB(ctx, r.db).WithContext(ctx).
		Where("external_transaction_id = ?", txID).
		First(&payment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &payment, err
}

func (r *paymentRepository) FindPendingByBookingID(ctx context.Context, bookingID uint) (*entity.Payment, error) {
	var payment entity.Payment
	err := DB(ctx, r.db).WithContext(ctx).
		Where("booking_id = ? AND status = ?", bookingID, entity.PaymentPending).
		First(&payment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &payment, err
}

func (r *paymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	return DB(ctx, r.db).WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	return DB(ctx, r.db).WithContext(ctx).Save(payment).Error
}

func (r *paymentRepository) UpdateManyPendingToSuccess(ctx context.Context, bookingID uint, adminNote string) error {
	return DB(ctx, r.db).WithContext(ctx).
		Model(&entity.Payment{}).
		Where("booking_id = ? AND status = ?", bookingID, entity.PaymentPending).
		Updates(map[string]interface{}{
			"status":     entity.PaymentPaid,
			"admin_note": adminNote,
		}).Error
}
