package entity

import "time"

type IcalSyncStatus string

const (
	IcalIdle    IcalSyncStatus = "idle"
	IcalSyncing IcalSyncStatus = "syncing"
	IcalError   IcalSyncStatus = "error"
)

type IcalLink struct {
	ID     uint `json:"id"`
	RoomID uint `json:"room_id"`

	Platform  string  `json:"platform"`
	ImportURL *string `json:"import_url,omitempty"`
	ExportURL *string `json:"export_url,omitempty"`

	LastSyncedAt *time.Time     `json:"last_synced_at,omitempty"`
	SyncStatus   IcalSyncStatus `json:"sync_status"`
	SyncError    *string        `json:"sync_error,omitempty"`

	IsActive bool `json:"is_active"`

	BaseModel

	Room Room `json:"-"`
}
