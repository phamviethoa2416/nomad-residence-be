package repository

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
	"time"
)

type BlockedDateRepository interface {
	FindByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) ([]entity.BlockedDate, error)
	BulkCreate(ctx context.Context, dates []entity.BlockedDate) error
	DeleteByRoomAndDate(ctx context.Context, roomID uint, date time.Time) error
	DeleteByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) error
	DeleteByRoomAndSource(ctx context.Context, roomID uint, source string) error
}
