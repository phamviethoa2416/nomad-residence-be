package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"nomad-residence-be/pkg/utils"

	"gorm.io/gorm"
)

type RoomRepository interface {
	FindAll(ctx context.Context, filter filter.RoomFilter) ([]entity.Room, int64, error)
	FindByID(ctx context.Context, id uint) (*entity.Room, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Room, error)
	Create(ctx context.Context, room *entity.Room) error
	Update(ctx context.Context, room *entity.Room) error
	Delete(ctx context.Context, id uint) error

	AddImages(ctx context.Context, images []entity.RoomImage) error
	DeleteImage(ctx context.Context, imageID uint) error
	ResetPrimaryImages(ctx context.Context, roomID uint) error
	UpdateImageSortOrder(ctx context.Context, imageID uint, sortOrder int) error

	AddAmenities(ctx context.Context, amenities []entity.RoomAmenity) error
	DeleteAmenity(ctx context.Context, amenityID uint) error
	ReplaceAmenities(ctx context.Context, roomID uint, amenities []entity.RoomAmenity) error
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) FindAll(ctx context.Context, filter filter.RoomFilter) ([]entity.Room, int64, error) {
	db := DB(ctx, r.db).WithContext(ctx).Model(&entity.Room{})

	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.RoomType != "" {
		db = db.Where("room_type = ?", filter.RoomType)
	}
	if filter.City != "" {
		db = db.Where("city = ?", filter.City)
	}
	if filter.District != "" {
		db = db.Where("district = ?", filter.District)
	}
	if filter.MinPrice != nil {
		db = db.Where("base_price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		db = db.Where("base_price <= ?", *filter.MaxPrice)
	}
	if filter.MinGuests > 0 {
		db = db.Where("max_guests >= ?", filter.MinGuests)
	}
	if filter.MaxGuests > 0 {
		db = db.Where("max_guests <= ?", filter.MaxGuests)
	}
	if len(filter.Amenities) > 0 {
		amenitySubQuery := DB(ctx, r.db).WithContext(ctx).
			Model(&entity.RoomAmenity{}).
			Select("room_id").
			Where("amenity_id IN ?", filter.Amenities).
			Group("room_id").
			Having("COUNT(DISTINCT amenity_id) = ?", len(filter.Amenities))
		db = db.Where("id IN (?)", amenitySubQuery)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, limit := utils.NormalizePage(filter.Page, filter.Limit)
	offset := (page - 1) * limit

	var rooms []entity.Room
	err := db.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_primary = ?", true).Limit(1)
		}).
		Order("sort_order ASC, id ASC").
		Offset(offset).Limit(limit).
		Find(&rooms).Error

	return rooms, total, err
}

func (r *roomRepository) FindByID(ctx context.Context, id uint) (*entity.Room, error) {
	var room entity.Room
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Amenities", func(db *gorm.DB) *gorm.DB {
			return db.Order("category ASC")
		}).
		First(&room, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &room, err
}

func (r *roomRepository) FindBySlug(ctx context.Context, slug string) (*entity.Room, error) {
	var room entity.Room
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, sort_order ASC")
		}).
		Preload("Amenities", func(db *gorm.DB) *gorm.DB {
			return db.Order("category ASC")
		}).
		Where("slug = ?", slug).
		First(&room).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &room, err
}

func (r *roomRepository) Create(ctx context.Context, room *entity.Room) error {
	return DB(ctx, r.db).WithContext(ctx).Create(room).Error
}

func (r *roomRepository) Update(ctx context.Context, room *entity.Room) error {
	return DB(ctx, r.db).WithContext(ctx).Save(room).Error
}

func (r *roomRepository) Delete(ctx context.Context, id uint) error {
	return DB(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Room{}).
			Where("id = ?", id).
			Update("status", entity.RoomStatusInactive).Error; err != nil {
			return err
		}

		if err := tx.Where("id = ?", id).Delete(&entity.Room{}).Error; err != nil {
			return err
		}

		if err := tx.Where("room_id = ?", id).Delete(&entity.RoomImage{}).Error; err != nil {
			return err
		}

		if err := tx.Where("room_id = ?", id).Delete(&entity.RoomAmenity{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *roomRepository) AddImages(ctx context.Context, images []entity.RoomImage) error {
	if len(images) == 0 {
		return nil
	}

	return DB(ctx, r.db).WithContext(ctx).Create(&images).Error
}

func (r *roomRepository) DeleteImage(ctx context.Context, imageID uint) error {
	return DB(ctx, r.db).WithContext(ctx).
		Delete(&entity.RoomImage{}, imageID).Error
}

func (r *roomRepository) ResetPrimaryImages(ctx context.Context, roomID uint) error {
	return DB(ctx, r.db).WithContext(ctx).
		Model(&entity.RoomImage{}).
		Where("room_id = ?", roomID).
		Update("is_primary", false).Error
}

func (r *roomRepository) UpdateImageSortOrder(ctx context.Context, imageID uint, sortOrder int) error {
	return DB(ctx, r.db).WithContext(ctx).
		Model(&entity.RoomImage{}).
		Where("id = ?", imageID).
		Update("sort_order", sortOrder).Error
}

func (r *roomRepository) AddAmenities(ctx context.Context, amenities []entity.RoomAmenity) error {
	if len(amenities) == 0 {
		return nil
	}
	return DB(ctx, r.db).WithContext(ctx).Create(&amenities).Error
}

func (r *roomRepository) DeleteAmenity(ctx context.Context, amenityID uint) error {
	return DB(ctx, r.db).WithContext(ctx).
		Delete(&entity.RoomAmenity{}, amenityID).Error
}

func (r *roomRepository) ReplaceAmenities(ctx context.Context, roomID uint, amenities []entity.RoomAmenity) error {
	return DB(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("room_id = ?", roomID).Delete(&entity.RoomAmenity{}).Error; err != nil {
			return err
		}
		if len(amenities) == 0 {
			return nil
		}
		return tx.Create(&amenities).Error
	})
}
