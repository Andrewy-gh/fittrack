package aichat

import (
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestMapChatWorkoutRowsRegroupsFlatRows(t *testing.T) {
	rows := []db.ListWorkoutsWithSetsForChatRow{
		chatWorkoutRow(2, "2026-07-03", "upper", "Bench Press", 1, 1, numericWeight(t, "185"), 5, "working"),
		chatWorkoutRow(2, "2026-07-03", "upper", "Bench Press", 1, 2, numericWeight(t, "195"), 3, "working"),
		chatWorkoutRow(2, "2026-07-03", "upper", "Barbell Row", 2, 1, numericWeight(t, "155.5"), 8, "working"),
		chatWorkoutRow(1, "2026-06-30", "lower", "Back Squat", 1, 1, pgtype.Numeric{}, 10, "warmup"),
	}

	workouts, err := mapChatWorkoutRows(rows)
	if err != nil {
		t.Fatalf("mapChatWorkoutRows() error = %v", err)
	}
	if len(workouts) != 2 {
		t.Fatalf("mapChatWorkoutRows() returned %d workouts, want 2", len(workouts))
	}
	if workouts[0].Date != "2026-07-03" || workouts[0].Focus != "upper" {
		t.Fatalf("first workout = %#v", workouts[0])
	}
	if len(workouts[0].Exercises) != 2 {
		t.Fatalf("first workout exercises = %#v, want 2 exercises", workouts[0].Exercises)
	}
	if got := workouts[0].Exercises[0].Sets; len(got) != 2 || got[0] != "185x5 working" || got[1] != "195x3 working" {
		t.Fatalf("bench sets = %#v", got)
	}
	if got := workouts[0].Exercises[1].Sets; len(got) != 1 || got[0] != "155.5x8 working" {
		t.Fatalf("row sets = %#v", got)
	}
	if got := workouts[1].Exercises[0].Sets; len(got) != 1 || got[0] != "10 reps warmup" {
		t.Fatalf("bodyweight sets = %#v", got)
	}
}

func TestNormalizeWorkoutHistoryFilterCapsLastN(t *testing.T) {
	if got := normalizeWorkoutHistoryFilter(WorkoutHistoryFilter{}).LastN; got != defaultChatWorkoutLimit {
		t.Fatalf("default LastN = %d, want %d", got, defaultChatWorkoutLimit)
	}
	if got := normalizeWorkoutHistoryFilter(WorkoutHistoryFilter{LastN: 99}).LastN; got != maxChatWorkoutLimit {
		t.Fatalf("capped LastN = %d, want %d", got, maxChatWorkoutLimit)
	}
	if got := normalizeWorkoutHistoryFilter(WorkoutHistoryFilter{LastN: 12}).LastN; got != 12 {
		t.Fatalf("LastN = %d, want 12", got)
	}
}

func chatWorkoutRow(workoutID int32, date string, focus string, exercise string, exerciseOrder int32, setOrder int32, weight pgtype.Numeric, reps int32, setType string) db.ListWorkoutsWithSetsForChatRow {
	parsed, _ := time.Parse("2006-01-02", date)
	return db.ListWorkoutsWithSetsForChatRow{
		WorkoutID:     workoutID,
		Date:          pgtype.Timestamptz{Time: parsed, Valid: true},
		WorkoutFocus:  pgtype.Text{String: focus, Valid: true},
		ExerciseName:  pgtype.Text{String: exercise, Valid: true},
		ExerciseOrder: pgtype.Int4{Int32: exerciseOrder, Valid: true},
		SetOrder:      pgtype.Int4{Int32: setOrder, Valid: true},
		Weight:        weight,
		Reps:          pgtype.Int4{Int32: reps, Valid: true},
		SetType:       pgtype.Text{String: setType, Valid: true},
	}
}

func numericWeight(t *testing.T, value string) pgtype.Numeric {
	t.Helper()
	var numeric pgtype.Numeric
	if err := numeric.Scan(value); err != nil {
		t.Fatalf("scan numeric weight: %v", err)
	}
	return numeric
}
