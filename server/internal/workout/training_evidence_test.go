package workout

import (
	"context"
	"errors"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/exercisecatalog"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type evidenceQuerierStub struct {
	rows   []db.ListTrainingEvidenceRow
	err    error
	calls  int
	params []db.ListTrainingEvidenceParams
}

func (s *evidenceQuerierStub) ListTrainingEvidence(_ context.Context, params db.ListTrainingEvidenceParams) ([]db.ListTrainingEvidenceRow, error) {
	s.calls++
	s.params = append(s.params, params)
	if s.err != nil {
		return nil, s.err
	}
	return s.rows, nil
}

func TestEvidenceServiceSummarizeAssemblesDeterministicRoleEvidence(t *testing.T) {
	now := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
	pushups := "Pushups"
	bench := "Dumbbell_Bench_Press"
	hammerCurls := "Hammer_Curls"
	stub := &evidenceQuerierStub{rows: []db.ListTrainingEvidenceRow{
		evidenceRow(2, time.Date(2026, time.January, 5, 18, 0, 0, 0, time.UTC), 12, &bench, 1),
		evidenceRow(1, time.Date(2026, time.January, 5, 8, 0, 0, 0, time.UTC), 11, &pushups, 2),
		evidenceRow(3, time.Date(2026, time.January, 6, 8, 0, 0, 0, time.UTC), 13, nil, 4),
		evidenceRow(4, time.Date(2026, time.January, 6, 18, 0, 0, 0, time.UTC), 14, &hammerCurls, 2),
	}}
	service := NewEvidenceService(stub, func() time.Time { return now })

	actual, err := service.Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{
		StartDate: "2026-01-04",
		Timezone:  "UTC",
	})
	require.NoError(t, err)

	assert.Equal(t, EvidenceActivity{
		LoggedWorkoutCount:      4,
		WorkingSetSessionCount:  4,
		WorkingSetLocalDayCount: 2,
	}, actual.Activity)
	assert.Equal(t, EvidenceClassification{
		TotalWorkingSetCount:        9,
		ClassifiedWorkingSetCount:   5,
		UnclassifiedWorkingSetCount: 4,
		CoverageStatus:              EvidenceCoveragePartial,
	}, actual.Classification)
	assert.Equal(t, []MuscleEvidence{
		{Muscle: "biceps", Role: "primary", WorkingSetCount: 2, WorkingSetSessionCount: 1, WorkingSetLocalDayCount: 1},
		{Muscle: "chest", Role: "primary", WorkingSetCount: 3, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 1},
		{Muscle: "shoulders", Role: "secondary", WorkingSetCount: 3, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 1},
		{Muscle: "triceps", Role: "secondary", WorkingSetCount: 3, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 1},
	}, actual.Muscles)
	assert.NotContains(t, actual.Muscles, MuscleEvidence{Muscle: "biceps", Role: "secondary"}, "empty catalog secondary lists add no secondary evidence")
	assert.Equal(t, EvidenceBasis{
		ClassificationBasis:      "current_exercise_catalog_link",
		CalculationPolicyVersion: "training-evidence-v1",
		CatalogVersion:           exercisecatalog.Version(),
	}, actual.Basis)
	assert.Contains(t, actual.Limitations, "Unclassified working sets are unknown muscle training, not zero training for every muscle.")
	assert.Equal(t, 1, stub.calls)
}

func TestEvidenceServiceSummarizeDeduplicatesSameMuscleWithinOneWorkout(t *testing.T) {
	now := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
	pushups := "Pushups"
	bench := "Dumbbell_Bench_Press"
	stub := &evidenceQuerierStub{rows: []db.ListTrainingEvidenceRow{
		evidenceRow(1, time.Date(2026, time.January, 5, 8, 0, 0, 0, time.UTC), 11, &pushups, 2),
		evidenceRow(1, time.Date(2026, time.January, 5, 8, 0, 0, 0, time.UTC), 12, &bench, 3),
	}}

	actual, err := NewEvidenceService(stub, fixedEvidenceClock(now)).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-04", Timezone: "UTC"})
	require.NoError(t, err)
	assert.Equal(t, EvidenceActivity{LoggedWorkoutCount: 1, WorkingSetSessionCount: 1, WorkingSetLocalDayCount: 1}, actual.Activity)
	assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 5, ClassifiedWorkingSetCount: 5, CoverageStatus: EvidenceCoverageComplete}, actual.Classification)
	assert.Equal(t, []MuscleEvidence{
		{Muscle: "chest", Role: "primary", WorkingSetCount: 5, WorkingSetSessionCount: 1, WorkingSetLocalDayCount: 1},
		{Muscle: "shoulders", Role: "secondary", WorkingSetCount: 5, WorkingSetSessionCount: 1, WorkingSetLocalDayCount: 1},
		{Muscle: "triceps", Role: "secondary", WorkingSetCount: 5, WorkingSetSessionCount: 1, WorkingSetLocalDayCount: 1},
	}, actual.Muscles)
}

