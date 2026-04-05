package request

import (
	"nomad-residence-be/pkg/validator"

	"github.com/shopspring/decimal"
)

type PricingRoomParams struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type PricingRuleParams struct {
	RuleID uint `uri:"ruleId" binding:"required,gt=0"`
}

type CreatePricingRuleRequest struct {
	RoomID        uint            `json:"room_id"        binding:"required"`
	Name          *string         `json:"name"           binding:"omitempty,max=255"`
	RuleType      string          `json:"rule_type"      binding:"required,oneof=date_range day_of_week"`
	DateFrom      *string         `json:"date_from"      binding:"omitempty,datetime=2006-01-02"`
	DateTo        *string         `json:"date_to"        binding:"omitempty,datetime=2006-01-02"`
	DayOfWeek     []int           `json:"day_of_week"    binding:"omitempty,dive,min=0,max=6"`
	PriceModifier decimal.Decimal `json:"price_modifier" binding:"required"`
	ModifierType  string          `json:"modifier_type"  binding:"required,oneof=fixed percent"`
	Priority      int             `json:"priority"`
	IsActive      *bool           `json:"is_active"`
}

func (r *CreatePricingRuleRequest) ApplyDefaults() {
	if r.ModifierType == "" {
		r.ModifierType = "fixed"
	}
	if r.DayOfWeek == nil {
		r.DayOfWeek = []int{}
	}
	if r.IsActive == nil {
		t := true
		r.IsActive = &t
	}
}

func (r *CreatePricingRuleRequest) Validate() (string, string) {
	if r.DateFrom != nil && r.DateTo != nil {
		return validator.ValidateDateRangeFromTo(*r.DateFrom, *r.DateTo)
	}
	return "", ""
}

type UpdatePricingRuleRequest struct {
	Name          *string          `json:"name"           binding:"omitempty,max=255"`
	RuleType      *string          `json:"rule_type"      binding:"omitempty,oneof=date_range day_of_week"`
	DateFrom      *string          `json:"date_from"      binding:"omitempty,datetime=2006-01-02"`
	DateTo        *string          `json:"date_to"        binding:"omitempty,datetime=2006-01-02"`
	DayOfWeek     []int            `json:"day_of_week"    binding:"omitempty,dive,min=0,max=6"`
	PriceModifier *decimal.Decimal `json:"price_modifier"`
	ModifierType  *string          `json:"modifier_type"  binding:"omitempty,oneof=fixed percent"`
	Priority      *int             `json:"priority"       binding:"omitempty,min=0"`
	IsActive      *bool            `json:"is_active"`
}

func (r *UpdatePricingRuleRequest) Validate() (string, string) {
	if r.DateFrom != nil && r.DateTo != nil {
		return validator.ValidateDateRangeFromTo(*r.DateFrom, *r.DateTo)
	}
	return "", ""
}
