package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettingRepository interface {
	FindAll(ctx context.Context) ([]entity.Setting, error)
	FindByKey(ctx context.Context, key string) (*entity.Setting, error)
	Upsert(ctx context.Context, setting *entity.Setting) error
	Delete(ctx context.Context, key string) error
}

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) FindAll(ctx context.Context) ([]entity.Setting, error) {
	var settings []entity.Setting
	err := DB(ctx, r.db).WithContext(ctx).
		Order("key ASC").
		Find(&settings).Error
	return settings, err
}

func (r *settingRepository) FindByKey(ctx context.Context, key string) (*entity.Setting, error) {
	var setting entity.Setting
	err := DB(ctx, r.db).WithContext(ctx).
		Where("key = ?", key).
		First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &setting, err
}

func (r *settingRepository) Upsert(ctx context.Context, setting *entity.Setting) error {
	return DB(ctx, r.db).WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "description", "updated_at"}),
		}).
		Create(setting).Error
}

func (r *settingRepository) Delete(ctx context.Context, key string) error {
	return DB(ctx, r.db).WithContext(ctx).
		Where("key = ?", key).
		Delete(&entity.Setting{}).Error
}
