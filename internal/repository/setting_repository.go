package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/repository/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) *settingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) FindAll(ctx context.Context) ([]entity.Setting, error) {
	var settings []model.Setting
	err := DB(ctx, r.db).WithContext(ctx).
		Order("key ASC").
		Find(&settings).Error
	if err != nil {
		return nil, err
	}
	return model.SettingsToDomain(settings), nil
}

func (r *settingRepository) FindByKey(ctx context.Context, key string) (*entity.Setting, error) {
	var m model.Setting
	err := DB(ctx, r.db).WithContext(ctx).
		Where("key = ?", key).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *settingRepository) Upsert(ctx context.Context, setting *entity.Setting) error {
	m := model.SettingFromDomain(setting)
	if err := DB(ctx, r.db).WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "description", "updated_at"}),
		}).
		Create(m).Error; err != nil {
		return err
	}
	*setting = *m.ToDomain()
	return nil
}

func (r *settingRepository) Delete(ctx context.Context, key string) error {
	return DB(ctx, r.db).WithContext(ctx).
		Where("key = ?", key).
		Delete(&model.Setting{}).Error
}
