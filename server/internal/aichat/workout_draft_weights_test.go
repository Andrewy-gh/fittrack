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
		"Dumbbell Bench Press":          reported(50),
		"Dumbbell Row":                  reported(50),
		"Barbell Row":                   reported(95),
		"Resistance Band Dumbbell Curl": reported(50),
		"Kettlebell Swing":              reported(25),
		"Overhead Press":                reported(65),
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
		{"name":"Overhead Press","sets":[{"reps":4,"setType":"warmup"},{"reps":8,"setType":"working"}]},
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
			`"Overhead Press","sets":[{"reps":4,"setType":"warmup"},`, `"Overhead Press","sets":[`, 1)},
		{name: "wrong reps on blank warmup", payload: strings.Replace(payload,
			`"Overhead Press","sets":[{"reps":4,"setType":"warmup"}`, `"Overhead Press","sets":[{"reps":5,"setType":"warmup"}`, 1)},
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
					{"equipment": "dumbbell", "weight": 35.0},
					{"equipment": "barbell", "weight": 120.0},
					fixedLimit,
				},
				"loadChoices": []map[string]any{
					{"exercise": "Barbell Bench Press", "kind": "requested", "weight": 65.0, "setType": "warmup"},
					{"exercise": "Barbell Bench Press", "kind": "requested", "weight": 115.0, "setType": "working", "reps": 8},
					{"exercise": "Barbell Bench Press", "kind": "blank", "setType": "working", "last": true},
					{"exercise": "Pull-Ups", "kind": "requested", "weight": 0.0},
					{"exercise": "Overhead Press", "kind": "blank", "setType": "warmup", "reps": 4},
					{"exercise": "Overhead Press", "kind": "structure", "setType": "all", "equipment": "barbell"},
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
	weight := draft.Exercises[0].Sets[1].Weight
	if weight != nil {
		t.Errorf("dumbbell work above the 35 lb cap received %.0f lb", *weight)
	}
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
	for _, known := range []bool{true, false} {
		t.Run(fmt.Sprintf("known injuries=%t", known), func(t *testing.T) {
			ctx := context.Background()
			generations := 0
			g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string {
				generations++
				return ordinaryDraftJSON
			})
			calls := 0
			genkit.DefineModel(g, "chat", &ai.ModelOptions{Supports: &ai.ModelSupports{Tools: true, Multiturn: true, SystemRole: true}},
				func(_ context.Context, req *ai.ModelRequest, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
					calls++
					if calls == 2 {
						require.Zero(t, generations)
						last := req.Messages[len(req.Messages)-1]
						require.Contains(t, fmt.Sprint(last.Content[0].ToolResponse.Output), "injuries field was omitted")
						if !known {
							return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewTextPart("Any current injuries?")), FinishReason: ai.FinishReasonStop}, nil
						}
					}
					if calls == 3 {
						return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewTextPart("Here is your draft.")), FinishReason: ai.FinishReasonStop}, nil
					}
					input := map[string]any{"sessionDuration": 10, "workoutFocus": "upper body", "equipment": "bodyweight"}
					if calls == 2 {
						input["injuries"] = "none"
					}
					call := ai.NewToolRequestPart(&ai.ToolRequest{Name: workoutDraftToolName, Ref: fmt.Sprint(calls), Input: input})
					return &ai.ModelResponse{Message: ai.NewModelMessage(call), FinishReason: ai.FinishReasonStop}, nil
				})
			r := &GenkitRuntime{available: true, modelName: "chat", g: g, workoutDraftTool: defineWorkoutDraftTool(g, "synthetic", nil)}
			prompt := "Make a 10-minute bodyweight upper body workout."
			if known {
				prompt += " No injuries."
			}
			done, err := r.StreamChat(ctx, prompt, nil, func(string) error { return nil })
			require.NoError(t, err)
			if known {
				require.NotNil(t, done.WorkoutDraft)
				require.Equal(t, 1, generations)
			} else {
				require.Nil(t, done.WorkoutDraft)
				require.Equal(t, "Any current injuries?", done.Text)
				require.Zero(t, generations)
			}
		})
	}
}
