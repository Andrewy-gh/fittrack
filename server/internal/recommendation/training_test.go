package recommendation_test

import (
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/stretchr/testify/require"
)

func TestTrainingPlanUsesPerformanceAndProfile(t *testing.T) {
	now := time.Now().UTC()
	weight := 100.0
	sets := []recommendation.WorkingSet{{Weight: &weight, Reps: 8}, {Weight: &weight, Reps: 8}}
	context := recommendation.TrainingContext{
		Sessions:   []recommendation.Session{{Date: now.Add(-time.Hour), Sets: sets}},
		Goal:       "hypertrophy",
		Experience: "intermediate",
	}
	normal := recommendation.RecommendTraining(now, context, "normal")
	require.Equal(t, 2, normal.Plan.Sets)
	require.Equal(t, 8, normal.Plan.Reps)
	require.Equal(t, 100.0, *normal.Plan.Weight)
	great := recommendation.RecommendTraining(now, context, "great")
	require.Equal(t, 9, great.Plan.Reps)
	easy := recommendation.RecommendTraining(now, context, "sluggish")
	require.Equal(t, 1, easy.Plan.Sets)
	require.Equal(t, 6, easy.Plan.Reps)

	context.Goal = "strength"
	context.Sessions = append(context.Sessions, recommendation.Session{Date: now.Add(-2 * time.Hour), Sets: sets})
	progress := recommendation.RecommendTraining(now, context, "normal")
	require.Equal(t, 102.5, *progress.Plan.Weight)
	require.Equal(t, 5, progress.Plan.Reps)

	context.Sessions[1].Date = context.Sessions[0].Date
	sameTime := recommendation.RecommendTraining(now, context, "normal")
	require.Equal(t, 100.0, *sameTime.Plan.Weight)
	require.Equal(t, 8, sameTime.Plan.Reps)

	progress.ExerciseID = 1
	require.NoError(t, recommendation.ValidateSnapshot(recommendation.Snapshot{ExerciseName: "Press", Recommendation: progress}))

	context.HasLimitations = true
	require.Nil(t, recommendation.RecommendTraining(now, context, "normal").Plan)
}

func TestTrainingPlanFallsBackWithoutUsableHistory(t *testing.T) {
	now := time.Now().UTC()
	weight := 50.0
	contexts := []recommendation.TrainingContext{
		{Goal: "endurance", Experience: "beginner"},
		{Goal: "endurance", Experience: "beginner", Sessions: []recommendation.Session{{Date: now.Add(-recommendation.Recency - time.Hour), Sets: []recommendation.WorkingSet{{Weight: &weight, Reps: 8}}}}},
		{Goal: "endurance", Experience: "beginner", Sessions: []recommendation.Session{{Date: now.Add(-time.Hour)}}},
	}
	for _, context := range contexts {
		result := recommendation.RecommendTraining(now, context, "normal")
		require.Equal(t, 2, result.Plan.Sets)
		require.Equal(t, 12, result.Plan.Reps)
		require.Nil(t, result.Plan.Weight)
	}
}

func TestTrainingPlanDoesNotInventLoads(t *testing.T) {
	now := time.Now().UTC()
	weight := 50.0
	context := recommendation.TrainingContext{Sessions: []recommendation.Session{{
		Date: now.Add(-time.Hour),
		Sets: []recommendation.WorkingSet{{Weight: &weight, Reps: 8}, {Reps: 6}},
	}}}
	result := recommendation.RecommendTraining(now, context, "great")
	require.Nil(t, result.Plan.Weight)
	require.Equal(t, 6, result.Plan.Reps)

	context.Avoided = true
	require.Nil(t, recommendation.RecommendTraining(now, context, "normal").Plan)
}
