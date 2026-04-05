package filter

import (
	"nomad-residence-be/internal/domain/entity"
)

type BookingFilter struct {
	Status      entity.BookingStatus
	RoomID      *uint
	GuestPhone  string
	CheckinFrom string
	CheckinTo   string
	Page        int
	Limit       int
}
