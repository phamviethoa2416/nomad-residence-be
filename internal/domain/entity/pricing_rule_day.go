package entity

type PricingRuleDay struct {
	ID     uint `gorm:"primaryKey"`
	RuleID uint `gorm:"not null;index"`
	Day    int  `gorm:"check:day >= 0 AND day <= 6"`

	Rule PricingRule `gorm:"foreignKey:RuleID"`
}
