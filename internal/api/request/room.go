package request

import (
	"nomad-residence-be/pkg/validator"

	"github.com/shopspring/decimal"
)

type ListRoomsQuery struct {
	Checkin  string `form:"checkin"    binding:"omitempty,datetime=2006-01-02"`
	Checkout string `form:"checkout"   binding:"omitempty,datetime=2006-01-02"`
	Guests   int    `form:"guests"     binding:"omitempty,min=1"`
	RoomType string `form:"room_type"`
	MinPrice int    `form:"min_price"  binding:"omitempty,min=0"`
	MaxPrice int    `form:"max_price"  binding:"omitempty,min=0"`
	Sort     string `form:"sort"       binding:"omitempty,oneof=sort_order price_asc price_desc newest"`
	Page     int    `form:"page"       binding:"omitempty,min=1"`
	Limit    int    `form:"limit"      binding:"omitempty,min=1,max=50"`
}

func (q *ListRoomsQuery) ApplyDefaults() {
	if q.Sort == "" {
		q.Sort = "sort_order"
	}
	// Clamp page >= 1
	if q.Page <= 0 {
		q.Page = 1
	}
	// Clamp limit: 1 ≤ limit ≤ 50, default 12
	if q.Limit <= 0 {
		q.Limit = 12
	} else if q.Limit > 50 {
		q.Limit = 50
	}
}

func (q *ListRoomsQuery) Validate() (string, string) {
	return validator.ValidateDateRange(q.Checkin, q.Checkout)
}

type GetRoomDetailQuery struct {
	Checkin  string `form:"checkin"  binding:"omitempty,datetime=2006-01-02"`
	Checkout string `form:"checkout" binding:"omitempty,datetime=2006-01-02"`
}

func (q *GetRoomDetailQuery) Validate() (string, string) {
	return validator.ValidateDateRange(q.Checkin, q.Checkout)
}

type GetRoomCalendarQuery struct {
	From string `form:"from" binding:"omitempty,datetime=2006-01-02"`
	To   string `form:"to"   binding:"omitempty,datetime=2006-01-02"`
}

func (q *GetRoomCalendarQuery) Validate() (string, string) {
	return validator.ValidateDateRangeFromTo(q.From, q.To)
}

type RoomIDParam struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type ImageIDParam struct {
	ImageID uint `uri:"imageId" binding:"required,gt=0"`
}

type CreateRoomRequest struct {
	Name               string           `json:"name"                binding:"required,min=1,max=255"`
	Slug               string           `json:"slug"                binding:"required,slug"`
	RoomType           string           `json:"room_type"           binding:"required,min=1,max=50"`
	Description        *string          `json:"description"         binding:"omitempty"`
	ShortDescription   *string          `json:"short_description"   binding:"omitempty,max=500"`
	MaxGuests          int              `json:"max_guests"          binding:"required,min=1"`
	NumBedrooms        int              `json:"num_bedrooms"        binding:"omitempty,min=0"`
	NumBathrooms       int              `json:"num_bathrooms"       binding:"omitempty,min=0"`
	NumBeds            int              `json:"num_beds"            binding:"omitempty,min=0"`
	Area               *decimal.Decimal `json:"area"                binding:"omitempty,gt=0"`
	Address            string           `json:"address"             binding:"required,min=1"`
	District           *string          `json:"district"            binding:"omitempty,max=100"`
	City               string           `json:"city"                binding:"omitempty,max=100"`
	Latitude           *decimal.Decimal `json:"latitude"            binding:"omitempty,gte=-90,lte=90"`
	Longitude          *decimal.Decimal `json:"longitude"           binding:"omitempty,gte=-180,lte=180"`
	BasePrice          decimal.Decimal  `json:"base_price"          binding:"required,gt=0"`
	CleaningFee        decimal.Decimal  `json:"cleaning_fee"        binding:"gte=0"`
	CheckinTime        string           `json:"checkin_time"        binding:"omitempty,time_hhmm"`
	CheckoutTime       string           `json:"checkout_time"       binding:"omitempty,time_hhmm"`
	MinNights          int              `json:"min_nights"          binding:"omitempty,min=1"`
	MaxNights          int              `json:"max_nights"          binding:"omitempty,min=1"`
	HouseRules         *string          `json:"house_rules"         binding:"omitempty"`
	CancellationPolicy *string          `json:"cancellation_policy" binding:"omitempty"`
	Status             string           `json:"status"              binding:"omitempty,oneof=active inactive"`
	SortOrder          int              `json:"sort_order"          binding:"omitempty"`
}

