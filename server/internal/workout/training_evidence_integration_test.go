package workout

import (
	"context"
	"os"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvidenceServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("requires an explicitly configured isolated test database")
	}

	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	queries := db.New(pool)
	ownerID := "test-user-a"
	otherID := "test-user-b"

	t.Run("uses actual SQL bounds, preserves empty workouts, and excludes foreign joins", func(t *testing.T) {
		ownerDBCtx := setTestUserContext(context.Background(), t, pool, ownerID)
		ownerCtx := user.WithContext(ownerDBCtx, ownerID)
		otherDBCtx := setTestUserContext(context.Background(), t, pool, otherID)
		otherExerciseID := insertEvidenceExercise(t, pool, otherDBCtx, otherID, "Other user exercise", stringPointer("Hammer_Curls"))
		ownerDBCtx = setTestUserContext(context.Background(), t, pool, ownerID)
		ownerCtx = user.WithContext(ownerDBCtx, ownerID)
		ownerExerciseID := insertEvidenceExercise(t, pool, ownerDBCtx, ownerID, "Owner pushups", stringPointer("Pushups"))

		startWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.January, 4, 0, 0, 0, 0, time.UTC))
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, startWorkoutID, ownerExerciseID, nil, 5, "working", 1, 1)
		// The set belongs to the owner workout but points at another user's exercise. It is legal
		// in the old schema and must not be counted or exposed by the evidence query.
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, startWorkoutID, otherExerciseID, 10.0, 5, "working", 2, 1)

		emptyWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.January, 5, 9, 0, 0, 0, time.UTC))
		require.NotZero(t, emptyWorkoutID)
		warmupWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.January, 6, 9, 0, 0, 0, time.UTC))
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, warmupWorkoutID, ownerExerciseID, 20.0, 8, "warmup", 1, 1)
		unexpectedTypeWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.January, 7, 9, 0, 0, 0, time.UTC))
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, unexpectedTypeWorkoutID, ownerExerciseID, 20.0, 8, "Working", 1, 1)
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, unexpectedTypeWorkoutID, ownerExerciseID, 20.0, 8, "working ", 1, 2)
		nowWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC))
		// Exact lowercase working rows count even when stored weight/reps bypass request validation.
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, nowWorkoutID, ownerExerciseID, 0.0, 0, "working", 1, 1)
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, nowWorkoutID, ownerExerciseID, nil, -2, "working", 1, 2)

		endWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.January, 11, 0, 0, 0, 0, time.UTC))
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, endWorkoutID, ownerExerciseID, 20.0, 8, "working", 1, 1)
		futureWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.January, 10, 12, 0, 0, 1000, time.UTC))
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, futureWorkoutID, ownerExerciseID, 20.0, 8, "working", 1, 1)
		beforeStartWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.January, 3, 23, 59, 59, 0, time.UTC))
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, beforeStartWorkoutID, ownerExerciseID, 20.0, 8, "working", 1, 1)

		observedAt := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
		service := NewEvidenceService(queries, fixedEvidenceClock(observedAt))
		actual, err := service.Summarize(ownerCtx, TrainingEvidenceRequest{StartDate: "2026-01-04", Timezone: "UTC"})
		require.NoError(t, err)
		assert.True(t, actual.Period.Partial)
		assert.Equal(t, EvidenceActivity{LoggedWorkoutCount: 5, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 2}, actual.Activity)
		assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 3, ClassifiedWorkingSetCount: 3, CoverageStatus: EvidenceCoverageComplete}, actual.Classification)
		assert.Equal(t, []MuscleEvidence{
			{Muscle: "chest", Role: "primary", WorkingSetCount: 3, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 2},
			{Muscle: "shoulders", Role: "secondary", WorkingSetCount: 3, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 2},
			{Muscle: "triceps", Role: "secondary", WorkingSetCount: 3, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 2},
		}, actual.Muscles)

		rawRows, err := queries.ListTrainingEvidence(ownerDBCtx, db.ListTrainingEvidenceParams{
			UserID:     ownerID,
			StartAt:    pgtype.Timestamptz{Time: actual.Period.StartAt, Valid: true},
			EndAt:      pgtype.Timestamptz{Time: actual.Period.EndAt, Valid: true},
			ObservedAt: pgtype.Timestamptz{Time: observedAt, Valid: true},
		})
		require.NoError(t, err)
		require.Len(t, rawRows, 5, "the outer joins retain empty and non-working workouts")
		workingSetsByWorkout := make(map[int32]int32, len(rawRows))
		for _, row := range rawRows {
			assert.NotEqual(t, otherExerciseID, row.ExerciseID.Int32, "foreign exercise IDs must not leak through the join")
			workingSetsByWorkout[row.WorkoutID] += row.WorkingSetCount
		}
		assert.Equal(t, int32(0), workingSetsByWorkout[unexpectedTypeWorkoutID], "Working and working-space are not exact working rows")
		assert.Equal(t, int32(2), workingSetsByWorkout[nowWorkoutID], "zero and negative reps do not suppress exact working rows")
		assert.NotContains(t, workingSetsByWorkout, futureWorkoutID, "a workout one microsecond after observedAt is excluded")

		otherDBCtx = setTestUserContext(context.Background(), t, pool, otherID)
		otherEvidence, err := service.Summarize(user.WithContext(otherDBCtx, otherID), TrainingEvidenceRequest{StartDate: "2026-01-04", Timezone: "UTC"})
		require.NoError(t, err)
		assert.Equal(t, EvidenceActivity{}, otherEvidence.Activity)
		assert.Equal(t, EvidenceClassification{CoverageStatus: EvidenceCoverageNotApplicable}, otherEvidence.Classification)
	})

	t.Run("retains workouts with only malformed cross-owner sets", func(t *testing.T) {
		ownerDBCtx := setTestUserContext(context.Background(), t, pool, ownerID)
		ownerCtx := user.WithContext(ownerDBCtx, ownerID)
		otherDBCtx := setTestUserContext(context.Background(), t, pool, otherID)
		ownerExerciseID := insertEvidenceExercise(t, pool, ownerDBCtx, ownerID, "Owner isolation exercise", stringPointer("Pushups"))
		otherExerciseID := insertEvidenceExercise(t, pool, otherDBCtx, otherID, "Other isolation exercise", stringPointer("Hammer_Curls"))

		malformedWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
		// An A-owned set can point at B's exercise in this schema; the query must exclude it.
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, malformedWorkoutID, otherExerciseID, 10.0, 5, "working", 1, 1)
		// The migration-owner test fixture can create the inverse malformed row without changing
		// production RLS: B's set points at A's exercise and A's workout.
		migrationOwnerCtx := context.Background()
		insertEvidenceSet(t, pool, migrationOwnerCtx, otherID, malformedWorkoutID, ownerExerciseID, 10.0, 5, "working", 2, 1)

		otherWorkoutID := insertEvidenceWorkout(t, pool, otherDBCtx, otherID, time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC))
		insertEvidenceSet(t, pool, otherDBCtx, otherID, otherWorkoutID, otherExerciseID, 10.0, 5, "working", 1, 1)

		observedAt := time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
		service := NewEvidenceService(queries, fixedEvidenceClock(observedAt))
		request := TrainingEvidenceRequest{StartDate: "2026-04-01", Timezone: "UTC"}

		ownerEvidence, err := service.Summarize(ownerCtx, request)
		require.NoError(t, err)
		assert.Equal(t, EvidenceActivity{LoggedWorkoutCount: 1}, ownerEvidence.Activity)
		assert.Equal(t, EvidenceClassification{CoverageStatus: EvidenceCoverageNotApplicable}, ownerEvidence.Classification)
		assert.Empty(t, ownerEvidence.Muscles)

		rawOwnerRows, err := queries.ListTrainingEvidence(ownerDBCtx, db.ListTrainingEvidenceParams{
			UserID:     ownerID,
			StartAt:    pgtype.Timestamptz{Time: ownerEvidence.Period.StartAt, Valid: true},
			EndAt:      pgtype.Timestamptz{Time: ownerEvidence.Period.EndAt, Valid: true},
			ObservedAt: pgtype.Timestamptz{Time: observedAt, Valid: true},
		})
		require.NoError(t, err)
		require.Len(t, rawOwnerRows, 1)
		assert.False(t, rawOwnerRows[0].ExerciseID.Valid)
		assert.Zero(t, rawOwnerRows[0].WorkingSetCount)

		otherEvidence, err := service.Summarize(user.WithContext(otherDBCtx, otherID), request)
		require.NoError(t, err)
		assert.Equal(t, EvidenceActivity{LoggedWorkoutCount: 1, WorkingSetSessionCount: 1, WorkingSetLocalDayCount: 1}, otherEvidence.Activity)
		assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 1, ClassifiedWorkingSetCount: 1, CoverageStatus: EvidenceCoverageComplete}, otherEvidence.Classification)
	})

	t.Run("uses actual DST bounds and excludes the exact local end", func(t *testing.T) {
		ownerDBCtx := setTestUserContext(context.Background(), t, pool, ownerID)
		ownerCtx := user.WithContext(ownerDBCtx, ownerID)
		exerciseID := insertEvidenceExercise(t, pool, ownerDBCtx, ownerID, "DST boundary exercise", stringPointer("Pushups"))

		startWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.March, 8, 5, 0, 0, 0, time.UTC))
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, startWorkoutID, exerciseID, 10.0, 5, "working", 1, 1)
		beforeEndWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.March, 15, 3, 59, 59, 999999000, time.UTC))
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, beforeEndWorkoutID, exerciseID, 10.0, 5, "working", 1, 1)
		atEndWorkoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.March, 15, 4, 0, 0, 0, time.UTC))
		insertEvidenceSet(t, pool, ownerDBCtx, ownerID, atEndWorkoutID, exerciseID, 10.0, 5, "working", 1, 1)

		observedAt := time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)
		service := NewEvidenceService(queries, fixedEvidenceClock(observedAt))
		evidence, err := service.Summarize(ownerCtx, TrainingEvidenceRequest{StartDate: "2026-03-08", Timezone: "America/New_York"})
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, time.March, 8, 5, 0, 0, 0, time.UTC), evidence.Period.StartAt)
		assert.Equal(t, time.Date(2026, time.March, 15, 4, 0, 0, 0, time.UTC), evidence.Period.EndAt)
		assert.False(t, evidence.Period.Partial)
		assert.Equal(t, EvidenceActivity{LoggedWorkoutCount: 2, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 2}, evidence.Activity)
		assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 2, ClassifiedWorkingSetCount: 2, CoverageStatus: EvidenceCoverageComplete}, evidence.Classification)

		rawRows, err := queries.ListTrainingEvidence(ownerDBCtx, db.ListTrainingEvidenceParams{
			UserID:     ownerID,
			StartAt:    pgtype.Timestamptz{Time: evidence.Period.StartAt, Valid: true},
			EndAt:      pgtype.Timestamptz{Time: evidence.Period.EndAt, Valid: true},
			ObservedAt: pgtype.Timestamptz{Time: observedAt, Valid: true},
		})
		require.NoError(t, err)
		require.Len(t, rawRows, 2)
		includedWorkoutIDs := make(map[int32]bool, len(rawRows))
		for _, row := range rawRows {
			includedWorkoutIDs[row.WorkoutID] = true
		}
		assert.Contains(t, includedWorkoutIDs, startWorkoutID)
		assert.Contains(t, includedWorkoutIDs, beforeEndWorkoutID)
		assert.NotContains(t, includedWorkoutIDs, atEndWorkoutID)
	})

	t.Run("recalculates historic counts from current links without a query cap", func(t *testing.T) {
		ownerDBCtx := setTestUserContext(context.Background(), t, pool, ownerID)
		ownerCtx := user.WithContext(ownerDBCtx, ownerID)
		exerciseID := insertEvidenceExercise(t, pool, ownerDBCtx, ownerID, "Historic custom exercise", nil)
		for hour := 0; hour < 21; hour++ {
			workoutID := insertEvidenceWorkout(t, pool, ownerDBCtx, ownerID, time.Date(2026, time.February, 2, hour, 0, 0, 0, time.UTC))
			weight := any(nil)
			if hour%2 == 1 {
				weight = 0.0
			}
			insertEvidenceSet(t, pool, ownerDBCtx, ownerID, workoutID, exerciseID, weight, 1, "working", 1, 1)
		}

		beforeHistory := evidenceSetHistory(t, pool, ownerDBCtx, exerciseID)
		service := NewEvidenceService(queries, fixedEvidenceClock(time.Date(2026, time.February, 20, 0, 0, 0, 0, time.UTC)))
		request := TrainingEvidenceRequest{StartDate: "2026-02-01", Timezone: "UTC"}

		unclassified, err := service.Summarize(ownerCtx, request)
		require.NoError(t, err)
		assert.Equal(t, EvidenceActivity{LoggedWorkoutCount: 21, WorkingSetSessionCount: 21, WorkingSetLocalDayCount: 1}, unclassified.Activity)
		assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 21, UnclassifiedWorkingSetCount: 21, CoverageStatus: EvidenceCoverageUnclassified}, unclassified.Classification)
		assert.Empty(t, unclassified.Muscles)

		_, err = pool.Exec(ownerDBCtx, "UPDATE exercise SET catalog_id = $1 WHERE id = $2 AND user_id = $3", "Pushups", exerciseID, ownerID)
		require.NoError(t, err)
		pushups, err := service.Summarize(ownerCtx, request)
		require.NoError(t, err)
		assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 21, ClassifiedWorkingSetCount: 21, CoverageStatus: EvidenceCoverageComplete}, pushups.Classification)
		assert.Equal(t, MuscleEvidence{Muscle: "chest", Role: "primary", WorkingSetCount: 21, WorkingSetSessionCount: 21, WorkingSetLocalDayCount: 1}, evidenceMuscle(t, pushups, "chest", "primary"))

		_, err = pool.Exec(ownerDBCtx, "UPDATE exercise SET catalog_id = $1 WHERE id = $2 AND user_id = $3", "Hammer_Curls", exerciseID, ownerID)
		require.NoError(t, err)
		hammerCurls, err := service.Summarize(ownerCtx, request)
		require.NoError(t, err)
		assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 21, ClassifiedWorkingSetCount: 21, CoverageStatus: EvidenceCoverageComplete}, hammerCurls.Classification)
		assert.Equal(t, []MuscleEvidence{{Muscle: "biceps", Role: "primary", WorkingSetCount: 21, WorkingSetSessionCount: 21, WorkingSetLocalDayCount: 1}}, hammerCurls.Muscles)

		_, err = pool.Exec(ownerDBCtx, "UPDATE exercise SET catalog_id = NULL WHERE id = $1 AND user_id = $2", exerciseID, ownerID)
		require.NoError(t, err)
		cleared, err := service.Summarize(ownerCtx, request)
		require.NoError(t, err)
		assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 21, UnclassifiedWorkingSetCount: 21, CoverageStatus: EvidenceCoverageUnclassified}, cleared.Classification)
		assert.Empty(t, cleared.Muscles)
		assert.Equal(t, beforeHistory, evidenceSetHistory(t, pool, ownerDBCtx, exerciseID), "classification changes preserve stored set history")
	})
}

