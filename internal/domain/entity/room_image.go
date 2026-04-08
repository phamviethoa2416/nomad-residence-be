package entity

type RoomImage struct {
	ID     uint `json:"id"`
	RoomID uint `json:"room_id"`

	URL       string  `json:"url"`
	AltText   *string `json:"alt_text,omitempty"`
	IsPrimary bool    `json:"is_primary"`
	SortOrder int     `json:"sort_order"`

	BaseModel

	Room Room `json:"-"`
}
