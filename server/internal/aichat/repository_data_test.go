package aichat

import (
	"fmt"
	"reflect"
	"strings"
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

func TestFormatChatWorkoutDateUsesUTCCalendarDay(t *testing.T) {
	utcMinusFour := time.FixedZone("UTC-4", -4*60*60)
	scannedLocalTime := time.Date(2026, 6, 30, 20, 0, 0, 0, utcMinusFour)

	if got := formatChatWorkoutDate(scannedLocalTime); got != "2026-07-01" {
		t.Fatalf("formatChatWorkoutDate() = %q, want 2026-07-01", got)
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

func TestNormalizeExerciseStatsWindow(t *testing.T) {
	tests := map[string]string{
		"":             "3m",
		"three months": "3m",
		"1 year":       "1y",
		"ALL-TIME":     "all",
		"nonsense":     "3m",
	}
	for input, want := range tests {
		if got := normalizeExerciseStatsWindow(input); got != want {
			t.Fatalf("normalizeExerciseStatsWindow(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCompactExerciseStatsTrendKeepsFirstAndLast(t *testing.T) {
	var points []exerciseStatsTrendRow
	for i := 0; i < 12; i++ {
		points = append(points, exerciseStatsTrendRow{
			WorkoutID:       int32(i + 1),
			WorkoutDay:      time.Date(2026, 1, i+1, 0, 0, 0, 0, time.UTC),
			SessionBestE1RM: float64(i + 100),
		})
	}

	compact := compactExerciseStatsTrend(points, 8)

	if len(compact) != 8 {
		t.Fatalf("compact len = %d, want 8", len(compact))
	}
	if compact[0].Date != "2026-01-01" || compact[len(compact)-1].Date != "2026-01-12" {
		t.Fatalf("compact endpoints = %#v ... %#v", compact[0], compact[len(compact)-1])
	}
}

func TestCompactExerciseStatsTrendMaxPointsOneKeepsLast(t *testing.T) {
	points := []exerciseStatsTrendRow{
		{
			WorkoutID:       1,
			WorkoutDay:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			SessionBestE1RM: 100,
		},
		{
			WorkoutID:       2,
			WorkoutDay:      time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC),
			SessionBestE1RM: 120,
		},
	}

	compact := compactExerciseStatsTrend(points, 1)

	if len(compact) != 1 {
		t.Fatalf("compact len = %d, want 1", len(compact))
	}
	if compact[0].WorkoutID != 2 || compact[0].Date != "2026-01-12" {
		t.Fatalf("compact[0] = %#v, want last point", compact[0])
	}
}

func TestFilterLastThreeMonthsUsesNow(t *testing.T) {
	now := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	points := []exerciseStatsTrendRow{
		{WorkoutDay: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{WorkoutDay: time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC)},
		{WorkoutDay: time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC)},
		{WorkoutDay: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
	}

	filtered := filterLastThreeMonths(points, now)

	if len(filtered) != 2 || !filtered[0].WorkoutDay.Equal(points[2].WorkoutDay) || !filtered[1].WorkoutDay.Equal(points[3].WorkoutDay) {
		t.Fatalf("filtered = %#v", filtered)
	}
}

func TestDecodeProfileStringArrayHandlesNull(t *testing.T) {
	values, err := decodeProfileStringArray(nil)
	if err != nil {
		t.Fatalf("decodeProfileStringArray() error = %v", err)
	}
	if values != nil {
		t.Fatalf("decodeProfileStringArray(nil) = %#v, want nil", values)
	}
}

func TestDecodeProfileStringArrayCleansValues(t *testing.T) {
	values, err := decodeProfileStringArray([]byte(`[" dumbbells ","", "bench", "bench"]`))
	if err != nil {
		t.Fatalf("decodeProfileStringArray() error = %v", err)
	}
	if !reflect.DeepEqual(values, []string{"dumbbells", "bench"}) {
		t.Fatalf("decodeProfileStringArray() = %#v", values)
	}
}

func TestCleanProfileStringListCapsAndDeduplicates(t *testing.T) {
	values := []string{" dumbbells ", "DUMBBELLS", "", strings.Repeat("x", 140)}
	for i := 0; i < 25; i++ {
		values = append(values, fmt.Sprintf("item-%02d", i))
	}

	got := cleanProfileStringList(values)

	if len(got) != 20 {
		t.Fatalf("cleanProfileStringList() len = %d, want 20", len(got))
	}
	if got[0] != "dumbbells" {
		t.Fatalf("cleanProfileStringList()[0] = %q, want dumbbells", got[0])
	}
	if len(got[1]) != 120 {
		t.Fatalf("cleanProfileStringList()[1] len = %d, want 120", len(got[1]))
	}
}

func TestHasTrainingProfileContent(t *testing.T) {
	if hasTrainingProfileContent(nil) {
		t.Fatal("nil profile should not have content")
	}
	if hasTrainingProfileContent(&TrainingProfile{}) {
		t.Fatal("empty profile should not have content")
	}
	if !hasTrainingProfileContent(&TrainingProfile{AvailableEquipment: []string{"dumbbells"}}) {
		t.Fatal("profile with equipment should have content")
	}
	if !hasTrainingProfileContent(&TrainingProfile{MovementLimitationsRecorded: true}) {
		t.Fatal("profile with explicitly recorded no movement limitations should have content")
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
