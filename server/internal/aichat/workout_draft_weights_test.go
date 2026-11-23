package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/stretchr/testify/assert"
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
	// UTC is already May 2; the browser is still on May 1.
	ctx := withDraftTimezone(user.WithContext(context.Background(), "draft-owner"), "America/Los_Angeles")
	now := time.Date(2026, 5, 2, 1, 30, 0, 0, time.UTC).In(draftTimezone(ctx))
	ctx = context.WithValue(ctx, draftClockKey{}, now)
	recent := now.AddDate(0, 0, -4)
	older := now.AddDate(0, 0, -10)
	working := func(weight float64, reps int) workout.SetInput {
		return workout.SetInput{Weight: &weight, Reps: reps, SetType: "working"}
	}
	latest := DraftWeightSession{Date: recent, Sets: []workout.SetInput{working(115, 8), working(115, 8)}}
	previous := DraftWeightSession{Date: older, Sets: []workout.SetInput{working(95, 8), working(95, 8)}}
	warmup := workout.SetInput{Weight: new(115.0), Reps: 8, SetType: "warmup"}

	for _, tc := range []struct {
		name          string
		focus         string
		sessions      []DraftWeightSession
		expected      *float64
		date          time.Time
		requestedDate string
	}{
		{name: "one session repeats weight without progression", sessions: []DraftWeightSession{latest}, expected: new(115.0)},
		{name: "latest session wins over lighter earlier session", sessions: []DraftWeightSession{latest, previous}, expected: new(115.0)},
		{name: "more reps in latest session do not carry over", sessions: []DraftWeightSession{{Date: recent, Sets: []workout.SetInput{working(115, 10)}}, previous}},
		{name: "old generated date cannot shift the rolling window", date: now.AddDate(0, 0, -100), sessions: []DraftWeightSession{latest, {Date: now.AddDate(0, 0, -110), Sets: []workout.SetInput{working(65, 8)}}}, expected: new(115.0)},
		{name: "today ignores an incorrectly backdated model output", date: now.AddDate(0, 0, -100), requestedDate: "today", sessions: []DraftWeightSession{latest}, expected: new(115.0)},
		{name: "future date cannot age out current history", date: now.AddDate(0, 0, 100), requestedDate: now.AddDate(0, 0, 100).Format(time.DateOnly), sessions: []DraftWeightSession{latest}, expected: new(115.0)},
		{name: "explicitly backdated draft excludes later logged sets", date: now.AddDate(0, 0, -6), requestedDate: now.AddDate(0, 0, -6).Format(time.DateOnly), sessions: []DraftWeightSession{latest}},
		{name: "warmup does not fill working weight", sessions: []DraftWeightSession{{Date: recent, Sets: []workout.SetInput{warmup}}, previous}},
		{name: "deload leaves weights blank despite matching history", focus: "upper body deload", sessions: []DraftWeightSession{latest, previous}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			date := tc.date
			if date.IsZero() {
				date = now
			}
			payload := `{"date":"` + date.Format(time.RFC3339) + `","exercises":[{"name":"Bench Press","sets":[{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]}]}`
			g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string { return payload })
			tool := defineWorkoutDraftTool(g, "synthetic", fixedDraftWeightHistory{"Bench Press": tc.sessions})
			focus := tc.focus
			if focus == "" {
				focus = "upper body"
			}
			out, err := tool.RunRaw(ctx, map[string]any{
				"fitnessLevel": "intermediate", "fitnessGoal": "strength",
				"equipment": "full gym", "sessionDuration": 10, "workoutFocus": focus, "injuries": "none", "workoutDate": tc.requestedDate,
			})
			require.NoError(t, err)
			raw, err := json.Marshal(out)
			require.NoError(t, err)
			var draft workout.CreateWorkoutRequest
			require.NoError(t, json.Unmarshal(raw, &draft))
			if tc.requestedDate == "today" {
				require.Equal(t, "2026-05-01T18:30:00-07:00", draft.Date)
			}
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
		"Dumbbell Bench Press":          reported(50),
		"Dumbbell Row":                  reported(50),
		"Barbell Row":                   reported(95),
		"Resistance Band Dumbbell Curl": reported(50),
		"Kettlebell Swing":              reported(25),
		"Barbell Overhead Press":        reported(65),
	}
	history["Barbell Row"] = append(history["Barbell Row"], DraftWeightSession{
		Date: now.AddDate(0, 0, -8), Sets: history["Barbell Row"][0].Sets,
	})

	payload := `{"date":"` + now.Format(time.RFC3339) + `","exercises":[
		{"name":"Barbell Bench Press","sets":[{"reps":5,"setType":"warmup"},{"reps":8,"setType":"working"},{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]},
		{"name":"Dumbbell Bench Press","sets":[{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]},
		{"name":"Barbell Row","sets":[{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]},
		{"name":"Pull-Ups","sets":[{"reps":4,"setType":"warmup"},{"reps":8,"setType":"working"}]},
		{"name":"Dumbbell Row","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Barbell Overhead Press","sets":[{"reps":4,"setType":"warmup"},{"reps":8,"setType":"working"}]},
		{"name":"Resistance Band Dumbbell Curl","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Kettlebell Swing","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Shoulder Press","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Leg Press Machine","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Incline Press","sets":[{"reps":8,"setType":"working"}]}
	]}`
	rejectedDrafts := []struct {
		name                string
		payload             string
		fixedWithoutCeiling bool
	}{
		{name: "missing warmup with unspecified reps", payload: strings.Replace(payload,
			`{"reps":5,"setType":"warmup"},`, "", 1)},
		{name: "missing explicitly blank warmup", fixedWithoutCeiling: true, payload: strings.Replace(payload,
			`"Barbell Overhead Press","sets":[{"reps":4,"setType":"warmup"},`, `"Barbell Overhead Press","sets":[`, 1)},
		{name: "wrong reps on blank warmup", payload: strings.Replace(payload,
			`"Barbell Overhead Press","sets":[{"reps":4,"setType":"warmup"}`, `"Barbell Overhead Press","sets":[{"reps":5,"setType":"warmup"}`, 1)},
		{name: "wrong reps on one governed working set", payload: strings.Replace(payload,
			`{"reps":8,"setType":"working"}`, `{"reps":7,"setType":"working"}`, 1)},
		{name: "wrong reps on last set with only weight overridden", payload: strings.Replace(payload,
			`{"reps":8,"setType":"working"}]},`, `{"reps":6,"setType":"working"}]},`, 1)},
	}
	for _, rejected := range rejectedDrafts {
		t.Run(rejected.name, func(t *testing.T) {
			attempts := 0
			g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string {
				attempts++
				if attempts == 1 {
					return rejected.payload
				}
				return payload
			})
			tool := defineWorkoutDraftTool(g, "synthetic", history)
			fixedLimit := map[string]any{"equipment": "kettlebell", "fixed": true}
			if !rejected.fixedWithoutCeiling {
				fixedLimit["weight"] = 35.0
			}
			out, err := tool.RunRaw(ctx, map[string]any{
				"fitnessLevel": "intermediate", "fitnessGoal": "strength",
				"equipment": "full gym", "sessionDuration": 25, "workoutFocus": "upper body", "injuries": "none",
				"exerciseWeightLimits": []map[string]any{
					{"exercise": "Dumbbell Bench Press", "equipment": "dumbbell", "weight": 35.0},
					{"exercise": "Barbell Row", "weight": 95.0},
					{"equipment": "Smith machine", "weight": 55.0},
					{"equipment": "cable machine", "weight": 50.0},
					{"equipment": "dumbbell", "weight": 35.0},
					{"equipment": "barbell", "weight": 120.0},
					fixedLimit,
				},
				"loadChoices": []map[string]any{
					{"exercise": "Barbell Bench Press", "kind": "requested", "weight": 65.0, "setType": "warmup"},
					{"exercise": "Barbell Bench Press", "kind": "requested", "weight": 115.0, "setType": "working", "reps": 8},
					{"exercise": "Barbell Bench Press", "kind": "blank", "setType": "working", "last": true},
					{"exercise": "Pull-Ups", "kind": "requested", "weight": 0.0},
					{"exercise": "Barbell Overhead Press", "kind": "blank", "setType": "warmup", "reps": 4},
					{"exercise": "Barbell Overhead Press", "kind": "structure", "setType": "all", "equipment": "barbell"},
					{"exercise": "Shoulder Press", "kind": "requested", "setType": "working", "equipment": "smith", "weight": 65.0},
					{"exercise": "Leg Press Machine", "kind": "requested", "weight": 125.0},
					{"exercise": "Incline Press", "kind": "requested", "weight": 65.0},
				},
			})
			require.NoError(t, err)
			raw, err := json.Marshal(out)
			require.NoError(t, err)
			var draft workout.CreateWorkoutRequest
			require.NoError(t, json.Unmarshal(raw, &draft))
			require.Len(t, draft.Exercises, 11)
			bench, dumbbells, rows, pullups := draft.Exercises[0], draft.Exercises[1], draft.Exercises[2], draft.Exercises[3]
			require.Len(t, bench.Sets, 4)
			require.Equal(t, 5, bench.Sets[0].Reps)
			require.Equal(t, new(65.0), bench.Sets[0].Weight)
			require.Equal(t, new(115.0), bench.Sets[1].Weight)
			require.Equal(t, new(115.0), bench.Sets[2].Weight)
			require.Nil(t, bench.Sets[3].Weight)
			require.Equal(t, 8, bench.Sets[3].Reps)
			require.Nil(t, draft.Exercises[4].Sets[0].Weight)
			require.Len(t, draft.Exercises[5].Sets, 2)
			require.Equal(t, "warmup", draft.Exercises[5].Sets[0].SetType)
			require.Equal(t, 4, draft.Exercises[5].Sets[0].Reps)
			require.Nil(t, draft.Exercises[5].Sets[0].Weight)
			assert.Equal(t, new(65.0), draft.Exercises[5].Sets[1].Weight)
			assert.NotNil(t, draft.Exercises[5].Sets[1].WeightSourceDate)
			require.Nil(t, draft.Exercises[6].Sets[0].Weight)
			require.Nil(t, draft.Exercises[7].Sets[0].Weight)
			require.Nil(t, draft.Exercises[8].Sets[0].Weight)
			require.Equal(t, new(125.0), draft.Exercises[9].Sets[0].Weight)
			require.Equal(t, new(65.0), draft.Exercises[10].Sets[0].Weight)
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
			require.Equal(t, 2, attempts)
		})
	}
}

