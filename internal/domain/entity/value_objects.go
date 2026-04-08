package entity

import "github.com/shopspring/decimal"

type DailyPrice struct {
	Date  string          `json:"date"`
	Price decimal.Decimal `json:"price"`
}

type PricingRuleDay struct {
	ID     uint `json:"id"`
	RuleID uint `json:"rule_id"`
	Day    int  `json:"day"`

	Rule PricingRule `json:"-"`
}

type PriceBreakdown struct {
	NumNights   int             `json:"num_nights"`
	BaseTotal   decimal.Decimal `json:"base_total"`
	CleaningFee decimal.Decimal `json:"cleaning_fee"`
	Discount    decimal.Decimal `json:"discount"`
	Total       decimal.Decimal `json:"total"`
	DailyPrices []DailyPrice    `json:"daily_prices"`
}

type CalendarDay struct {
	Date      string `json:"date"`
	Available bool   `json:"available"`
	Source    string `json:"source,omitempty"`
}

type IcalSyncResult struct {
	LinkID   uint   `json:"link_id"`
	RoomID   uint   `json:"room_id"`
	Platform string `json:"platform"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Imported int    `json:"imported"`
}

type ImageOrder struct {
	ID        uint
	SortOrder int
}
