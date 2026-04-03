package entity

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
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
	ID     uint `gorm:"primaryKey"`
	RoomID uint `gorm:"not null;index"`

	Name *string

	RuleType PricingRuleType `gorm:"type:varchar(30);not null;index"`

	DateFrom  *time.Time     `gorm:"type:date;index:idx_date_range"`
	DateTo    *time.Time     `gorm:"type:date;index:idx_date_range"`
	DayOfWeek datatypes.JSON `gorm:"column:day_of_week;type:jsonb"      json:"day_of_week"`

	PriceModifier decimal.Decimal `gorm:"type:decimal(12,0);not null"`
	ModifierType  ModifierType    `gorm:"type:varchar(10);default:'fixed'"`

	Priority int  `gorm:"default:0"`
	IsActive bool `gorm:"default:true"`

	BaseModel

	Room Room `gorm:"foreignKey:RoomID" json:"-"`
}
