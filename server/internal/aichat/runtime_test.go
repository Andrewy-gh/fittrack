package aichat

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type stubFeatureAccessReader struct{}

func (stubFeatureAccessReader) ListCurrentUserAccess(context.Context) ([]featureaccess.FeatureAccessGrant, error) {
	return nil, nil
}

func TestConfiguredAPIKeyEnvVar(t *testing.T) {
	tests := []struct {
		name      string
		geminiKey string
		googleKey string
		want      string
	}{
		{
			name:      "prefers gemini key to match googlegenai plugin precedence",
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
			name: "reports no configured key",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(geminiAPIKeyEnvVar, tt.geminiKey)
			t.Setenv(googleAPIKeyEnvVar, tt.googleKey)

			if got := configuredAPIKeyEnvVar(); got != tt.want {
				t.Fatalf("configuredAPIKeyEnvVar() = %q, want %q", got, tt.want)
			}
		})
	}
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
	t.Setenv(geminiAPIKeyEnvVar, "")

	runtime := NewGenkitRuntime(context.Background(), stubFeatureAccessReader{})

	if runtime == nil {
		t.Fatal("NewGenkitRuntime() returned nil")
	}
	if runtime.Available() {
		t.Fatal("NewGenkitRuntime() should leave runtime unavailable without a configured key")
	}
}

func TestNewGenkitRuntimeSkipsAvailabilityWhenGenkitPanics(t *testing.T) {
	t.Setenv(geminiAPIKeyEnvVar, "configured")
	t.Setenv(googleAPIKeyEnvVar, "")

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

func TestActiveFeaturesToolNameIsGeminiCompatible(t *testing.T) {
	validToolNamePattern := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]{0,63}$`)

	if !validToolNamePattern.MatchString(activeFeaturesToolName) {
		t.Fatalf("activeFeaturesToolName = %q, want Gemini-compatible tool name", activeFeaturesToolName)
	}
	if strings.Contains(activeFeaturesToolName, "/") {
		t.Fatalf("activeFeaturesToolName = %q, slash is not allowed in Gemini tool names", activeFeaturesToolName)
	}
}

func TestToolCallGuardCachesFeatureSnapshot(t *testing.T) {
	featureAccess := &countingFeatureAccessReader{
		grants: []featureaccess.FeatureAccessGrant{
			{FeatureKey: "beta"},
			{FeatureKey: "alpha"},
		},
	}
	guard := toolGuardFromContext(withToolGuard(context.Background()))

	first, err := guard.listActiveFeatures(context.Background(), featureAccess)
	if err != nil {
		t.Fatalf("first listActiveFeatures() error = %v", err)
	}
	second, err := guard.listActiveFeatures(context.Background(), featureAccess)
	if err != nil {
		t.Fatalf("second listActiveFeatures() error = %v", err)
	}

	if featureAccess.calls != 1 {
		t.Fatalf("feature access calls = %d, want 1", featureAccess.calls)
	}
	if got := first.FeatureKeys; len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("unexpected first snapshot: %#v", got)
	}
	if got := second.FeatureKeys; len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("unexpected second snapshot: %#v", got)
	}
	first.FeatureKeys[0] = "mutated"
	if second.FeatureKeys[0] != "alpha" {
		t.Fatal("tool guard should return cloned cached snapshots")
	}
}

func TestToolCallGuardCachesErrors(t *testing.T) {
	expectedErr := errors.New("boom")
	featureAccess := &countingFeatureAccessReader{err: expectedErr}
	guard := toolGuardFromContext(withToolGuard(context.Background()))

	_, err := guard.listActiveFeatures(context.Background(), featureAccess)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("first listActiveFeatures() error = %v, want %v", err, expectedErr)
	}
	_, err = guard.listActiveFeatures(context.Background(), featureAccess)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("second listActiveFeatures() error = %v, want %v", err, expectedErr)
	}
	if featureAccess.calls != 1 {
		t.Fatalf("feature access calls = %d, want 1", featureAccess.calls)
	}
}

func TestPromptsReferenceActiveFeaturesToolName(t *testing.T) {
	structuredPrompt := buildStructuredPrompt("test prompt")
	chatPrompt := buildChatSystemPrompt()

	if !strings.Contains(structuredPrompt, activeFeaturesToolName) {
		t.Fatalf("buildStructuredPrompt() = %q, want tool name %q", structuredPrompt, activeFeaturesToolName)
	}
	if !strings.Contains(chatPrompt, activeFeaturesToolName) {
		t.Fatalf("buildChatSystemPrompt() = %q, want tool name %q", chatPrompt, activeFeaturesToolName)
	}
}

type countingFeatureAccessReader struct {
	grants []featureaccess.FeatureAccessGrant
	err    error
	calls  int
}

func (r *countingFeatureAccessReader) ListCurrentUserAccess(context.Context) ([]featureaccess.FeatureAccessGrant, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	return r.grants, nil
}
