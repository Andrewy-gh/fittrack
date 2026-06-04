package aichat

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/firebase/genkit/go/genkit"
)

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

	runtime := NewGenkitRuntime(context.Background())

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

	runtime := NewGenkitRuntime(context.Background())

	if runtime == nil {
		t.Fatal("NewGenkitRuntime() returned nil")
	}
	if runtime.Available() {
		t.Fatal("NewGenkitRuntime() should leave runtime unavailable after init panic")
	}
	if runtime.g != nil {
		t.Fatal("NewGenkitRuntime() should not retain a genkit instance after init panic")
	}
	if runtime.workoutDraftTool != nil {
		t.Fatal("NewGenkitRuntime() should not retain a workout draft tool after init panic")
	}
}

func TestWorkoutDraftToolNameIsGeminiCompatible(t *testing.T) {
	validToolNamePattern := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]{0,63}$`)

	if !validToolNamePattern.MatchString(workoutDraftToolName) {
		t.Fatalf("workoutDraftToolName = %q, want Gemini-compatible tool name", workoutDraftToolName)
	}
	if strings.Contains(workoutDraftToolName, "/") {
		t.Fatalf("workoutDraftToolName = %q, slash is not allowed in Gemini tool names", workoutDraftToolName)
	}
}

func TestPromptsReferenceToolNames(t *testing.T) {
	structuredPrompt := buildStructuredPrompt("test prompt")
	chatPrompt := buildChatSystemPrompt()

	if strings.Contains(structuredPrompt, "list_active_features") {
		t.Fatalf("buildStructuredPrompt() = %q, should not reference active feature tools", structuredPrompt)
	}
	if strings.Contains(chatPrompt, "list_active_features") {
		t.Fatalf("buildChatSystemPrompt() = %q, should not reference active feature tools", chatPrompt)
	}
	if !strings.Contains(chatPrompt, workoutDraftToolName) {
		t.Fatalf("buildChatSystemPrompt() = %q, want workout tool name %q", chatPrompt, workoutDraftToolName)
	}
}
