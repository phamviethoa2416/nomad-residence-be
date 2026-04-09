package admin

import (
	"errors"
	"log/slog"
	"nomad-residence-be/internal/api/handler/public"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"

	"github.com/gin-gonic/gin"
)

type IcalHandler struct {
	icalUsecase port.IcalUsecase
	icalPubH    *public.IcalHandler
	logger      *slog.Logger
}

func NewIcalHandler(icalUsecase port.IcalUsecase, icalPubH *public.IcalHandler, logger *slog.Logger) *IcalHandler {
	return &IcalHandler{icalUsecase: icalUsecase, icalPubH: icalPubH, logger: logger}
}

func (h *IcalHandler) ListIcalLinks(c *gin.Context) {
	var params request.IcalRoomParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	links, err := h.icalUsecase.GetLinksByRoomID(c.Request.Context(), params.ID)
	if err != nil {
		h.logger.Error("list_ical_links failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, links)
}

func (h *IcalHandler) AddIcalLink(c *gin.Context) {
	var params request.IcalRoomParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.AddIcalLinkRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	exportURL := h.icalPubH.GenerateExportURL(params.ID)
	link := &entity.IcalLink{
		RoomID:    params.ID,
		Platform:  body.Platform,
		ImportURL: body.ImportURL,
		ExportURL: &exportURL,
		IsActive:  true,
	}

	if err := h.icalUsecase.CreateLink(c.Request.Context(), link); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("add_ical_link failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(201, link)
}

func (h *IcalHandler) DeleteIcalLink(c *gin.Context) {
	var params request.IcalLinkParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.icalUsecase.DeleteLink(c.Request.Context(), params.LinkID); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("delete_ical_link failed", slog.Uint64("link_id", uint64(params.LinkID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, gin.H{"message": "Đã xóa liên kết iCal"})
}

func (h *IcalHandler) SyncIcalLink(c *gin.Context) {
	var params request.IcalLinkParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result, err := h.icalUsecase.SyncSingleLinkByID(c.Request.Context(), params.LinkID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("sync_ical_link failed", slog.Uint64("link_id", uint64(params.LinkID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, result)
}

func (h *IcalHandler) GetExportURL(c *gin.Context) {
	var params request.IcalRoomParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"export_url": h.icalPubH.GenerateExportURL(params.ID)})
}
