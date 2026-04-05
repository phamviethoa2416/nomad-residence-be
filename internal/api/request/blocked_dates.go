package request

import "nomad-residence-be/pkg/validator"

type BlockedDateRoomParams struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type BlockedDateIDParam struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type ListBlockedDatesQuery struct {
	From string `form:"from" binding:"required,datetime=2006-01-02"`
	To   string `form:"to"   binding:"omitempty,datetime=2006-01-02"`
}

func (q *ListBlockedDatesQuery) Validate() (string, string) {
	return validator.ValidateDateRangeFromTo(q.From, q.To)
}

type BlockDatesRequest struct {
	Dates  []string `json:"dates"  binding:"required,min=1,dive,datetime=2006-01-02"`
	Reason *string  `json:"reason" binding:"omitempty,max=255"`
}

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
