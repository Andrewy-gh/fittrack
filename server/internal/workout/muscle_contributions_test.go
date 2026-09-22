package workout

import (
	"context"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"
)

type muscleStub struct {
	rows   []db.ListMuscleContributionSetsRow
	params db.ListMuscleContributionSetsParams
}

func (s *muscleStub) ListMuscleContributionSets(_ context.Context, p db.ListMuscleContributionSetsParams) ([]db.ListMuscleContributionSetsRow, error) {
	s.params = p
	return s.rows, nil
}
func muscleRow(id int32, catalog string) db.ListMuscleContributionSetsRow {
	return db.ListMuscleContributionSetsRow{WorkoutID: 1, WorkoutDate: pgTimestamp(time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)), ExerciseID: id, ExerciseName: catalog, CatalogID: pgtype.Text{String: catalog, Valid: catalog != ""}, SetType: "working", Reps: 8}
}
func TestMuscleContributionsWorkedExample(t *testing.T) {
	stub := &muscleStub{}
	for range 6 {
		stub.rows = append(stub.rows, muscleRow(1, "Barbell_Bench_Press_-_Medium_Grip"))
	}
	for range 4 {
		stub.rows = append(stub.rows, muscleRow(2, "Triceps_Pushdown"))
	}
	stub.rows = append(stub.rows, muscleRow(3, ""), muscleRow(4, "Barbell_Full_Squat"), muscleRow(5, "Plank"))
	invalid := muscleRow(1, "Barbell_Bench_Press_-_Medium_Grip")
	invalid.Reps = 0
	stub.rows = append(stub.rows, invalid)
	warmup := muscleRow(1, "Barbell_Bench_Press_-_Medium_Grip")
	warmup.SetType = "warmup"
	stub.rows = append(stub.rows, warmup)
	result, err := NewMuscleContributionService(stub, func() time.Time { return time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC) }).Summarize(user.WithContext(context.Background(), "owner"), TrainingEvidenceRequest{"2026-09-14", "UTC"})
	require.NoError(t, err)
	require.Equal(t, "owner", stub.params.UserID)
	require.Equal(t, 13, result.WorkingSets)
	require.Equal(t, 1, result.InvalidWorkingSets)
	require.Equal(t, 1, result.UnclassifiedSets)
	require.Equal(t, 1, result.UnreviewedSets)
	require.Equal(t, MuscleContribution{Muscle: "triceps", DirectSets: 4, IndirectSets: 6, UnresolvedSets: 3, InvalidMappedSets: 1}, result.Muscles[1])
	require.Equal(t, MuscleContribution{Muscle: "hamstrings", UnresolvedSets: 13, InvalidUnresolvedSets: 1}, result.Muscles[4])
	require.Equal(t, 1, result.Muscles[3].DirectSets) // No side doubling for logged sets.
	require.Len(t, result.Exercises, 5)
}
func TestMuscleContributionsHTTPAndDST(t *testing.T) {
	for _, tc := range []struct {
		query  string
		authed bool
		status int
	}{
		{"?startDate=2026-03-08&timezone=America/New_York", true, 200},
		{"?startDate=2026-03-08&timezone=UTC", false, 401},
		{"?startDate=2026-03-08", true, 400},
		{"?startDate=2026-02-30&timezone=UTC", true, 400},
		{"?startDate=2027-01-01&timezone=UTC", true, 400},
	} {
		t.Run(tc.query, func(t *testing.T) {
			stub := &muscleStub{}
			handler := NewMuscleContributionHandler(NewMuscleContributionService(stub, func() time.Time { return time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC) }), slog.New(slog.NewTextHandler(io.Discard, nil)))
			req := httptest.NewRequest("GET", "/api/analytics/muscle-contributions"+tc.query, nil)
			if tc.authed {
				req = req.WithContext(user.WithContext(req.Context(), "owner"))
			}
			rec := httptest.NewRecorder()
			handler.Get(rec, req)
			require.Equal(t, tc.status, rec.Code)
			if tc.status == 200 {
				require.Equal(t, 167*time.Hour, stub.params.EndAt.Time.Sub(stub.params.StartAt.Time))
				require.Contains(t, rec.Body.String(), `"partial":true`)
				require.Contains(t, rec.Body.String(), `"exercises":[]`)
			}
		})
	}
}
func TestMuscleContributionInvalidWeights(t *testing.T) {
	for _, value := range []string{"-1", "NaN", "Infinity", "-Infinity"} {
		t.Run(value, func(t *testing.T) {
			row := muscleRow(1, "Barbell_Curl")
			require.NoError(t, row.Weight.Scan(value))
			result, err := NewMuscleContributionService(&muscleStub{rows: []db.ListMuscleContributionSetsRow{row}}, func() time.Time { return time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC) }).Summarize(user.WithContext(context.Background(), "owner"), TrainingEvidenceRequest{"2026-09-14", "UTC"})
			require.NoError(t, err)
			require.Zero(t, result.WorkingSets)
			require.Equal(t, 1, result.Muscles[2].InvalidMappedSets)
		})
	}
}
