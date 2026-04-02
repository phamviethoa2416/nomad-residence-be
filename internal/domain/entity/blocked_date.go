package entity

import "time"

type BlockSource string

const (
	BlockManual  BlockSource = "manual"
	BlockBooking BlockSource = "booking"
	BlockIcal    BlockSource = "ical"
)

type BlockedDate struct {
	ID     uint        `gorm:"primaryKey"`
	RoomID uint        `gorm:"not null;uniqueIndex:idx_room_date_src"`
	Date   time.Time   `gorm:"type:date;not null;uniqueIndex:idx_room_date_src;index:idx_room_date"`
	Source BlockSource `gorm:"type:varchar(30);not null;uniqueIndex:idx_room_date_src;index"`

	SourceRef *string `gorm:"type:varchar(255)"`
	Reason    *string `gorm:"type:varchar(255)"`

	BaseModel

	Room Room `gorm:"foreignKey:RoomID" json:"-"`
}
