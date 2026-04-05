package validator

var Messages = map[string]string{
	"required":    "Trường này là bắt buộc",
	"required_if": "Trường này là bắt buộc trong trường hợp này",
	"email":       "Email không hợp lệ",
	"min":         "Giá trị quá nhỏ",
	"max":         "Giá trị quá lớn",
	"oneof":       "Giá trị không nằm trong danh sách cho phép",
	"datetime":    "Định dạng ngày không hợp lệ (cần: YYYY-MM-DD)",
	"url":         "URL không hợp lệ",
	"numeric":     "Phải là số",
	"alphanum":    "Chỉ được chứa chữ cái và số",
	"gte":         "Giá trị phải lớn hơn hoặc bằng giá trị tối thiểu",
	"lte":         "Giá trị phải nhỏ hơn hoặc bằng giá trị tối đa",
	"gt":          "Giá trị phải lớn hơn 0",
	"positive":    "Phải là số dương",
	"dive":        "Một hoặc nhiều phần tử trong danh sách không hợp lệ",

	"booking_code":      "Mã đặt phòng không hợp lệ",
	"time_hhmm":         "Định dạng giờ không hợp lệ (cần: HH:MM)",
	"slug":              "Slug chỉ được chứa chữ thường, số và dấu gạch ngang",
	"password_strength": "Mật khẩu phải chứa ít nhất 1 chữ hoa, 1 chữ thường và 1 chữ số",
	"date_gte_today":    "Ngày không thể là ngày trong quá khứ",
}

func Message(tag string) string {
	if msg, ok := Messages[tag]; ok {
		return msg
	}
	return "Giá trị không hợp lệ"
}
