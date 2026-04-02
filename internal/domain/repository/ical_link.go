package repository

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
)

type IcalLinkRepository interface {
	FindByRoomID(ctx context.Context, roomID uint) ([]entity.IcalLink, error)
	FindByID(ctx context.Context, id uint) (*entity.IcalLink, error)
	FindActiveImportLinks(ctx context.Context) ([]entity.IcalLink, error)
	Create(ctx context.Context, link *entity.IcalLink) error
	Update(ctx context.Context, link *entity.IcalLink) error
	Delete(ctx context.Context, id uint) error
}
