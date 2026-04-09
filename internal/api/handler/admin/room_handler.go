package admin

import (
	"errors"
	"log/slog"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomUsecase port.RoomUsecase
	logger      *slog.Logger
}

func NewRoomHandler(roomUsecase port.RoomUsecase, logger *slog.Logger) *RoomHandler {
	return &RoomHandler{roomUsecase: roomUsecase, logger: logger}
}

func (h *RoomHandler) ListRooms(c *gin.Context) {
	rooms, _, err := h.roomUsecase.ListRooms(c.Request.Context(), filter.RoomFilter{
		Page:  1,
		Limit: 500,
	})
	if err != nil {
		h.logger.Error("list_rooms failed", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, rooms)
}

func (h *RoomHandler) GetRoom(c *gin.Context) {
	var params request.RoomIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	room, err := h.roomUsecase.GetRoomByID(c.Request.Context(), params.ID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("get_room failed", slog.Uint64("id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, room)
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var body request.CreateRoomRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	body.ApplyDefaults()

	room := &entity.Room{
		Name:               body.Name,
		Slug:               body.Slug,
		RoomType:           entity.RoomType(body.RoomType),
		Description:        body.Description,
		ShortDescription:   body.ShortDescription,
		MaxGuests:          body.MaxGuests,
		NumBedrooms:        body.NumBedrooms,
		NumBathrooms:       body.NumBathrooms,
		NumBeds:            body.NumBeds,
		Area:               body.Area,
		Address:            &body.Address,
		District:           body.District,
		City:               body.City,
		Latitude:           body.Latitude,
		Longitude:          body.Longitude,
		BasePrice:          body.BasePrice,
		CleaningFee:        body.CleaningFee,
		MinNights:          body.MinNights,
		MaxNights:          body.MaxNights,
		HouseRules:         body.HouseRules,
		CancellationPolicy: body.CancellationPolicy,
		Status:             entity.RoomStatus(body.Status),
		SortOrder:          body.SortOrder,
	}

	if err := h.roomUsecase.CreateRoom(c.Request.Context(), room); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("create_room failed", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(201, room)
}

func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	var params request.RoomIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.UpdateRoomRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.roomUsecase.GetRoomByID(c.Request.Context(), params.ID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("update_room: get failed", slog.Uint64("id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	if body.Name != nil {
		existing.Name = *body.Name
	}
	if body.Slug != nil {
		existing.Slug = *body.Slug
	}
	if body.RoomType != nil {
		existing.RoomType = entity.RoomType(*body.RoomType)
	}
	if body.Description != nil {
		existing.Description = body.Description
	}
	if body.ShortDescription != nil {
		existing.ShortDescription = body.ShortDescription
	}
	if body.MaxGuests != nil {
		existing.MaxGuests = *body.MaxGuests
	}
	if body.NumBedrooms != nil {
		existing.NumBedrooms = *body.NumBedrooms
	}
	if body.NumBathrooms != nil {
		existing.NumBathrooms = *body.NumBathrooms
	}
	if body.NumBeds != nil {
		existing.NumBeds = *body.NumBeds
	}
	if body.Area != nil {
		existing.Area = body.Area
	}
	if body.Address != nil {
		existing.Address = body.Address
	}
	if body.District != nil {
		existing.District = body.District
	}
	if body.City != nil {
		existing.City = *body.City
	}
	if body.Latitude != nil {
		existing.Latitude = body.Latitude
	}
	if body.Longitude != nil {
		existing.Longitude = body.Longitude
	}
	if body.BasePrice != nil {
		existing.BasePrice = *body.BasePrice
	}
	if body.CleaningFee != nil {
		existing.CleaningFee = *body.CleaningFee
	}
	if body.MinNights != nil {
		existing.MinNights = *body.MinNights
	}
	if body.MaxNights != nil {
		existing.MaxNights = *body.MaxNights
	}
	if body.HouseRules != nil {
		existing.HouseRules = body.HouseRules
	}
	if body.CancellationPolicy != nil {
		existing.CancellationPolicy = body.CancellationPolicy
	}
	if body.Status != nil {
		existing.Status = entity.RoomStatus(*body.Status)
	}
	if body.SortOrder != nil {
		existing.SortOrder = *body.SortOrder
	}

	if err := h.roomUsecase.UpdateRoom(c.Request.Context(), existing); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("update_room failed", slog.Uint64("id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, existing)
}

func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	var params request.RoomIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.roomUsecase.DeleteRoom(c.Request.Context(), params.ID); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("delete_room failed", slog.Uint64("id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, gin.H{"message": "Xóa phòng thành công"})
}

func (h *RoomHandler) AddImage(c *gin.Context) {
	var params request.RoomIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.AddRoomImageRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	images := []entity.RoomImage{{
		URL:       body.URL,
		AltText:   body.AltText,
		IsPrimary: body.IsPrimary,
		SortOrder: body.SortOrder,
	}}

	if err := h.roomUsecase.AddImages(c.Request.Context(), params.ID, images); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("add_image failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(201, images[0])
}

func (h *RoomHandler) DeleteImage(c *gin.Context) {
	var params request.ImageIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.roomUsecase.DeleteImage(c.Request.Context(), params.ImageID); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("delete_image failed", slog.Uint64("image_id", uint64(params.ImageID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, gin.H{"message": "Đã xóa ảnh"})
}

func (h *RoomHandler) ReorderImages(c *gin.Context) {
	var params request.RoomIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.ReorderRoomImagesRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	orders := make([]entity.ImageOrder, len(body.Order))
	for i, o := range body.Order {
		orders[i].ID = o.ID
		orders[i].SortOrder = o.SortOrder
	}

	if err := h.roomUsecase.ReorderImages(c.Request.Context(), params.ID, orders); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("reorder_images failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, gin.H{"message": "Đã sắp xếp lại ảnh"})
}

func (h *RoomHandler) UpdateAmenities(c *gin.Context) {
	var params request.RoomIDParam
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var body request.UpdateAmenitiesRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	amenities := make([]entity.Amenity, len(body.Amenities))
	for i, a := range body.Amenities {
		amenities[i] = entity.Amenity{
			Name:     a.Name,
			Icon:     a.Icon,
			Category: a.Category,
		}
		if amenities[i].Category == "" {
			amenities[i].Category = "general"
		}
	}

	if err := h.roomUsecase.ReplaceAmenities(c.Request.Context(), params.ID, amenities); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message, "code": appErr.Code})
			return
		}
		h.logger.Error("update_amenities failed", slog.Uint64("room_id", uint64(params.ID)), slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	c.JSON(200, gin.H{"message": "Đã cập nhật tiện ích"})
}
