package aichat

import (
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/firebase/genkit/go/ai"
)

func TestBuildChatSystemPromptIncludesWorkoutGuardrails(t *testing.T) {
	prompt := buildChatSystemPrompt()

	requiredSnippets := []string{
		"ask at most 3 short, focused follow-up questions",
		"call the " + workoutDraftToolName + " tool immediately",
		"Do not list specific exercises, sets, or reps in plain text before",
		"only way to produce a structured workout draft",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("buildChatSystemPrompt() missing %q\nprompt=%s", snippet, prompt)
		}
	}
}

func TestBuildWorkoutGenerationPromptIncludesFitTrackContract(t *testing.T) {
	prompt := buildWorkoutGenerationPrompt(WorkoutGenerationToolInput{}, time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC))

	requiredSnippets := []string{
		`"date": "RFC3339 timestamp"`,
		`"workoutFocus": "optional string"`,
		`"setType": "warmup" | "working"`,
		`"date" is always required and must be RFC3339.`,
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("buildWorkoutGenerationPrompt() missing %q\nprompt=%s", snippet, prompt)
		}
	}
}

func TestBuildWorkoutGenerationPromptUsesUserLocalLanguageForRelativeDates(t *testing.T) {
	loc := time.FixedZone("EDT", -4*60*60)
	prompt := buildWorkoutGenerationPrompt(
		WorkoutGenerationToolInput{WorkoutDate: "tomorrow"},
		time.Date(2026, 4, 20, 23, 30, 0, 0, loc),
	)

	if !strings.Contains(prompt, `interpret it from the user's local day rather than UTC`) {
		t.Fatalf("buildWorkoutGenerationPrompt() = %q, want user-local relative date guidance", prompt)
	}
	if strings.Contains(prompt, `relative to`) {
		t.Fatalf("buildWorkoutGenerationPrompt() = %q, should not anchor relative dates to a server timestamp", prompt)
	}
}

func TestExtractWorkoutDraftFromHistoryValidatesCurrentWorkoutContract(t *testing.T) {
	history := []*ai.Message{
		ai.NewMessage(ai.RoleTool, nil, ai.NewToolResponsePart(&ai.ToolResponse{
			Name: workoutDraftToolName,
			Output: map[string]any{
				"date":         "2026-04-20T12:00:00Z",
				"workoutFocus": "pull",
				"exercises": []map[string]any{
					{
						"name": "Chest Supported Row",
						"sets": []map[string]any{
							{
								"reps":    10,
								"setType": "working",
							},
						},
					},
				},
			},
		})),
	}

	draft, err := extractWorkoutDraftFromHistory(history)
	if err != nil {
		t.Fatalf("extractWorkoutDraftFromHistory() error = %v", err)
	}
	if draft == nil {
		t.Fatal("extractWorkoutDraftFromHistory() returned nil draft")
	}
	if draft.Date != "2026-04-20T12:00:00Z" {
		t.Fatalf("draft.Date = %q, want RFC3339 value", draft.Date)
	}
	if len(draft.Exercises) != 1 || draft.Exercises[0].Name != "Chest Supported Row" {
		t.Fatalf("unexpected exercises: %#v", draft.Exercises)
	}
}

func TestExtractWorkoutDraftFromHistoryRejectsInvalidContract(t *testing.T) {
	history := []*ai.Message{
		ai.NewMessage(ai.RoleTool, nil, ai.NewToolResponsePart(&ai.ToolResponse{
			Name: workoutDraftToolName,
			Output: map[string]any{
				"workoutFocus": "legs",
				"exercises": []map[string]any{
					{
						"name": "Goblet Squat",
						"sets": []map[string]any{
							{
								"reps":    10,
								"setType": "working",
							},
						},
					},
				},
			},
		})),
	}

	draft, err := extractWorkoutDraftFromHistory(history)
	if err == nil {
		t.Fatalf("extractWorkoutDraftFromHistory() = %#v, want validation error", draft)
	}
	if !strings.Contains(err.Error(), "Date") {
		t.Fatalf("extractWorkoutDraftFromHistory() error = %v, want missing date validation", err)
	}
}

func TestFinalizeAssistantTextSuppressesExerciseDumpWhenWorkoutDraftExists(t *testing.T) {
	draft := &workout.CreateWorkoutRequest{
		Date: "2026-04-20T12:00:00Z",
		Exercises: []workout.ExerciseInput{
			{
				Name: "Bench Press",
				Sets: []workout.SetInput{
					{Reps: 5, SetType: "working"},
				},
			},
			{
				Name: "Incline Dumbbell Press",
				Sets: []workout.SetInput{
					{Reps: 10, SetType: "working"},
				},
			},
		},
	}

	text := finalizeAssistantText("1. Bench Press\n2. Incline Dumbbell Press", draft)
	if text != workoutDraftSummaryMessage {
		t.Fatalf("finalizeAssistantText() = %q, want %q", text, workoutDraftSummaryMessage)
	}
}
