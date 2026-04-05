package response

import (
	"time"

	"github.com/shopspring/decimal"
)

type RoomImageResponse struct {
	ID        uint    `json:"id"`
	URL       string  `json:"url"`
	AltText   *string `json:"alt_text,omitempty"`
	IsPrimary bool    `json:"is_primary"`
	SortOrder int     `json:"sort_order"`
}

type RoomAmenityResponse struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Icon     *string `json:"icon,omitempty"`
	Category string  `json:"category"`
}

type RoomResponse struct {
	ID                 uint                  `json:"id"`
	Name               string                `json:"name"`
	Slug               string                `json:"slug"`
	RoomType           string                `json:"room_type"`
	Description        *string               `json:"description,omitempty"`
	ShortDescription   *string               `json:"short_description,omitempty"`
	MaxGuests          int                   `json:"max_guests"`
	NumBedrooms        int                   `json:"num_bedrooms"`
	NumBathrooms       int                   `json:"num_bathrooms"`
	NumBeds            int                   `json:"num_beds"`
	Area               *decimal.Decimal      `json:"area,omitempty"`
	Address            *string               `json:"address,omitempty"`
	District           *string               `json:"district,omitempty"`
	City               string                `json:"city"`
	Latitude           *decimal.Decimal      `json:"latitude,omitempty"`
	Longitude          *decimal.Decimal      `json:"longitude,omitempty"`
	BasePrice          decimal.Decimal       `json:"base_price"`
	CleaningFee        decimal.Decimal       `json:"cleaning_fee"`
	CheckinTime        string                `json:"checkin_time"`
	CheckoutTime       string                `json:"checkout_time"`
	MinNights          int                   `json:"min_nights"`
	MaxNights          int                   `json:"max_nights"`
	HouseRules         *string               `json:"house_rules,omitempty"`
	CancellationPolicy *string               `json:"cancellation_policy,omitempty"`
	Status             string                `json:"status"`
	SortOrder          int                   `json:"sort_order"`
	Images             []RoomImageResponse   `json:"images,omitempty"`
	Amenities          []RoomAmenityResponse `json:"amenities,omitempty"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at"`
}

type RoomSummaryResponse struct {
	ID               uint               `json:"id"`
	Name             string             `json:"name"`
	Slug             string             `json:"slug"`
	RoomType         string             `json:"room_type"`
	ShortDescription *string            `json:"short_description,omitempty"`
	MaxGuests        int                `json:"max_guests"`
	NumBedrooms      int                `json:"num_bedrooms"`
	BasePrice        decimal.Decimal    `json:"base_price"`
	CleaningFee      decimal.Decimal    `json:"cleaning_fee"`
	Status           string             `json:"status"`
	PrimaryImage     *RoomImageResponse `json:"primary_image,omitempty"`
}

type RoomAvailabilityResponse struct {
	Available   bool            `json:"available"`
	NumNights   int             `json:"num_nights"`
	BaseTotal   decimal.Decimal `json:"base_total"`
	CleaningFee decimal.Decimal `json:"cleaning_fee"`
	TotalAmount decimal.Decimal `json:"total_amount"`
}
