package middlewares

import (
	"errors"
	"net/http"
	"nomad-residence-be/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Errors  interface{} `json:"errors,omitempty"`
}

type ValidationFieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		handleError(c, err)
	}
}

func handleError(c *gin.Context, err error) {
	var v validator.ValidationErrors
	if errors.As(err, &v) {
		fields := make([]ValidationFieldError, 0, len(v))
		for _, f := range v {
			fields = append(fields, ValidationFieldError{
				Field:   utils.ToSnakeCase(f.Field()),
				Message: validationMessage(f),
			})
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Dữ liệu đầu vào không hợp lệ",
			Code:    "VALIDATION_ERROR",
			Errors:  fields,
		})
		return
	}

	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, ErrorResponse{
			Success: false,
			Message: appErr.Message,
			Code:    appErr.Code,
		})
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Success: false,
			Message: "Không tìm thấy dữ liệu",
			Code:    "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Message: "Lỗi server, vui lòng thử lại sau",
		Code:    "INTERNAL_SERVER_ERROR",
	})
}

func abortWithError(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()

	handleError(c, err)
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Trường này là bắt buộc"
	case "email":
		return "Email không hợp lệ"
	case "min":
		return "Giá trị quá nhỏ (tối thiểu " + fe.Param() + ")"
	case "max":
		return "Giá trị quá lớn (tối đa " + fe.Param() + ")"
	case "oneof":
		return "Giá trị không hợp lệ, phải là một trong: " + fe.Param()
	case "datetime":
		return "Định dạng ngày không hợp lệ (cần: YYYY-MM-DD)"
	case "url":
		return "URL không hợp lệ"
	case "required_if":
		return "Trường này là bắt buộc trong trường hợp này"
	default:
		return "Giá trị không hợp lệ"
	}
}
