package middlewares

import (
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

func Timeout(d time.Duration) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(d),
		timeout.WithResponse(func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, ErrorResponse{
				Success: false,
				Message: "Yêu cầu mất quá nhiều thời gian, vui lòng thử lại",
				Code:    "REQUEST_TIMEOUT",
			})
		}),
	)
}
