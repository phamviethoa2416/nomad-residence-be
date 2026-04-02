package repository

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
	"time"
)

type PricingRuleRepository interface {
	FindByRoomID(ctx context.Context, roomID uint) ([]entity.PricingRule, error)
	FindByID(ctx context.Context, id uint) (*entity.PricingRule, error)
	FindActiveRulesForRange(ctx context.Context, roomID uint, from, to time.Time) ([]entity.PricingRule, error)
	Create(ctx context.Context, rule *entity.PricingRule) error
	Update(ctx context.Context, rule *entity.PricingRule) error
	Delete(ctx context.Context, id uint) error
}
