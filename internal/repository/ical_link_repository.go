package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"

	"gorm.io/gorm"
)

type IcalLinkRepository interface {
	FindByRoomID(ctx context.Context, roomID uint) ([]entity.IcalLink, error)
	FindByID(ctx context.Context, id uint) (*entity.IcalLink, error)
	FindActiveImportLinks(ctx context.Context) ([]entity.IcalLink, error)

	Create(ctx context.Context, link *entity.IcalLink) error
	Update(ctx context.Context, link *entity.IcalLink) error
	Delete(ctx context.Context, id uint) error
}

type icalLinkRepository struct {
	db *gorm.DB
}

func NewIcalLinkRepository(db *gorm.DB) IcalLinkRepository {
	return &icalLinkRepository{db: db}
}

func (r *icalLinkRepository) FindByRoomID(ctx context.Context, roomID uint) ([]entity.IcalLink, error) {
	var links []entity.IcalLink
	err := DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Find(&links).Error
	return links, err
}

func (r *icalLinkRepository) FindByID(ctx context.Context, id uint) (*entity.IcalLink, error) {
	var link entity.IcalLink
	err := DB(ctx, r.db).WithContext(ctx).First(&link, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &link, err
}

func (r *icalLinkRepository) FindActiveImportLinks(ctx context.Context) ([]entity.IcalLink, error) {
	var links []entity.IcalLink
	err := DB(ctx, r.db).WithContext(ctx).
		Where("is_active = ? AND import_url IS NOT NULL", true).
		Preload("Room", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name")
		}).
		Order("id ASC").
		Find(&links).Error
	return links, err
}

func (r *icalLinkRepository) Create(ctx context.Context, link *entity.IcalLink) error {
	return DB(ctx, r.db).WithContext(ctx).Create(link).Error
}

func (r *icalLinkRepository) Update(ctx context.Context, link *entity.IcalLink) error {
	return DB(ctx, r.db).WithContext(ctx).Save(link).Error
}

func (r *icalLinkRepository) Delete(ctx context.Context, id uint) error {
	return DB(ctx, r.db).WithContext(ctx).Delete(&entity.IcalLink{}, id).Error
}
