package recommendation_test

import (
	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTrainingPlanUsesPerformanceAndProfile(t *testing.T) {
	now := time.Now().UTC()
	weight := 100.0
	sets := []recommendation.WorkingSet{{Weight: &weight, Reps: 8}, {Weight: &weight, Reps: 8}}
	context := recommendation.TrainingContext{Previous: &recommendation.Session{WorkoutID: 1, Date: now.Add(-time.Hour), WorkingSets: 2}, Sessions: [][]recommendation.WorkingSet{sets}, Goal: "hypertrophy", Experience: "intermediate"}
	normal := recommendation.RecommendTraining(now, context, "normal")
	require.Equal(t, 2, normal.Plan.Sets)
	require.Equal(t, 8, normal.Plan.Reps)
	require.Equal(t, 100.0, *normal.Plan.Weight)
	great := recommendation.RecommendTraining(now, context, "great")
	require.Equal(t, 9, great.Plan.Reps)
	require.Equal(t, 100.0, *great.Plan.Weight)
	easy := recommendation.RecommendTraining(now, context, "sluggish")
	require.Equal(t, 1, easy.Plan.Sets)
	require.Equal(t, 6, easy.Plan.Reps)
	context.Goal = "strength"
	context.Sessions = append(context.Sessions, sets)
	progress := recommendation.RecommendTraining(now, context, "normal")
	require.Equal(t, 102.5, *progress.Plan.Weight)
	require.Equal(t, 5, progress.Plan.Reps)
	progress.ExerciseID = 1
	require.NoError(t, recommendation.ValidateSnapshot(recommendation.Snapshot{ExerciseName: "Press", Recommendation: progress}))
	context.Goal = "hypertrophy"
	require.Equal(t, 100.0, *recommendation.RecommendTraining(now, context, "normal").Plan.Weight)
	context.HasLimitations = true
	require.Nil(t, recommendation.RecommendTraining(now, context, "normal").Plan)
}

func TestTrainingPlanDoesNotInventLoads(t *testing.T) {
	now := time.Now().UTC()
	context := recommendation.TrainingContext{Goal: "endurance", Experience: "beginner"}
	result := recommendation.RecommendTraining(now, context, "normal")
	require.Equal(t, 2, result.Plan.Sets)
	require.Equal(t, 12, result.Plan.Reps)
	require.Nil(t, result.Plan.Weight)
	weight := 50.0
	context.Previous = &recommendation.Session{Date: now.Add(-recommendation.Recency - time.Hour), WorkingSets: 2}
	context.Sessions = [][]recommendation.WorkingSet{{{Weight: &weight, Reps: 8}}}
	require.Nil(t, recommendation.RecommendTraining(now, context, "great").Plan.Weight)
	context.Previous.Date = now.Add(-time.Hour)
	context.Sessions[0] = append(context.Sessions[0], recommendation.WorkingSet{Reps: 6})
	result = recommendation.RecommendTraining(now, context, "great")
	require.Nil(t, result.Plan.Weight)
	require.Equal(t, 6, result.Plan.Reps)
	context.Avoided = true
	require.Nil(t, recommendation.RecommendTraining(now, context, "normal").Plan)
}

func TestLowEnergyAdjustsFallbackWithoutDoubleReducingSavedSets(t *testing.T) {
	now := time.Now().UTC()
	for _, previous := range []*recommendation.Session{
		nil,
		{Date: now.Add(-recommendation.Recency - time.Hour), WorkingSets: 3},
		{Date: now.Add(-time.Hour), WorkingSets: 0},
	} {
		context := recommendation.TrainingContext{Goal: "endurance", Experience: "intermediate", Previous: previous}
		normal := recommendation.RecommendTraining(now, context, "normal")
		low := recommendation.RecommendTraining(now, context, "sluggish")
		require.Equal(t, 3, normal.Plan.Sets)
		require.Equal(t, 12, normal.Plan.Reps)
		require.Equal(t, 2, low.Plan.Sets)
		require.Equal(t, 10, low.Plan.Reps)
		require.Nil(t, low.Plan.Weight)
		context.Baseline = &recommendation.Range{Min: 3, Max: 4}
		low = recommendation.RecommendTraining(now, context, "sluggish")
		require.Equal(t, 2, low.Plan.Sets)
		require.Equal(t, 10, low.Plan.Reps)
	}
}
