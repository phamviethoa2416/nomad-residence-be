package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/repository/model"
	"time"

	"gorm.io/gorm"
)

type pricingRuleRepository struct {
	db *gorm.DB
}

func NewPricingRuleRepository(db *gorm.DB) *pricingRuleRepository {
	return &pricingRuleRepository{db: db}
}

func (r *pricingRuleRepository) FindByRoomID(ctx context.Context, roomID uint) ([]entity.PricingRule, error) {
	var rules []model.PricingRule
	err := DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("priority DESC, created_at DESC").
		Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return model.PricingRulesToDomain(rules), nil
}

func (r *pricingRuleRepository) FindByID(ctx context.Context, id uint) (*entity.PricingRule, error) {
	var m model.PricingRule
	err := DB(ctx, r.db).WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *pricingRuleRepository) FindActiveRulesForRange(
	ctx context.Context,
	roomID uint,
	from, to time.Time,
) ([]entity.PricingRule, error) {
	var rules []model.PricingRule

	err := DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ? AND is_active = ?", roomID, true).
		Where(
			r.db.Where("date_from IS NULL OR date_from <= ?", to).
				Where("date_to IS NULL OR date_to >= ?", from),
		).
		Order("priority DESC, created_at DESC").
		Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return model.PricingRulesToDomain(rules), nil
}

func (r *pricingRuleRepository) Create(ctx context.Context, rule *entity.PricingRule) error {
	m := model.PricingRuleFromDomain(rule)
	if err := DB(ctx, r.db).WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*rule = *m.ToDomain()
	return nil
}

func (r *pricingRuleRepository) Update(ctx context.Context, rule *entity.PricingRule) error {
	m := model.PricingRuleFromDomain(rule)
	if err := DB(ctx, r.db).WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	*rule = *m.ToDomain()
	return nil
}

func (r *pricingRuleRepository) Delete(ctx context.Context, id uint) error {
	return DB(ctx, r.db).WithContext(ctx).Delete(&model.PricingRule{}, id).Error
}
