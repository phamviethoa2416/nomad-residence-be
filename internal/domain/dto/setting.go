package dto

import "time"

type UpsertSettingRequest struct {
	Value       interface{} `json:"value"       binding:"required"`
	Description *string     `json:"description" binding:"omitempty,max=255"`
}

type SettingResponse struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Description *string     `json:"description,omitempty"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
