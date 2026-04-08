package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/repository/model"

	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *paymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) FindByID(ctx context.Context, id uint) (*entity.Payment, error) {
	var m model.Payment
	err := DB(ctx, r.db).WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *paymentRepository) FindByBookingID(ctx context.Context, bookingID uint) ([]entity.Payment, error) {
	var payments []model.Payment
	err := DB(ctx, r.db).WithContext(ctx).
		Where("booking_id = ?", bookingID).
		Order("created_at DESC").
		Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return model.PaymentsToDomain(payments), nil
}

func (r *paymentRepository) FindByQRTransactionID(ctx context.Context, txID string) (*entity.Payment, error) {
	var m model.Payment
	err := DB(ctx, r.db).WithContext(ctx).
		Where("external_transaction_id = ?", txID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *paymentRepository) FindPendingByBookingID(ctx context.Context, bookingID uint) (*entity.Payment, error) {
	var m model.Payment
	err := DB(ctx, r.db).WithContext(ctx).
		Where("booking_id = ? AND status = ?", bookingID, entity.PaymentPending).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *paymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	m := model.PaymentFromDomain(payment)
	if err := DB(ctx, r.db).WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*payment = *m.ToDomain()
	return nil
}

func (r *paymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	m := model.PaymentFromDomain(payment)
	if err := DB(ctx, r.db).WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	*payment = *m.ToDomain()
	return nil
}

func (r *paymentRepository) UpdateManyPendingToSuccess(ctx context.Context, bookingID uint, adminNote string) error {
	return DB(ctx, r.db).WithContext(ctx).
		Model(&model.Payment{}).
		Where("booking_id = ? AND status = ?", bookingID, entity.PaymentPending).
		Updates(map[string]interface{}{
			"status":     entity.PaymentPaid,
			"admin_note": adminNote,
		}).Error
}
