// Package config provides application configuration loading and defaults.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/wb-go/wbf/config"
)

// Config holds the application-wide configuration state.
type Config struct {
	Env      string   `mapstructure:"env"`
	HTTP     HTTP     `mapstructure:"http"`
	Postgres Postgres `mapstructure:"postgres"`
	Logging  Logging  `mapstructure:"logging"`
	Redis    Redis    `mapstructure:"redis"`
	RabbitMQ RabbitMQ `mapstructure:"rabbitmq"`
	Mailer   SMTP     `mapstructure:"mailer"`
	Telegram Telegram `mapstructure:"telegram"`
}

// HTTP contains settings for the HTTP server.
type HTTP struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type Telegram struct {
	TelegramToken string `mapstructure:"-"`
	ChatID        int64  `mapstructure:"-"`
}

// Postgres encapsulates database connection settings.
type Postgres struct {
	ConnectionURL string `mapstructure:"connection_url"`
}

// Logging defines the verbosity and format of application logs.
type Logging struct {
	Level string `mapstructure:"level"`
}

// Redis contains Redis connection settings.
type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// RabbitMQ contains RabbitMQ connection and queue settings.
type RabbitMQ struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	QueueName    string `mapstructure:"queue_name"`
	DLXName      string `mapstructure:"dlx_name"`
	DLQName      string `mapstructure:"dlq_name"`
	ExchangeName string `mapstructure:"exchange_name"`
}

// SMTP contains SMTP settings.
type SMTP struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Email       string        `mapstructure:"-"`
	Password    string        `mapstructure:"-"`
	UseTLS      bool          `mapstructure:"use_tls"`
	SendTimeout time.Duration `mapstructure:"send_timeout"`
}

// Load initializes the config registry, applies defaults, and loads
// configuration files and environment variables.
func Load(configPath string) (*Config, error) {
	c := config.New()

	setDefaults(c)

	_ = c.LoadEnvFiles(".env")

	if err := c.LoadConfigFiles(configPath); err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
	}

	c.EnableEnv("APP")

	var cfg Config
	if err := c.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.Telegram.TelegramToken = os.Getenv("APP_TELEGRAM_TOKEN")

	chatID, err := mustInt64Env("APP_CHAT_ID")
	if err != nil {
		return nil, err
	}
	cfg.Telegram.ChatID = chatID

	cfg.Mailer.Email = os.Getenv("APP_MAILER_EMAIL")
	cfg.Mailer.Password = os.Getenv("APP_MAILER_PASSWORD")

	return &cfg, nil
}

// setDefaults populates the configuration with sensible default values.
func setDefaults(c *config.Config) {
	c.SetDefault("env", "development")
	c.SetDefault("http.port", ":8080")
	c.SetDefault("http.read_timeout", "10s")
	c.SetDefault("http.write_timeout", "10s")
	c.SetDefault("http.idle_timeout", "60s")
	c.SetDefault("http.shutdown_timeout", "8s")
	c.SetDefault("logging.level", "info")
}
func mustInt64Env(key string) (int64, error) {
	v := os.Getenv(key)
	if v == "" {
		return 0, fmt.Errorf("%s is required", key)
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return n, nil
}
