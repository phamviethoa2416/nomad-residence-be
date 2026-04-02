package dto

import "time"

type CreateIcalLinkRequest struct {
	RoomID    uint    `json:"room_id"    binding:"required"`
	Platform  string  `json:"platform"   binding:"required,max=50"`
	ImportURL *string `json:"import_url"`
	ExportURL *string `json:"export_url"`
}

type UpdateIcalLinkRequest struct {
	ImportURL *string `json:"import_url"`
	ExportURL *string `json:"export_url"`
	IsActive  *bool   `json:"is_active"`
}

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
