package repository

import (
	"context"
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/repository/model"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type blockedDateRepository struct {
	db *gorm.DB
}

func NewBlockedDateRepository(db *gorm.DB) *blockedDateRepository {
	return &blockedDateRepository{db: db}
}

func (r *blockedDateRepository) FindByID(ctx context.Context, id uint) (*entity.BlockedDate, error) {
	var m model.BlockedDate
	err := DB(ctx, r.db).WithContext(ctx).First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *blockedDateRepository) FindByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) ([]entity.BlockedDate, error) {
	db := DB(ctx, r.db).WithContext(ctx).
		Where("date >= ? AND date < ?", from, to)

	if roomID != 0 {
		db = db.Where("room_id = ?", roomID)
	}

	var dates []model.BlockedDate
	err := db.Order("room_id ASC, date ASC").Find(&dates).Error
	if err != nil {
		return nil, err
	}
	return model.BlockedDatesToDomain(dates), nil
}

func (r *blockedDateRepository) BulkCreate(ctx context.Context, dates []entity.BlockedDate) error {
	if len(dates) == 0 {
		return nil
	}
	models := model.BlockedDatesFromDomain(dates)
	return DB(ctx, r.db).WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(&models, 100).Error
}

func (r *blockedDateRepository) DeleteByRoomAndDate(ctx context.Context, roomID uint, date time.Time) error {
	return DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ? AND date = ?", roomID, date).
		Delete(&model.BlockedDate{}).Error
}

func (r *blockedDateRepository) DeleteByRoomAndRange(ctx context.Context, roomID uint, from, to time.Time) error {
	return DB(ctx, r.db).WithContext(ctx).
		Where("room_id = ? AND date >= ? AND date < ?", roomID, from, to).
		Delete(&model.BlockedDate{}).Error
}

func (r *blockedDateRepository) DeleteByRoomAndSource(ctx context.Context, roomID uint, source string) error {
	db := DB(ctx, r.db).WithContext(ctx)

	if idx := strings.Index(source, ":"); idx > 0 {
		src := source[:idx]
		ref := source[idx+1:]
		return db.
			Where("room_id = ? AND source = ? AND source_ref = ?", roomID, src, ref).
			Delete(&model.BlockedDate{}).Error
	}

	return db.
		Where("room_id = ? AND source = ?", roomID, source).
		Delete(&model.BlockedDate{}).Error
}

func (r *blockedDateRepository) GetUnavailableRoomIDs(ctx context.Context, checkin, checkout time.Time) ([]uint, error) {
	type result struct {
		RoomID uint
	}
	var rows []result
	err := DB(ctx, r.db).WithContext(ctx).
		Model(&model.BlockedDate{}).
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
