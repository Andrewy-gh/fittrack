package aichateval

import (
	"context"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
)

func TestFixtureChatDataReaderUsesFixtureUser(t *testing.T) {
	reader := NewFixtureChatDataReader()

	workouts, err := reader.ListWorkoutsWithSets(context.Background(), FixtureUserID, aichat.WorkoutHistoryFilter{
		ExerciseName: "Back Squat",
		LastN:        1,
	})
	if err != nil {
		t.Fatalf("ListWorkoutsWithSets() error = %v", err)
	}
	if len(workouts) != 1 {
		t.Fatalf("ListWorkoutsWithSets() returned %d workouts, want 1", len(workouts))
	}
	if workouts[0].Date != "2026-06-30" {
		t.Fatalf("fixture workout date = %q, want 2026-06-30", workouts[0].Date)
	}

	names, err := reader.ResolveExerciseNames(context.Background(), FixtureUserID, "squat")
	if err != nil {
		t.Fatalf("ResolveExerciseNames() error = %v", err)
	}
	if len(names) == 0 {
		t.Fatal("ResolveExerciseNames() returned no fixture squat matches")
	}

	snapshot, err := reader.TrainingSnapshot(context.Background(), FixtureUserID)
	if err != nil {
		t.Fatalf("TrainingSnapshot() error = %v", err)
	}
	if snapshot == nil || snapshot.LastWorkoutDate != "2026-07-03" {
		t.Fatalf("TrainingSnapshot() = %#v, want last workout 2026-07-03", snapshot)
	}
}

func TestFixtureChatDataReaderDoesNotReturnSeedDataForOtherUsers(t *testing.T) {
	reader := NewFixtureChatDataReader()

	workouts, err := reader.ListWorkoutsWithSets(context.Background(), "other-user", aichat.WorkoutHistoryFilter{})
	if err != nil {
		t.Fatalf("ListWorkoutsWithSets() error = %v", err)
	}
	if len(workouts) != 0 {
		t.Fatalf("ListWorkoutsWithSets() returned %d workouts for another user, want 0", len(workouts))
	}

	names, err := reader.ResolveExerciseNames(context.Background(), "other-user", "bench")
	if err != nil {
		t.Fatalf("ResolveExerciseNames() error = %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("ResolveExerciseNames() returned %d names for another user, want 0", len(names))
	}

	snapshot, err := reader.TrainingSnapshot(context.Background(), "other-user")
	if err != nil {
		t.Fatalf("TrainingSnapshot() error = %v", err)
	}
	if snapshot != nil {
		t.Fatalf("TrainingSnapshot() = %#v for another user, want nil", snapshot)
	}
}
