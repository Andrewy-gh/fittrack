package aichat

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/require"
)

func TestDraftToolLoggedWeights(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed draft history test in short mode")
	}
	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	defer cleanup()
	const owner, other = "aichat-logged-draft-owner", "aichat-logged-draft-other"
	for _, id := range []string{owner, other} {
		seedAIChatRepositoryTestUser(t, pool, id)
	}
	now := time.Now().UTC()
	day := func(days int) string { return now.AddDate(0, 0, days).Format(time.DateOnly) }
	working := func(lb float64, reps int32) chatDataSetSeed {
		return chatDataSetSeed{weight: new(lb), reps: reps, setType: "working"}
	}
	seedAIChatDataWorkout(t, pool, owner, day(-7), "push", []chatDataExerciseSeed{
		{name: "Barbell Bench Press", sets: []chatDataSetSeed{working(100, 8), working(85, 10)}},
	})
	seedAIChatDataWorkout(t, pool, owner, day(-2), "push", []chatDataExerciseSeed{
		{name: "Barbell Bench Press", sets: []chatDataSetSeed{
			working(115, 8), working(115, 8), working(145, 3), working(150, 3), working(120, 5),
			{weight: new(65.0), reps: 6, setType: "warmup"},
			{reps: 5, setType: "working"},
		}},
		{name: "Bench Press", sets: []chatDataSetSeed{working(60, 8)}},
		{name: "Barbell or Dumbbell Press", sets: []chatDataSetSeed{working(90, 8)}},
	})
	seedAIChatDataWorkout(t, pool, owner, day(7), "push", []chatDataExerciseSeed{
		{name: "Barbell Bench Press", sets: []chatDataSetSeed{working(190, 8)}},
	})
	seedAIChatDataWorkout(t, pool, owner, day(-100), "push", []chatDataExerciseSeed{
		{name: "Kettlebell Press", sets: []chatDataSetSeed{working(35, 8)}},
	})
	seedAIChatDataWorkout(t, pool, other, day(-1), "push", []chatDataExerciseSeed{
		{name: "Barbell Bench Press", sets: []chatDataSetSeed{working(200, 8)}},
		{name: "Dumbbell Bench Press", sets: []chatDataSetSeed{working(80, 8)}},
	})

	const payload = `{"date":"DATE","exercises":[
		{"name":"Barbell Bench Press","sets":[
			{"reps":8,"setType":"working"},{"reps":6,"setType":"warmup"},
			{"reps":8,"setType":"warmup"},{"reps":10,"setType":"working"},
			{"reps":3,"setType":"working"},{"reps":5,"setType":"working"},
			{"reps":8,"setType":"working","weight":0},{"reps":8,"setType":"working","weight":75}]},
		{"name":"Dumbbell Bench Press","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Bench Press","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Barbell or Dumbbell Press","sets":[{"reps":8,"setType":"working"}]},
		{"name":"Kettlebell Press","sets":[{"reps":8,"setType":"working"}]}]}`
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := user.WithContext(context.Background(), owner)
	for _, tc := range []struct {
		name, injuries        string
		enabled               bool
		wantBench, wantWarmup *float64
	}{
		{name: "known equipment", injuries: "none", enabled: true, wantBench: new(115.0), wantWarmup: new(65.0)},
		{name: "uncertain or changed equipment, or weight opt-out", injuries: "none"},
		{name: "active limitation", injuries: "knee pain", enabled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requested := now.Add(time.Hour).Format(time.RFC3339)
			g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string {
				return strings.Replace(payload, "DATE", requested, 1)
			})
			reader := NewRepository(logger, db.New(pool), pool).(draftHistoryReader)
			result, err := defineWorkoutDraftTool(g, "synthetic", reader).RunRaw(ctx, map[string]any{
				"sessionDuration": 25, "workoutFocus": "push", "equipment": "full gym", "injuries": tc.injuries, "useLoggedWeights": tc.enabled,
			})
			require.NoError(t, err)
			raw, err := json.Marshal(result)
			require.NoError(t, err)
			var draft workout.CreateWorkoutRequest
			require.NoError(t, json.Unmarshal(raw, &draft))
			require.Equal(t, requested, draft.Date)
			require.Len(t, draft.Exercises, 5)
			for i, want := range []*float64{tc.wantBench, tc.wantWarmup, nil, nil, nil, nil, new(0.0), new(75.0)} {
				require.Equal(t, want, draft.Exercises[0].Sets[i].Weight, "set %d", i+1)
			}
			for _, e := range draft.Exercises[1:] {
				require.Nil(t, e.Sets[0].Weight, e.Name)
			}
		})
	}
}
