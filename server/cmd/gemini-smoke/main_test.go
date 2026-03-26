package main

import (
	"errors"
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

func TestIsQuotaOr429Error(t *testing.T) {
	raw := "Error 429, Message: quota exceeded, Status: RESOURCE_EXHAUSTED"

	if !isQuotaOr429Error(raw) {
		t.Fatalf("expected quota error to be detected")
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
