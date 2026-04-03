package usecase

import (
	"context"
	"encoding/json"
	"nomad-residence-be/internal/domain/dto"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/repository"
	"nomad-residence-be/pkg/utils"
)

type PricingUsecase struct {
	roomRepo repository.RoomRepository
	ruleRepo repository.PricingRuleRepository
}

func NewPricingUsecase(
	roomRepo repository.RoomRepository,
	ruleRepo repository.PricingRuleRepository,
) *PricingUsecase {
	return &PricingUsecase{
		roomRepo: roomRepo,
		ruleRepo: ruleRepo,
	}
}

func (u *PricingUsecase) GetRules(ctx context.Context, roomID uint) ([]entity.PricingRule, error) {
	return u.ruleRepo.FindByRoomID(ctx, roomID)
}

func (u *PricingUsecase) CreateRule(ctx context.Context, req dto.CreatePricingRuleRequest) (*entity.PricingRule, error) {
	dowJson, _ := json.Marshal(req.DayOfWeek)

	rule := &entity.PricingRule{
		RoomID:        req.RoomID,
		Name:          req.Name,
		RuleType:      req.RuleType,
		PriceModifier: req.PriceModifier,
		ModifierType:  req.ModifierType,
		Priority:      req.Priority,
		DayOfWeek:     dowJson,
	}

	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	} else {
		rule.IsActive = true
	}

	if req.DateFrom != nil {
		t, err := utils.ParseDate(*req.DateFrom)
		if err != nil {
			return nil, utils.New(400, "INVALID_DATE", "Ngày đến không hợp lệ")
		}
		rule.DateFrom = &t
	}

	if req.DateTo != nil {
		t, err := utils.ParseDate(*req.DateTo)
		if err != nil {
			return nil, utils.New(400, "INVALID_DATE", "Ngày đi không hợp lệ")
		}
		rule.DateTo = &t
	}

	if err := u.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (u *PricingUsecase) UpdateRule(ctx context.Context, ruleID uint, req dto.UpdatePricingRuleRequest) (*entity.PricingRule, error) {
	rule, err := u.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, utils.ErrPricingRuleNotFound
	}

	if req.Name != nil {
		rule.Name = req.Name
	}
	if req.PriceModifier != nil {
		rule.PriceModifier = *req.PriceModifier
	}
	if req.ModifierType != nil {
		rule.ModifierType = *req.ModifierType
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.DayOfWeek != nil {
		dowJSON, _ := json.Marshal(req.DayOfWeek)
		rule.DayOfWeek = dowJSON
	}
	if req.DateFrom != nil {
		t, _ := utils.ParseDate(*req.DateFrom)
		rule.DateFrom = &t
	}
	if req.DateTo != nil {
		t, _ := utils.ParseDate(*req.DateTo)
		rule.DateTo = &t
	}

	if err := u.ruleRepo.Update(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (u *PricingUsecase) DeleteRule(ctx context.Context, ruleID uint) error {
	rule, err := u.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return err
	}
	if rule == nil {
		return utils.ErrPricingRuleNotFound
	}
	return u.ruleRepo.Delete(ctx, ruleID)
}
