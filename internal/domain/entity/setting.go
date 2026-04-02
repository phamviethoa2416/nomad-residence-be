package entity

import (
	"time"

	"gorm.io/datatypes"
)

type Setting struct {
	Key   string         `gorm:"primaryKey;type:varchar(100)"`
	Value datatypes.JSON `gorm:"type:jsonb;not null"`

	Description *string
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