func TestEvidenceServiceSummarizeKeepsEmptyAndWarmupOnlyWorkoutsOutOfWorkingFrequency(t *testing.T) {
	pushups := "Pushups"
	stub := &evidenceQuerierStub{rows: []db.ListTrainingEvidenceRow{
		evidenceEmptyWorkoutRow(1, time.Date(2026, time.January, 4, 9, 0, 0, 0, time.UTC)),
		evidenceRow(2, time.Date(2026, time.January, 5, 9, 0, 0, 0, time.UTC), 11, &pushups, 0),
		evidenceRow(3, time.Date(2026, time.January, 6, 9, 0, 0, 0, time.UTC), 12, nil, 0),
	}}
	service := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)))

	actual, err := service.Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-04", Timezone: "UTC"})
	require.NoError(t, err)
	assert.Equal(t, EvidenceActivity{LoggedWorkoutCount: 3}, actual.Activity)
	assert.Equal(t, EvidenceClassification{CoverageStatus: EvidenceCoverageNotApplicable}, actual.Classification)
	assert.Empty(t, actual.Muscles)
}

func TestEvidenceServiceSummarizeReturnsNoWorkoutsAsNoLoggedWorkouts(t *testing.T) {
	stub := &evidenceQuerierStub{}
	actual, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC))).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-04", Timezone: "UTC"})
	require.NoError(t, err)
	assert.Equal(t, EvidenceActivity{}, actual.Activity)
	assert.Equal(t, EvidenceClassification{CoverageStatus: EvidenceCoverageNotApplicable}, actual.Classification)
	assert.Empty(t, actual.Muscles)
}

func TestEvidenceServiceSummarizeReportsAllUnclassifiedAndFullCoverageSeparately(t *testing.T) {
	now := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
	t.Run("all unclassified working sets", func(t *testing.T) {
		stub := &evidenceQuerierStub{rows: []db.ListTrainingEvidenceRow{
			evidenceRow(1, time.Date(2026, time.January, 5, 8, 0, 0, 0, time.UTC), 11, nil, 3),
			evidenceRow(2, time.Date(2026, time.January, 5, 18, 0, 0, 0, time.UTC), 12, nil, 2),
		}}
		actual, err := NewEvidenceService(stub, fixedEvidenceClock(now)).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-04", Timezone: "UTC"})
		require.NoError(t, err)
		assert.Equal(t, EvidenceActivity{LoggedWorkoutCount: 2, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 1}, actual.Activity)
		assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 5, UnclassifiedWorkingSetCount: 5, CoverageStatus: EvidenceCoverageUnclassified}, actual.Classification)
		assert.Empty(t, actual.Muscles)
	})

	t.Run("all classified working sets", func(t *testing.T) {
		pushups := "Pushups"
		hammerCurls := "Hammer_Curls"
		stub := &evidenceQuerierStub{rows: []db.ListTrainingEvidenceRow{
			evidenceRow(1, time.Date(2026, time.January, 5, 8, 0, 0, 0, time.UTC), 11, &pushups, 2),
			evidenceRow(2, time.Date(2026, time.January, 6, 18, 0, 0, 0, time.UTC), 12, &hammerCurls, 1),
		}}
		actual, err := NewEvidenceService(stub, fixedEvidenceClock(now)).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-04", Timezone: "UTC"})
		require.NoError(t, err)
		assert.Equal(t, EvidenceClassification{TotalWorkingSetCount: 3, ClassifiedWorkingSetCount: 3, CoverageStatus: EvidenceCoverageComplete}, actual.Classification)
	})
}

