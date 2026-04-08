package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/repository/model"

	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *adminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) FindAll(ctx context.Context) ([]entity.Admin, error) {
	var models []model.Admin
	err := DB(ctx, r.db).WithContext(ctx).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	return model.AdminsToDomain(models), nil
}

func (r *adminRepository) FindByID(ctx context.Context, id uint) (*entity.Admin, error) {
	var m model.Admin
	err := DB(ctx, r.db).WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *adminRepository) FindByEmail(ctx context.Context, email string) (*entity.Admin, error) {
	var m model.Admin
	err := DB(ctx, r.db).WithContext(ctx).
		Where("email = ?", email).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *adminRepository) Create(ctx context.Context, admin *entity.Admin) error {
	m := model.AdminFromDomain(admin)
	if err := DB(ctx, r.db).WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*admin = *m.ToDomain()
	return nil
}

func (r *adminRepository) Update(ctx context.Context, admin *entity.Admin) error {
	m := model.AdminFromDomain(admin)
	if err := DB(ctx, r.db).WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	*admin = *m.ToDomain()
	return nil
}

func (r *adminRepository) Delete(ctx context.Context, id uint) error {
	return DB(ctx, r.db).WithContext(ctx).Delete(&model.Admin{}, id).Error
}
