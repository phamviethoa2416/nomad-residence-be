package entity

type RoomAmenity struct {
	RoomID    uint `gorm:"primaryKey"`
	AmenityID uint `gorm:"primaryKey"`

	Room    Room    `gorm:"foreignKey:RoomID"`
	Amenity Amenity `gorm:"foreignKey:AmenityID"`
}
