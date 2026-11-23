package aichat

import (
	"context"
	"encoding/json"
	"strings"
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

func TestWorkoutDraftToolMatchesLatestWorkingSets(t *testing.T) {
	ctx := user.WithContext(context.Background(), "draft-owner")
	now := time.Now()
	recent := now.AddDate(0, 0, -4)
	older := now.AddDate(0, 0, -10)
	working := func(weight float64, reps int) workout.SetInput {
		return workout.SetInput{Weight: &weight, Reps: reps, SetType: "working"}
	}
	latest := DraftWeightSession{Date: recent, Sets: []workout.SetInput{working(115, 8), working(115, 8)}}
	previous := DraftWeightSession{Date: older, Sets: []workout.SetInput{working(95, 8), working(95, 8)}}
	warmup := workout.SetInput{Weight: new(115.0), Reps: 8, SetType: "warmup"}

	for _, tc := range []struct {
		name     string
		focus    string
		sessions []DraftWeightSession
		expected *float64
	}{
		{name: "one session repeats weight without progression", sessions: []DraftWeightSession{latest}, expected: new(115.0)},
		{name: "latest session wins over lighter earlier session", sessions: []DraftWeightSession{latest, previous}, expected: new(115.0)},
		{name: "more reps in latest session do not carry over", sessions: []DraftWeightSession{{Date: recent, Sets: []workout.SetInput{working(115, 10)}}, previous}},
		{name: "warmup does not fill working weight", sessions: []DraftWeightSession{{Date: recent, Sets: []workout.SetInput{warmup}}, previous}},
		{name: "deload leaves weights blank despite matching history", focus: "upper body deload", sessions: []DraftWeightSession{latest, previous}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := `{"date":"` + now.Format(time.RFC3339) + `","exercises":[{"name":"Bench Press","sets":[{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]}]}`
			g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string { return payload })
			tool := defineWorkoutDraftTool(g, "synthetic", fixedDraftWeightHistory{"Bench Press": tc.sessions})
			focus := tc.focus
			if focus == "" {
				focus = "upper body"
			}
			out, err := tool.RunRaw(ctx, map[string]any{
				"fitnessLevel": "intermediate", "fitnessGoal": "strength",
				"equipment": "full gym", "sessionDuration": 10, "workoutFocus": focus, "injuries": "none",
			})
			require.NoError(t, err)
			raw, err := json.Marshal(out)
			require.NoError(t, err)
			var draft workout.CreateWorkoutRequest
			require.NoError(t, json.Unmarshal(raw, &draft))
			require.Len(t, draft.Exercises, 1)
			exercise := draft.Exercises[0]
			require.Len(t, exercise.Sets, 2)
			require.Nil(t, exercise.DraftProgression)
			for _, set := range exercise.Sets {
				require.Equal(t, tc.expected, set.Weight)
				if tc.expected == nil {
					require.Nil(t, set.WeightSourceDate)
				} else {
					require.Equal(t, new(recent.Format(time.RFC3339)), set.WeightSourceDate)
				}
			}
		})
	}
}

func TestWorkoutDraftToolPreservesSetOverridesAndEquipmentLimits(t *testing.T) {
	ctx := user.WithContext(context.Background(), "draft-owner")
	now := time.Now()
	reported := func(weight float64) []DraftWeightSession {
		return []DraftWeightSession{{Date: now.AddDate(0, 0, -5), Sets: []workout.SetInput{
			{Weight: &weight, Reps: 8, SetType: "working"},
			{Weight: &weight, Reps: 8, SetType: "working"},
		}}}
	}
	history := fixedDraftWeightHistory{
		"Dumbbell Bench Press": reported(50),
		"Barbell Row":          reported(95),
	}
	history["Barbell Row"] = append(history["Barbell Row"], DraftWeightSession{
		Date: now.AddDate(0, 0, -8), Sets: history["Barbell Row"][0].Sets,
	})

	payload := `{"date":"` + now.Format(time.RFC3339) + `","exercises":[
		{"name":"Barbell Bench Press","sets":[{"reps":5,"setType":"warmup"},{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]},
		{"name":"Dumbbell Bench Press","sets":[{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]},
		{"name":"Barbell Row","sets":[{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]},
		{"name":"Pull-Ups","sets":[{"reps":4,"setType":"warmup"},{"reps":8,"setType":"working"}]}
	]}`
	attempts := 0
	g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string {
		attempts++
		if attempts == 1 {
			return strings.Replace(payload, `"reps":5,"setType":"warmup"`, `"reps":10,"setType":"warmup"`, 1)
		}
		return payload
	})
	tool := defineWorkoutDraftTool(g, "synthetic", history)
	out, err := tool.RunRaw(ctx, map[string]any{
		"fitnessLevel": "intermediate", "fitnessGoal": "strength",
		"equipment": "full gym", "sessionDuration": 25, "workoutFocus": "upper body", "injuries": "none",
		"exerciseWeightLimits": []map[string]any{
			{"exercise": "Dumbbell Bench Press", "weight": 35.0},
			{"exercise": "Barbell Row", "weight": 95.0},
		},
		"loadChoices": []map[string]any{
			{"exercise": "Barbell Bench Press", "kind": "requested", "weight": 65.0, "reps": 5, "setType": "warmup", "set": 1},
			{"exercise": "Barbell Bench Press", "kind": "requested", "weight": 115.0, "setType": "working"},
			{"exercise": "Barbell Bench Press", "kind": "blank", "setType": "working", "last": true},
			{"exercise": "Pull-Ups", "kind": "requested", "weight": 0.0},
		},
	})
	require.NoError(t, err)
	raw, err := json.Marshal(out)
	require.NoError(t, err)
	var draft workout.CreateWorkoutRequest
	require.NoError(t, json.Unmarshal(raw, &draft))
	require.Len(t, draft.Exercises, 4)
	bench, dumbbells, rows, pullups := draft.Exercises[0], draft.Exercises[1], draft.Exercises[2], draft.Exercises[3]
	require.Len(t, bench.Sets, 3)
	require.Equal(t, 5, bench.Sets[0].Reps)
	require.Equal(t, new(65.0), bench.Sets[0].Weight)
	require.Equal(t, new(115.0), bench.Sets[1].Weight)
	require.Nil(t, bench.Sets[2].Weight)
	for _, set := range dumbbells.Sets {
		require.Nil(t, set.Weight)
	}
	for _, set := range rows.Sets {
		require.Equal(t, new(95.0), set.Weight)
	}
	require.Nil(t, rows.DraftProgression)
	for _, set := range pullups.Sets {
		require.Equal(t, new(0.0), set.Weight)
	}
}
