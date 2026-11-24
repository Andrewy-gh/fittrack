package aichat

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"time"
	"unicode"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

type draftSetKey struct {
	name    string
	reps    int
	setType string
}

type draftHistoryReader interface {
	draftLoggedWeights(ctx context.Context, userID string, names []string, start, end time.Time) (map[draftSetKey]float64, error)
}

func (r *repository) draftLoggedWeights(ctx context.Context, userID string, names []string, start, end time.Time) (map[draftSetKey]float64, error) {
	rows, err := r.queries.ListDraftLoggedWeights(ctx, db.ListDraftLoggedWeightsParams{
		UserID: userID, ExerciseNames: names, HistoryStart: timePtrToPg(&start), HistoryEnd: timePtrToPg(&end),
	})
	if err != nil {
		return nil, err
	}
	weights := make(map[draftSetKey]float64, len(rows))
	for _, row := range rows {
		load, err := row.Weight.Float64Value()
		if err == nil && load.Valid && load.Float64 > 0 && load.Float64 <= 999999999.9 {
			weights[draftSetKey{row.Name, int(row.Reps), row.SetType}] = load.Float64
		}
	}
	return weights, nil
}

// Only explicitly named, conventional free-weight variants qualify. A generic
// lift or machine name does not establish which equipment produced the load.
func loggedWeightExercise(name, equipment string) bool {
	words := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
	if len(words) < 2 {
		return false
	}
	switch words[0] {
	case "barbell", "dumbbell", "kettlebell":
	default:
		return false
	}
	// A free-form equipment description is not an assurance that loads
	// are unrestricted (e.g. "dumbbells up to 25 lb"). Admit only plain
	// known contexts and match a named equipment type to the exercise.
	switch normalizeQualityText(equipment) {
	case "gym", "full gym":
	case words[0], words[0] + "s":
	default:
		return false
	}
	for _, word := range words[1:] {
		switch word {
		case "or", "and", "barbell", "barbells", "dumbbell", "dumbbells", "kettlebell", "kettlebells", "smith", "machine", "machines", "cable", "cables", "band", "bands", "resistance", "bodyweight":
			return false
		}
	}
	return !strings.ContainsAny(name, "/&+()")
}

func attachLoggedDraftWeights(ctx context.Context, reader draftHistoryReader, input WorkoutGenerationToolInput, draft *workout.CreateWorkoutRequest, now time.Time) {
	userID, ok := user.Current(ctx)
	if !ok || reader == nil || !input.UseLoggedWeights || activeInjuryText(input.Injuries) != "" || isEquipmentOptionalWorkout(input) {
		return
	}

	names := make(map[string]struct{})
	for _, e := range draft.Exercises {
		if loggedWeightExercise(e.Name, input.Equipment) && slices.ContainsFunc(e.Sets, func(s workout.SetInput) bool { return s.Weight == nil }) {
			names[e.Name] = struct{}{}
		}
	}
	if len(names) == 0 {
		return
	}

	planned, err := time.Parse(time.RFC3339, draft.Date)
	if err != nil {
		return
	}
	start, end := now.Add(-90*24*time.Hour), now
	if planned.Before(end) {
		end = planned
	}
	if end.Before(start) {
		return
	}
	lookup, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	weights, err := reader.draftLoggedWeights(lookup, userID, slices.Sorted(maps.Keys(names)), start, end)
	if err != nil {
		slog.WarnContext(ctx, "workout draft history unavailable", "error", err)
		return
	}
	for i := range draft.Exercises {
		e := &draft.Exercises[i]
		for j := range e.Sets {
			s := &e.Sets[j]
			if weight, found := weights[draftSetKey{e.Name, s.Reps, s.SetType}]; found && s.Weight == nil {
				s.Weight = &weight
			}
		}
	}
}
