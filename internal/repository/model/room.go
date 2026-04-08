package model

import (
	"nomad-residence-be/internal/domain/entity"

	"github.com/shopspring/decimal"
)

type Room struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null"`
	Slug string `gorm:"type:varchar(255);not null;uniqueIndex"`

	RoomType entity.RoomType `gorm:"type:varchar(50);not null;index"`

	Description      *string `gorm:"type:text"`
	ShortDescription *string `gorm:"type:varchar(500)"`

	MaxGuests    int `gorm:"not null;default:2;check:max_guests > 0"`
	NumBedrooms  int `gorm:"not null;default:1;check:num_bedrooms >= 0"`
	NumBathrooms int `gorm:"not null;default:1;check:num_bathrooms >= 0"`
	NumBeds      int `gorm:"not null;default:1;check:num_beds > 0"`

	Area *decimal.Decimal `gorm:"type:decimal(6,1);check:area >= 0"`

	Address  *string `gorm:"type:text"`
	District *string `gorm:"type:varchar(100);index"`
	City     string  `gorm:"type:varchar(100);default:'Hà Nội';index"`

	Latitude  *decimal.Decimal `gorm:"type:decimal(10,7)"`
	Longitude *decimal.Decimal `gorm:"type:decimal(10,7)"`

	BasePrice   decimal.Decimal `gorm:"type:decimal(12,0);not null;check:base_price >= 0;index"`
	CleaningFee decimal.Decimal `gorm:"type:decimal(12,0);default:0;check:cleaning_fee >= 0"`

	MinNights int `gorm:"default:1;check:min_nights > 0"`
	MaxNights int `gorm:"default:30;check:max_nights >= min_nights"`

	CheckinTime  string `gorm:"type:varchar(10);default:'14:00'"`
	CheckoutTime string `gorm:"type:varchar(10);default:'12:00'"`

	HouseRules         *string `gorm:"type:text"`
	CancellationPolicy *string `gorm:"type:text"`

	Status    entity.RoomStatus `gorm:"type:varchar(20);default:'active';index"`
	SortOrder int               `gorm:"default:0;index"`

	Images    []RoomImage   `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE"`
	Amenities []RoomAmenity `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE"`

	Bookings     []Booking     `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE"`
	BlockedDates []BlockedDate `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE"`
	PricingRules []PricingRule `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE"`
	IcalLinks    []IcalLink    `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE"`

	BaseModel
}

func (Room) TableName() string { return "rooms" }

func RoomFromDomain(e *entity.Room) *Room {
	if e == nil {
		return nil
	}
	return &Room{
		ID: e.ID, Name: e.Name, Slug: e.Slug, RoomType: e.RoomType,
		Description: e.Description, ShortDescription: e.ShortDescription,
		MaxGuests: e.MaxGuests, NumBedrooms: e.NumBedrooms,
		NumBathrooms: e.NumBathrooms, NumBeds: e.NumBeds,
		Area: e.Area, Address: e.Address, District: e.District, City: e.City,
		Latitude: e.Latitude, Longitude: e.Longitude,
		BasePrice: e.BasePrice, CleaningFee: e.CleaningFee,
		MinNights: e.MinNights, MaxNights: e.MaxNights,
		CheckinTime: e.CheckinTime, CheckoutTime: e.CheckoutTime,
		HouseRules: e.HouseRules, CancellationPolicy: e.CancellationPolicy,
		Status: e.Status, SortOrder: e.SortOrder,
		BaseModel: baseFromDomain(e.BaseModel),
	}
}

func (m *Room) ToDomain() *entity.Room {
	if m == nil {
		return nil
	}
	r := &entity.Room{
		ID: m.ID, Name: m.Name, Slug: m.Slug, RoomType: m.RoomType,
		Description: m.Description, ShortDescription: m.ShortDescription,
		MaxGuests: m.MaxGuests, NumBedrooms: m.NumBedrooms,
		NumBathrooms: m.NumBathrooms, NumBeds: m.NumBeds,
		Area: m.Area, Address: m.Address, District: m.District, City: m.City,
		Latitude: m.Latitude, Longitude: m.Longitude,
		BasePrice: m.BasePrice, CleaningFee: m.CleaningFee,
		MinNights: m.MinNights, MaxNights: m.MaxNights,
		CheckinTime: m.CheckinTime, CheckoutTime: m.CheckoutTime,
		HouseRules: m.HouseRules, CancellationPolicy: m.CancellationPolicy,
		Status: m.Status, SortOrder: m.SortOrder,
		BaseModel: m.BaseModel.toDomainBase(),
	}
	for i := range m.Images {
		r.Images = append(r.Images, *m.Images[i].ToDomain())
	}
	for i := range m.Amenities {
		r.Amenities = append(r.Amenities, *m.Amenities[i].ToDomain())
	}
	return r
}

func RoomsToDomain(models []Room) []entity.Room {
	result := make([]entity.Room, len(models))
	for i := range models {
		result[i] = *models[i].ToDomain()
	}
	return result
}
