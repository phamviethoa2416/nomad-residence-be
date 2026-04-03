package repository

import (
	"context"
	"nomad-residence-be/internal/domain/dto"
	"nomad-residence-be/internal/domain/entity"
)

type RoomRepository interface {
	FindAll(ctx context.Context, filter dto.RoomFilterRequest) ([]entity.Room, int64, error)
	FindByID(ctx context.Context, id uint) (*entity.Room, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Room, error)
	Create(ctx context.Context, room *entity.Room) error
	Update(ctx context.Context, room *entity.Room) error
	Delete(ctx context.Context, id uint) error

	AddImages(ctx context.Context, images []entity.RoomImage) error
	DeleteImage(ctx context.Context, imageID uint) error

	AddAmenities(ctx context.Context, amenities []entity.RoomAmenity) error
	DeleteAmenity(ctx context.Context, amenityID uint) error
	ReplaceAmenities(ctx context.Context, roomID uint, amenities []entity.RoomAmenity) error
}
