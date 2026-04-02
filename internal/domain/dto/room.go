package dto

import (
	"nomad-residence-be/internal/domain/entity"
	"time"

	"github.com/shopspring/decimal"
)

type CreateRoomRequest struct {
	Name     string          `json:"name" binding:"required,max=255"`
	Slug     string          `json:"slug" binding:"required,max=255"`
	RoomType entity.RoomType `json:"room_type" binding:"required"`

	Description      *string `json:"description"`
	ShortDescription *string `json:"short_description" binding:"omitempty,max=500"`

	MaxGuests    int              `json:"max_guests" binding:"required,min=1"`
	NumBedrooms  int              `json:"num_bedrooms" binding:"omitempty,min=0"`
	NumBathrooms int              `json:"num_bathrooms" binding:"omitempty,min=0"`
	NumBeds      int              `json:"num_beds" binding:"omitempty,min=0"`
	Area         *decimal.Decimal `json:"area" binding:"omitempty,gt=0"`

	Address   *string          `json:"address"`
	District  *string          `json:"district" binding:"omitempty,max=100"`
	City      string           `json:"city" binding:"required,max=100"`
	Latitude  *decimal.Decimal `json:"latitude" binding:"omitempty,gte=-90,lte=90"`
	Longitude *decimal.Decimal `json:"longitude" binding:"omitempty,gte=-180,lte=180"`

	BasePrice   decimal.Decimal `json:"base_price" binding:"required,gt=0"`
	CleaningFee decimal.Decimal `json:"cleaning_fee" binding:"gte=0"`

	MinNights int `json:"min_nights" binding:"omitempty,min=1"`
	MaxNights int `json:"max_nights" binding:"omitempty,min=1"`

	HouseRules         *string `json:"house_rules"`
	CancellationPolicy *string `json:"cancellation_policy"`

	SortOrder int `json:"sort_order"`
}

type UpdateRoomRequest struct {
	Name     *string          `json:"name" binding:"omitempty,max=255"`
	RoomType *entity.RoomType `json:"room_type" binding:"omitempty"`

	Description      *string `json:"description"`
	ShortDescription *string `json:"short_description" binding:"omitempty,max=500"`

	MaxGuests    *int             `json:"max_guests" binding:"omitempty,min=1"`
	NumBedrooms  *int             `json:"num_bedrooms" binding:"omitempty,min=0"`
	NumBathrooms *int             `json:"num_bathrooms" binding:"omitempty,min=0"`
	NumBeds      *int             `json:"num_beds" binding:"omitempty,min=0"`
	Area         *decimal.Decimal `json:"area" binding:"omitempty,gt=0"`

	Address   *string          `json:"address"`
	District  *string          `json:"district" binding:"omitempty,max=100"`
	City      *string          `json:"city" binding:"omitempty,max=100"`
	Latitude  *decimal.Decimal `json:"latitude" binding:"omitempty,gte=-90,lte=90"`
	Longitude *decimal.Decimal `json:"longitude" binding:"omitempty,gte=-180,lte=180"`

	BasePrice   *decimal.Decimal `json:"base_price" binding:"omitempty,gt=0"`
	CleaningFee *decimal.Decimal `json:"cleaning_fee" binding:"omitempty,gte=0"`

	MinNights *int `json:"min_nights" binding:"omitempty,min=1"`
	MaxNights *int `json:"max_nights" binding:"omitempty,min=1"`

	HouseRules         *string `json:"house_rules"`
	CancellationPolicy *string `json:"cancellation_policy"`

	Status    *entity.RoomStatus `json:"status" binding:"omitempty,oneof=active inactive"`
	SortOrder *int               `json:"sort_order"`
}

type RoomFilterRequest struct {
	RoomType  entity.RoomType   `form:"room_type"`
	Status    entity.RoomStatus `form:"status" binding:"omitempty,oneof=active inactive"`
	City      string            `form:"city"`
	District  string            `form:"district"`
	MinPrice  *decimal.Decimal  `form:"min_price"`
	MaxPrice  *decimal.Decimal  `form:"max_price"`
	Amenities []uint            `form:"amenities"`
	MinGuests int               `form:"min_guests" binding:"omitempty,min=1"`
	MaxGuests int               `form:"max_guests" binding:"omitempty,min=1"`
	Page      int               `form:"page" binding:"omitempty,min=1"`
	Limit     int               `form:"limit" binding:"omitempty,min=1,max=100"`
}

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
	RoomType           entity.RoomType       `json:"room_type"`
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
	Status             entity.RoomStatus     `json:"status"`
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
	RoomType         entity.RoomType    `json:"room_type"`
	ShortDescription *string            `json:"short_description,omitempty"`
	MaxGuests        int                `json:"max_guests"`
	NumBedrooms      int                `json:"num_bedrooms"`
	BasePrice        decimal.Decimal    `json:"base_price"`
	CleaningFee      decimal.Decimal    `json:"cleaning_fee"`
	Status           entity.RoomStatus  `json:"status"`
	PrimaryImage     *RoomImageResponse `json:"primary_image,omitempty"`
}

type RoomAvailabilityRequest struct {
	CheckinDate  string `form:"checkin_date"  binding:"required,datetime=2006-01-02"`
	CheckoutDate string `form:"checkout_date" binding:"required,datetime=2006-01-02"`
	NumGuests    int    `form:"num_guests"    binding:"omitempty,min=1"`
}

type RoomAvailabilityResponse struct {
	Available   bool            `json:"available"`
	NumNights   int             `json:"num_nights"`
	BaseTotal   decimal.Decimal `json:"base_total"`
	CleaningFee decimal.Decimal `json:"cleaning_fee"`
	TotalAmount decimal.Decimal `json:"total_amount"`
}
