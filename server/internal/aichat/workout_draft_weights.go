package aichat

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"math"
	"slices"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

type DraftWeightSession struct {
	Date time.Time
	Sets []workout.SetInput
}

type draftWeightHistoryReader interface {
	LatestDraftExerciseSets(ctx context.Context, userID string, names []string, after, before time.Time) (map[string][]DraftWeightSession, error)
}

func (r *repository) LatestDraftExerciseSets(ctx context.Context, userID string, names []string, after, before time.Time) (map[string][]DraftWeightSession, error) {
	rows, err := r.queries.ListLatestDraftExerciseSets(ctx, db.ListLatestDraftExerciseSetsParams{
		UserID: userID, ExerciseNames: names, AfterDate: timePtrToPg(&after), BeforeDate: timePtrToPg(&before),
	})
	if err != nil {
		return nil, fmt.Errorf("list latest draft exercise sets: %w", err)
	}

	history := make(map[string][]DraftWeightSession, len(names))
	lastWorkout := make(map[string]int32, len(names))
	for _, row := range rows {
		sessions := history[row.ExerciseName]
		if len(sessions) == 0 || lastWorkout[row.ExerciseName] != row.WorkoutID {
			sessions = append(sessions, DraftWeightSession{Date: row.WorkoutDate.Time})
			lastWorkout[row.ExerciseName] = row.WorkoutID
		}
		set := workout.SetInput{Reps: int(row.Reps), SetType: row.SetType}
		if weight, err := row.Weight.Float64Value(); err == nil && weight.Valid && validDraftWeight(weight.Float64) {
			set.Weight = &weight.Float64
		}
		sessions[len(sessions)-1].Sets = append(sessions[len(sessions)-1].Sets, set)
		history[row.ExerciseName] = sessions
	}
	return history, nil
}