func TestEvidenceServiceSummarizeUsesExplicitLocalCalendarBoundsAndCapturesClockOnce(t *testing.T) {
	now := time.Date(2026, time.March, 11, 15, 0, 0, 0, time.UTC)
	clockCalls := 0
	stub := &evidenceQuerierStub{}
	service := NewEvidenceService(stub, func() time.Time {
		clockCalls++
		return now
	})

	actual, err := service.Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{
		StartDate: "2026-03-08",
		Timezone:  "America/New_York",
	})
	require.NoError(t, err)
	require.Len(t, stub.params, 1)
	assert.Equal(t, 1, clockCalls)
	assert.Equal(t, "evidence-user", stub.params[0].UserID)
	assert.Equal(t, time.Date(2026, time.March, 8, 5, 0, 0, 0, time.UTC), stub.params[0].StartAt.Time)
	assert.Equal(t, time.Date(2026, time.March, 15, 4, 0, 0, 0, time.UTC), stub.params[0].EndAt.Time)
	assert.Equal(t, now, stub.params[0].ObservedAt.Time)
	assert.Equal(t, "2026-03-15", actual.Period.EndDate)
	assert.True(t, actual.Period.Partial)
}

func TestEvidenceServiceSummarizeUsesCalendarDaysAcrossFallDSTAndTimezoneOffsets(t *testing.T) {
	t.Run("fall DST", func(t *testing.T) {
		observedAt := time.Date(2026, time.November, 10, 12, 0, 0, 0, time.UTC)
		stub := &evidenceQuerierStub{}
		actual, err := NewEvidenceService(stub, fixedEvidenceClock(observedAt)).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-11-01", Timezone: "America/New_York"})
		require.NoError(t, err)
		require.Len(t, stub.params, 1)
		assert.Equal(t, time.Date(2026, time.November, 1, 4, 0, 0, 0, time.UTC), stub.params[0].StartAt.Time)
		assert.Equal(t, time.Date(2026, time.November, 8, 5, 0, 0, 0, time.UTC), stub.params[0].EndAt.Time)
		assert.Equal(t, stub.params[0].StartAt.Time, actual.Period.StartAt)
		assert.Equal(t, stub.params[0].EndAt.Time, actual.Period.EndAt)
		assert.False(t, actual.Period.Partial)
	})

	t.Run("Tokyo local dates", func(t *testing.T) {
		observedAt := time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
		stub := &evidenceQuerierStub{}
		actual, err := NewEvidenceService(stub, fixedEvidenceClock(observedAt)).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-05", Timezone: "Asia/Tokyo"})
		require.NoError(t, err)
		require.Len(t, stub.params, 1)
		assert.Equal(t, time.Date(2026, time.January, 4, 15, 0, 0, 0, time.UTC), stub.params[0].StartAt.Time)
		assert.Equal(t, time.Date(2026, time.January, 11, 15, 0, 0, 0, time.UTC), stub.params[0].EndAt.Time)
		assert.Equal(t, "2026-01-05", actual.Period.StartDate)
		assert.Equal(t, "2026-01-12", actual.Period.EndDate)
	})
}

func TestEvidenceServiceSummarizeUsesRequestedTimezoneForLocalWorkingDays(t *testing.T) {
	pushups := "Pushups"
	stub := &evidenceQuerierStub{rows: []db.ListTrainingEvidenceRow{
		evidenceRow(1, time.Date(2026, time.January, 4, 15, 30, 0, 0, time.UTC), 11, &pushups, 1),
		evidenceRow(2, time.Date(2026, time.January, 5, 14, 30, 0, 0, time.UTC), 12, &pushups, 1),
	}}
	actual, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC))).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-05", Timezone: "Asia/Tokyo"})
	require.NoError(t, err)
	assert.Equal(t, EvidenceActivity{LoggedWorkoutCount: 2, WorkingSetSessionCount: 2, WorkingSetLocalDayCount: 1}, actual.Activity)
}