func insertEvidenceWorkout(t *testing.T, pool *pgxpool.Pool, ctx context.Context, userID string, date time.Time) int32 {
	t.Helper()
	var workoutID int32
	err := pool.QueryRow(ctx, "INSERT INTO workout (date, user_id) VALUES ($1, $2) RETURNING id", date, userID).Scan(&workoutID)
	require.NoError(t, err)
	return workoutID
}

func insertEvidenceExercise(t *testing.T, pool *pgxpool.Pool, ctx context.Context, userID, name string, catalogID *string) int32 {
	t.Helper()
	var exerciseID int32
	err := pool.QueryRow(ctx, "INSERT INTO exercise (name, catalog_id, user_id) VALUES ($1, $2, $3) RETURNING id", name, catalogID, userID).Scan(&exerciseID)
	require.NoError(t, err)
	return exerciseID
}

func insertEvidenceSet(t *testing.T, pool *pgxpool.Pool, ctx context.Context, userID string, workoutID, exerciseID int32, weight any, reps int32, setType string, exerciseOrder, setOrder int32) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type, user_id, exercise_order, set_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, exerciseID, workoutID, weight, reps, setType, userID, exerciseOrder, setOrder)
	require.NoError(t, err)
}

func evidenceSetHistory(t *testing.T, pool *pgxpool.Pool, ctx context.Context, exerciseID int32) []int32 {
	t.Helper()
	var workoutIDs []int32
	err := pool.QueryRow(ctx, "SELECT array_agg(workout_id ORDER BY id) FROM \"set\" WHERE exercise_id = $1", exerciseID).Scan(&workoutIDs)
	require.NoError(t, err)
	return workoutIDs
}

func evidenceMuscle(t *testing.T, evidence *TrainingEvidence, muscle, role string) MuscleEvidence {
	t.Helper()
	for _, item := range evidence.Muscles {
		if item.Muscle == muscle && item.Role == role {
			return item
		}
	}
	t.Fatalf("missing %s %s evidence", role, muscle)
	return MuscleEvidence{}
}

func stringPointer(value string) *string {
	return &value
}
