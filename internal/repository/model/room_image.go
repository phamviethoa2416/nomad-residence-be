package model

import "nomad-residence-be/internal/domain/entity"

type RoomImage struct {
	ID     uint `gorm:"primaryKey"`
	RoomID uint `gorm:"not null;index"`

	URL       string  `gorm:"type:varchar(500);not null"`
	AltText   *string `gorm:"type:varchar(255)"`
	IsPrimary bool    `gorm:"default:false"`
	SortOrder int     `gorm:"default:0"`

	BaseModel

	Room Room `gorm:"foreignKey:RoomID"`
}

func RoomImageFromDomain(e *entity.RoomImage) *RoomImage {
	if e == nil {
		return nil
	}
	return &RoomImage{
		ID: e.ID, RoomID: e.RoomID,
		URL: e.URL, AltText: e.AltText,
		IsPrimary: e.IsPrimary, SortOrder: e.SortOrder,
		BaseModel: baseFromDomain(e.BaseModel),
	}
}

func RoomImagesFromDomain(entities []entity.RoomImage) []RoomImage {
	result := make([]RoomImage, len(entities))
	for i := range entities {
		result[i] = *RoomImageFromDomain(&entities[i])
	}
	return result
}

func (m *RoomImage) ToDomain() *entity.RoomImage {
	if m == nil {
		return nil
	}
	return &entity.RoomImage{
		ID: m.ID, RoomID: m.RoomID,
		URL: m.URL, AltText: m.AltText,
		IsPrimary: m.IsPrimary, SortOrder: m.SortOrder,
		BaseModel: m.BaseModel.toDomainBase(),
	}
}
