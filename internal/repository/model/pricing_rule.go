package model

import (
	"nomad-residence-be/internal/domain/entity"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type PricingRule struct {
	ID     uint `gorm:"primaryKey"`
	RoomID uint `gorm:"not null;index"`

	Name *string

	RuleType entity.PricingRuleType `gorm:"type:varchar(30);not null;index"`

	DateFrom  *time.Time     `gorm:"type:date;index:idx_date_range"`
	DateTo    *time.Time     `gorm:"type:date;index:idx_date_range"`
	DayOfWeek datatypes.JSON `gorm:"column:day_of_week;type:jsonb"`

	PriceModifier decimal.Decimal     `gorm:"type:decimal(12,0);not null"`
	ModifierType  entity.ModifierType `gorm:"type:varchar(10);default:'fixed'"`

	Priority int  `gorm:"default:0"`
	IsActive bool `gorm:"default:true"`

	BaseModel

	Room Room `gorm:"foreignKey:RoomID"`
}

func PricingRuleFromDomain(e *entity.PricingRule) *PricingRule {
	if e == nil {
		return nil
	}
	return &PricingRule{
		ID: e.ID, RoomID: e.RoomID, Name: e.Name,
		RuleType: e.RuleType,
		DateFrom: e.DateFrom, DateTo: e.DateTo,
		DayOfWeek:     toJSON(e.DayOfWeek),
		PriceModifier: e.PriceModifier, ModifierType: e.ModifierType,
		Priority: e.Priority, IsActive: e.IsActive,
		BaseModel: baseFromDomain(e.BaseModel),
	}
}

func (m *PricingRule) ToDomain() *entity.PricingRule {
	if m == nil {
		return nil
	}
	return &entity.PricingRule{
		ID: m.ID, RoomID: m.RoomID, Name: m.Name,
		RuleType: m.RuleType,
		DateFrom: m.DateFrom, DateTo: m.DateTo,
		DayOfWeek:     fromJSON(m.DayOfWeek),
		PriceModifier: m.PriceModifier, ModifierType: m.ModifierType,
		Priority: m.Priority, IsActive: m.IsActive,
		BaseModel: m.BaseModel.toDomainBase(),
	}
}

func PricingRulesToDomain(models []PricingRule) []entity.PricingRule {
	result := make([]entity.PricingRule, len(models))
	for i := range models {
		result[i] = *models[i].ToDomain()
	}
	return result
}

type PricingRuleDay struct {
	ID     uint `gorm:"primaryKey"`
	RuleID uint `gorm:"not null;index"`
	Day    int  `gorm:"check:day >= 0 AND day <= 6"`

	Rule PricingRule `gorm:"foreignKey:RuleID"`
}

func (m *PricingRuleDay) ToDomain() *entity.PricingRuleDay {
	if m == nil {
		return nil
	}
	return &entity.PricingRuleDay{ID: m.ID, RuleID: m.RuleID, Day: m.Day}
}
