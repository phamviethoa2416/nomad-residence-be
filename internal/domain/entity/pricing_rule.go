package entity

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type ModifierType string

const (
	ModifierFixed   ModifierType = "fixed"
	ModifierPercent ModifierType = "percent"
)

type PricingRuleType string

const (
	RuleDateRange PricingRuleType = "date_range"
	RuleDayOfWeek PricingRuleType = "day_of_week"
)

type PricingRule struct {
	ID     uint `json:"id"`
	RoomID uint `json:"room_id"`

	Name *string `json:"name,omitempty"`

	RuleType PricingRuleType `json:"rule_type"`

	DateFrom  *time.Time      `json:"date_from,omitempty"`
	DateTo    *time.Time      `json:"date_to,omitempty"`
	DayOfWeek json.RawMessage `json:"day_of_week"`

	PriceModifier decimal.Decimal `json:"price_modifier"`
	ModifierType  ModifierType    `json:"modifier_type"`

	Priority int  `json:"priority"`
	IsActive bool `json:"is_active"`

	BaseModel

	Room Room `json:"-"`
}
