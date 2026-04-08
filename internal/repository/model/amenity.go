package model

import "nomad-residence-be/internal/domain/entity"

type Amenity struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Icon     *string
	Category string `gorm:"type:varchar(50);index"`

	BaseModel
}

func (m *Amenity) ToDomain() *entity.Amenity {
	if m == nil {
		return nil
	}
	return &entity.Amenity{
		ID: m.ID, Name: m.Name, Icon: m.Icon, Category: m.Category,
		BaseModel: m.BaseModel.toDomainBase(),
	}
}
