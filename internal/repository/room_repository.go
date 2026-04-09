package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"nomad-residence-be/internal/repository/model"
	"nomad-residence-be/pkg/utils"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *roomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) FindAll(ctx context.Context, f filter.RoomFilter) ([]entity.Room, int64, error) {
	db := r.applyRoomFilter(ctx, DB(ctx, r.db).WithContext(ctx).Model(&model.Room{}), f)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, limit := utils.NormalizePage(f.Page, f.Limit)
	offset := (page - 1) * limit

	var rooms []model.Room
	err := db.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_primary = ?", true).Limit(1)
		}).
		Order("sort_order ASC, id ASC").
		Offset(offset).Limit(limit).
		Find(&rooms).Error
	if err != nil {
		return nil, 0, err
	}

	return model.RoomsToDomain(rooms), total, nil
}

func (r *roomRepository) FindAvailable(
	ctx context.Context,
	f filter.RoomFilter,
	checkin, checkout time.Time,
) ([]entity.Room, int64, error) {
	db := r.applyRoomFilter(ctx, DB(ctx, r.db).WithContext(ctx).Model(&model.Room{}), f)
	db = db.Where(
		`NOT EXISTS (
			SELECT 1
			FROM blocked_dates bd
			WHERE bd.room_id = rooms.id
			AND bd.date >= ? AND bd.date < ?
		)`,
		checkin,
		checkout,
	)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, limit := utils.NormalizePage(f.Page, f.Limit)
	offset := (page - 1) * limit

	var rooms []model.Room
	err := db.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_primary = ?", true).Limit(1)
		}).
		Order("sort_order ASC, id ASC").
		Offset(offset).Limit(limit).
		Find(&rooms).Error
	if err != nil {
		return nil, 0, err
	}

	return model.RoomsToDomain(rooms), total, nil
}

func (r *roomRepository) applyRoomFilter(ctx context.Context, db *gorm.DB, f filter.RoomFilter) *gorm.DB {
	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	if f.RoomType != "" {
		db = db.Where("room_type = ?", f.RoomType)
	}
	if f.City != "" {
		db = db.Where("city = ?", f.City)
	}
	if f.District != "" {
		db = db.Where("district = ?", f.District)
	}
	if f.MinPrice != nil {
		db = db.Where("base_price >= ?", *f.MinPrice)
	}
	if f.MaxPrice != nil {
		db = db.Where("base_price <= ?", *f.MaxPrice)
	}
	if f.MinGuests > 0 {
		db = db.Where("max_guests >= ?", f.MinGuests)
	}
	if f.MaxGuests > 0 {
		db = db.Where("max_guests <= ?", f.MaxGuests)
	}
	if len(f.Amenities) > 0 {
		amenitySubQuery := DB(ctx, r.db).WithContext(ctx).
			Model(&model.RoomAmenity{}).
			Select("room_id").
			Where("amenity_id IN ?", f.Amenities).
			Group("room_id").
			Having("COUNT(DISTINCT amenity_id) = ?", len(f.Amenities))
		db = db.Where("id IN (?)", amenitySubQuery)
	}
	return db
}

func (r *roomRepository) FindByID(ctx context.Context, id uint) (*entity.Room, error) {
	var m model.Room
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Amenities").
		Preload("Amenities.Amenity", func(db *gorm.DB) *gorm.DB { return db.Order("category ASC, name ASC") }).
		First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *roomRepository) FindBySlug(ctx context.Context, slug string) (*entity.Room, error) {
	var m model.Room
	err := DB(ctx, r.db).WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, sort_order ASC")
		}).
		Preload("Amenities").
		Preload("Amenities.Amenity", func(db *gorm.DB) *gorm.DB { return db.Order("category ASC, name ASC") }).
		Where("slug = ?", slug).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *roomRepository) Create(ctx context.Context, room *entity.Room) error {
	m := model.RoomFromDomain(room)
	if err := DB(ctx, r.db).WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*room = *m.ToDomain()
	return nil
}

func (r *roomRepository) Update(ctx context.Context, room *entity.Room) error {
	m := model.RoomFromDomain(room)
	if err := DB(ctx, r.db).WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	*room = *m.ToDomain()
	return nil
}

func (r *roomRepository) Delete(ctx context.Context, id uint) error {
	return DB(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Room{}).
			Where("id = ?", id).
			Update("status", entity.RoomStatusInactive).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&model.Room{}).Error; err != nil {
			return err
		}
		if err := tx.Where("room_id = ?", id).Delete(&model.RoomImage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("room_id = ?", id).Delete(&model.RoomAmenity{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *roomRepository) AddImages(ctx context.Context, images []entity.RoomImage) error {
	if len(images) == 0 {
		return nil
	}
	models := model.RoomImagesFromDomain(images)
	return DB(ctx, r.db).WithContext(ctx).Create(&models).Error
}

func (r *roomRepository) DeleteImage(ctx context.Context, imageID uint) error {
	return DB(ctx, r.db).WithContext(ctx).Delete(&model.RoomImage{}, imageID).Error
}

func (r *roomRepository) ResetPrimaryImages(ctx context.Context, roomID uint) error {
	return DB(ctx, r.db).WithContext(ctx).
		Model(&model.RoomImage{}).
		Where("room_id = ?", roomID).
		Update("is_primary", false).Error
}

func (r *roomRepository) UpdateImageSortOrder(ctx context.Context, imageID uint, sortOrder int) error {
	return DB(ctx, r.db).WithContext(ctx).
		Model(&model.RoomImage{}).
		Where("id = ?", imageID).
		Update("sort_order", sortOrder).Error
}

func (r *roomRepository) AddAmenities(ctx context.Context, amenities []entity.RoomAmenity) error {
	if len(amenities) == 0 {
		return nil
	}
	models := model.RoomAmenitiesFromDomain(amenities)
	return DB(ctx, r.db).WithContext(ctx).Create(&models).Error
}

func (r *roomRepository) DeleteAmenity(ctx context.Context, amenityID uint) error {
	return DB(ctx, r.db).WithContext(ctx).
		Where("amenity_id = ?", amenityID).
		Delete(&model.RoomAmenity{}).Error
}

func (r *roomRepository) ReplaceAmenities(ctx context.Context, roomID uint, amenities []entity.Amenity) error {
	return DB(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("room_id = ?", roomID).Delete(&model.RoomAmenity{}).Error; err != nil {
			return err
		}
		if len(amenities) == 0 {
			return nil
		}

		joins := make([]model.RoomAmenity, 0, len(amenities))
		for _, a := range amenities {
			if a.Name == "" {
				continue
			}

			am := model.Amenity{
				Name:     a.Name,
				Icon:     a.Icon,
				Category: a.Category,
			}
			if am.Category == "" {
				am.Category = "general"
			}

			if err := tx.
				Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "name"}},
					DoUpdates: clause.AssignmentColumns([]string{"icon", "category", "updated_at"}),
				}).
				Create(&am).Error; err != nil {
				return err
			}
			if am.ID == 0 {
				if err := tx.Select("id").Where("name = ?", am.Name).First(&am).Error; err != nil {
					return err
				}
			}

			joins = append(joins, model.RoomAmenity{
				RoomID:    roomID,
				AmenityID: am.ID,
			})
		}

		if len(joins) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&joins).Error
	})
}