func TestWorkoutDraftToolKeepsReportedLoadWhenLoggedHistoryIsSkipped(t *testing.T) {
	ctx := user.WithContext(context.Background(), "draft-owner")
	date := time.Now().Format(time.RFC3339)
	for _, tc := range []struct {
		name, equipment, exercise, setType string
		reps                               int
		weight                             float64
	}{
		{name: "dumbbell exception", equipment: "adjustable dumbbells", exercise: "Dumbbell Floor Press", setType: "working", reps: 8, weight: 30},
		{name: "trap bar exception", equipment: "a trap bar and plates", exercise: "Trap Bar Deadlift", setType: "working", reps: 8, weight: 65},
		{name: "sandbag exception", equipment: "an adjustable sandbag", exercise: "Sandbag Squat", setType: "working", reps: 8, weight: 40},
		{name: "weighted vest exception", equipment: "a weighted vest", exercise: "Weighted Push-Up", setType: "working", reps: 8, weight: 15},
		{name: "comparable warmup", equipment: "adjustable dumbbells", exercise: "Dumbbell Floor Press", setType: "warmup", reps: 5, weight: 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string {
				return fmt.Sprintf(`{"date":%q,"exercises":[
					{"name":%q,"sets":[{"reps":%d,"setType":%q},{"reps":10,"setType":"working"}]},
					{"name":"Reverse Lunge","sets":[{"reps":8,"setType":"working"}]}
				]}`, date, tc.exercise, tc.reps, tc.setType)
			})
			history := fixedDraftWeightHistory{"Reverse Lunge": {{Date: time.Now().AddDate(0, 0, -1), Sets: []workout.SetInput{
				{Weight: new(35.0), Reps: 8, SetType: "working"},
			}}}}
			tool := defineWorkoutDraftTool(g, "synthetic", history)
			out, err := tool.RunRaw(ctx, map[string]any{
				"skipHistoryWeights": true,
				"equipment":          "no equipment except " + tc.equipment, "sessionDuration": 10, "workoutFocus": "full body", "injuries": "none",
				"loadChoices": []map[string]any{
					{"exercise": tc.exercise, "kind": "reported", "weight": tc.weight, "reps": tc.reps, "setType": tc.setType},
				},
			})
			require.NoError(t, err)
			raw, err := json.Marshal(out)
			require.NoError(t, err)
			var draft workout.CreateWorkoutRequest
			require.NoError(t, json.Unmarshal(raw, &draft))
			require.Len(t, draft.Exercises, 2)
			require.Equal(t, &tc.weight, draft.Exercises[0].Sets[0].Weight)
			require.Nil(t, draft.Exercises[0].Sets[1].Weight)
			require.Nil(t, draft.Exercises[1].Sets[0].Weight)
			require.False(t, draft.Exercises[1].Sets[0].DraftBodyweight)
		})
	}
}

