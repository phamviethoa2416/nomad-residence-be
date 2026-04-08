package entity

import "github.com/shopspring/decimal"

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
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`

	RoomType RoomType `json:"room_type"`

	Description      *string `json:"description,omitempty"`
	ShortDescription *string `json:"short_description,omitempty"`

	MaxGuests    int `json:"max_guests"`
	NumBedrooms  int `json:"num_bedrooms"`
	NumBathrooms int `json:"num_bathrooms"`
	NumBeds      int `json:"num_beds"`

	Area *decimal.Decimal `json:"area,omitempty"`

	Address  *string `json:"address,omitempty"`
	District *string `json:"district,omitempty"`
	City     string  `json:"city"`

	Latitude  *decimal.Decimal `json:"latitude,omitempty"`
	Longitude *decimal.Decimal `json:"longitude,omitempty"`

	BasePrice   decimal.Decimal `json:"base_price"`
	CleaningFee decimal.Decimal `json:"cleaning_fee"`

	MinNights int `json:"min_nights"`
	MaxNights int `json:"max_nights"`

	CheckinTime  string `json:"checkin_time"`
	CheckoutTime string `json:"checkout_time"`

	HouseRules         *string `json:"house_rules,omitempty"`
	CancellationPolicy *string `json:"cancellation_policy,omitempty"`

	Status    RoomStatus `json:"status"`
	SortOrder int        `json:"sort_order"`

	Images    []RoomImage   `json:"images,omitempty"`
	Amenities []RoomAmenity `json:"amenities,omitempty"`

	Bookings     []Booking     `json:"-"`
	BlockedDates []BlockedDate `json:"-"`
	PricingRules []PricingRule `json:"-"`
	IcalLinks    []IcalLink    `json:"-"`

	BaseModel
}
