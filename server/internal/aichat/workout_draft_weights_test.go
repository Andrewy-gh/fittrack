package aichat

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/require"
)

type fixedDraftWeightHistory map[string][]DraftWeightSession

func (h fixedDraftWeightHistory) LatestDraftExerciseSets(_ context.Context, owner string, names []string, after, before time.Time) (map[string][]DraftWeightSession, error) {
	result := make(map[string][]DraftWeightSession)
	if owner == "draft-owner" {
		for _, name := range names {
			for _, session := range h[name] {
				if !session.Date.Before(after) && !session.Date.After(before) {
					result[name] = append(result[name], session)
				}
			}
		}
	}
	return result, nil
}

func TestWorkoutDraftToolAppliesScopedLoadsAndOffersProgression(t *testing.T) {
	ctx := user.WithContext(context.Background(), "draft-owner")
	now := time.Now()
	latest := now.AddDate(0, 0, -5)
	older := now.AddDate(0, 0, -9)
	stale := now.AddDate(0, 0, -92)
	working := func(weight float64) workout.SetInput {
		return workout.SetInput{Weight: &weight, Reps: 8, SetType: "working"}
	}
	history := fixedDraftWeightHistory{
		"Bench Press":   {{Date: latest, Sets: []workout.SetInput{working(115), working(115)}}, {Date: older, Sets: []workout.SetInput{working(115), working(115)}}},
		"Dumbbell Row":  {{Date: latest, Sets: []workout.SetInput{working(30)}}},
		"Pull-Ups":      {{Date: latest, Sets: []workout.SetInput{working(20)}}},
		"Dumbbell Curl": {{Date: latest, Sets: []workout.SetInput{working(20), working(25)}}},
		"Reverse Fly":   {{Date: stale, Sets: []workout.SetInput{working(15)}}},
	}
	payload := `{"date":"` + now.Format(time.RFC3339) + `","exercises":[
		{"name":"Bench Press","sets":[{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]},
		{"name":"Dumbbell Row","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Pull-Ups","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Dumbbell Curl","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Reverse Fly","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Triceps Pushdown","sets":[{"reps":8,"setType":"working"}]}
	]}`
	g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string { return payload })
	tool := defineWorkoutDraftTool(g, "synthetic", history)
	out, err := tool.RunRaw(ctx, map[string]any{
		"fitnessLevel": "intermediate", "fitnessGoal": "strength",
		"equipment": "full gym", "sessionDuration": 35, "workoutFocus": "upper body", "injuries": "none",
		"loadChoices": []map[string]any{
			{"exercise": "Dumbbell Row", "kind": "blank"},
			{"exercise": "Pull-Ups", "kind": "requested", "weight": 0.0},
			{"exercise": "Triceps Pushdown", "kind": "reported", "weight": 35.0, "reps": 8},
		},
	})
	require.NoError(t, err)
	raw, err := json.Marshal(out)
	require.NoError(t, err)
	var draft workout.CreateWorkoutRequest
	require.NoError(t, json.Unmarshal(raw, &draft))
	bench := draft.Exercises[0]
	for _, set := range bench.Sets {
		require.Equal(t, new(115.0), set.Weight)
		require.Equal(t, new(latest.Format(time.RFC3339)), set.WeightSourceDate)
	}
	require.NotNil(t, bench.DraftProgression)
	require.Equal(t, new(117.5), bench.DraftProgression.Weight)
	require.Equal(t, 5, bench.DraftProgression.Reps)
	require.Nil(t, draft.Exercises[1].Sets[0].Weight)
	require.Equal(t, new(0.0), draft.Exercises[2].Sets[0].Weight)
	require.Nil(t, draft.Exercises[2].Sets[0].WeightSourceDate)
	require.Nil(t, draft.Exercises[3].Sets[0].Weight)
	require.Nil(t, draft.Exercises[4].Sets[0].Weight)
	require.Equal(t, new(35.0), draft.Exercises[5].Sets[0].Weight)
}