func TestWorkoutDraftToolKeepsWorkingSetEquipmentSeparate(t *testing.T) {
	ctx := user.WithContext(context.Background(), "draft-owner")
	now := time.Now()
	g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string {
		return fmt.Sprintf(`{"date":%q,"exercises":[{"name":"Overhead Press","sets":[{"reps":5,"setType":"warmup"},{"reps":8,"setType":"working"}]}]}`, now.Format(time.RFC3339))
	})
	history := fixedDraftWeightHistory{"Overhead Press": {{Date: now.AddDate(0, 0, -1), Sets: []workout.SetInput{
		{Weight: new(65.0), Reps: 8, SetType: "working"},
	}}}}
	tool := defineWorkoutDraftTool(g, "synthetic", history)
	out, err := tool.RunRaw(ctx, map[string]any{
		"equipment":       "barbell warm-up, then dumbbells up to 35 lb for working sets",
		"sessionDuration": 10, "workoutFocus": "upper body", "injuries": "none",
		"exerciseWeightLimits": []map[string]any{{"equipment": "dumbbell", "weight": 35.0}},
		"loadChoices":          []map[string]any{{"exercise": "Overhead Press", "kind": "blank", "equipment": "barbell", "setType": "warmup", "reps": 5}},
	})
	require.NoError(t, err)
	raw, err := json.Marshal(out)
	require.NoError(t, err)
	var draft workout.CreateWorkoutRequest
	require.NoError(t, json.Unmarshal(raw, &draft))
	require.Len(t, draft.Exercises, 2)
	assert.Equal(t, "Barbell Overhead Press", draft.Exercises[0].Name)
	require.Len(t, draft.Exercises[1].Sets, 1)
	assert.Nil(t, draft.Exercises[1].Sets[0].Weight, "a warm-up assignment must not establish working-set equipment or bypass its cap")
}

