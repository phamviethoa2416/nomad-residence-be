package middlewares

import (
	apperrors "nomad-residence-be/pkg/utils"

	"github.com/gin-gonic/gin"
)

func Authorize(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		admin := GetAdminContext(c)
		if admin == nil {
			abortWithError(c, apperrors.New(
				401,
				"UNAUTHORIZED",
				"Vui lòng đăng nhập"),
			)
			return
		}

		if len(allowedRoles) == 0 {
			c.Next()
			return
		}

		for _, role := range allowedRoles {
			if admin.Role == role {
				c.Next()
				return
			}
		}

		abortWithError(c, apperrors.New(
			403,
			"FORBIDDEN",
			"Bạn không có quyền truy cập"),
		)
	}
}
