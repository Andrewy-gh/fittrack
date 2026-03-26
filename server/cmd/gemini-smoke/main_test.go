package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatResponseCondensesWhitespace(t *testing.T) {
	got := formatResponse("  FitTrack\n   smoke\t test   ok  ")

	if got != "FitTrack smoke test ok" {
		t.Fatalf("formatResponse() = %q", got)
	}
}

func TestFormatResponseTruncatesLongOutput(t *testing.T) {
	got := formatResponse(strings.Repeat("a", 200))

	if len(got) != 160 {
		t.Fatalf("expected truncated output length 160, got %d", len(got))
	}

	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected truncated output to end with ellipsis, got %q", got)
	}
}

func TestGetModelNameDefaultsToGemini25Flash(t *testing.T) {
	t.Setenv("GEMINI_MODEL", "")

	if got := getModelName(); got != defaultGeminiModel {
		t.Fatalf("getModelName() = %q", got)
	}
}

func TestGetModelNameUsesOverride(t *testing.T) {
	const override = "googleai/gemini-2.0-flash"
	t.Setenv("GEMINI_MODEL", override)

	if got := getModelName(); got != override {
		t.Fatalf("getModelName() = %q", got)
	}
}

func TestConfiguredAPIKeyEnvVar(t *testing.T) {
	tests := []struct {
		name      string
		geminiKey string
		googleKey string
		want      string
	}{
		{
			name:      "prefers gemini api key",
			geminiKey: "gemini-key",
			googleKey: "google-key",
			want:      geminiAPIKeyEnvVar,
		},
		{
			name:      "accepts google api key",
			googleKey: "google-key",
			want:      googleAPIKeyEnvVar,
		},
		{
			name:      "reports missing key",
			geminiKey: "",
			googleKey: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsetEnvForTest(t, geminiAPIKeyEnvVar, googleAPIKeyEnvVar)
			if tt.geminiKey != "" {
				t.Setenv(geminiAPIKeyEnvVar, tt.geminiKey)
			}
			if tt.googleKey != "" {
				t.Setenv(googleAPIKeyEnvVar, tt.googleKey)
			}

			if got := configuredAPIKeyEnvVar(); got != tt.want {
				t.Fatalf("configuredAPIKeyEnvVar() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadLocalEnvRespectsPriorityAndSetenv(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ".env.local"), "GEMINI_MODEL=local-model\n")
	writeTestFile(t, filepath.Join(dir, ".env"), "GOOGLE_API_KEY=from-dotenv\nGEMINI_MODEL=env-model\n")
	writeTestFile(t, filepath.Join(dir, "setenv.sh"), "export GEMINI_API_KEY=from-setenv\nexport GEMINI_MODEL=setenv-model\n")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd error = %v", err)
		}
	})

	unsetEnvForTest(t, geminiAPIKeyEnvVar, googleAPIKeyEnvVar, "GEMINI_MODEL")

	if err := loadLocalEnv(); err != nil {
		t.Fatalf("loadLocalEnv() error = %v", err)
	}

	if got := os.Getenv(geminiAPIKeyEnvVar); got != "from-setenv" {
		t.Fatalf("expected %s from setenv.sh, got %q", geminiAPIKeyEnvVar, got)
	}
	if got := os.Getenv(googleAPIKeyEnvVar); got != "from-dotenv" {
		t.Fatalf("expected %s from .env, got %q", googleAPIKeyEnvVar, got)
	}
	if got := os.Getenv("GEMINI_MODEL"); got != "local-model" {
		t.Fatalf("expected GEMINI_MODEL from .env.local, got %q", got)
	}
}

func TestIsQuotaOr429Error(t *testing.T) {
	raw := "Error 429, Message: quota exceeded, Status: RESOURCE_EXHAUSTED"

	if !isQuotaOr429Error(raw) {
		t.Fatalf("expected quota error to be detected")
	}
}

func TestFormatRunErrorForTimeoutIncludesDeadline(t *testing.T) {
	got := formatRunError(context.DeadlineExceeded, defaultGeminiModel)

	for _, want := range []string{
		"timed out after 20s",
		defaultGeminiModel,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected timeout error to contain %q, got %q", want, got)
		}
	}
}

func TestFormatRunErrorForQuotaIncludesActionableGuidance(t *testing.T) {
	raw := "Error 429, Message: quota exceeded, Status: RESOURCE_EXHAUSTED"

	got := formatRunError(errors.New(raw), "googleai/gemini-2.0-flash")

	for _, want := range []string{
		"Google accepted the API key",
		"default smoke-test model",
		`$env:GEMINI_MODEL="googleai/gemini-2.5-flash"; go run ./cmd/gemini-smoke`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected formatted error to contain %q, got %q", want, got)
		}
	}
}

func unsetEnvForTest(t *testing.T, keys ...string) {
	t.Helper()

	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("os.Unsetenv(%q) error = %v", key, err)
		}

		t.Cleanup(func() {
			if !ok {
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("os.Unsetenv(%q) cleanup error = %v", key, err)
				}
				return
			}
			if err := os.Setenv(key, value); err != nil {
				t.Fatalf("os.Setenv(%q) cleanup error = %v", key, err)
			}
		})
	}
}

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}
