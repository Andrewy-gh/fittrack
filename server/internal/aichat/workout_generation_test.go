package aichat

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

func TestBuildChatSystemPromptIncludesWorkoutGuardrails(t *testing.T) {
	prompt := buildChatSystemPrompt()

	requiredSnippets := []string{
		"Ask at most 3 short, focused follow-up questions",
		"MVP-ready inputs are workout focus, session duration, enough equipment or workout context",
		"do not treat it as a hard blocker",
		"Equipment is optional for mobility, rehab, prehab, stretching, or warm-up requests",
		"Resistance bands, foam rollers, sticks, and similar tools",
		"If injury status is missing, ask once before generating",
		"Do not infer \"none\" from silence in the initial request",
		"Use injuries=\"none\" only when the user explicitly says they have no injuries",
		"When the user answers a follow-up, combine that answer with the earlier visible workout request",
		"previous message only asked about injuries and the user now confirms no injuries",
		"does not mention injuries, ask about injuries before calling the " + workoutDraftToolName + " tool",
		"Do not ask scheduling, frequency, or future-date questions",
		"call the " + workoutDraftToolName + " tool immediately",
		"Do not list specific exercises, sets, or reps in plain text before",
		"only way to produce a structured workout draft",
		"If the user adds a new pain, injury, or movement limitation after a workout draft",
		"ask one focused follow-up about the painful movements, ranges, or triggers",
		"swap anything that bothers my knee/elbow/shoulder/back/wrist",
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
		"Equipment is optional for mobility, rehab, prehab, stretching, or warm-up requests",
		"injury status",
		"Use injuries=none only when the user explicitly reports no injuries",
		"continues without answering after one injury-status question",
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

func TestValidateWorkoutGenerationToolInputAllowsMobilityWithoutEquipment(t *testing.T) {
	input := WorkoutGenerationToolInput{
		SessionDuration: 15,
		WorkoutFocus:    "mobility and rehab for low-back stiffness",
		Injuries:        "No acute injury or numbness. Gentle low-back stiffness.",
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

func TestGenerateWorkoutDraftWithRepairsQualityFailureOnce(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "hypertrophy",
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "pull",
		Injuries:        "none",
	}
	firstDraft := validDraftWithExercises(
		draftExercise("Lat Pulldown", workingSet(10), workingSet(10)),
	)
	repairedDraft := validDraftWithExercises(
		draftExercise("Pull-Up", warmupSet(6), warmupSet(6), workingSet(8), workingSet(8), workingSet(8)),
		draftExercise("Chest Supported Row", workingSet(10), workingSet(10), workingSet(10)),
		draftExercise("Seated Cable Row", workingSet(12), workingSet(12)),
		draftExercise("Incline Dumbbell Curl", workingSet(12), workingSet(12)),
	)
	prompts := make([]string, 0, 2)
	drafts := []*workout.CreateWorkoutRequest{firstDraft, repairedDraft}
	generator := func(_ context.Context, _ *genkit.Genkit, _ string, _ string, userPrompt string) (*workout.CreateWorkoutRequest, error) {
		prompts = append(prompts, userPrompt)
		return drafts[len(prompts)-1], nil
	}

	draft, err := generateWorkoutDraftWith(
		context.Background(),
		nil,
		"test-model",
		input,
		time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC),
		generator,
	)
	if err != nil {
		t.Fatalf("generateWorkoutDraftWith() error = %v, want nil", err)
	}
	if draft != repairedDraft {
		t.Fatalf("generateWorkoutDraftWith() draft = %#v, want repaired draft", draft)
	}
	if len(prompts) != 2 {
		t.Fatalf("generator called %d times, want 2", len(prompts))
	}
	if !strings.Contains(prompts[1], "failed deterministic quality validation") {
		t.Fatalf("repair prompt = %q, want quality validation context", prompts[1])
	}
	if !strings.Contains(prompts[1], "expected at least") {
		t.Fatalf("repair prompt = %q, want deterministic validator feedback", prompts[1])
	}
	if !strings.Contains(prompts[1], "- Session duration: 45 minutes") || !strings.Contains(prompts[1], "- Workout focus: pull") {
		t.Fatalf("repair prompt = %q, want original request context", prompts[1])
	}
}

func TestGenerateWorkoutDraftWithDoesNotRetryGenerationErrors(t *testing.T) {
	expectedErr := errors.New("model unavailable")
	calls := 0
	generator := func(_ context.Context, _ *genkit.Genkit, _ string, _ string, _ string) (*workout.CreateWorkoutRequest, error) {
		calls++
		return nil, expectedErr
	}

	draft, err := generateWorkoutDraftWith(
		context.Background(),
		nil,
		"test-model",
		WorkoutGenerationToolInput{Equipment: "full gym", SessionDuration: 45, WorkoutFocus: "pull", Injuries: "none"},
		time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC),
		generator,
	)
	if err == nil {
		t.Fatalf("generateWorkoutDraftWith() draft = %#v, error = nil, want generation error", draft)
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("generateWorkoutDraftWith() error = %v, want %v", err, expectedErr)
	}
	if calls != 1 {
		t.Fatalf("generator called %d times, want 1", calls)
	}
}

func TestGenerateWorkoutDraftWithDoesNotRetrySchemaValidationErrors(t *testing.T) {
	calls := 0
	generator := func(_ context.Context, _ *genkit.Genkit, _ string, _ string, _ string) (*workout.CreateWorkoutRequest, error) {
		calls++
		return &workout.CreateWorkoutRequest{
			Date: "2026-04-20T12:00:00Z",
		}, nil
	}

	draft, err := generateWorkoutDraftWith(
		context.Background(),
		nil,
		"test-model",
		WorkoutGenerationToolInput{Equipment: "full gym", SessionDuration: 45, WorkoutFocus: "pull", Injuries: "none"},
		time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC),
		generator,
	)
	if err == nil {
		t.Fatalf("generateWorkoutDraftWith() draft = %#v, error = nil, want validation error", draft)
	}
	if !strings.Contains(err.Error(), "validate workout draft") {
		t.Fatalf("generateWorkoutDraftWith() error = %v, want schema validation context", err)
	}
	if calls != 1 {
		t.Fatalf("generator called %d times, want 1", calls)
	}
}

