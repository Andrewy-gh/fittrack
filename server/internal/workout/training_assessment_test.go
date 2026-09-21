package workout

import (
	"context"
	"errors"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type assessmentStub struct {
	rows                 []db.ListExerciseAssessmentSetsRow
	lookupErr, errorRows error
	owner                string
	params               db.ListExerciseAssessmentSetsParams
}

func (s *assessmentStub) GetExercise(_ context.Context, p db.GetExerciseParams) (db.GetExerciseRow, error) {
	s.owner = p.UserID
	return db.GetExerciseRow{ID: p.ID}, s.lookupErr
}
func (s *assessmentStub) ListExerciseAssessmentSets(_ context.Context, p db.ListExerciseAssessmentSetsParams) ([]db.ListExerciseAssessmentSetsRow, error) {
	s.params = p
	return s.rows, s.errorRows
}
func assessmentRow(workoutID int32, reps int32, setType string) db.ListExerciseAssessmentSetsRow {
	return db.ListExerciseAssessmentSetsRow{WorkoutID: workoutID, WorkoutDate: pgTimestamp(time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)), Reps: reps, SetType: setType}
}
func TestAssessmentSessionComparisons(t *testing.T) {
	for _, tc := range []struct {
		name   string
		rows   []db.ListExerciseAssessmentSetsRow
		counts []int
		states []string
		met    int
	}{
		{"one plus five not averaged", []db.ListExerciseAssessmentSetsRow{assessmentRow(1, 5, "working"), assessmentRow(2, 5, "working"), assessmentRow(2, 5, "working"), assessmentRow(2, 5, "working"), assessmentRow(2, 5, "working"), assessmentRow(2, 5, "working")}, []int{1, 5}, []string{"below_reference", "reference_met"}, 1},
		{"null weight and single session sufficient for set comparison", []db.ListExerciseAssessmentSetsRow{assessmentRow(1, 5, "working"), assessmentRow(1, 5, "working")}, []int{2}, []string{"reference_met"}, 1},
		{"warmups are not working", []db.ListExerciseAssessmentSetsRow{assessmentRow(1, 5, "warmup")}, []int{0}, []string{"cannot_assess"}, 0},
		{"invalid row suppresses affected session", []db.ListExerciseAssessmentSetsRow{assessmentRow(1, 0, "working"), assessmentRow(1, 5, "working"), assessmentRow(1, 5, "working")}, []int{2}, []string{"cannot_assess"}, 0},
		{"no exercise sessions", nil, []int{}, []string{}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &assessmentStub{rows: tc.rows}
			service := NewAssessmentService(stub, func() time.Time { return time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC) })
			result, err := service.Assess(user.WithContext(context.Background(), "owner"), 3, TrainingEvidenceRequest{StartDate: "2026-09-14", Timezone: "UTC"})
			require.NoError(t, err)
			require.Len(t, result.Sessions, len(tc.counts))
			require.Equal(t, tc.met, result.SessionsMeetingReference)
			for i, session := range result.Sessions {
				require.Equal(t, tc.counts[i], session.WorkingSets)
				require.Equal(t, tc.states[i], session.State)
			}
			require.Equal(t, "owner", stub.owner)
			require.Equal(t, "owner", stub.params.UserID)
			require.Equal(t, int32(3), stub.params.ExerciseID)
		})
	}
}
func TestAssessmentPeriodAndErrors(t *testing.T) {
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	stub := &assessmentStub{rows: []db.ListExerciseAssessmentSetsRow{assessmentRow(1, 5, "working"), assessmentRow(1, 5, "working")}}
	service := NewAssessmentService(stub, func() time.Time { return now })
	ctx := user.WithContext(context.Background(), "owner")
	result, err := service.Assess(ctx, 1, TrainingEvidenceRequest{StartDate: "2026-03-08", Timezone: "America/New_York"})
	require.NoError(t, err)
	require.True(t, result.Period.Partial)
	require.Equal(t, 167*time.Hour, result.Period.EndAt.Sub(result.Period.StartAt))
	require.Equal(t, now, stub.params.ObservedAt.Time)
	require.Equal(t, "period_in_progress", result.Sessions[0].ReasonCode)
	require.Zero(t, result.SessionsMeetingReference)
	_, err = service.Assess(context.Background(), 1, TrainingEvidenceRequest{})
	require.Error(t, err)
	_, err = service.Assess(ctx, 1, TrainingEvidenceRequest{StartDate: "2026-03-20", Timezone: "UTC"})
	require.ErrorIs(t, err, ErrFutureTrainingEvidencePeriod)
	stub.lookupErr = pgx.ErrNoRows
	_, err = service.Assess(ctx, 1, TrainingEvidenceRequest{StartDate: "2026-03-01", Timezone: "UTC"})
	require.ErrorContains(t, err, "not found")
	stub.lookupErr = nil
	stub.errorRows = errors.New("database failed")
	_, err = service.Assess(ctx, 1, TrainingEvidenceRequest{StartDate: "2026-03-01", Timezone: "UTC"})
	require.ErrorContains(t, err, "database failed")
}
func TestAssessmentInvalidWeights(t *testing.T) {
	for _, value := range []string{"0", "-1", "NaN"} {
		t.Run(value, func(t *testing.T) {
			row := assessmentRow(1, 5, "working")
			var weight pgtype.Numeric
			require.NoError(t, weight.Scan(value))
			row.Weight = weight
			result, err := NewAssessmentService(&assessmentStub{rows: []db.ListExerciseAssessmentSetsRow{row, row}}, func() time.Time { return time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC) }).Assess(user.WithContext(context.Background(), "owner"), 1, TrainingEvidenceRequest{StartDate: "2026-09-14", Timezone: "UTC"})
			require.NoError(t, err)
			if value == "0" {
				require.Equal(t, "reference_met", result.Sessions[0].State)
			} else {
				require.Equal(t, "invalid_set_data", result.Sessions[0].ReasonCode)
				require.Equal(t, 2, result.Sessions[0].InvalidWorkingSets)
			}
		})
	}
}
