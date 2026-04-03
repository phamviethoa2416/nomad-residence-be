package config

import "time"

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

type RateLimitConfig struct {
	RequestsPerWindow int
	Window            time.Duration
	Message           string
}
