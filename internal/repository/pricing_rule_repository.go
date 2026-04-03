package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"time"

	"gorm.io/gorm"
)

type PricingRuleRepository interface {
	FindByRoomID(ctx context.Context, roomID uint) ([]entity.PricingRule, error)
	FindByID(ctx context.Context, id uint) (*entity.PricingRule, error)
	FindActiveRulesForRange(ctx context.Context, roomID uint, from, to time.Time) ([]entity.PricingRule, error)

	Create(ctx context.Context, rule *entity.PricingRule) error
	Update(ctx context.Context, rule *entity.PricingRule) error
	Delete(ctx context.Context, id uint) error
}

type pricingRuleRepository struct {
	db *gorm.DB
}

func NewPricingRuleRepository(db *gorm.DB) PricingRuleRepository {
	return &pricingRuleRepository{db: db}
}

func (r *pricingRuleRepository) FindByRoomID(ctx context.Context, roomID uint) ([]entity.PricingRule, error) {
	var rules []entity.PricingRule
	err := DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("priority DESC, created_at DESC").
		Find(&rules).Error
	return rules, err
}

func (r *pricingRuleRepository) FindByID(ctx context.Context, id uint) (*entity.PricingRule, error) {
	var rule entity.PricingRule
	err := DB(ctx, r.db).WithContext(ctx).First(&rule, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rule, err
}

func (r *pricingRuleRepository) FindActiveRulesForRange(
	ctx context.Context,
	roomID uint,
	from, to time.Time,
) ([]entity.PricingRule, error) {
	var rules []entity.PricingRule

	err := DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ? AND is_active = ?", roomID, true).
		Where(
			r.db.Where("date_from IS NULL OR date_from <= ?", to).
				Where("date_to IS NULL OR date_to >= ?", from),
		).
		Order("priority DESC, created_at DESC").
		Find(&rules).Error

	return rules, err
}

func (r *pricingRuleRepository) Create(ctx context.Context, rule *entity.PricingRule) error {
	return DB(ctx, r.db).WithContext(ctx).Create(rule).Error
}

func (r *pricingRuleRepository) Update(ctx context.Context, rule *entity.PricingRule) error {
	return DB(ctx, r.db).WithContext(ctx).Save(rule).Error
}

func (r *pricingRuleRepository) Delete(ctx context.Context, id uint) error {
	return DB(ctx, r.db).WithContext(ctx).Delete(&entity.PricingRule{}, id).Error
}
