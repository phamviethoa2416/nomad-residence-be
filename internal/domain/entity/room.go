package entity

import (
	"github.com/shopspring/decimal"
)

type RoomStatus string

const (
	RoomStatusActive   RoomStatus = "active"
	RoomStatusInactive RoomStatus = "inactive"
)

type RoomType string

const (
	RoomTypeSingle RoomType = "single"
	RoomTypeDouble RoomType = "double"
	RoomTypeSuite  RoomType = "suite"
	RoomTypeVilla  RoomType = "villa"
)

type Room struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	Slug string `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`

	RoomType RoomType `gorm:"type:varchar(50);not null;index" json:"room_type"`

	Description      *string `gorm:"type:text" json:"description,omitempty"`
	ShortDescription *string `gorm:"type:varchar(500)" json:"short_description,omitempty"`

	MaxGuests    int `gorm:"not null;default:2;check:max_guests > 0" json:"max_guests"`
	NumBedrooms  int `gorm:"not null;default:1;check:num_bedrooms >= 0" json:"num_bedrooms"`
	NumBathrooms int `gorm:"not null;default:1;check:num_bathrooms >= 0" json:"num_bathrooms"`
	NumBeds      int `gorm:"not null;default:1;check:num_beds > 0" json:"num_beds"`

	Area *decimal.Decimal `gorm:"type:decimal(6,1);check:area >= 0" json:"area,omitempty"`

	Address  *string `gorm:"type:text" json:"address,omitempty"`
	District *string `gorm:"type:varchar(100);index" json:"district,omitempty"`
	City     string  `gorm:"type:varchar(100);default:'Hà Nội';index" json:"city"`

	Latitude  *decimal.Decimal `gorm:"type:decimal(10,7)" json:"latitude,omitempty"`
	Longitude *decimal.Decimal `gorm:"type:decimal(10,7)" json:"longitude,omitempty"`

	BasePrice   decimal.Decimal `gorm:"type:decimal(12,0);not null;check:base_price >= 0;index" json:"base_price"`
	CleaningFee decimal.Decimal `gorm:"type:decimal(12,0);default:0;check:cleaning_fee >= 0" json:"cleaning_fee"`

	MinNights int `gorm:"default:1;check:min_nights > 0" json:"min_nights"`
	MaxNights int `gorm:"default:30;check:max_nights >= min_nights" json:"max_nights"`

	CheckinTime  string `gorm:"type:varchar(10);default:'14:00'" json:"checkin_time"`
	CheckoutTime string `gorm:"type:varchar(10);default:'12:00'" json:"checkout_time"`

	HouseRules         *string `gorm:"type:text" json:"house_rules,omitempty"`
	CancellationPolicy *string `gorm:"type:text" json:"cancellation_policy,omitempty"`

	Status    RoomStatus `gorm:"type:varchar(20);default:'active';index" json:"status"`
	SortOrder int        `gorm:"default:0;index" json:"sort_order"`

	Images    []RoomImage   `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"images,omitempty"`
	Amenities []RoomAmenity `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"amenities,omitempty"`

	Bookings     []Booking     `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"-"`
	BlockedDates []BlockedDate `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"-"`
	PricingRules []PricingRule `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"-"`
	IcalLinks    []IcalLink    `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"-"`

	BaseModel
}

func (Room) TableName() string {
	return "rooms"
}
