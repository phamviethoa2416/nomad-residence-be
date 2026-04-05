package response

import "time"

type BlockedDateResponse struct {
	ID        uint      `json:"id"`
	RoomID    uint      `json:"room_id"`
	Date      time.Time `json:"date"`
	Source    string    `json:"source"`
	SourceRef *string   `json:"source_ref,omitempty"`
	Reason    *string   `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
