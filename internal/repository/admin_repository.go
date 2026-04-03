package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"

	"gorm.io/gorm"
)

type AdminRepository interface {
	FindAll(ctx context.Context) ([]entity.Admin, error)
	FindByID(ctx context.Context, id uint) (*entity.Admin, error)
	FindByEmail(ctx context.Context, email string) (*entity.Admin, error)
	Create(ctx context.Context, admin *entity.Admin) error
	Update(ctx context.Context, admin *entity.Admin) error
	Delete(ctx context.Context, id uint) error
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}
func (r *adminRepository) FindAll(ctx context.Context) ([]entity.Admin, error) {
	var admins []entity.Admin
	err := DB(ctx, r.db).WithContext(ctx).
		Order("created_at ASC").
		Find(&admins).Error
	return admins, err
}

func (r *adminRepository) FindByID(ctx context.Context, id uint) (*entity.Admin, error) {
	var admin entity.Admin
	err := DB(ctx, r.db).WithContext(ctx).First(&admin, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &admin, err
}

func (r *adminRepository) FindByEmail(ctx context.Context, email string) (*entity.Admin, error) {
	var admin entity.Admin
	err := DB(ctx, r.db).WithContext(ctx).
		Where("email = ?", email).
		First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &admin, err
}

func (r *adminRepository) Create(ctx context.Context, admin *entity.Admin) error {
	return DB(ctx, r.db).WithContext(ctx).Create(admin).Error
}

func (r *adminRepository) Update(ctx context.Context, admin *entity.Admin) error {
	return DB(ctx, r.db).WithContext(ctx).Save(admin).Error
}

func (r *adminRepository) Delete(ctx context.Context, id uint) error {
	return DB(ctx, r.db).WithContext(ctx).Delete(&entity.Admin{}, id).Error
}
