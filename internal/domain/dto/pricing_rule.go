package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreatePricingRuleRequest struct {
	RoomID        uint            `json:"room_id"        binding:"required"`
	Name          *string         `json:"name"           binding:"omitempty,max=255"`
	RuleType      string          `json:"rule_type"      binding:"required,oneof=date_range day_of_week seasonal"`
	DateFrom      *string         `json:"date_from"      binding:"omitempty,datetime=2006-01-02"`
	DateTo        *string         `json:"date_to"        binding:"omitempty,datetime=2006-01-02"`
	DayOfWeek     []int           `json:"day_of_week"    binding:"omitempty,dive,min=0,max=6"`
	PriceModifier decimal.Decimal `json:"price_modifier" binding:"required"`
	ModifierType  string          `json:"modifier_type"  binding:"required,oneof=fixed percent"`
	Priority      int             `json:"priority"`
	IsActive      *bool           `json:"is_active"`
}

type UpdatePricingRuleRequest struct {
	Name          *string          `json:"name"           binding:"omitempty,max=255"`
	DateFrom      *string          `json:"date_from"      binding:"omitempty,datetime=2006-01-02"`
	DateTo        *string          `json:"date_to"        binding:"omitempty,datetime=2006-01-02"`
	DayOfWeek     []int            `json:"day_of_week"    binding:"omitempty,dive,min=0,max=6"`
	PriceModifier *decimal.Decimal `json:"price_modifier"`
	ModifierType  *string          `json:"modifier_type"  binding:"omitempty,oneof=fixed percent"`
	Priority      *int             `json:"priority"`
	IsActive      *bool            `json:"is_active"`
}

type PricingRuleResponse struct {
	ID            uint            `json:"id"`
	RoomID        uint            `json:"room_id"`
	Name          *string         `json:"name,omitempty"`
	RuleType      string          `json:"rule_type"`
	DateFrom      *time.Time      `json:"date_from,omitempty"`
	DateTo        *time.Time      `json:"date_to,omitempty"`
	DayOfWeek     []int           `json:"day_of_week"`
	PriceModifier decimal.Decimal `json:"price_modifier"`
	ModifierType  string          `json:"modifier_type"`
	Priority      int             `json:"priority"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
