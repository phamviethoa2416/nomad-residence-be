package entity

import "time"

type BlockSource string

const (
	BlockManual  BlockSource = "manual"
	BlockBooking BlockSource = "booking"
	BlockIcal    BlockSource = "ical"
)

type BlockedDate struct {
	ID     uint        `json:"id"`
	RoomID uint        `json:"room_id"`
	Date   time.Time   `json:"date"`
	Source BlockSource `json:"source"`

	SourceRef *string `json:"source_ref,omitempty"`
	Reason    *string `json:"reason,omitempty"`

	BaseModel

	Room Room `json:"-"`
}
