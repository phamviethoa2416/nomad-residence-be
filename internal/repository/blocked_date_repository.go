package repository

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BlockedDateRepository interface {
	FindByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) ([]entity.BlockedDate, error)
	BulkCreate(ctx context.Context, dates []entity.BlockedDate) error
	DeleteByRoomAndDate(ctx context.Context, roomID uint, date time.Time) error
	DeleteByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) error
	DeleteByRoomAndSource(ctx context.Context, roomID uint, source string) error
	GetUnavailableRoomIDs(ctx context.Context, checkin, checkout time.Time) ([]uint, error)
}

type blockedDateRepository struct {
	db *gorm.DB
}

func NewBlockedDateRepository(db *gorm.DB) BlockedDateRepository {
	return &blockedDateRepository{db: db}
}

func (r *blockedDateRepository) FindByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) ([]entity.BlockedDate, error) {
	db := DB(ctx, r.db).WithContext(ctx).
		Where("date >= ? AND date < ?", from, to)

	if roomID != 0 {
		db = db.Where("room_id = ?", roomID)
	}

	var dates []entity.BlockedDate
	err := db.Order("room_id ASC, date ASC").Find(&dates).Error
	return dates, err
}

func (r *blockedDateRepository) BulkCreate(ctx context.Context, dates []entity.BlockedDate) error {
	if len(dates) == 0 {
		return nil
	}

	return DB(ctx, r.db).WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(&dates, 100).Error
}

func (r *blockedDateRepository) DeleteByRoomAndDate(ctx context.Context, roomID uint, date time.Time) error {
	return DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ? AND date = ?", roomID, date).
		Delete(&entity.BlockedDate{}).Error
}

func (r *blockedDateRepository) DeleteByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) error {
	return DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ? AND date >= ? AND date < ?", roomID, from, to).
		Delete(&entity.BlockedDate{}).Error
}

func (r *blockedDateRepository) DeleteByRoomAndSource(ctx context.Context, roomID uint, source string) error {
	db := DB(ctx, r.db).WithContext(ctx)

	if idx := strings.Index(source, ":"); idx > 0 {
		src := source[:idx]
		ref := source[idx+1:]
		return db.
			Where("room_id = ? AND source = ? AND source_ref = ?", roomID, src, ref).
			Delete(&entity.BlockedDate{}).Error
	}

	return db.
		Where("room_id = ? AND source = ?", roomID, source).
		Delete(&entity.BlockedDate{}).Error
}

func (r *blockedDateRepository) GetUnavailableRoomIDs(ctx context.Context, checkin, checkout time.Time) ([]uint, error) {
	type result struct {
		RoomID uint
	}
	var rows []result
	err := DB(ctx, r.db).WithContext(ctx).
		Model(&entity.BlockedDate{}).
		Select("DISTINCT room_id").
		Where("date >= ? AND date < ?", checkin, checkout).
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.RoomID
	}
	return ids, nil
}
