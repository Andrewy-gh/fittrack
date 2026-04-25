package aichat

import (
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

func TestValidateWorkoutDraftQualityRejectsUnderScopedHypertrophyDraft(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "hypertrophy",
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "pull",
		Injuries:        "none",
	}
	draft := validDraftWithExercises(
		draftExercise("Lat Pulldown", workingSet(10), workingSet(10)),
		draftExercise("Cable Row", workingSet(12), workingSet(12)),
	)

	err := validateWorkoutDraftQuality(input, draft)
	if err == nil {
		t.Fatal("validateWorkoutDraftQuality() error = nil, want under-scoped draft error")
	}
	if !strings.Contains(err.Error(), "working sets") {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want working set issue", err)
	}
}

func TestValidateWorkoutDraftQualityScalesMinimumWorkToRequestedDuration(t *testing.T) {
	shortInput := WorkoutGenerationToolInput{
		FitnessGoal:     "hypertrophy",
		Equipment:       "dumbbells",
		SessionDuration: 25,
		WorkoutFocus:    "upper body",
		Injuries:        "none",
	}
	longInput := shortInput
	longInput.SessionDuration = 60

	if got := minimumWorkingSets(shortInput); got != 5 {
		t.Fatalf("minimumWorkingSets(shortInput) = %d, want 5", got)
	}
	if got := minimumWorkingSets(longInput); got != 12 {
		t.Fatalf("minimumWorkingSets(longInput) = %d, want 12", got)
	}
}

func TestValidateWorkoutDraftQualityAcceptsReasonableHypertrophyDraft(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "hypertrophy",
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "pull",
		Injuries:        "none",
	}
	draft := validDraftWithExercises(
		draftExercise("Pull-Up", warmupSet(6), warmupSet(6), workingSet(8), workingSet(8), workingSet(8)),
		draftExercise("Chest Supported Row", workingSet(10), workingSet(10), workingSet(10)),
		draftExercise("Seated Cable Row", workingSet(12), workingSet(12)),
		draftExercise("Incline Dumbbell Curl", workingSet(12), workingSet(12)),
	)

	if err := validateWorkoutDraftQuality(input, draft); err != nil {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want nil", err)
	}
}

func TestValidateWorkoutDraftQualityRejectsUnavailableEquipment(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:      "general fitness",
		Equipment:        "bodyweight only",
		SessionDuration:  30,
		WorkoutFocus:     "full body",
		SpaceConstraints: "home",
		Injuries:         "none",
	}
	draft := validDraftWithExercises(
		draftExercise("Barbell Back Squat", workingSet(8), workingSet(8), workingSet(8)),
	)

	err := validateWorkoutDraftQuality(input, draft)
	if err == nil {
		t.Fatal("validateWorkoutDraftQuality() error = nil, want unavailable equipment error")
	}
	if !strings.Contains(err.Error(), "barbell") {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want barbell issue", err)
	}
}

func TestValidateWorkoutDraftQualityRejectsExerciseConflictingWithInjury(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "strength",
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "legs",
		Injuries:        "knee pain",
	}
	draft := validDraftWithExercises(
		draftExercise("Back Squat", warmupSet(5), workingSet(5), workingSet(5), workingSet(5)),
		draftExercise("Leg Press", workingSet(8), workingSet(8), workingSet(8)),
		draftExercise("Hamstring Curl", workingSet(10), workingSet(10)),
	)

	err := validateWorkoutDraftQuality(input, draft)
	if err == nil {
		t.Fatal("validateWorkoutDraftQuality() error = nil, want injury conflict error")
	}
	if !strings.Contains(err.Error(), "injury") {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want injury issue", err)
	}
}

func TestValidateWorkoutDraftQualityAllowsExplicitBeginnerLowerVolume(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessLevel:    "beginner",
		FitnessGoal:     "strength",
		Equipment:       "dumbbells, bench",
		SessionDuration: 45,
		WorkoutFocus:    "beginner upper body",
		Injuries:        "none",
	}
	draft := validDraftWithExercises(
		draftExercise("Dumbbell Bench Press", workingSet(8), workingSet(8)),
		draftExercise("One-Arm Dumbbell Row", workingSet(10), workingSet(10)),
	)

	if err := validateWorkoutDraftQuality(input, draft); err != nil {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want nil for explicit beginner request", err)
	}
}