func applyDraftWeights(ctx context.Context, reader draftWeightHistoryReader, input WorkoutGenerationToolInput, draft *workout.CreateWorkoutRequest, now time.Time) {
	bodyweight := input.BodyweightOnly || hasAnyQualityTerm(normalizeQualityText(input.Equipment), "bodyweight only", "no equipment") ||
		strings.EqualFold(strings.TrimSpace(input.Equipment), "bodyweight")
	suppress := bodyweight || input.SkipHistoryWeights || isEquipmentOptionalWorkout(input) || activeInjuryText(input.Injuries) != "" ||
		hasAnyQualityTerm(normalizeQualityText(input.FitnessGoal+" "+input.WorkoutFocus), "deload", "recovery", "reduced load")

	eligible := make(map[string]struct{})
	for i := range draft.Exercises {
		exercise := &draft.Exercises[i]
		for j := range exercise.Sets {
			set := &exercise.Sets[j]
			choice := draftLoadChoice(input.LoadChoices, exercise.Name, *set, j+1)
			switch {
			case choice.Kind == "requested":
				if allowedDraftWeight(choice.Weight, input.MaxWeight) {
					set.Weight = choice.Weight
				}
			case choice.Kind == "blank":
				continue
			case suppress:
				set.DraftBodyweight = bodyweight
			case choice.Kind == "reported":
				if allowedDraftWeight(choice.Weight, input.MaxWeight) {
					set.Weight = choice.Weight
				}
			case reader != nil:
				eligible[exercise.Name] = struct{}{}
			}
		}
	}

	userID, ok := user.Current(ctx)
	if !ok || len(eligible) == 0 {
		return
	}
	plannedDate, _ := time.Parse(time.RFC3339, draft.Date)
	before := now
	if !plannedDate.IsZero() && plannedDate.Before(before) {
		before = plannedDate
	}
	after := plannedDate.AddDate(0, 0, -90)
	if plannedDate.IsZero() {
		after = before.AddDate(0, 0, -90)
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	history, err := reader.LatestDraftExerciseSets(lookupCtx, userID, slices.Sorted(maps.Keys(eligible)), after, before)
	if err != nil {
		slog.Warn("could not look up weights for workout draft", "error", err)
		return
	}

	for i := range draft.Exercises {
		exercise := &draft.Exercises[i]
		sessions := history[exercise.Name]
		if len(sessions) == 0 {
			continue
		}
		latest := sessions[0]
		sourceDate := latest.Date.Format(time.RFC3339)
		for j := range exercise.Sets {
			set := &exercise.Sets[j]
			if set.Weight != nil || draftLoadChoice(input.LoadChoices, exercise.Name, *set, j+1).Kind != "" {
				continue
			}
			weight := matchingDraftWeight(latest.Sets, *set)
			if allowedDraftWeight(weight, input.MaxWeight) {
				set.Weight = weight
				set.WeightSourceDate = &sourceDate
			}
		}
		exercise.DraftProgression = draftProgression(input, exercise.Sets, sessions, before)
	}
}

func draftLoadChoice(choices []workoutDraftLoadChoice, name string, set workout.SetInput, position int) workoutDraftLoadChoice {
	var blank, reported, requested workoutDraftLoadChoice
	for _, choice := range choices {
		if !strings.EqualFold(strings.TrimSpace(choice.Exercise), name) ||
			(choice.Set != 0 && choice.Set != position) {
			continue
		}
		setType := strings.ToLower(strings.TrimSpace(choice.SetType))
		if setType == "" && choice.Kind != "blank" && choice.Set == 0 {
			setType = "working"
		}
		if setType != "" && setType != "all" && setType != set.SetType {
			continue
		}
		switch choice.Kind {
		case "requested":
			if requested.Kind == "" || choice.Set != 0 || (requested.Set == 0 && choice.SetType != "") {
				requested = choice
			}
		case "blank":
			blank = choice
		case "reported":
			if set.SetType == "working" && set.Reps == choice.Reps {
				reported = choice
			}
		}
	}
	if requested.Kind != "" {
		return requested
	}
	if blank.Kind != "" {
		return blank
	}
	return reported
}

func matchingDraftWeight(sets []workout.SetInput, target workout.SetInput) *float64 {
	var weight *float64
	for _, prior := range sets {
		if prior.SetType != target.SetType || prior.Reps != target.Reps {
			continue
		}
		if prior.Weight == nil || *prior.Weight == 0 || (weight != nil && *weight != *prior.Weight) {
			return nil
		}
		weight = prior.Weight
	}
	return weight
}

func validDraftWeight(weight float64) bool {
	return !math.IsNaN(weight) && weight >= 0 && weight <= 999999999.9
}

func allowedDraftWeight(weight, cap *float64) bool {
	return weight != nil && validDraftWeight(*weight) && (cap == nil || *weight <= *cap)
}

func draftProgression(input WorkoutGenerationToolInput, draftSets []workout.SetInput, history []DraftWeightSession, now time.Time) *recommendation.Plan {
	context := recommendation.TrainingContext{Goal: strings.ToLower(strings.TrimSpace(input.FitnessGoal)), Experience: strings.ToLower(strings.TrimSpace(input.FitnessLevel))}
	for _, session := range history {
		working := recommendation.Session{Date: session.Date}
		for _, set := range session.Sets {
			if set.SetType == "working" {
				working.Sets = append(working.Sets, recommendation.WorkingSet{Weight: set.Weight, Reps: set.Reps})
			}
		}
		context.Sessions = append(context.Sessions, working)
	}
	result := recommendation.RecommendTraining(now, context, "normal")
	plan := result.Plan
	if plan == nil || !allowedDraftWeight(plan.Weight, input.MaxWeight) {
		return nil
	}
	workingCount := 0
	for _, set := range draftSets {
		if set.SetType == "working" {
			if set.WeightSourceDate == nil || set.Weight == nil || *plan.Weight <= *set.Weight || plan.Reps >= set.Reps {
				return nil
			}
			workingCount++
		}
	}
	if workingCount != plan.Sets {
		return nil
	}
	return plan
}
