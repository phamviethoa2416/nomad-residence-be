package model

import (
	"nomad-residence-be/internal/domain/entity"
	"time"

	"gorm.io/datatypes"
)

type Setting struct {
	Key   string         `gorm:"primaryKey;type:varchar(100)"`
	Value datatypes.JSON `gorm:"type:jsonb;not null"`

	Description *string
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func SettingFromDomain(e *entity.Setting) *Setting {
	if e == nil {
		return nil
	}
	return &Setting{
		Key:         e.Key,
		Value:       toJSON(e.Value),
		Description: e.Description,
		UpdatedAt:   e.UpdatedAt,
	}
}

func (m *Setting) ToDomain() *entity.Setting {
	if m == nil {
		return nil
	}
	return &entity.Setting{
		Key:         m.Key,
		Value:       fromJSON(m.Value),
		Description: m.Description,
		UpdatedAt:   m.UpdatedAt,
	}
}

func SettingsToDomain(models []Setting) []entity.Setting {
	result := make([]entity.Setting, len(models))
	for i := range models {
		result[i] = *models[i].ToDomain()
	}
	return result
}
