package aichat

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
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

	runtime := NewGenkitRuntime(context.Background(), nil)

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

	runtime := NewGenkitRuntime(context.Background(), nil)

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
	if !validToolNamePattern.MatchString(getExerciseStatsToolName) {
		t.Fatalf("getExerciseStatsToolName = %q, want Gemini-compatible tool name", getExerciseStatsToolName)
	}
}

func TestPromptsReferenceToolNames(t *testing.T) {
	structuredPrompt := buildStructuredPrompt("test prompt")
	chatPrompt := buildChatSystemPrompt(nil, nil, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC), true)

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

func TestBuildChatSystemPromptComposesDataSections(t *testing.T) {
	snapshot := &TrainingSnapshot{
		LastWorkoutDate: "2026-07-03",
		WorkoutsLast30D: 7,
		TopExercises:    []string{"Bench Press", "Back Squat"},
	}
	prompt := buildChatSystemPrompt(snapshot, nil, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC), true)

	for _, snippet := range []string{
		"call the " + getWorkoutsToolName + " tool",
		"Default to " + getWorkoutsToolName + " for personal workout-history questions",
		"use " + getExerciseStatsToolName + " only for all-time bests",
		"do not call data tools unless the user explicitly references past training",
		"Current date: 2026-07-06.",
		"User training snapshot:",
		"Last workout: 2026-07-03",
		"Workouts in last 30 days: 7",
		"Most frequent exercises: Bench Press, Back Squat",
	} {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("buildChatSystemPrompt() missing %q\nprompt=%s", snippet, prompt)
		}
	}
}

func TestBuildChatSystemPromptOmitsSnapshotAndDataToolWhenReaderNil(t *testing.T) {
	prompt := buildChatSystemPrompt(nil, nil, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC), false)

	if strings.Contains(prompt, "User training snapshot:") {
		t.Fatalf("buildChatSystemPrompt() included snapshot without reader: %s", prompt)
	}
	if strings.Contains(prompt, getWorkoutsToolName) {
		t.Fatalf("buildChatSystemPrompt() referenced data tool without reader: %s", prompt)
	}
	if !strings.Contains(prompt, "Do not call data tools for general fitness knowledge.") {
		t.Fatalf("buildChatSystemPrompt() missing nil-reader data routing rule: %s", prompt)
	}
}

func TestBuildChatSystemPromptComposesTrainingProfileSection(t *testing.T) {
	profile := &TrainingProfile{
		PrimaryGoal:                     "hypertrophy",
		ExperienceLevel:                 "intermediate",
		PreferredSessionDurationMinutes: 45,
		UsualTrainingLocation:           "home",
		AvailableEquipment:              []string{"adjustable dumbbells", "bench"},
		AvoidedExercises:                []string{"burpees"},
		MovementLimitations:             []string{"no overhead pressing"},
	}
	prompt := buildChatSystemPrompt(nil, profile, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC), true)

	for _, snippet := range []string{
		"User training profile:",
		"Goal: hypertrophy",
		"Experience: intermediate",
		"Preferred duration: 45 minutes",
		"Usual location: home",
		"Available equipment: adjustable dumbbells, bench",
		"Avoided exercises: burpees",
		"Movement limitations: no overhead pressing",
		"Treat user profile values as defaults",
		"The user's current message always overrides the profile",
		"none-stated profile value is sufficient injury status",
		"Call the " + updateTrainingProfileToolName + " tool only for durable training facts",
		"Do not call it for one-off session details",
		"unless you called the " + updateTrainingProfileToolName + " tool in this same turn",
	} {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("buildChatSystemPrompt() missing %q\nprompt=%s", snippet, prompt)
		}
	}
}

func TestIsEmptyChatModelResponse(t *testing.T) {
	cases := []struct {
		name         string
		resp         *ai.ModelResponse
		streamedText string
		want         bool
	}{
		{
			name: "nil response",
			resp: nil,
			want: true,
		},
		{
			name: "empty candidate with no tool calls",
			resp: &ai.ModelResponse{Request: &ai.ModelRequest{}},
			want: true,
		},
		{
			name: "final text present",
			resp: &ai.ModelResponse{
				Request: &ai.ModelRequest{},
				Message: ai.NewModelMessage(ai.NewTextPart("Do you have any injuries?")),
			},
			want: false,
		},
		{
			name:         "streamed text only",
			resp:         &ai.ModelResponse{Request: &ai.ModelRequest{}},
			streamedText: "partial answer",
			want:         false,
		},
		{
			name: "tool call in history without text",
			resp: &ai.ModelResponse{
				Request: &ai.ModelRequest{
					Messages: []*ai.Message{
						ai.NewModelMessage(ai.NewToolRequestPart(&ai.ToolRequest{Name: getWorkoutsToolName})),
					},
				},
				Message: ai.NewModelMessage(ai.NewTextPart("")),
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isEmptyChatModelResponse(tc.resp, tc.streamedText); got != tc.want {
				t.Fatalf("isEmptyChatModelResponse() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestChatToolsOmitsDataToolWhenReaderNil(t *testing.T) {
	draft := fakeTool{name: workoutDraftToolName}
	workouts := fakeTool{name: getWorkoutsToolName}
	stats := fakeTool{name: getExerciseStatsToolName}
	profile := fakeTool{name: updateTrainingProfileToolName}

	withData := (&GenkitRuntime{workoutDraftTool: draft, getWorkoutsTool: workouts, getExerciseStatsTool: stats, updateProfileTool: profile}).chatTools()
	if len(withData) != 4 ||
		withData[0].Name() != workoutDraftToolName ||
		withData[1].Name() != getWorkoutsToolName ||
		withData[2].Name() != getExerciseStatsToolName ||
		withData[3].Name() != updateTrainingProfileToolName {
		t.Fatalf("chatTools() with data = %#v, want draft then workout data then exercise stats then profile update", withData)
	}

	withoutData := (&GenkitRuntime{workoutDraftTool: draft}).chatTools()
	if len(withoutData) != 1 || withoutData[0].Name() != workoutDraftToolName {
		t.Fatalf("chatTools() without data = %#v, want only draft tool", withoutData)
	}
}

type fakeTool struct {
	name string
}

func (f fakeTool) Name() string { return f.name }

func (f fakeTool) Definition() *ai.ToolDefinition { return &ai.ToolDefinition{Name: f.name} }

func (f fakeTool) RunRaw(context.Context, any) (any, error) { return nil, nil }

func (f fakeTool) RunRawMultipart(context.Context, any) (*ai.MultipartToolResponse, error) {
	return &ai.MultipartToolResponse{}, nil
}

func (f fakeTool) Respond(*ai.Part, any, *ai.RespondOptions) *ai.Part { return nil }

func (f fakeTool) Restart(*ai.Part, *ai.RestartOptions) *ai.Part { return nil }

func (f fakeTool) Register(api.Registry) {}
