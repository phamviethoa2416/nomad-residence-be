package dto

import "time"

type CreateBlockedDateRequest struct {
	RoomID    uint    `json:"room_id"    binding:"required"`
	Date      string  `json:"date"       binding:"required,datetime=2006-01-02"`
	Source    string  `json:"source"     binding:"required,max=30"`
	SourceRef *string `json:"source_ref" binding:"omitempty,max=255"`
	Reason    *string `json:"reason"     binding:"omitempty,max=255"`
}

type CreateBlockedDateRangeRequest struct {
	RoomID   uint    `json:"room_id"    binding:"required"`
	DateFrom string  `json:"date_from"  binding:"required,datetime=2006-01-02"`
	DateTo   string  `json:"date_to"    binding:"required,datetime=2006-01-02"`
	Reason   *string `json:"reason"     binding:"omitempty,max=255"`
}

type BlockedDateResponse struct {
	ID        uint      `json:"id"`
	RoomID    uint      `json:"room_id"`
	Date      time.Time `json:"date"`
	Source    string    `json:"source"`
	SourceRef *string   `json:"source_ref,omitempty"`
	Reason    *string   `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
