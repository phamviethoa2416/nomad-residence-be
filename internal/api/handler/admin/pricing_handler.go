package admin

import (
	"errors"
	"log/slog"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"

	"github.com/gin-gonic/gin"
)

type PricingHandler struct {
	pricingUsecase port.PricingUsecase
	logger         *slog.Logger
}

func NewPricingHandler(pricingUsecase port.PricingUsecase, logger *slog.Logger) *PricingHandler {
	return &PricingHandler{pricingUsecase: pricingUsecase, logger: logger}
}

func (h *PricingHandler) ListRules(c *gin.Context) {
	var params request.PricingRoomParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	rules, err := h.pricingUsecase.GetRulesByRoomID(c.Request.Context(), params.ID)
	if err != nil {
		h.logger.Error("list_pricing_rules failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, rules)
}

func (h *PricingHandler) CreateRule(c *gin.Context) {
	var params request.PricingRoomParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.CreatePricingRuleRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	body.ApplyDefaults()
	if field, msg := body.Validate(); field != "" {
		c.JSON(400, gin.H{"error": msg, "field": field})
		return
	}

	rule := &entity.PricingRule{
		RoomID:       params.ID,
		Name:         body.Name,
		RuleType:     entity.PricingRuleType(body.RuleType),
		ModifierType: entity.ModifierType(body.ModifierType),
		Priority:     body.Priority,
		IsActive:     *body.IsActive,
	}

	rule.PriceModifier = body.PriceModifier

	if body.DateFrom != nil {
		if t, err := utils.ParseDate(*body.DateFrom); err == nil {
			rule.DateFrom = &t
		}
	}
	if body.DateTo != nil {
		if t, err := utils.ParseDate(*body.DateTo); err == nil {
			rule.DateTo = &t
		}
	}

	if err := h.pricingUsecase.CreateRule(c.Request.Context(), rule); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("create_pricing_rule failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(201, rule)
}

func (h *PricingHandler) UpdateRule(c *gin.Context) {
	var params request.PricingRuleParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.UpdatePricingRuleRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if field, msg := body.Validate(); field != "" {
		c.JSON(400, gin.H{"error": msg, "field": field})
		return
	}

	existing, err := h.pricingUsecase.GetRuleByID(c.Request.Context(), params.RuleID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("update_pricing_rule: get failed", slog.Uint64("rule_id", uint64(params.RuleID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	// Apply partial update
	if body.Name != nil {
		existing.Name = body.Name
	}
	if body.RuleType != nil {
		existing.RuleType = entity.PricingRuleType(*body.RuleType)
	}
	if body.PriceModifier != nil {
		existing.PriceModifier = *body.PriceModifier
	}
	if body.ModifierType != nil {
		existing.ModifierType = entity.ModifierType(*body.ModifierType)
	}
	if body.Priority != nil {
		existing.Priority = *body.Priority
	}
	if body.IsActive != nil {
		existing.IsActive = *body.IsActive
	}
	if body.DateFrom != nil {
		if t, err := utils.ParseDate(*body.DateFrom); err == nil {
			existing.DateFrom = &t
		}
	}
	if body.DateTo != nil {
		if t, err := utils.ParseDate(*body.DateTo); err == nil {
			existing.DateTo = &t
		}
	}
	if body.DayOfWeek != nil {
		existing.DayOfWeek = marshalDayOfWeek(body.DayOfWeek)
	}

	if err := h.pricingUsecase.UpdateRule(c.Request.Context(), existing); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("update_pricing_rule failed", slog.Uint64("rule_id", uint64(params.RuleID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, existing)
}

func (h *PricingHandler) DeleteRule(c *gin.Context) {
	var params request.PricingRuleParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.pricingUsecase.DeleteRule(c.Request.Context(), params.RuleID); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("delete_pricing_rule failed", slog.Uint64("rule_id", uint64(params.RuleID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, gin.H{"message": "Đã xóa rule giá"})
}

func marshalDayOfWeek(days []int) []byte {
	if len(days) == 0 {
		return []byte("[]")
	}
	b := []byte("[")
	for i, d := range days {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, byte('0'+d))
	}
	b = append(b, ']')
	return b
}
