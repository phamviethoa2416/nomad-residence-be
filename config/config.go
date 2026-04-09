package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	CORS         CORSConfig         `mapstructure:"cors"`
	RateLimit    RateLimitConfig    `mapstructure:"rate_limit"`
	Scheduler    SchedulerConfig    `mapstructure:"scheduler"`
	Notification NotificationConfig `mapstructure:"notification"`
	VietQR       VietQRConfig       `mapstructure:"vietqr"`
}

type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	Environment     string        `mapstructure:"environment"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

type JWTConfig struct {
	Secret   string        `mapstructure:"secret"`
	TokenTTL time.Duration `mapstructure:"token_ttl"`
}

type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

type RateLimitConfig struct {
	RequestsPerWindow int           `mapstructure:"requests_per_window"`
	Window            time.Duration `mapstructure:"window"`
	Message           string        `mapstructure:"message"`
}

type SchedulerConfig struct {
	BookingJobCron          string `mapstructure:"booking_job_cron"`
	IcalSyncIntervalMinutes int    `mapstructure:"ical_sync_interval_minutes"`
}

type NotificationConfig struct {
	Email    EmailConfig    `mapstructure:"email"`
	Telegram TelegramConfig `mapstructure:"telegram"`
}

type EmailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type TelegramConfig struct {
	BotToken string `mapstructure:"bot_token"`
	ChatID   string `mapstructure:"chat_id"`
}

type VietQRConfig struct {
	BankBin      string `mapstructure:"bank_bin"`
	BankName     string `mapstructure:"bank_name"`
	AccountNo    string `mapstructure:"account_no"`
	AccountName  string `mapstructure:"account_name"`
	Template     string `mapstructure:"template"`
	WebhookToken string `mapstructure:"webhook_token"`
}

func LoadConfig(path string) (*AppConfig, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config: %w", err)
	}

	if err := validate(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.environment", "dev")
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.shutdown_timeout", "5s")

	// Database
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "5m")
	v.SetDefault("database.conn_max_idle_time", "2m")

	// JWT
	v.SetDefault("jwt.token_ttl", "24h")

	// CORS
	v.SetDefault("cors.allow_origins", []string{"*"})
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE"})
	v.SetDefault("cors.allow_headers", []string{"*"})
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.max_age", 86400)

	// Rate limit
	v.SetDefault("rate_limit.requests_per_window", 100)
	v.SetDefault("rate_limit.window", "1m")
	v.SetDefault("rate_limit.message", "Too many requests")

	// Scheduler
	v.SetDefault("scheduler.booking_job_cron", "0 * * * *")
	v.SetDefault("scheduler.ical_sync_interval_minutes", 30)

	// VietQR
	v.SetDefault("vietqr.template", "compact2")
}

func validate(cfg *AppConfig) error {
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Server.Port == 0 {
		return fmt.Errorf("server port is required")
	}

	return nil
}
