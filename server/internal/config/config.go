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
	Port           int    `validate:"omitempty,min=1,max=65535"`
	LogLevel       string `validate:"omitempty,oneof=debug info warn error"`
	Environment    string `validate:"omitempty,oneof=development staging production"`
	RateLimitRPM   int    `validate:"omitempty,min=1"`
	AllowedOrigins string `validate:"omitempty"`

	// Database connection pool settings (optional)
	DBMaxConns    int32  `validate:"omitempty,min=1"`
	DBMinConns    int32  `validate:"omitempty,min=0"`
	DBMaxConnIdle string `validate:"omitempty"`
	DBMaxConnLife string `validate:"omitempty"`
	DBHealthCheck string `validate:"omitempty"`

	// Metrics basic auth (optional)
	MetricsUsername string `validate:"omitempty"`
	MetricsPassword string `validate:"omitempty"`

	// Optional AI chat recovery wiring
	InngestEventKey   string `validate:"omitempty"`
	InngestSigningKey string `validate:"omitempty"`

	// Optional local-development E2E auth bootstrap
	LocalE2EAuthEnabled     bool   `validate:"omitempty"`
	LocalE2EAuthUserID      string `validate:"omitempty"`
	LocalE2EAuthEmail       string `validate:"omitempty,email"`
	LocalE2EAuthDisplayName string `validate:"omitempty"`
}

// Load reads configuration from environment variables and validates it
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		ProjectID:           os.Getenv("PROJECT_ID"),
		Port:                getEnvInt("PORT", 8080),
		LogLevel:            getEnvString("LOG_LEVEL", "info"),
		Environment:         getEnvString("ENVIRONMENT", "development"),
		RateLimitRPM:        getEnvInt("RATE_LIMIT_RPM", 100),
		AllowedOrigins:      os.Getenv("ALLOWED_ORIGINS"),
		DBMaxConns:          int32(getEnvInt("DB_MAX_CONNS", 15)),
		DBMinConns:          int32(getEnvInt("DB_MIN_CONNS", 2)),
		DBMaxConnIdle:       getEnvString("DB_MAX_CONN_IDLE", "30s"),
		DBMaxConnLife:       getEnvString("DB_MAX_CONN_LIFE", "30m"),
		DBHealthCheck:       getEnvString("DB_HEALTHCHECK", "30s"),
		MetricsUsername:     os.Getenv("METRICS_USERNAME"),
		MetricsPassword:     os.Getenv("METRICS_PASSWORD"),
		InngestEventKey:     os.Getenv("INNGEST_EVENT_KEY"),
		InngestSigningKey:   os.Getenv("INNGEST_SIGNING_KEY"),
		LocalE2EAuthEnabled: getEnvBool("E2E_LOCAL_AUTH_ENABLED", false),
		LocalE2EAuthUserID:  getEnvString("E2E_LOCAL_AUTH_USER_ID", "local-e2e-user"),
		LocalE2EAuthEmail: getEnvString(
			"E2E_LOCAL_AUTH_EMAIL",
			"local-e2e-user@example.test",
		),
		LocalE2EAuthDisplayName: getEnvString(
			"E2E_LOCAL_AUTH_DISPLAY_NAME",
			"Local E2E User",
		),
	}

	// Validate configuration
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) AIChatRecoveryConfigured() bool {
	return strings.TrimSpace(c.InngestEventKey) != "" &&
		strings.TrimSpace(c.InngestSigningKey) != ""
}

func (c *Config) LocalE2EAuthConfigured() bool {
	return c.Environment == "development" &&
		c.LocalE2EAuthEnabled &&
		strings.TrimSpace(c.LocalE2EAuthUserID) != "" &&
		strings.TrimSpace(c.LocalE2EAuthEmail) != ""
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

func getEnvBool(key string, defaultValue bool) bool {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		switch strings.ToLower(value) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		default:
			log.Printf("config: failed to parse %s=%q as boolean, using default: %t", key, value, defaultValue)
		}
	}
	return defaultValue
}
