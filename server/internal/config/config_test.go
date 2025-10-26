package config

import (
	"os"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Set required environment variables
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
	os.Setenv("PROJECT_ID", "test-project-123")
	defer cleanupEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.DatabaseURL != "postgresql://user:pass@localhost:5432/testdb" {
		t.Errorf("expected DatabaseURL to be set, got: %s", cfg.DatabaseURL)
	}

	if cfg.ProjectID != "test-project-123" {
		t.Errorf("expected ProjectID to be 'test-project-123', got: %s", cfg.ProjectID)
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	// Only set PROJECT_ID, leave DATABASE_URL empty
	os.Setenv("PROJECT_ID", "test-project-123")
	os.Unsetenv("DATABASE_URL")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing, got nil")
	}
}

func TestLoad_MissingProjectID(t *testing.T) {
	// Only set DATABASE_URL, leave PROJECT_ID empty
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
	os.Unsetenv("PROJECT_ID")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when PROJECT_ID is missing, got nil")
	}
}

func TestLoad_InvalidDatabaseURL(t *testing.T) {
	// Set invalid URL format
	os.Setenv("DATABASE_URL", "not-a-valid-url")
	os.Setenv("PROJECT_ID", "test-project-123")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is invalid, got nil")
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
	os.Setenv("PROJECT_ID", "test-project-123")
	os.Setenv("LOG_LEVEL", "invalid-level")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when LOG_LEVEL is invalid, got nil")
	}
}

func TestLoad_InvalidEnvironment(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
	os.Setenv("PROJECT_ID", "test-project-123")
	os.Setenv("ENVIRONMENT", "invalid-env")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when ENVIRONMENT is invalid, got nil")
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
	os.Setenv("PROJECT_ID", "test-project-123")
	defer cleanupEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check default values
	if cfg.Port != 8080 {
		t.Errorf("expected default Port to be 8080, got: %d", cfg.Port)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected default LogLevel to be 'info', got: %s", cfg.LogLevel)
	}

	if cfg.Environment != "development" {
		t.Errorf("expected default Environment to be 'development', got: %s", cfg.Environment)
	}

	if cfg.RateLimitRPM != 100 {
		t.Errorf("expected default RateLimitRPM to be 100, got: %d", cfg.RateLimitRPM)
	}

	if cfg.DBMaxConns != 15 {
		t.Errorf("expected default DBMaxConns to be 15, got: %d", cfg.DBMaxConns)
	}

	if cfg.DBMinConns != 2 {
		t.Errorf("expected default DBMinConns to be 2, got: %d", cfg.DBMinConns)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
	os.Setenv("PROJECT_ID", "test-project-123")
	os.Setenv("PORT", "9000")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("RATE_LIMIT_RPM", "200")
	defer cleanupEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Port != 9000 {
		t.Errorf("expected Port to be 9000, got: %d", cfg.Port)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel to be 'debug', got: %s", cfg.LogLevel)
	}

	if cfg.Environment != "production" {
		t.Errorf("expected Environment to be 'production', got: %s", cfg.Environment)
	}

	if cfg.RateLimitRPM != 200 {
		t.Errorf("expected RateLimitRPM to be 200, got: %d", cfg.RateLimitRPM)
	}
}

func TestGetAllowedOrigins_EmptyString(t *testing.T) {
	cfg := &Config{AllowedOrigins: ""}
	origins := cfg.GetAllowedOrigins()

	if len(origins) != 0 {
		t.Errorf("expected empty slice, got: %v", origins)
	}
}

func TestGetAllowedOrigins_SingleOrigin(t *testing.T) {
	cfg := &Config{AllowedOrigins: "https://example.com"}
	origins := cfg.GetAllowedOrigins()

	if len(origins) != 1 {
		t.Fatalf("expected 1 origin, got: %d", len(origins))
	}

	if origins[0] != "https://example.com" {
		t.Errorf("expected 'https://example.com', got: %s", origins[0])
	}
}

func TestGetAllowedOrigins_MultipleOrigins(t *testing.T) {
	cfg := &Config{AllowedOrigins: "https://example.com,https://app.example.com,http://localhost:3000"}
	origins := cfg.GetAllowedOrigins()

	if len(origins) != 3 {
		t.Fatalf("expected 3 origins, got: %d", len(origins))
	}

	expected := []string{"https://example.com", "https://app.example.com", "http://localhost:3000"}
	for i, exp := range expected {
		if origins[i] != exp {
			t.Errorf("expected origins[%d] to be '%s', got: %s", i, exp, origins[i])
		}
	}
}

func TestGetAllowedOrigins_WithWhitespace(t *testing.T) {
	cfg := &Config{AllowedOrigins: "https://example.com , https://app.example.com  ,  http://localhost:3000"}
	origins := cfg.GetAllowedOrigins()

	if len(origins) != 3 {
		t.Fatalf("expected 3 origins, got: %d", len(origins))
	}

	// All whitespace should be trimmed
	expected := []string{"https://example.com", "https://app.example.com", "http://localhost:3000"}
	for i, exp := range expected {
		if origins[i] != exp {
			t.Errorf("expected origins[%d] to be '%s', got: %s", i, exp, origins[i])
		}
	}
}

func TestGetAllowedOrigins_WithEmptyValues(t *testing.T) {
	cfg := &Config{AllowedOrigins: "https://example.com,,https://app.example.com"}
	origins := cfg.GetAllowedOrigins()

	// Empty values should be filtered out
	if len(origins) != 2 {
		t.Fatalf("expected 2 origins, got: %d", len(origins))
	}

	expected := []string{"https://example.com", "https://app.example.com"}
	for i, exp := range expected {
		if origins[i] != exp {
			t.Errorf("expected origins[%d] to be '%s', got: %s", i, exp, origins[i])
		}
	}
}

func TestLoad_ValidLogLevels(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range validLevels {
		t.Run("LogLevel_"+level, func(t *testing.T) {
			os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
			os.Setenv("PROJECT_ID", "test-project-123")
			os.Setenv("LOG_LEVEL", level)
			defer cleanupEnv()

			cfg, err := Load()
			if err != nil {
				t.Fatalf("expected no error for valid log level '%s', got: %v", level, err)
			}

			if cfg.LogLevel != level {
				t.Errorf("expected LogLevel to be '%s', got: %s", level, cfg.LogLevel)
			}
		})
	}
}

func TestLoad_ValidEnvironments(t *testing.T) {
	validEnvs := []string{"development", "staging", "production"}

	for _, env := range validEnvs {
		t.Run("Environment_"+env, func(t *testing.T) {
			os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
			os.Setenv("PROJECT_ID", "test-project-123")
			os.Setenv("ENVIRONMENT", env)
			defer cleanupEnv()

			cfg, err := Load()
			if err != nil {
				t.Fatalf("expected no error for valid environment '%s', got: %v", env, err)
			}

			if cfg.Environment != env {
				t.Errorf("expected Environment to be '%s', got: %s", env, cfg.Environment)
			}
		})
	}
}

// cleanupEnv removes all test environment variables
func cleanupEnv() {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("PROJECT_ID")
	os.Unsetenv("PORT")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("RATE_LIMIT_RPM")
	os.Unsetenv("ALLOWED_ORIGINS")
	os.Unsetenv("DB_MAX_CONNS")
	os.Unsetenv("DB_MIN_CONNS")
	os.Unsetenv("DB_MAX_CONN_IDLE")
	os.Unsetenv("DB_MAX_CONN_LIFE")
	os.Unsetenv("DB_HEALTHCHECK")
}
