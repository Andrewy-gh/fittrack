package aichat

import (
	"context"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type stubFeatureAccessReader struct{}

func (stubFeatureAccessReader) ListCurrentUserAccess(context.Context) ([]featureaccess.FeatureAccessGrant, error) {
	return nil, nil
}

func TestResolveModelNameDefaultsToGemini25Flash(t *testing.T) {
	t.Setenv("GEMINI_MODEL", "")

	if got := resolveModelName(); got != defaultModelName {
		t.Fatalf("resolveModelName() = %q, want %q", got, defaultModelName)
	}
}

func TestResolveModelNameUsesOverride(t *testing.T) {
	const override = "googleai/gemini-2.0-flash"
	t.Setenv("GEMINI_MODEL", override)

	if got := resolveModelName(); got != override {
		t.Fatalf("resolveModelName() = %q, want %q", got, override)
	}
}

func TestNewGenkitRuntimeReturnsUnavailableWithoutGoogleAPIKey(t *testing.T) {
	t.Setenv(googleAPIKeyEnvVar, "")

	runtime := NewGenkitRuntime(context.Background(), stubFeatureAccessReader{})

	if runtime == nil {
		t.Fatal("NewGenkitRuntime() returned nil")
	}
	if runtime.Available() {
		t.Fatal("NewGenkitRuntime() should leave runtime unavailable without a configured key")
	}
}

func TestNewGenkitRuntimeSkipsAvailabilityWhenGenkitPanics(t *testing.T) {
	t.Setenv(googleAPIKeyEnvVar, "configured")

	original := genkitInit
	genkitInit = func(context.Context, ...genkit.GenkitOption) *genkit.Genkit {
		panic("boom")
	}
	t.Cleanup(func() {
		genkitInit = original
	})

	runtime := NewGenkitRuntime(context.Background(), stubFeatureAccessReader{})

	if runtime == nil {
		t.Fatal("NewGenkitRuntime() returned nil")
	}
	if runtime.Available() {
		t.Fatal("NewGenkitRuntime() should leave runtime unavailable after init panic")
	}
	if runtime.g != nil {
		t.Fatal("NewGenkitRuntime() should not retain a genkit instance after init panic")
	}
	if runtime.tool != (ai.Tool)(nil) {
		t.Fatal("NewGenkitRuntime() should not retain a tool after init panic")
	}
}
