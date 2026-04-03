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

type SchedulerConfig struct {
	BookingJobCron          string
	IcalSyncIntervalMinutes int
}

type NotificationConfig struct {
	Email    EmailConfig
	Telegram TelegramConfig
}

type EmailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type TelegramConfig struct {
	BotToken string
	ChatID   string
}
