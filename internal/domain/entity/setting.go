package entity

import (
	"encoding/json"
	"time"
)

type Setting struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`

	Description *string   `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}
