package request

type IcalExportParams struct {
	RoomID uint `uri:"room_id" binding:"required,gt=0"`
}

type IcalExportQuery struct {
	Token string `form:"token" binding:"required,min=1"`
}

type IcalRoomParams struct {
	ID uint `uri:"id" binding:"required,gt=0"`
}

type IcalLinkParams struct {
	LinkID uint `uri:"linkId" binding:"required,gt=0"`
}

type AddIcalLinkRequest struct {
	Platform  string  `json:"platform"   binding:"required,min=1,max=50"`
	ImportURL *string `json:"import_url" binding:"omitempty,url"`
}

type CreateIcalLinkRequest struct {
	RoomID    uint    `json:"room_id"    binding:"required"`
	Platform  string  `json:"platform"   binding:"required,max=50"`
	ImportURL *string `json:"import_url"`
	ExportURL *string `json:"export_url"`
}

type UpdateIcalLinkRequest struct {
	ImportURL *string `json:"import_url"`
	ExportURL *string `json:"export_url"`
	IsActive  *bool   `json:"is_active"`
}
