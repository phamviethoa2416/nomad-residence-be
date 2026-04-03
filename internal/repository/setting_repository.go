package repository

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
)

type SettingRepository interface {
	FindAll(ctx context.Context) ([]entity.Setting, error)
	FindByKey(ctx context.Context, key string) (*entity.Setting, error)
	Upsert(ctx context.Context, setting *entity.Setting) error
	Delete(ctx context.Context, key string) error
}
