package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-Id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		existing := c.GetHeader(HeaderRequestID)
		id := existing
		if id == "" {
			id = uuid.NewString()
		}

		c.Set("request_id", id)
		c.Header(HeaderRequestID, id)

		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	if id, ok := c.Get("request_id"); ok {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}