func TestWorkoutDraftToolRequiresRequestedWarmupWithReportedLoad(t *testing.T) {
	ctx := user.WithContext(context.Background(), "draft-owner")
	date := time.Now().Format(time.RFC3339)
	attempt := 0
	g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string {
		attempt++
		warmup := ""
		if attempt > 1 {
			warmup = `{"reps":5,"setType":"warmup"},`
		}
		return fmt.Sprintf(`{"date":%q,"exercises":[{"name":"Bench Press","sets":[%s{"reps":8,"setType":"working"}]}]}`, date, warmup)
	})
	tool := defineWorkoutDraftTool(g, "synthetic", nil)
	out, err := tool.RunRaw(ctx, map[string]any{
		"equipment": "full gym", "sessionDuration": 10, "workoutFocus": "upper body", "injuries": "none",
		"skipHistoryWeights": true,
		"loadChoices": []map[string]any{
			{"exercise": "Bench Press", "kind": "reported", "setType": "warmup", "reps": 5, "weight": 65.0},
			{"exercise": "Bench Press", "kind": "structure", "setType": "warmup", "reps": 5},
		},
	})
	require.NoError(t, err)
	raw, err := json.Marshal(out)
	require.NoError(t, err)
	var draft workout.CreateWorkoutRequest
	require.NoError(t, json.Unmarshal(raw, &draft))
	require.Equal(t, "warmup", draft.Exercises[0].Sets[0].SetType, "the requested warm-up was dropped")
	require.Equal(t, new(65.0), draft.Exercises[0].Sets[0].Weight)
	require.Equal(t, 2, attempt)
}
func TestChatRecoversMissingInjuryFieldWithoutAssumingInjuryStatus(t *testing.T) {
	for _, tc := range []struct {
		answer     string
		earlier    string
		wantDraft  bool
		limitation string
	}{
		{answer: "No injuries.", wantDraft: true},
		{answer: "Make a 10-minute bodyweight upper body workout, no injuries. I'm a beginner.", wantDraft: true},
		{answer: "Go ahead."},
		{answer: "My training partner has no injuries. Go ahead."},
		{answer: "I have a friend with no injuries. Go ahead."},
		{answer: "I have no injuries, and I'm not sure about my fitness level.", wantDraft: true},
		{answer: "I also have shoulder pain.", earlier: "I have knee pain.", wantDraft: true, limitation: "i have knee pain; i also have shoulder pain"},
		{answer: "Go ahead, but I can't confirm I have no injuries yet."},
		{answer: "Go ahead. I haven't confirmed that I have no injuries."},
		{answer: "I don't know my fitness level, but I have no injuries.", wantDraft: true},
		{answer: "I don't have any injuries.", wantDraft: true},
		{answer: "No injuries, but I can't raise my right arm overhead.", wantDraft: true, limitation: "i can't raise my right arm overhead"},
		{answer: "No injuries, but I can't fully raise my right arm overhead.", wantDraft: true, limitation: "i can't fully raise my right arm overhead"},
		{answer: "I have knee pain. I have shoulder pain and overhead pressing hurts.", wantDraft: true, limitation: "i have knee pain; i have shoulder pain and overhead pressing hurts"},
		{answer: "It's mild front-of-knee discomfort, no sharp pain.", wantDraft: true, limitation: "mild front-of-knee discomfort"},
	} {
		t.Run(tc.answer, func(t *testing.T) {
			ctx := context.Background()
			generations := 0
			g := syntheticDraftModel(ctx, func(req *ai.ModelRequest) string {
				generations++
				if tc.limitation != "" {
					prompt := ""
					for _, message := range req.Messages {
						prompt += message.Text()
					}
					require.Contains(t, prompt, tc.limitation)
				}
				return ordinaryDraftJSON
			})
			calls := 0
			genkit.DefineModel(g, "chat", &ai.ModelOptions{Supports: &ai.ModelSupports{Tools: true, Multiturn: true, SystemRole: true}},
				func(_ context.Context, req *ai.ModelRequest, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
					calls++
					if calls == 2 && !tc.wantDraft {
						require.Zero(t, generations)
						last := req.Messages[len(req.Messages)-1]
						require.Contains(t, fmt.Sprint(last.Content[0].ToolResponse.Output), "injuries field was omitted or does not contain confirmed information")
						return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewTextPart("Any current injuries?")), FinishReason: ai.FinishReasonStop}, nil
					}
					if calls == 2 {
						return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewTextPart("Here is your draft.")), FinishReason: ai.FinishReasonStop}, nil
					}
					input := map[string]any{"sessionDuration": 10, "workoutFocus": "upper body", "equipment": "bodyweight"}
					if !tc.wantDraft {
						input["injuries"] = "none"
					}
					call := ai.NewToolRequestPart(&ai.ToolRequest{Name: workoutDraftToolName, Ref: fmt.Sprint(calls), Input: input})
					return &ai.ModelResponse{Message: ai.NewModelMessage(call), FinishReason: ai.FinishReasonStop}, nil
				})
			r := &GenkitRuntime{available: true, modelName: "chat", g: g, workoutDraftTool: defineWorkoutDraftTool(g, "synthetic", nil)}
			history := []RuntimeChatMessage{
				{Role: "user", Text: "Make a 10-minute bodyweight upper body workout."},
				{Role: "assistant", Text: "Any current injuries?"},
			}
			if tc.earlier != "" {
				history = append(history, RuntimeChatMessage{Role: "user", Text: tc.earlier})
			}
			done, err := r.StreamChat(ctx, tc.answer, history, func(string) error { return nil })
			require.NoError(t, err)
			if tc.wantDraft {
				require.Equal(t, 1, generations, "calls=%d, text=%q, tools=%v", calls, done.Text, done.ToolCalls)
				require.NotNil(t, done.WorkoutDraft)
			} else {
				require.Nil(t, done.WorkoutDraft)
				require.Equal(t, "Any current injuries?", done.Text)
				require.Zero(t, generations)
			}
		})
	}
}

