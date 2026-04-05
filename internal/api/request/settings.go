package request

type UpdateSettingsRequest struct {
	Key   *string `json:"key"   binding:"omitempty,min=1"`
	Value *string `json:"value"`

	Settings []SettingItem `json:"settings" binding:"omitempty,dive"`
}

type SettingItem struct {
	Key   string `json:"key"   binding:"required,min=1"`
	Value string `json:"value"`
}

func (r *UpdateSettingsRequest) Validate() (string, string) {
	hasBatch := len(r.Settings) > 0
	hasSingle := r.Key != nil && r.Value != nil

	if !hasBatch && !hasSingle {
		return "settings", "Thiếu cấu hình thiết lập đơn lẻ hoặc mảng settings"
	}
	return "", ""
}

type UpsertSettingRequest struct {
	Value       interface{} `json:"value"       binding:"required"`
	Description *string     `json:"description" binding:"omitempty,max=255"`
}
