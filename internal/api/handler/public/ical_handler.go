package public

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IcalHandler struct {
	icalUsecase port.IcalUsecase
	tokenSecret string
	logger      *slog.Logger
}

func NewIcalHandler(icalUsecase port.IcalUsecase, tokenSecret string, logger *slog.Logger) *IcalHandler {
	return &IcalHandler{icalUsecase: icalUsecase, tokenSecret: tokenSecret, logger: logger}
}

func (h *IcalHandler) ExportIcal(c *gin.Context) {
	var params request.IcalExportParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var query request.IcalExportQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !h.verifyToken(params.RoomID, query.Token) {
		c.JSON(403, gin.H{
			"error": apperrors.New(403, "FORBIDDEN_ICAL", "Mã xác thực (token) không hợp lệ").Message,
			"code":  "FORBIDDEN_ICAL",
		})
		return
	}

	icsContent, err := h.icalUsecase.ExportRoomIcal(c.Request.Context(), params.RoomID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("failed to generate ical", slog.Uint64("room_id", uint64(params.RoomID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Không thể tạo iCal cho phòng này"})
		return
	}

	if icsContent == "" {
		c.JSON(500, gin.H{"error": "Không thể tạo iCal cho phòng này", "code": "ICAL_GENERATION_FAILED"})
		return
	}

	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="room-%d.ics"`, params.RoomID))
	c.Header("Cache-Control", "public, max-age=300")
	c.String(200, "%s", icsContent)
}

func (h *IcalHandler) GenerateExportURL(roomID uint) string {
	mac := hmac.New(sha256.New, []byte(h.tokenSecret))
	mac.Write([]byte(strconv.FormatUint(uint64(roomID), 10)))
	token := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("/api/v1/ical/%d/calendar.ics?token=%s", roomID, token)
}

func (h *IcalHandler) verifyToken(roomID uint, token string) bool {
	if token == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.tokenSecret))
	mac.Write([]byte(strconv.FormatUint(uint64(roomID), 10)))
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}
