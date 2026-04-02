package entity

import "time"

type IcalSyncStatus string

const (
	IcalIdle    IcalSyncStatus = "idle"
	IcalSyncing IcalSyncStatus = "syncing"
	IcalError   IcalSyncStatus = "error"
)

type IcalLink struct {
	ID     uint `gorm:"primaryKey"`
	RoomID uint `gorm:"not null;index"`

	Platform string `gorm:"type:varchar(50);not null"`

	ImportURL *string `gorm:"type:text"`
	ExportURL *string `gorm:"type:text"`

	LastSyncedAt *time.Time

	SyncStatus IcalSyncStatus `gorm:"type:varchar(20);default:'idle'"`
	SyncError  *string

	IsActive bool `gorm:"default:true"`

	BaseModel

	Room Room `gorm:"foreignKey:RoomID" json:"-"`
}