func TestEvidenceServiceSummarizeValidatesPeriodsAndFutureStarts(t *testing.T) {
	invalidRequests := []TrainingEvidenceRequest{
		{StartDate: "", Timezone: "UTC"},
		{StartDate: "2026-1-04", Timezone: "UTC"},
		{StartDate: "2026-02-30", Timezone: "UTC"},
		{StartDate: " 2026-01-04", Timezone: "UTC"},
		{StartDate: "2026-01-04", Timezone: ""},
		{StartDate: "2026-01-04", Timezone: "Local"},
		{StartDate: "2026-01-04", Timezone: "Not/AZone"},
	}
	for _, request := range invalidRequests {
		t.Run(request.StartDate+"/"+request.Timezone, func(t *testing.T) {
			stub := &evidenceQuerierStub{}
			_, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC))).Summarize(user.WithContext(context.Background(), "evidence-user"), request)
			require.ErrorIs(t, err, ErrInvalidTrainingEvidencePeriod)
			assert.Zero(t, stub.calls)
		})
	}

	t.Run("rejects a date whose local midnight cannot round-trip", func(t *testing.T) {
		stub := &evidenceQuerierStub{}
		_, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2012, time.January, 10, 0, 0, 0, 0, time.UTC))).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2011-12-30", Timezone: "Pacific/Apia"})
		require.ErrorIs(t, err, ErrInvalidTrainingEvidencePeriod)
		assert.Zero(t, stub.calls)
	})

	t.Run("rejects a date whose computed end midnight cannot round-trip", func(t *testing.T) {
		stub := &evidenceQuerierStub{}
		_, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2012, time.January, 10, 0, 0, 0, 0, time.UTC))).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2011-12-23", Timezone: "Pacific/Apia"})
		require.ErrorIs(t, err, ErrInvalidTrainingEvidencePeriod)
		assert.Zero(t, stub.calls)
	})

	t.Run("allows start exactly at observed time and marks it partial", func(t *testing.T) {
		stub := &evidenceQuerierStub{}
		actual, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC))).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-02", Timezone: "UTC"})
		require.NoError(t, err)
		assert.True(t, actual.Period.Partial)
		assert.Equal(t, 1, stub.calls)
	})

	t.Run("rejects a wholly future period", func(t *testing.T) {
		stub := &evidenceQuerierStub{}
		_, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2026, time.January, 1, 23, 59, 59, 0, time.UTC))).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-02", Timezone: "UTC"})
		require.ErrorIs(t, err, ErrFutureTrainingEvidencePeriod)
		assert.Zero(t, stub.calls)
	})
}

func TestEvidenceServiceSummarizeDoesNotReturnPartialResultsOnErrors(t *testing.T) {
	t.Run("requires an authenticated user", func(t *testing.T) {
		stub := &evidenceQuerierStub{}
		actual, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC))).Summarize(context.Background(), TrainingEvidenceRequest{StartDate: "2026-01-01", Timezone: "UTC"})
		var unauthorized *apperrors.Unauthorized
		require.ErrorAs(t, err, &unauthorized)
		assert.Nil(t, actual)
		assert.Zero(t, stub.calls)
	})

	t.Run("propagates query failure without evidence", func(t *testing.T) {
		queryErr := errors.New("database unavailable")
		stub := &evidenceQuerierStub{err: queryErr}
		actual, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC))).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-01", Timezone: "UTC"})
		require.ErrorIs(t, err, queryErr)
		assert.Nil(t, actual)
		assert.Equal(t, 1, stub.calls)
	})

	t.Run("fails explicitly for a nonembedded current catalog link", func(t *testing.T) {
		unknownCatalog := "not-in-embedded-catalog"
		stub := &evidenceQuerierStub{rows: []db.ListTrainingEvidenceRow{
			evidenceRow(1, time.Date(2026, time.January, 2, 8, 0, 0, 0, time.UTC), 11, &unknownCatalog, 0),
		}}
		actual, err := NewEvidenceService(stub, fixedEvidenceClock(time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC))).Summarize(user.WithContext(context.Background(), "evidence-user"), TrainingEvidenceRequest{StartDate: "2026-01-01", Timezone: "UTC"})
		require.ErrorIs(t, err, ErrInconsistentExerciseCatalog)
		assert.Nil(t, actual)
	})
}

func evidenceRow(workoutID int32, workoutDate time.Time, exerciseID int32, catalogID *string, workingSetCount int32) db.ListTrainingEvidenceRow {
	catalog := pgtype.Text{}
	if catalogID != nil {
		catalog = pgtype.Text{String: *catalogID, Valid: true}
	}
	return db.ListTrainingEvidenceRow{
		WorkoutID:       workoutID,
		WorkoutDate:     pgtype.Timestamptz{Time: workoutDate, Valid: true},
		ExerciseID:      pgtype.Int4{Int32: exerciseID, Valid: true},
		CatalogID:       catalog,
		WorkingSetCount: workingSetCount,
	}
}

func evidenceEmptyWorkoutRow(workoutID int32, workoutDate time.Time) db.ListTrainingEvidenceRow {
	return db.ListTrainingEvidenceRow{
		WorkoutID:   workoutID,
		WorkoutDate: pgtype.Timestamptz{Time: workoutDate, Valid: true},
	}
}

func fixedEvidenceClock(now time.Time) func() time.Time {
	return func() time.Time { return now }
}
