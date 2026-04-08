package entity

type Amenity struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Icon     *string `json:"icon,omitempty"`
	Category string  `json:"category"`

	BaseModel
}
