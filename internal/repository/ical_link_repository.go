package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/repository/model"

	"gorm.io/gorm"
)

type icalLinkRepository struct {
	db *gorm.DB
}

func NewIcalLinkRepository(db *gorm.DB) *icalLinkRepository {
	return &icalLinkRepository{db: db}
}

func (r *icalLinkRepository) FindByRoomID(ctx context.Context, roomID uint) ([]entity.IcalLink, error) {
	var links []model.IcalLink
	err := DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&links).Error
	if err != nil {
		return nil, err
	}
	return model.IcalLinksToDomain(links), nil
}

func (r *icalLinkRepository) FindByID(ctx context.Context, id uint) (*entity.IcalLink, error) {
	var m model.IcalLink
	err := DB(ctx, r.db).WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *icalLinkRepository) FindActiveImportLinks(ctx context.Context) ([]entity.IcalLink, error) {
	var links []model.IcalLink
	err := DB(ctx, r.db).WithContext(ctx).
		Where("is_active = ? AND import_url IS NOT NULL", true).
		Preload("Room", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		Order("id ASC").
		Find(&links).Error
	if err != nil {
		return nil, err
	}
	return model.IcalLinksToDomain(links), nil
}

func (r *icalLinkRepository) Create(ctx context.Context, link *entity.IcalLink) error {
	m := model.IcalLinkFromDomain(link)
	if err := DB(ctx, r.db).WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*link = *m.ToDomain()
	return nil
}

func (r *icalLinkRepository) Update(ctx context.Context, link *entity.IcalLink) error {
	m := model.IcalLinkFromDomain(link)
	if err := DB(ctx, r.db).WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	*link = *m.ToDomain()
	return nil
}

func (r *icalLinkRepository) Delete(ctx context.Context, id uint) error {
	return DB(ctx, r.db).WithContext(ctx).Delete(&model.IcalLink{}, id).Error
}