func TestWorkoutDraftToolHonorsLoadAcrossIndependentRepAndEquipmentScopes(t *testing.T) {
	ctx := context.Background()
	attempt := 0
	g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string {
		attempt++
		extraSet := ""
		if attempt > 1 {
			extraSet = `{"reps":8,"setType":"working"},`
		}
		return fmt.Sprintf(`{"date":"2026-11-23T00:00:00Z","exercises":[{
			"name":"Barbell Bench Press","sets":[{"reps":5,"setType":"warmup"},{"reps":8,"setType":"working"},%s{"reps":6,"setType":"working"}]
		},{"name":"Dumbbell Floor Press","sets":[{"reps":8,"setType":"working"},{"reps":6,"setType":"working"}]}]}`, extraSet)
	})
	tool := defineWorkoutDraftTool(g, "synthetic", nil)
	out, err := tool.RunRaw(ctx, map[string]any{
		"equipment": "bench, barbells up to 45 lb and dumbbells up to 60 lb", "sessionDuration": 10, "workoutFocus": "upper body", "injuries": "none",
		"exerciseWeightLimits": []map[string]any{
			{"equipment": "barbell", "weight": 45.0},
			{"equipment": "dumbbell", "weight": 60.0},
		},
		"loadChoices": []map[string]any{
			{"exercise": "Barbell Bench Press", "kind": "requested", "setType": "warmup", "equipment": "barbell", "weight": 45.0},
			{"exercise": "Barbell Bench Press", "kind": "requested", "setType": "working", "equipment": "dumbbell", "weight": 55.0, "reps": 8},
			{"exercise": "Barbell Bench Press", "kind": "structure", "setType": "working", "count": 3},
			{"exercise": "Barbell Bench Press", "kind": "structure", "setType": "working", "last": true, "reps": 6},
			{"exercise": "Dumbbell Floor Press", "kind": "requested", "setType": "working", "reps": 8, "weight": 55.0},
			{"exercise": "Dumbbell Floor Press", "kind": "structure", "setType": "working", "last": true, "reps": 6},
		},
	})
	require.NoError(t, err)
	raw, err := json.Marshal(out)
	require.NoError(t, err)
	var draft workout.CreateWorkoutRequest
	require.NoError(t, json.Unmarshal(raw, &draft))
	require.Len(t, draft.Exercises, 3)
	require.Equal(t, "Barbell Bench Press", draft.Exercises[0].Name)
	require.Len(t, draft.Exercises[0].Sets, 1)
	assert.Equal(t, new(45.0), draft.Exercises[0].Sets[0].Weight)
	require.Equal(t, "Dumbbell Bench Press", draft.Exercises[1].Name)
	sets := draft.Exercises[1].Sets
	require.Len(t, sets, 3)
	for i, reps := range []int{8, 8, 6} {
		assert.Equal(t, new(55.0), sets[i].Weight)
		assert.Equal(t, reps, sets[i].Reps)
	}
	require.Equal(t, 2, attempt)
	require.Equal(t, new(55.0), draft.Exercises[2].Sets[1].Weight)
}

