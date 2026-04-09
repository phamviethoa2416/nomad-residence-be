package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/port"
	"nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"time"

	"github.com/shopspring/decimal"
)

type PricingUsecase struct {
	pricingRepo port.PricingRuleRepository
	roomRepo    port.RoomRepository
	bookingRepo port.BookingRepository
	logger      *slog.Logger
}

func NewPricingUsecase(
	pricingRepo port.PricingRuleRepository,
	roomRepo port.RoomRepository,
	bookingRepo port.BookingRepository,
	logger *slog.Logger,
) *PricingUsecase {
	return &PricingUsecase{
		pricingRepo: pricingRepo,
		roomRepo:    roomRepo,
		bookingRepo: bookingRepo,
		logger:      logger,
	}
}

func (uc *PricingUsecase) CalculatePrice(
	ctx context.Context,
	roomID uint,
	checkin, checkout time.Time,
) (*entity.PriceBreakdown, error) {
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.ErrRoomNotFound
	}

	available, err := uc.bookingRepo.IsAvailable(ctx, roomID, checkin, checkout, nil)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, errors.ErrRoomNotAvailable
	}

	return uc.CalculatePriceForRoom(ctx, room, checkin, checkout)
}

func (uc *PricingUsecase) CalculatePriceForRoom(
	ctx context.Context,
	room *entity.Room,
	checkin, checkout time.Time,
) (*entity.PriceBreakdown, error) {
	checkin = utils.TruncateToDate(checkin)
	checkout = utils.TruncateToDate(checkout)

	numNights := utils.NightsBetween(checkin, checkout)
	if numNights <= 0 {
		return nil, errors.ErrInvalidDates
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
	dailyPrices := make([]entity.DailyPrice, 0, len(dates))
	baseTotal := decimal.Zero

	for _, date := range dates {
		nightPrice := uc.resolvePriceForDate(room.BasePrice, rules, date)
		dailyPrices = append(dailyPrices, entity.DailyPrice{
			Date:  utils.FormatDate(date),
			Price: nightPrice,
		})
		baseTotal = baseTotal.Add(nightPrice)
	}

	total := baseTotal.Add(room.CleaningFee)

	return &entity.PriceBreakdown{
		NumNights:   numNights,
		BaseTotal:   baseTotal,
		CleaningFee: room.CleaningFee,
		Discount:    decimal.Zero,
		Total:       total,
		DailyPrices: dailyPrices,
	}, nil
}

func (uc *PricingUsecase) resolvePriceForDate(
	basePrice decimal.Decimal,
	rules []entity.PricingRule,
	date time.Time,
) decimal.Decimal {
	price := basePrice
	weekday := int(date.Weekday())
	fixedApplied := false

	for _, rule := range rules {
		if !uc.ruleMatchesDate(rule, date, weekday) {
			continue
		}

		switch rule.ModifierType {
		case entity.ModifierFixed:
			if fixedApplied {
				continue
			}
			price = rule.PriceModifier
			fixedApplied = true
		case entity.ModifierPercent:
			adjustment := basePrice.Mul(rule.PriceModifier).Div(decimal.NewFromInt(100))
			price = price.Add(adjustment)
		}
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

func (uc *PricingUsecase) GetRulesByRoomID(ctx context.Context, roomID uint) ([]entity.PricingRule, error) {
	return uc.pricingRepo.FindByRoomID(ctx, roomID)
}

func (uc *PricingUsecase) GetRuleByID(ctx context.Context, id uint) (*entity.PricingRule, error) {
	rule, err := uc.pricingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, errors.ErrPricingRuleNotFound
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
		return errors.ErrPricingRuleNotFound
	}
	return uc.pricingRepo.Delete(ctx, id)
}
