package entity

type Amenity struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Icon     *string
	Category string `gorm:"type:varchar(50);index"`

	BaseModel
}
