package model

import "nomad-residence-be/internal/domain/entity"

type RoomAmenity struct {
	RoomID    uint `gorm:"primaryKey"`
	AmenityID uint `gorm:"primaryKey"`

	Room    Room    `gorm:"foreignKey:RoomID"`
	Amenity Amenity `gorm:"foreignKey:AmenityID"`
}

func RoomAmenityFromDomain(e *entity.RoomAmenity) *RoomAmenity {
	if e == nil {
		return nil
	}
	return &RoomAmenity{RoomID: e.RoomID, AmenityID: e.AmenityID}
}

func RoomAmenitiesFromDomain(entities []entity.RoomAmenity) []RoomAmenity {
	result := make([]RoomAmenity, len(entities))
	for i := range entities {
		result[i] = *RoomAmenityFromDomain(&entities[i])
	}
	return result
}

func (m *RoomAmenity) ToDomain() *entity.RoomAmenity {
	if m == nil {
		return nil
	}
	d := &entity.RoomAmenity{RoomID: m.RoomID, AmenityID: m.AmenityID}
	if am := m.Amenity.ToDomain(); am != nil {
		d.Amenity = *am
	}
	return d
}
