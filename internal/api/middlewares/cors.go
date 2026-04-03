package middlewares

import (
	"fmt"
	"net/http"
	"nomad-residence-be/config"
	"strings"

	"github.com/gin-gonic/gin"
)

func DefaultCORSConfig() config.CORSConfig {
	return config.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	allowMethodsStr := strings.Join(cfg.AllowMethods, ", ")
	allowHeadersStr := strings.Join(cfg.AllowHeaders, ", ")
	exposeHeadersStr := strings.Join(cfg.ExposeHeaders, ", ")
	maxAgeStr := ""
	if cfg.MaxAge > 0 {
		maxAgeStr = fmt.Sprintf("%d", cfg.MaxAge)
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigin := resolveAllowedOrigin(cfg.AllowOrigins, origin, cfg.AllowCredentials)

		if allowedOrigin == "" && origin != "" {
			c.Next()
			return
		}

		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}
		c.Header("Access-Control-Allow-Methods", allowMethodsStr)
		c.Header("Access-Control-Allow-Headers", allowHeadersStr)
		if exposeHeadersStr != "" {
			c.Header("Access-Control-Expose-Headers", exposeHeadersStr)
		}
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			if maxAgeStr != "" {
				c.Header("Access-Control-Max-Age", maxAgeStr)
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func resolveAllowedOrigin(allowed []string, origin string, withCredentials bool) string {
	for _, o := range allowed {
		if o == "*" {
			if withCredentials {
				return origin
			}
			return "*"
		}
		if o == origin {
			return origin
		}
	}
	return ""
}
