package middlewares

import (
	"net/http"
	"nomad-residence-be/config"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func DefaultRateLimitConfig() config.RateLimitConfig {
	return config.RateLimitConfig{
		RequestsPerWindow: 100,
		Window:            time.Minute,
		Message:           "Quá nhiều yêu cầu, vui lòng thử lại sau",
	}
}

func StrictRateLimitConfig() config.RateLimitConfig {
	return config.RateLimitConfig{
		RequestsPerWindow: 10,
		Window:            time.Minute,
		Message:           "Quá nhiều yêu cầu, vui lòng thử lại sau",
	}
}

type ipEntry struct {
	count   int
	resetAt time.Time
	mu      sync.Mutex
}

func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	store := &sync.Map{}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			store.Range(func(key, value any) bool {
				entry := value.(*ipEntry)
				entry.mu.Lock()
				expired := now.After(entry.resetAt)
				entry.mu.Unlock()
				if expired {
					store.Delete(key)
				}
				return true
			})
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		raw, _ := store.LoadOrStore(ip, &ipEntry{
			resetAt: time.Now().Add(cfg.Window),
		})
		entry := raw.(*ipEntry)

		entry.mu.Lock()
		now := time.Now()

		if now.After(entry.resetAt) {
			entry.count = 0
			entry.resetAt = now.Add(cfg.Window)
			store.Store(ip, entry)
		}

		entry.count++
		count := entry.count
		resetAt := entry.resetAt
		entry.mu.Unlock()

		remaining := cfg.RequestsPerWindow - count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerWindow))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.Itoa(int(resetAt.Unix())))

		if count > cfg.RequestsPerWindow {
			c.Header("Retry-After", strconv.Itoa(int(time.Until(resetAt).Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, ErrorResponse{
				Success: false,
				Message: cfg.Message,
				Code:    "RATE_LIMIT_EXCEEDED",
			})
			return
		}

		c.Next()
	}
}
