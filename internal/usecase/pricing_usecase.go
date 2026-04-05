package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/repository"
	apperrors "nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"time"

	"github.com/shopspring/decimal"
)

type DailyPrice struct {
	Date  string          `json:"date"`
	Price decimal.Decimal `json:"price"`
}

type PriceBreakdown struct {
	NumNights   int             `json:"num_nights"`
	BaseTotal   decimal.Decimal `json:"base_total"`
	CleaningFee decimal.Decimal `json:"cleaning_fee"`
	Discount    decimal.Decimal `json:"discount"`
	Total       decimal.Decimal `json:"total"`
	DailyPrices []DailyPrice    `json:"daily_prices"`
}

type PricingUsecase struct {
	pricingRepo repository.PricingRuleRepository
	roomRepo    repository.RoomRepository
	bookingRepo repository.BookingRepository
	logger      *slog.Logger
}

func NewPricingUsecase(
	pricingRepo repository.PricingRuleRepository,
	roomRepo repository.RoomRepository,
	bookingRepo repository.BookingRepository,
	logger *slog.Logger,
) *PricingUsecase {
	return &PricingUsecase{
		pricingRepo: pricingRepo,
		roomRepo:    roomRepo,
		bookingRepo: bookingRepo,
		logger:      logger,
	}
}

// CalculatePrice looks up a room and checks availability before calculating the price.
// This is the public entry point for price queries (e.g. from API handlers).
// Internal callers that already verified availability should use CalculatePriceForRoom.
func (uc *PricingUsecase) CalculatePrice(
	ctx context.Context,
	roomID uint,
	checkin, checkout time.Time,
) (*PriceBreakdown, error) {
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, apperrors.ErrRoomNotFound
	}

	available, err := uc.bookingRepo.IsAvailable(ctx, roomID, checkin, checkout, nil)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, apperrors.ErrRoomNotAvailable
	}

	return uc.CalculatePriceForRoom(ctx, room, checkin, checkout)
}

func (uc *PricingUsecase) CalculatePriceForRoom(
	ctx context.Context,
	room *entity.Room,
	checkin, checkout time.Time,
) (*PriceBreakdown, error) {
	checkin = utils.TruncateToDate(checkin)
	checkout = utils.TruncateToDate(checkout)

	numNights := utils.NightsBetween(checkin, checkout)
	if numNights <= 0 {
		return nil, apperrors.ErrInvalidDates
	}

	rules, err := uc.pricingRepo.FindActiveRulesForRange(ctx, room.ID, checkin, checkout)
	if err != nil {
		uc.logger.Error("failed to fetch pricing rules",
			slog.Uint64("room_id", uint64(room.ID)),
			slog.Any("error", err),
		)
		return nil, err
	}

	dates := utils.GetDateRange(checkin, checkout)
	dailyPrices := make([]DailyPrice, 0, len(dates))
	baseTotal := decimal.Zero

	for _, date := range dates {
		nightPrice := uc.resolvePriceForDate(room.BasePrice, rules, date)
		dailyPrices = append(dailyPrices, DailyPrice{
			Date:  utils.FormatDate(date),
			Price: nightPrice,
		})
		baseTotal = baseTotal.Add(nightPrice)
	}

	total := baseTotal.Add(room.CleaningFee)

	return &PriceBreakdown{
		NumNights:   numNights,
		BaseTotal:   baseTotal,
		CleaningFee: room.CleaningFee,
		Discount:    decimal.Zero,
		Total:       total,
		DailyPrices: dailyPrices,
	}, nil
}

// resolvePriceForDate applies active pricing rules to a specific date and returns
// the final night price. Rules are evaluated in priority order (highest first).
// A date_range rule with fixed modifier replaces the base price entirely;
// a percent modifier adjusts the base price by the given percentage.
// day_of_week rules work similarly but match on the weekday.
func (uc *PricingUsecase) resolvePriceForDate(
	basePrice decimal.Decimal,
	rules []entity.PricingRule,
	date time.Time,
) decimal.Decimal {
	price := basePrice
	weekday := int(date.Weekday())

	for _, rule := range rules {
		if !uc.ruleMatchesDate(rule, date, weekday) {
			continue
		}

		price = uc.applyModifier(basePrice, rule)
		break
	}

	if price.LessThan(decimal.Zero) {
		return decimal.Zero
	}
	return price
}

func (uc *PricingUsecase) ruleMatchesDate(rule entity.PricingRule, date time.Time, weekday int) bool {
	switch rule.RuleType {
	case entity.RuleDateRange:
		if rule.DateFrom != nil && date.Before(*rule.DateFrom) {
			return false
		}
		if rule.DateTo != nil && date.After(*rule.DateTo) {
			return false
		}
		return true

	case entity.RuleDayOfWeek:
		if len(rule.DayOfWeek) == 0 {
			return false
		}
		var days []int
		if err := json.Unmarshal(rule.DayOfWeek, &days); err != nil {
			return false
		}
		for _, d := range days {
			if d == weekday {
				return true
			}
		}
		return false
	}

	return false
}

func (uc *PricingUsecase) applyModifier(basePrice decimal.Decimal, rule entity.PricingRule) decimal.Decimal {
	switch rule.ModifierType {
	case entity.ModifierFixed:
		return rule.PriceModifier
	case entity.ModifierPercent:
		adjustment := basePrice.Mul(rule.PriceModifier).Div(decimal.NewFromInt(100))
		return basePrice.Add(adjustment)
	default:
		return basePrice
	}
}

// CRUD operations for pricing rules

func (uc *PricingUsecase) GetRulesByRoomID(ctx context.Context, roomID uint) ([]entity.PricingRule, error) {
	return uc.pricingRepo.FindByRoomID(ctx, roomID)
}

func (uc *PricingUsecase) GetRuleByID(ctx context.Context, id uint) (*entity.PricingRule, error) {
	rule, err := uc.pricingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, apperrors.ErrPricingRuleNotFound
	}
	return rule, nil
}

func (uc *PricingUsecase) CreateRule(ctx context.Context, rule *entity.PricingRule) error {
	return uc.pricingRepo.Create(ctx, rule)
}

func (uc *PricingUsecase) UpdateRule(ctx context.Context, rule *entity.PricingRule) error {
	return uc.pricingRepo.Update(ctx, rule)
}

func (uc *PricingUsecase) DeleteRule(ctx context.Context, id uint) error {
	existing, err := uc.pricingRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperrors.ErrPricingRuleNotFound
	}
	return uc.pricingRepo.Delete(ctx, id)
}