func TestGenerateWorkoutDraftWithStopsAfterOneQualityRepairRetry(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "hypertrophy",
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "pull",
		Injuries:        "none",
	}
	calls := 0
	generator := func(_ context.Context, _ *genkit.Genkit, _ string, _ string, _ string) (*workout.CreateWorkoutRequest, error) {
		calls++
		return validDraftWithExercises(
			draftExercise("Lat Pulldown", workingSet(10), workingSet(10)),
		), nil
	}

	draft, err := generateWorkoutDraftWith(
		context.Background(),
		nil,
		"test-model",
		input,
		time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC),
		generator,
	)
	if err == nil {
		t.Fatalf("generateWorkoutDraftWith() draft = %#v, error = nil, want quality retry error", draft)
	}
	if !strings.Contains(err.Error(), "after repair retry") {
		t.Fatalf("generateWorkoutDraftWith() error = %v, want repair retry context", err)
	}
	if calls != 2 {
		t.Fatalf("generator called %d times, want 2", calls)
	}
}

func TestGenerateWorkoutDraftWithValidDraftPassesDirectly(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "hypertrophy",
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "pull",
		Injuries:        "none",
	}
	expectedDraft := validDraftWithExercises(
		draftExercise("Pull-Up", warmupSet(6), warmupSet(6), workingSet(8), workingSet(8), workingSet(8)),
		draftExercise("Chest Supported Row", workingSet(10), workingSet(10), workingSet(10)),
		draftExercise("Seated Cable Row", workingSet(12), workingSet(12)),
		draftExercise("Incline Dumbbell Curl", workingSet(12), workingSet(12)),
	)
	calls := 0
	generator := func(_ context.Context, _ *genkit.Genkit, _ string, _ string, _ string) (*workout.CreateWorkoutRequest, error) {
		calls++
		return expectedDraft, nil
	}

	draft, err := generateWorkoutDraftWith(
		context.Background(),
		nil,
		"test-model",
		input,
		time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC),
		generator,
	)
	if err != nil {
		t.Fatalf("generateWorkoutDraftWith() error = %v, want nil", err)
	}
	if draft != expectedDraft {
		t.Fatalf("generateWorkoutDraftWith() draft = %#v, want expected draft", draft)
	}
	if calls != 1 {
		t.Fatalf("generator called %d times, want 1", calls)
	}
}

func TestBuildWorkoutGenerationRepairPromptUsesDeterministicFeedback(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "hypertrophy",
		Equipment:       "bodyweight only",
		SessionDuration: 45,
		WorkoutFocus:    "upper body",
		Injuries:        "shoulder pain",
	}
	draft := validDraftWithExercises(
		draftExercise("Barbell Overhead Press", workingSet(10), workingSet(10)),
	)
	qualityErr := validateWorkoutDraftQuality(input, draft)
	if qualityErr == nil {
		t.Fatal("validateWorkoutDraftQuality() error = nil, want quality issue")
	}

	prompt := buildWorkoutGenerationRepairPrompt(input, qualityErr)
	for _, issue := range strings.Split(qualityErr.Error(), "; ") {
		if !strings.Contains(prompt, issue) {
			t.Fatalf("repair prompt = %q, want deterministic issue %q", prompt, issue)
		}
	}
	if !strings.Contains(prompt, "- Equipment: bodyweight only") || !strings.Contains(prompt, "- Injuries or limitations: shoulder pain") {
		t.Fatalf("repair prompt = %q, want original constraints", prompt)
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
