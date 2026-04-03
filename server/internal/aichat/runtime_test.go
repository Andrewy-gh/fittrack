package aichat

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

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

func TestDebugStreamDelayUsesPositiveMilliseconds(t *testing.T) {
	t.Setenv(debugStreamDelayEnvVar, "250")

	if got := debugStreamDelay(); got != 250*time.Millisecond {
		t.Fatalf("debugStreamDelay() = %v, want %v", got, 250*time.Millisecond)
	}
}

func TestDebugStreamDelayIgnoresInvalidValues(t *testing.T) {
	testCases := []string{"", "abc", "0", "-10"}

	for _, value := range testCases {
		t.Run(value, func(t *testing.T) {
			t.Setenv(debugStreamDelayEnvVar, value)

			if got := debugStreamDelay(); got != 0 {
				t.Fatalf("debugStreamDelay() = %v, want 0", got)
			}
		})
	}
}

func TestDebugForceRecoveryAfterChunksUsesPositiveCount(t *testing.T) {
	t.Setenv(debugForceRecoveryEnvVar, "3")

	if got := debugForceRecoveryAfterChunks(); got != 3 {
		t.Fatalf("debugForceRecoveryAfterChunks() = %d, want 3", got)
	}
}

func TestDebugForceRecoveryAfterChunksIgnoresInvalidValues(t *testing.T) {
	testCases := []string{"", "abc", "0", "-2"}

	for _, value := range testCases {
		t.Run(value, func(t *testing.T) {
			t.Setenv(debugForceRecoveryEnvVar, value)

			if got := debugForceRecoveryAfterChunks(); got != 0 {
				t.Fatalf("debugForceRecoveryAfterChunks() = %d, want 0", got)
			}
		})
	}
}

func TestForegroundStreamDebugEnabledUsesContextFlag(t *testing.T) {
	if foregroundStreamDebugEnabled(context.Background()) {
		t.Fatal("foregroundStreamDebugEnabled() should default to false")
	}

	if !foregroundStreamDebugEnabled(withForegroundStreamDebug(context.Background(), true)) {
		t.Fatal("foregroundStreamDebugEnabled() should return true when enabled in context")
	}

	if foregroundStreamDebugEnabled(withForegroundStreamDebug(context.Background(), false)) {
		t.Fatal("foregroundStreamDebugEnabled() should return false when disabled in context")
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

func TestSplitStreamTextUsesRuneChunks(t *testing.T) {
	got := splitStreamText("abcdefg", 3)
	want := []string{"abc", "def", "g"}
	if !slices.Equal(got, want) {
		t.Fatalf("splitStreamText() = %#v, want %#v", got, want)
	}

	got = splitStreamText("go🏋️chat", 2)
	want = []string{"go", "🏋️", "ch", "at"}
	if !slices.Equal(got, want) {
		t.Fatalf("splitStreamText() with emoji = %#v, want %#v", got, want)
	}
}

func TestEmitStreamTextSplitsWhenDebugDelayEnabled(t *testing.T) {
	t.Setenv(debugStreamDelayEnvVar, "5")

	originalSleep := sleepWithContext
	sleepCalls := 0
	sleepWithContext = func(context.Context, time.Duration) error {
		sleepCalls++
		return nil
	}
	t.Cleanup(func() {
		sleepWithContext = originalSleep
	})

	var chunks []string
	err := emitStreamText(context.Background(), "abcdefghijklmnop", func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	}, newStreamDebugState(false))
	if err != nil {
		t.Fatalf("emitStreamText() error = %v", err)
	}

	want := []string{"abcdefghijkl", "mnop"}
	if !slices.Equal(chunks, want) {
		t.Fatalf("emitStreamText() chunks = %#v, want %#v", chunks, want)
	}
	if sleepCalls != 1 {
		t.Fatalf("sleep calls = %d, want 1", sleepCalls)
	}
}

func TestEmitStreamTextForcesRecoveryAfterConfiguredChunkCount(t *testing.T) {
	originalSleep := sleepWithContext
	sleepWithContext = func(context.Context, time.Duration) error {
		return nil
	}
	t.Cleanup(func() {
		sleepWithContext = originalSleep
	})

	debugState := &streamDebugState{
		delay:                    time.Millisecond,
		forceRecoveryAfterChunks: 2,
	}

	var chunks []string
	err := emitStreamText(context.Background(), "abcdefghijklmnop", func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	}, debugState)
	if !errors.Is(err, ErrStreamDisconnected) {
		t.Fatalf("emitStreamText() error = %v, want %v", err, ErrStreamDisconnected)
	}

	want := []string{"abcdefghijkl", "mnop"}
	if !slices.Equal(chunks, want) {
		t.Fatalf("emitStreamText() chunks = %#v, want %#v", chunks, want)
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
