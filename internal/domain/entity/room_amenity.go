package entity

type RoomAmenity struct {
	RoomID    uint `json:"room_id"`
	AmenityID uint `json:"amenity_id"`

	Room    Room    `json:"-"`
	Amenity Amenity `json:"-"`
}
