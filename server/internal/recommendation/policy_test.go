package recommendation_test

import (
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/stretchr/testify/require"
)

func TestBaselineSelectionAndFallbacks(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	recent := &recommendation.Session{WorkoutID: 1, Date: now.Add(-24 * time.Hour), WorkingSets: 5}
	tests := []struct {
		name         string
		prescription *recommendation.Range
		history      *recommendation.Session
		readiness    string
		source       string
		want         *recommendation.Range
	}{
		{"explicit wins", &recommendation.Range{Min: 2, Max: 4}, recent, "normal", "prescription", &recommendation.Range{Min: 2, Max: 4}},
		{"history", nil, recent, "normal", "history", &recommendation.Range{Min: 5, Max: 5}},
		{"great unchanged", nil, recent, "great", "history", &recommendation.Range{Min: 5, Max: 5}},
		{"sluggish", nil, recent, "sluggish", "history", &recommendation.Range{Min: 4, Max: 4}},
		{"floor", &recommendation.Range{Min: 1, Max: 2}, nil, "sluggish", "prescription", &recommendation.Range{Min: 1, Max: 1}},
		{"explicit without history", &recommendation.Range{Min: 2, Max: 3}, nil, "normal", "prescription", &recommendation.Range{Min: 2, Max: 3}},
		{"empty", nil, nil, "normal", "none", nil},
		{"stale", nil, &recommendation.Session{Date: now.Add(-recommendation.Recency - time.Second), WorkingSets: 4}, "normal", "none", nil},
		{"exact threshold", nil, &recommendation.Session{Date: now.Add(-recommendation.Recency), WorkingSets: 4}, "normal", "history", &recommendation.Range{Min: 4, Max: 4}},
		{"future", nil, &recommendation.Session{Date: now.Add(time.Second), WorkingSets: 4}, "normal", "none", nil},
		{"warmups only", nil, &recommendation.Session{Date: now, WorkingSets: 0}, "normal", "none", nil},
		{"unsupported history", nil, &recommendation.Session{Date: now, WorkingSets: 21}, "normal", "none", nil},
		{"inverted", &recommendation.Range{Min: 4, Max: 2}, recent, "normal", "none", nil},
		{"zero", &recommendation.Range{Min: 0, Max: 2}, recent, "normal", "none", nil},
		{"too large", &recommendation.Range{Min: 1, Max: 21}, recent, "normal", "none", nil},
		{"unknown readiness", &recommendation.Range{Min: 2, Max: 3}, recent, "excellent", "none", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := recommendation.Recommend(now, test.prescription, test.history, test.readiness)
			require.Equal(t, test.source, got.Source)
			require.Equal(t, test.want, got.Range)
			require.NotEmpty(t, got.Explanation)
			require.Equal(t, recommendation.PolicyVersion, got.PolicyVersion)
		})
	}
}

func TestAllSupportedRangesStayBoundedAndDoNotIncrease(t *testing.T) {
	for low := 1; low <= recommendation.MaxSets; low++ {
		for high := low; high <= recommendation.MaxSets; high++ {
			original := recommendation.Range{Min: low, Max: high}
			normal := recommendation.Recommend(time.Now(), &original, nil, "normal")
			great := recommendation.Recommend(time.Now(), &original, nil, "great")
			sluggish := recommendation.Recommend(time.Now(), &original, nil, "sluggish")
			require.Equal(t, normal.Range, great.Range)
			require.True(t, sluggish.Range.Valid())
			require.LessOrEqual(t, sluggish.Range.Min, low)
			require.LessOrEqual(t, sluggish.Range.Max, high)
			require.Equal(t, recommendation.Range{Min: low, Max: high}, original)
		}
	}
}

func TestSnapshotValidationDoesNotRecomputeHistoricalRecommendations(t *testing.T) {
	result := recommendation.Recommend(time.Now(), &recommendation.Range{Min: 3, Max: 4}, nil, "sluggish")
	result.ExerciseID = 12
	snapshot := recommendation.Snapshot{ExerciseName: "Press", Recommendation: result, Feedback: "about_right"}
	require.NoError(t, recommendation.ValidateSnapshot(snapshot))
	snapshot.Feedback = "bad"
	require.Error(t, recommendation.ValidateSnapshot(snapshot))
	snapshot.Feedback = ""
	snapshot.Recommendation.Range = &recommendation.Range{Min: 3, Max: 4}
	require.Error(t, recommendation.ValidateSnapshot(snapshot))
}
