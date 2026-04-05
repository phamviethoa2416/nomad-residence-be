package response

import (
	"time"

	"github.com/shopspring/decimal"
)

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
