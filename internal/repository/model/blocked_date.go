package model

import (
	"nomad-residence-be/internal/domain/entity"
	"time"
)

type BlockedDate struct {
	ID     uint               `gorm:"primaryKey"`
	RoomID uint               `gorm:"not null;uniqueIndex:idx_room_date_src"`
	Date   time.Time          `gorm:"type:date;not null;uniqueIndex:idx_room_date_src;index:idx_room_date"`
	Source entity.BlockSource `gorm:"type:varchar(30);not null;uniqueIndex:idx_room_date_src;index"`

	SourceRef *string `gorm:"type:varchar(255)"`
	Reason    *string `gorm:"type:varchar(255)"`

	BaseModel

	Room Room `gorm:"foreignKey:RoomID"`
}

func BlockedDateFromDomain(e *entity.BlockedDate) *BlockedDate {
	if e == nil {
		return nil
	}
	return &BlockedDate{
		ID: e.ID, RoomID: e.RoomID, Date: e.Date, Source: e.Source,
		SourceRef: e.SourceRef, Reason: e.Reason,
		BaseModel: baseFromDomain(e.BaseModel),
	}
}

func BlockedDatesFromDomain(entities []entity.BlockedDate) []BlockedDate {
	result := make([]BlockedDate, len(entities))
	for i := range entities {
		result[i] = *BlockedDateFromDomain(&entities[i])
	}
	return result
}

func (m *BlockedDate) ToDomain() *entity.BlockedDate {
	if m == nil {
		return nil
	}
	return &entity.BlockedDate{
		ID: m.ID, RoomID: m.RoomID, Date: m.Date, Source: m.Source,
		SourceRef: m.SourceRef, Reason: m.Reason,
		BaseModel: m.BaseModel.toDomainBase(),
	}
}

func BlockedDatesToDomain(models []BlockedDate) []entity.BlockedDate {
	result := make([]entity.BlockedDate, len(models))
	for i := range models {
		result[i] = *models[i].ToDomain()
	}
	return result
}
