package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()

				logger.Error("panic_recovered",
					slog.Any("panic", r),
					slog.String("stack", string(stack)),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("request_id", GetRequestID(c)),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Lỗi server, vui lòng thử lại sau",
					Code:    "INTERNAL_SERVER_ERROR",
				})
			}
		}()
		c.Next()
	}
}

func DefaultRecovery() gin.HandlerFunc {
	return Recovery(slog.Default())
}