func TestWorkoutDraftToolKeepsEquipmentAndWorkingSetPositionsScoped(t *testing.T) {
	ctx := user.WithContext(context.Background(), "draft-owner")
	now := time.Now()
	attempts := 0
	g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string {
		attempts++
		row := ""
		if attempts > 1 {
			row = `,{"name":"Dumbbell Row","sets":[{"reps":8,"setType":"working"}]}`
		}
		return fmt.Sprintf(`{"date":%q,"exercises":[
			{"name":"Barbell Bench Press","sets":[{"reps":5,"setType":"warmup"},{"reps":8,"setType":"working"}]},
			{"name":"Deadlift","sets":[{"reps":8,"setType":"working"}]},
			{"name":"Hack Squat","sets":[{"reps":8,"setType":"working"}]},
			{"name":"Bench Press","sets":[{"reps":5,"setType":"warmup"},{"reps":8,"setType":"working"},{"reps":8,"setType":"working"},{"reps":8,"setType":"working"}]},
			{"name":"Lever Arm Hack Squat","sets":[{"reps":8,"setType":"working"}]},
			{"name":"Pec Deck Fly","sets":[{"reps":8,"setType":"working"}]},
			{"name":"Row","sets":[{"reps":8,"setType":"working"}]},
			{"name":"Barbell Row","sets":[{"reps":8,"setType":"working"}]},
			{"name":"Plate-Loaded Leg Press Machine","sets":[{"reps":8,"setType":"working"}]},
			{"name":"Pec Deck Reverse Fly","sets":[{"reps":8,"setType":"working"}]}%s
		]}`, now.Format(time.RFC3339), row)
	})
	history := fixedDraftWeightHistory{"Barbell Bench Press": {{Date: now.AddDate(0, 0, -1), Sets: []workout.SetInput{
		{Weight: new(45.0), Reps: 5, SetType: "warmup"},
		{Weight: new(45.0), Reps: 8, SetType: "working"},
	}}}}
	history["Lever Arm Hack Squat"] = history["Barbell Bench Press"]
	history["Pec Deck Fly"] = history["Barbell Bench Press"]
	history["Plate-Loaded Leg Press Machine"] = history["Barbell Bench Press"]
	tool := defineWorkoutDraftTool(g, "synthetic", history)
	out, err := tool.RunRaw(ctx, map[string]any{
		"equipment": "full gym", "sessionDuration": 20, "workoutFocus": "full body", "injuries": "none",
		"exerciseWeightLimits": []map[string]any{
			{"equipment": "trap bar", "weight": 135.0},
			{"equipment": "pivot arms", "weight": 100.0},
			{"equipment": "iso cradle", "weight": 30.0},
			{"equipment": "pec deck", "weight": 60.0},
			{"equipment": "dumbbell", "weight": 60.0},
		},
		"loadChoices": []map[string]any{
			{"exercise": "Barbell Bench Press", "kind": "structure", "equipment": "dumbbell", "setType": "working"},
			{"exercise": "Deadlift", "kind": "requested", "equipment": "trap bar", "setType": "working", "weight": 155.0},
			{"exercise": "Hack Squat", "kind": "requested", "equipment": "pivot arm", "setType": "working", "weight": 125.0},
			{"exercise": "Bench Press", "kind": "requested", "setType": "working", "weight": 115.0},
			{"exercise": "Bench Press", "kind": "blank", "setType": "working", "set": 2},
			{"exercise": "Lever Arm Hack Squat", "kind": "structure", "equipment": "Smith machine", "setType": "working"},
			{"exercise": "Pec Deck Fly", "kind": "structure", "equipment": "dumbbell", "setType": "working"},
			{"exercise": "Row", "kind": "structure", "equipment": "dumbbell", "setType": "working"},
			{"exercise": "Row", "kind": "reported", "equipment": "dumbbell", "reps": 8, "weight": 45.0},
			{"exercise": "Row", "kind": "reported", "equipment": "barbell", "reps": 8, "weight": 55.0},
			{"exercise": "Barbell Row", "kind": "structure", "equipment": "dumbbell", "setType": "working"},
			{"exercise": "Barbell Row", "kind": "reported", "equipment": "barbell", "reps": 8, "weight": 45.0},
			{"exercise": "Plate-Loaded Leg Press Machine", "kind": "structure", "equipment": "selectorized leg press machine", "setType": "working"},
			{"exercise": "Pec Deck Reverse Fly", "kind": "requested", "weight": 90.0},
			{"exercise": "Dumbbell Row", "kind": "blank", "setType": "all"},
		},
	})
	require.NoError(t, err)
	raw, err := json.Marshal(out)
	require.NoError(t, err)
	var draft workout.CreateWorkoutRequest
	require.NoError(t, json.Unmarshal(raw, &draft))
	assert.Equal(t, new(45.0), draft.Exercises[0].Sets[0].Weight)
	require.Len(t, draft.Exercises[0].Sets, 1)
	assert.Equal(t, "Dumbbell Bench Press", draft.Exercises[1].Name)
	assert.Nil(t, draft.Exercises[1].Sets[0].Weight, "barbell history does not apply to dumbbells")
	assert.Nil(t, draft.Exercises[2].Sets[0].Weight, "trap bar cap blocks the requested load")
	assert.Nil(t, draft.Exercises[3].Sets[0].Weight, "unclassified equipment cap blocks the requested load")
	bench := draft.Exercises[4].Sets
	assert.Equal(t, new(115.0), bench[1].Weight)
	assert.Nil(t, bench[2].Weight)
	assert.Equal(t, new(115.0), bench[3].Weight)
	require.Equal(t, 2, attempts)
	require.Len(t, draft.Exercises, 12)
	assert.Nil(t, draft.Exercises[5].Sets[0].Weight, "lever-arm history does not apply to Smith machine sets")
	assert.Nil(t, draft.Exercises[6].Sets[0].Weight, "unclassified logged equipment does not establish a dumbbell match")
	assert.Equal(t, new(45.0), draft.Exercises[7].Sets[0].Weight, "choose the comparable report and ignore unrelated unclassified caps")
	assert.Nil(t, draft.Exercises[8].Sets[0].Weight, "reported barbell load does not apply to dumbbell sets")
	assert.Equal(t, "Selectorized Leg Press Machine", draft.Exercises[9].Name)
	assert.Nil(t, draft.Exercises[9].Sets[0].Weight, "machines with different loading mechanisms do not share history")
	assert.Nil(t, draft.Exercises[10].Sets[0].Weight, "a specific equipment phrase in the name establishes an explicit cap")
	assert.Equal(t, "Dumbbell Row", draft.Exercises[11].Name)
	assert.Nil(t, draft.Exercises[11].Sets[0].Weight)
}
