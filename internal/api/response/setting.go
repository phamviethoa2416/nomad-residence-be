package response

import "time"

type SettingResponse struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Description *string     `json:"description,omitempty"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