func (r *CreateRoomRequest) ApplyDefaults() {
	if r.City == "" {
		r.City = "Hà Nội"
	}
	if r.CheckinTime == "" {
		r.CheckinTime = "14:00"
	}
	if r.CheckoutTime == "" {
		r.CheckoutTime = "12:00"
	}
	if r.MinNights <= 0 {
		r.MinNights = 1
	}
	if r.MaxNights <= 0 {
		r.MaxNights = 30
	}
	if r.Status == "" {
		r.Status = "active"
	}
	if r.NumBedrooms <= 0 {
		r.NumBedrooms = 1
	}
	if r.NumBathrooms <= 0 {
		r.NumBathrooms = 1
	}
	if r.NumBeds <= 0 {
		r.NumBeds = 1
	}
}

type UpdateRoomRequest struct {
	Name               *string          `json:"name"                binding:"omitempty,min=1,max=255"`
	Slug               *string          `json:"slug"                binding:"omitempty,slug"`
	RoomType           *string          `json:"room_type"           binding:"omitempty,min=1,max=50"`
	Description        *string          `json:"description"`
	ShortDescription   *string          `json:"short_description"   binding:"omitempty,max=500"`
	MaxGuests          *int             `json:"max_guests"          binding:"omitempty,min=1"`
	NumBedrooms        *int             `json:"num_bedrooms"        binding:"omitempty,min=0"`
	NumBathrooms       *int             `json:"num_bathrooms"       binding:"omitempty,min=0"`
	NumBeds            *int             `json:"num_beds"            binding:"omitempty,min=0"`
	Area               *decimal.Decimal `json:"area"                binding:"omitempty,gt=0"`
	Address            *string          `json:"address"             binding:"omitempty,min=1"`
	District           *string          `json:"district"            binding:"omitempty,max=100"`
	City               *string          `json:"city"                binding:"omitempty,max=100"`
	Latitude           *decimal.Decimal `json:"latitude"            binding:"omitempty,gte=-90,lte=90"`
	Longitude          *decimal.Decimal `json:"longitude"           binding:"omitempty,gte=-180,lte=180"`
	BasePrice          *decimal.Decimal `json:"base_price"          binding:"omitempty,gt=0"`
	CleaningFee        *decimal.Decimal `json:"cleaning_fee"        binding:"omitempty,gte=0"`
	CheckinTime        *string          `json:"checkin_time"        binding:"omitempty,time_hhmm"`
	CheckoutTime       *string          `json:"checkout_time"       binding:"omitempty,time_hhmm"`
	MinNights          *int             `json:"min_nights"          binding:"omitempty,min=1"`
	MaxNights          *int             `json:"max_nights"          binding:"omitempty,min=1"`
	HouseRules         *string          `json:"house_rules"`
	CancellationPolicy *string          `json:"cancellation_policy"`
	Status             *string          `json:"status"              binding:"omitempty,oneof=active inactive"`
	SortOrder          *int             `json:"sort_order"`
}

type AddRoomImageRequest struct {
	URL       string  `json:"url"        binding:"required,url,max=500"`
	AltText   *string `json:"alt_text"   binding:"omitempty,max=255"`
	IsPrimary bool    `json:"is_primary"`
	SortOrder int     `json:"sort_order" binding:"omitempty,min=0"`
}

type ReorderRoomImagesRequest struct {
	Order []ImageOrderItem `json:"order" binding:"required,min=1,dive"`
}

type ImageOrderItem struct {
	ID        uint `json:"id"         binding:"required,gt=0"`
	SortOrder int  `json:"sort_order" binding:"min=0"`
}

type UpdateAmenitiesRequest struct {
	Amenities []AmenityItem `json:"amenities" binding:"omitempty,dive"`
}

type AmenityItem struct {
	Name     string  `json:"name"     binding:"required,min=1,max=100"`
	Icon     *string `json:"icon"     binding:"omitempty,max=50"`
	Category string  `json:"category" binding:"omitempty,max=50"`
}

type RoomFilterRequest struct {
	RoomType  string           `form:"room_type"`
	Status    string           `form:"status"    binding:"omitempty,oneof=active inactive"`
	City      string           `form:"city"`
	District  string           `form:"district"`
	MinPrice  *decimal.Decimal `form:"min_price"`
	MaxPrice  *decimal.Decimal `form:"max_price"`
	Amenities []uint           `form:"amenities"`
	MinGuests int              `form:"min_guests" binding:"omitempty,min=1"`
	MaxGuests int              `form:"max_guests" binding:"omitempty,min=1"`
	Page      int              `form:"page"       binding:"omitempty,min=1"`
	Limit     int              `form:"limit"      binding:"omitempty,min=1,max=100"`
}

type RoomAvailabilityRequest struct {
	CheckinDate  string `form:"checkin_date"  binding:"required,datetime=2006-01-02"`
	CheckoutDate string `form:"checkout_date" binding:"required,datetime=2006-01-02"`
	NumGuests    int    `form:"num_guests"    binding:"omitempty,min=1"`
}
