package model

import (
	"nomad-residence-be/internal/domain/entity"
	"time"
)

type IcalLink struct {
	ID     uint `gorm:"primaryKey"`
	RoomID uint `gorm:"not null;index"`

	Platform  string  `gorm:"type:varchar(50);not null"`
	ImportURL *string `gorm:"type:text"`
	ExportURL *string `gorm:"type:text"`

	LastSyncedAt *time.Time
	SyncStatus   entity.IcalSyncStatus `gorm:"type:varchar(20);default:'idle'"`
	SyncError    *string

	IsActive bool `gorm:"default:true"`

	BaseModel

	Room Room `gorm:"foreignKey:RoomID"`
}

func IcalLinkFromDomain(e *entity.IcalLink) *IcalLink {
	if e == nil {
		return nil
	}
	return &IcalLink{
		ID: e.ID, RoomID: e.RoomID,
		Platform: e.Platform, ImportURL: e.ImportURL, ExportURL: e.ExportURL,
		LastSyncedAt: e.LastSyncedAt, SyncStatus: e.SyncStatus, SyncError: e.SyncError,
		IsActive:  e.IsActive,
		BaseModel: baseFromDomain(e.BaseModel),
	}
}

func (m *IcalLink) ToDomain() *entity.IcalLink {
	if m == nil {
		return nil
	}
	l := &entity.IcalLink{
		ID: m.ID, RoomID: m.RoomID,
		Platform: m.Platform, ImportURL: m.ImportURL, ExportURL: m.ExportURL,
		LastSyncedAt: m.LastSyncedAt, SyncStatus: m.SyncStatus, SyncError: m.SyncError,
		IsActive:  m.IsActive,
		BaseModel: m.BaseModel.toDomainBase(),
	}
	if m.Room.ID != 0 {
		l.Room = *m.Room.ToDomain()
	}
	return l
}

func IcalLinksToDomain(models []IcalLink) []entity.IcalLink {
	result := make([]entity.IcalLink, len(models))
	for i := range models {
		result[i] = *models[i].ToDomain()
	}
	return result
}