func TestValidateWorkoutDraftQualityRejectsBenchExerciseForBodyweightRequest(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "general fitness",
		Equipment:       "bodyweight",
		SessionDuration: 30,
		WorkoutFocus:    "upper body",
		Injuries:        "none",
	}
	draft := validDraftWithExercises(
		draftExercise("Bench Dip", workingSet(10), workingSet(10), workingSet(10)),
	)

	err := validateWorkoutDraftQuality(input, draft)
	if err == nil {
		t.Fatal("validateWorkoutDraftQuality() error = nil, want unavailable bench error")
	}
	if !strings.Contains(err.Error(), "bench") {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want bench issue", err)
	}
}

func TestValidateWorkoutDraftQualityAllowsBenchExerciseWhenBenchIsAvailable(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "general fitness",
		Equipment:       "bodyweight, bench",
		SessionDuration: 30,
		WorkoutFocus:    "upper body",
		Injuries:        "none",
	}
	draft := validDraftWithExercises(
		draftExercise("Bench Dip", workingSet(10), workingSet(10), workingSet(10)),
	)

	if err := validateWorkoutDraftQuality(input, draft); err != nil {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want nil when bench is available", err)
	}
}

func TestValidateWorkoutDraftQualityRejectsDumbbellBenchPressWithoutBench(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "general fitness",
		Equipment:       "dumbbells, no bench",
		SessionDuration: 30,
		WorkoutFocus:    "upper body",
		Injuries:        "none",
	}
	draft := validDraftWithExercises(
		draftExercise("Dumbbell Bench Press", workingSet(10), workingSet(10), workingSet(10)),
	)

	err := validateWorkoutDraftQuality(input, draft)
	if err == nil {
		t.Fatal("validateWorkoutDraftQuality() error = nil, want unavailable bench error")
	}
	if !strings.Contains(err.Error(), "bench") {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want bench issue", err)
	}
}

func TestValidateWorkoutDraftQualityAllowsDumbbellBenchPressWithDumbbellsAndBench(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "general fitness",
		Equipment:       "bodyweight, bench, dumbbells",
		SessionDuration: 30,
		WorkoutFocus:    "upper body",
		Injuries:        "none",
	}
	draft := validDraftWithExercises(
		draftExercise("Dumbbell Bench Press", workingSet(10), workingSet(10), workingSet(10)),
	)

	if err := validateWorkoutDraftQuality(input, draft); err != nil {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want nil when dumbbells and bench are available", err)
	}
}

func TestValidateWorkoutDraftQualityTreatsNoKneePainAsNoActiveInjury(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "strength",
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "legs",
		Injuries:        "no knee pain",
	}
	draft := validDraftWithExercises(
		draftExercise("Back Squat", workingSet(5), workingSet(5), workingSet(5), workingSet(5), workingSet(5), workingSet(5), workingSet(5), workingSet(5)),
	)

	if err := validateWorkoutDraftQuality(input, draft); err != nil {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want nil for negated knee pain", err)
	}
}

func TestValidateWorkoutDraftQualityTreatsHistoryOfKneePainAsActiveConstraint(t *testing.T) {
	input := WorkoutGenerationToolInput{
		FitnessGoal:     "strength",
		Equipment:       "full gym",
		SessionDuration: 45,
		WorkoutFocus:    "legs",
		Injuries:        "history of knee pain",
	}
	draft := validDraftWithExercises(
		draftExercise("Back Squat", workingSet(5), workingSet(5), workingSet(5), workingSet(5), workingSet(5), workingSet(5), workingSet(5), workingSet(5)),
	)

	err := validateWorkoutDraftQuality(input, draft)
	if err == nil {
		t.Fatal("validateWorkoutDraftQuality() error = nil, want knee history constraint")
	}
	if !strings.Contains(err.Error(), "injury") {
		t.Fatalf("validateWorkoutDraftQuality() error = %v, want injury issue", err)
	}
}

func validDraftWithExercises(exercises ...workout.ExerciseInput) *workout.CreateWorkoutRequest {
	focus := "test"
	return &workout.CreateWorkoutRequest{
		Date:         "2026-04-20T12:00:00Z",
		WorkoutFocus: &focus,
		Exercises:    exercises,
	}
}

func draftExercise(name string, sets ...workout.SetInput) workout.ExerciseInput {
	return workout.ExerciseInput{Name: name, Sets: sets}
}

func warmupSet(reps int) workout.SetInput {
	return workout.SetInput{Reps: reps, SetType: "warmup"}
}

func workingSet(reps int) workout.SetInput {
	return workout.SetInput{Reps: reps, SetType: "working"}
}
