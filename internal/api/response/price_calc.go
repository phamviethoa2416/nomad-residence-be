package response

import "github.com/shopspring/decimal"

type PriceCalcResult struct {
	NumNights   int             `json:"num_nights"`
	BaseTotal   decimal.Decimal `json:"base_total"`
	CleaningFee decimal.Decimal `json:"cleaning_fee"`
	Total       decimal.Decimal `json:"total"`
}
