package repository

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
)

type AdminRepository interface {
	FindAll(ctx context.Context) ([]entity.Admin, error)
	FindByID(ctx context.Context, id uint) (*entity.Admin, error)
	FindByEmail(ctx context.Context, email string) (*entity.Admin, error)
	Create(ctx context.Context, admin *entity.Admin) error
	Update(ctx context.Context, admin *entity.Admin) error
	Delete(ctx context.Context, id uint) error
}
