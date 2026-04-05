package filter

import (
	"nomad-residence-be/internal/domain/entity"

	"github.com/shopspring/decimal"
)

type RoomFilter struct {
	RoomType  entity.RoomType
	Status    entity.RoomStatus
	City      string
	District  string
	MinPrice  *decimal.Decimal
	MaxPrice  *decimal.Decimal
	Amenities []uint
	MinGuests int
	MaxGuests int
	Page      int
	Limit     int
}
