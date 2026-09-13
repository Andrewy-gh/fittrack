package exercise

import (
	"github.com/stretchr/testify/require"
	"math"
	"testing"
	"time"
)

func TestGoalCheckObservations(t *testing.T) {
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	window := goalCheckWindow(now, time.UTC)
	weight := 80.0
	sets := []goalCheckSet{}
	for i := 0; i < 6; i++ {
		day := now.AddDate(0, 0, -1)
		workout := int32(1)
		if i >= 3 {
			day = now.AddDate(0, 0, -3)
			workout = 2
		}
		sets = append(sets, goalCheckSet{ID: int32(i + 1), WorkoutID: workout, Date: day, Weight: &weight, Reps: 5 + i})
	}
	result := calculateGoalCheck(1, []string{"hypertrophy", "strength"}, window, time.UTC, sets, &GoalCheckReference{Value: 100, Origin: "workout_estimate"})
	require.True(t, result.Applicable)
	require.Equal(t, 2, result.Frequency.TrainingDays)
	require.Equal(t, "reference_met", result.Frequency.Status)
	require.Equal(t, 6, result.WorkingSets.Total)
	require.Equal(t, 3.0, *result.WorkingSets.Average)
	require.Equal(t, 5, *result.Loading.RepsMin)
	require.Equal(t, 10, *result.Loading.RepsMax)
	require.Equal(t, 80.0, *result.Loading.PercentMin)
	require.Len(t, result.Sessions, 2)
	// Same-day workouts remain separate but aren't two weekly exposures.
	for i := range sets {
		sets[i].Date = now
	}
	result = calculateGoalCheck(1, []string{"strength"}, window, time.UTC, sets, nil)
	require.Len(t, result.Sessions, 2)
	require.Equal(t, 1, result.Frequency.TrainingDays)
	require.Equal(t, "below_reference", result.Frequency.Status)
	require.Nil(t, result.Loading.PercentMin)
	// A varied distribution is retained; high averages have no upper penalty.
	sets[0].WorkoutID = 2
	sets[1].WorkoutID = 2
	result = calculateGoalCheck(1, []string{"strength"}, window, time.UTC, sets, nil)
	require.Equal(t, 1, result.Sessions[0].WorkingSets)
	require.Equal(t, 5, result.Sessions[1].WorkingSets)
	require.Equal(t, "reference_met", result.WorkingSets.Status)
}
func TestGoalCheckMissingAndInvalidData(t *testing.T) {
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	window := goalCheckWindow(now, time.UTC)
	for _, reference := range []*GoalCheckReference{nil, {Value: 0}, {Value: math.NaN()}} {
		result := calculateGoalCheck(1, []string{"strength"}, window, time.UTC, nil, reference)
		require.Empty(t, result.Sessions)
		require.Nil(t, result.WorkingSets.Average)
		require.Nil(t, result.Loading.Reference)
		require.Equal(t, "insufficient_data", result.Frequency.Status)
	}
	zero := 0.0
	sets := []goalCheckSet{{ID: 1, WorkoutID: 1, Date: window.Start, Reps: 5}, {ID: 2, WorkoutID: 1, Date: now, Reps: 8, Weight: &zero},
		{ID: 3, Date: now.Add(time.Second), Reps: 5}, {ID: 4, Date: window.Start.Add(-time.Second), Reps: 5}, {ID: 5, Date: now, Reps: 0}}
	sets = append(sets, sets[0])
	result := calculateGoalCheck(1, []string{"strength"}, window, time.UTC, sets, nil)
	require.Equal(t, 2, result.WorkingSets.Total)
	require.Equal(t, 0, result.Loading.SetsWithLoad)
	require.Nil(t, result.Loading.PercentMin)
	result = calculateGoalCheck(1, []string{"hypertrophy"}, window, time.UTC, sets, nil)
	require.False(t, result.Applicable)
	require.Equal(t, "insufficient_data", result.WorkingSets.Status)
	high := 120.0
	result = calculateGoalCheck(1, []string{"strength"}, window, time.UTC, []goalCheckSet{{ID: 1, WorkoutID: 1, Date: now, Reps: 1, Weight: &high}}, &GoalCheckReference{Value: 100})
	require.Equal(t, 120.0, *result.Loading.PercentMax)
}
func TestGoalCheckCalendarWindowAcrossDST(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, location)
	window := goalCheckWindow(now, location)
	require.Equal(t, "2026-03-04T00:00:00-05:00", window.Start.Format(time.RFC3339))
	require.Equal(t, 155*time.Hour, now.Sub(window.Start))
	sets := []goalCheckSet{{ID: 1, WorkoutID: 1, Date: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC), Reps: 5}, {ID: 2, WorkoutID: 2, Date: time.Date(2026, 3, 10, 6, 0, 0, 0, time.UTC), Reps: 5}}
	result := calculateGoalCheck(1, []string{"strength"}, window, location, sets, nil)
	require.Equal(t, 2, result.Frequency.TrainingDays)
}
