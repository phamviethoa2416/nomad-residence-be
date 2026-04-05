package response

import "time"

type IcalLinkResponse struct {
	ID           uint       `json:"id"`
	RoomID       uint       `json:"room_id"`
	Platform     string     `json:"platform"`
	ImportURL    *string    `json:"import_url,omitempty"`
	ExportURL    *string    `json:"export_url,omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	SyncStatus   string     `json:"sync_status"`
	SyncError    *string    `json:"sync_error,omitempty"`
	IsActive     bool       `json:"is_active"`
}
