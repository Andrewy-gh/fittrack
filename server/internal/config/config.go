package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Config holds all configuration values for the application
type Config struct {
	// Required fields
	DatabaseURL string `validate:"required,url"`
	ProjectID   string `validate:"required"`

	// Optional fields with defaults
	Port            int    `validate:"omitempty,min=1,max=65535"`
	LogLevel        string `validate:"omitempty,oneof=debug info warn error"`
	Environment     string `validate:"omitempty,oneof=development staging production"`
	RateLimitRPM    int    `validate:"omitempty,min=1"`
	AllowedOrigins  string `validate:"omitempty"`

	// Database connection pool settings (optional)
	DBMaxConns         int32  `validate:"omitempty,min=1"`
	DBMinConns         int32  `validate:"omitempty,min=0"`
	DBMaxConnIdle      string `validate:"omitempty"`
	DBMaxConnLife      string `validate:"omitempty"`
	DBHealthCheck      string `validate:"omitempty"`

	// Metrics basic auth (optional)
	MetricsUsername string `validate:"omitempty"`
	MetricsPassword string `validate:"omitempty"`
}

// Load reads configuration from environment variables and validates it
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		ProjectID:      os.Getenv("PROJECT_ID"),
		Port:           getEnvInt("PORT", 8080),
		LogLevel:       getEnvString("LOG_LEVEL", "info"),
		Environment:    getEnvString("ENVIRONMENT", "development"),
		RateLimitRPM:   getEnvInt("RATE_LIMIT_RPM", 100),
		AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
		DBMaxConns:     int32(getEnvInt("DB_MAX_CONNS", 15)),
		DBMinConns:     int32(getEnvInt("DB_MIN_CONNS", 2)),
		DBMaxConnIdle:  getEnvString("DB_MAX_CONN_IDLE", "30s"),
		DBMaxConnLife:  getEnvString("DB_MAX_CONN_LIFE", "30m"),
		DBHealthCheck:  getEnvString("DB_HEALTHCHECK", "30s"),
		MetricsUsername: os.Getenv("METRICS_USERNAME"),
		MetricsPassword: os.Getenv("METRICS_PASSWORD"),
	}

	// Validate configuration
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// GetAllowedOrigins parses the comma-separated ALLOWED_ORIGINS into a string slice
func (c *Config) GetAllowedOrigins() []string {
	if c.AllowedOrigins == "" {
		return []string{}
	}

	origins := strings.Split(c.AllowedOrigins, ",")
	result := make([]string, 0, len(origins))

	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// getEnvString returns the environment variable value or a default
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns the environment variable value as an int or a default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		} else {
			log.Printf("config: failed to parse %s=%q as integer: %v, using default: %d", key, value, err, defaultValue)
		}
	}
	return defaultValue
}
