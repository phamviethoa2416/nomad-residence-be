package middlewares

import (
	"errors"
	"net/http"
	apperrors "nomad-residence-be/pkg/errors"
	customValidator "nomad-residence-be/pkg/validator"

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
				Field:   f.Field(),
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

	var appErr *apperrors.AppError
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

func AbortWithValidationError(c *gin.Context, err error) {
	AbortWithError(c, err)
}

func AbortWithError(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
	handleError(c, err)
}

func validationMessage(fe validator.FieldError) string {
	msg := customValidator.Message(fe.Tag())

	switch fe.Tag() {
	case "min":
		return msg + " (tối thiểu " + fe.Param() + ")"
	case "max":
		return msg + " (tối đa " + fe.Param() + ")"
	case "oneof":
		return msg + ": " + fe.Param()
	}

	return msg
}
