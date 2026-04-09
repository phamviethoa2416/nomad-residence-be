package admin

import (
	"fmt"
	"log/slog"
	"math"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/port"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db          *gorm.DB
	settingRepo port.SettingRepository
	logger      *slog.Logger
}

func NewDashboardHandler(db *gorm.DB, settingRepo port.SettingRepository, logger *slog.Logger) *DashboardHandler {
	return &DashboardHandler{db: db, settingRepo: settingRepo, logger: logger}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now()

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())

	var (
		todayCheckins   int64
		todayCheckouts  int64
		activeBookings  int64
		totalRooms      int64
		pendingBookings int64
		icalSyncErrors  int64
	)

	if err := h.db.WithContext(ctx).Model(&entity.Booking{}).
		Where("checkin_date >= ? AND checkin_date < ? AND status = ?", today, tomorrow, "confirmed").
		Count(&todayCheckins).Error; err != nil {
		h.logger.Error("dashboard: today checkins", slog.Any("error", err))
	}

	if err := h.db.WithContext(ctx).Model(&entity.Booking{}).
		Where("checkout_date >= ? AND checkout_date < ? AND status = ?", today, tomorrow, "confirmed").
		Count(&todayCheckouts).Error; err != nil {
		h.logger.Error("dashboard: today checkouts", slog.Any("error", err))
	}

	if err := h.db.WithContext(ctx).Model(&entity.Booking{}).
		Where("status = ? AND checkin_date <= ? AND checkout_date > ?", "confirmed", today, today).
		Count(&activeBookings).Error; err != nil {
		h.logger.Error("dashboard: active bookings", slog.Any("error", err))
	}

	type monthAgg struct {
		TotalRevenue float64
		Count        int64
	}
	var agg monthAgg
	if err := h.db.WithContext(ctx).Model(&entity.Booking{}).
		Select("COALESCE(SUM(total_amount), 0) as total_revenue, COUNT(id) as count").
		Where("status IN ? AND checkin_date >= ? AND checkin_date <= ?",
			[]string{"confirmed", "completed"}, monthStart, monthEnd).
		Scan(&agg).Error; err != nil {
		h.logger.Error("dashboard: month aggregate", slog.Any("error", err))
	}

	type nightsResult struct {
		Total int64
	}
	var nightsRes nightsResult
	if err := h.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(num_nights), 0) AS total
		FROM bookings
		WHERE status IN ('confirmed','completed')
		  AND checkin_date <= ?
		  AND checkout_date >= ?
	`, monthEnd, monthStart).Scan(&nightsRes).Error; err != nil {
		h.logger.Error("dashboard: nights occupied", slog.Any("error", err))
	}

	if err := h.db.WithContext(ctx).Model(&entity.Room{}).
		Where("status = ?", "active").
		Count(&totalRooms).Error; err != nil {
		h.logger.Error("dashboard: total rooms", slog.Any("error", err))
	}

	if err := h.db.WithContext(ctx).Model(&entity.Booking{}).
		Where("status = ? AND expires_at > ?", "pending", now).
		Count(&pendingBookings).Error; err != nil {
		h.logger.Error("dashboard: pending bookings", slog.Any("error", err))
	}

	if err := h.db.WithContext(ctx).Model(&entity.IcalLink{}).
		Where("sync_status = ? AND is_active = ?", "error", true).
		Count(&icalSyncErrors).Error; err != nil {
		h.logger.Error("dashboard: ical sync errors", slog.Any("error", err))
	}

	daysInMonth := float64(time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day())
	totalRoomNights := float64(totalRooms) * daysInMonth
	occupancyRate := 0.0
	if totalRoomNights > 0 {
		occupancyRate = math.Round(float64(nightsRes.Total)/totalRoomNights*1000) / 10
	}

	c.JSON(200, gin.H{
		"today": gin.H{
			"checkins":        todayCheckins,
			"checkouts":       todayCheckouts,
			"active_bookings": activeBookings,
		},
		"month": gin.H{
			"revenue":        agg.TotalRevenue,
			"total_bookings": agg.Count,
			"occupancy_rate": occupancyRate,
		},
		"pending_actions": gin.H{
			"pending_bookings": pendingBookings,
			"ical_sync_errors": icalSyncErrors,
		},
	})
}

func (h *DashboardHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingRepo.FindAll(c.Request.Context())
	if err != nil {
		h.logger.Error("get_settings failed", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	result := make(map[string]interface{}, len(settings))
	for _, s := range settings {
		result[s.Key] = s.Value
	}

	c.JSON(200, result)
}

func (h *DashboardHandler) UpdateSettings(c *gin.Context) {
	var body request.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if field, msg := body.Validate(); field != "" {
		c.JSON(400, gin.H{"error": msg, "field": field})
		return
	}

	ctx := c.Request.Context()

	// Batch mode
	if len(body.Settings) > 0 {
		for _, item := range body.Settings {
			s := &entity.Setting{
				Key:   item.Key,
				Value: []byte(`"` + item.Value + `"`),
			}
			if err := h.settingRepo.Upsert(ctx, s); err != nil {
				h.logger.Error("update_settings: batch upsert failed", slog.String("key", item.Key), slog.Any("error", err))
				c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
				return
			}
		}
		c.JSON(200, gin.H{"message": fmt.Sprintf("Đã cập nhật %d cài đặt", len(body.Settings))})
		return
	}

	// Single mode
	s := &entity.Setting{
		Key:   *body.Key,
		Value: []byte(`"` + *body.Value + `"`),
	}
	if err := h.settingRepo.Upsert(ctx, s); err != nil {
		h.logger.Error("update_settings: upsert failed", slog.String("key", *body.Key), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, s)
}
