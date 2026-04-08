package model

import (
	"encoding/json"
	"nomad-residence-be/internal/domain/entity"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func baseFromDomain(b entity.BaseModel) BaseModel {
	m := BaseModel{
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
	if b.DeletedAt != nil {
		m.DeletedAt = gorm.DeletedAt{Time: *b.DeletedAt, Valid: true}
	}
	return m
}

func (b BaseModel) toDomainBase() entity.BaseModel {
	e := entity.BaseModel{
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
	if b.DeletedAt.Valid {
		e.DeletedAt = &b.DeletedAt.Time
	}
	return e
}

func toJSON(raw json.RawMessage) datatypes.JSON {
	if raw == nil {
		return nil
	}
	return datatypes.JSON(raw)
}

func fromJSON(j datatypes.JSON) json.RawMessage {
	if j == nil {
		return nil
	}
	return json.RawMessage(j)
}

func toJSONPtr(raw *json.RawMessage) *datatypes.JSON {
	if raw == nil {
		return nil
	}
	j := datatypes.JSON(*raw)
	return &j
}

func fromJSONPtr(j *datatypes.JSON) *json.RawMessage {
	if j == nil {
		return nil
	}
	raw := json.RawMessage(*j)
	return &raw
}
