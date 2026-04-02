package entity

type RoomImage struct {
	ID     uint `gorm:"primaryKey"`
	RoomID uint `gorm:"not null;index"`

	URL string `gorm:"type:varchar(500);not null"`

	AltText   *string `gorm:"type:varchar(255)"`
	IsPrimary bool    `gorm:"default:false"`
	SortOrder int     `gorm:"default:0"`

	BaseModel

	Room Room `gorm:"foreignKey:RoomID" json:"-"`
}
