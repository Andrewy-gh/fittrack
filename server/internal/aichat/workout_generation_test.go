package aichat

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/firebase/genkit/go/ai"
)

func TestBuildChatSystemPromptIncludesWorkoutGuardrails(t *testing.T) {
	prompt := buildChatSystemPrompt()

	requiredSnippets := []string{
		"Ask at most 3 short, focused follow-up questions",
		"MVP-ready inputs are workout focus, session duration, enough equipment or workout context",
		"do not treat it as a hard blocker",
		"If injury status is missing, ask once",
		"assume injuries are \"none\" and proceed",
		"Do not ask scheduling, frequency, or future-date questions",
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

func TestWorkoutDraftToolDescriptionMatchesMVPReadiness(t *testing.T) {
	requiredSnippets := []string{
		"workout focus",
		"session duration",
		"enough equipment or workout context",
		"injury status",
		"Fitness level improves weight assumptions but is not required",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(workoutDraftToolDescription, snippet) {
			t.Fatalf("workoutDraftToolDescription missing %q\ndescription=%s", snippet, workoutDraftToolDescription)
		}
	}
}

func TestWorkoutGenerationToolInputSchemaDoesNotRequireHelpfulButOptionalFields(t *testing.T) {
	inputType := reflect.TypeOf(WorkoutGenerationToolInput{})
	optionalFields := []string{"FitnessLevel", "FitnessGoal", "Equipment", "SpaceConstraints"}

	for _, fieldName := range optionalFields {
		field, ok := inputType.FieldByName(fieldName)
		if !ok {
			t.Fatalf("WorkoutGenerationToolInput missing field %s", fieldName)
		}
		if !strings.Contains(field.Tag.Get("json"), "omitempty") {
			t.Fatalf("%s json tag = %q, want omitempty", fieldName, field.Tag.Get("json"))
		}
	}

	injuriesField, ok := inputType.FieldByName("Injuries")
	if !ok {
		t.Fatal("WorkoutGenerationToolInput missing field Injuries")
	}
	if strings.Contains(injuriesField.Tag.Get("json"), "omitempty") {
		t.Fatalf("Injuries json tag = %q, should remain required", injuriesField.Tag.Get("json"))
	}
}

func TestValidateWorkoutGenerationToolInputAllowsMVPReadyRequestWithoutFitnessLevel(t *testing.T) {
	input := WorkoutGenerationToolInput{
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "hypertrophy pull day",
		Injuries:        "none",
	}

	if err := validateWorkoutGenerationToolInput(input); err != nil {
		t.Fatalf("validateWorkoutGenerationToolInput() error = %v, want nil", err)
	}
}

func TestValidateWorkoutGenerationToolInputAllowsWorkoutContextWithoutEquipment(t *testing.T) {
	input := WorkoutGenerationToolInput{
		SessionDuration:  20,
		WorkoutFocus:     "bodyweight full body",
		SpaceConstraints: "hotel room",
		Injuries:         "none",
	}

	if err := validateWorkoutGenerationToolInput(input); err != nil {
		t.Fatalf("validateWorkoutGenerationToolInput() error = %v, want nil", err)
	}
}

func TestValidateWorkoutGenerationToolInputAllowsFullySpecifiedRequest(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessLevel:     "intermediate",
		FitnessGoal:      "hypertrophy",
		Equipment:        "barbells, dumbbells, machines",
		SessionDuration:  60,
		WorkoutFocus:     "legs",
		SpaceConstraints: "gym",
		Injuries:         "none",
	}

	if err := validateWorkoutGenerationToolInput(input); err != nil {
		t.Fatalf("validateWorkoutGenerationToolInput() error = %v, want nil", err)
	}
}

func TestValidateWorkoutGenerationToolInputStillRequiresInjuryStatus(t *testing.T) {
	input := WorkoutGenerationToolInput{
		Equipment:       "dumbbells",
		SessionDuration: 30,
		WorkoutFocus:    "upper body",
	}

	err := validateWorkoutGenerationToolInput(input)
	if err == nil {
		t.Fatal("validateWorkoutGenerationToolInput() error = nil, want missing injuries error")
	}
	if !strings.Contains(err.Error(), "injuries") {
		t.Fatalf("validateWorkoutGenerationToolInput() error = %v, want injuries field", err)
	}
}

func TestValidateWorkoutGenerationToolInputRequiresFeasibleWorkoutContext(t *testing.T) {
	input := WorkoutGenerationToolInput{
		SessionDuration: 30,
		WorkoutFocus:    "legs",
		Injuries:        "none",
	}

	err := validateWorkoutGenerationToolInput(input)
	if err == nil {
		t.Fatal("validateWorkoutGenerationToolInput() error = nil, want missing context error")
	}
	if !strings.Contains(err.Error(), "equipment or spaceConstraints") {
		t.Fatalf("validateWorkoutGenerationToolInput() error = %v, want context field", err)
	}
}

func TestBuildWorkoutGenerationPromptIncludesFitTrackContract(t *testing.T) {
	prompt := buildWorkoutGenerationPrompt(WorkoutGenerationToolInput{}, time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC))

	requiredSnippets := []string{
		`"date": "RFC3339 timestamp"`,
		`"workoutFocus": "optional string"`,
		`"setType": "warmup" | "working"`,
		`"date" is always required and must be RFC3339.`,
		`If fitness level is unknown, prefer omitting weights instead of guessing aggressively.`,
		`Scale the draft to the requested session duration by estimating setup and transitions, set execution time, rest between sets, and warm-up or ramp-up needs when appropriate.`,
		`Do not satisfy a normal 40+ minute strength or hypertrophy request with a very small workout unless the user asked for minimal, beginner, rehab, warm-up, or low-volume work.`,
		`strength can use fewer exercises with longer rests and enough sets`,
		`endurance or circuit work should use shorter rests and higher density`,
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(prompt, snippet) {
			t.Fatalf("buildWorkoutGenerationPrompt() missing %q\nprompt=%s", snippet, prompt)
		}
	}
}

func TestBuildWorkoutGenerationUserPromptLabelsMissingFitnessLevelAsUnknown(t *testing.T) {
	prompt := buildWorkoutGenerationUserPrompt(WorkoutGenerationToolInput{
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "pull",
		Injuries:        "none",
	})

	if !strings.Contains(prompt, "- Fitness level: unknown") {
		t.Fatalf("buildWorkoutGenerationUserPrompt() = %q, want unknown fitness level", prompt)
	}
	if !strings.Contains(prompt, "- Goal: unknown") {
		t.Fatalf("buildWorkoutGenerationUserPrompt() = %q, want unknown goal", prompt)
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
