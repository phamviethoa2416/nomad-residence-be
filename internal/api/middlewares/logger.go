package middlewares

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		if size < 0 {
			size = 0
		}

		requestID := GetRequestID(c)
		clientIP := c.ClientIP()
		method := c.Request.Method

		fullPath := path
		if raw != "" {
			fullPath = path + "?" + raw
		}

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		attrs := []any{
			slog.String("method", method),
			slog.String("path", fullPath),
			slog.Int("status", status),
			slog.Duration("latency", latency.Round(time.Millisecond)),
			slog.String("ip", clientIP),
			slog.Int("size", size),
		}
		if requestID != "" {
			attrs = append(attrs, slog.String("request_id", requestID))
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.Last().Error()))
		}

		logger.Log(c.Request.Context(), level, "http_request", attrs...)
	}
}
