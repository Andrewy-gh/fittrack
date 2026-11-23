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
	equipment := normalizeQualityText(input.Equipment)
	bodyweight := input.BodyweightOnly || ((equipment == "bodyweight" ||
		hasAnyQualityTerm(equipment, "bodyweight only", "no equipment")) &&
		!hasExternalDraftEquipment(equipment))
	suppress := bodyweight || input.SuppressSuggestedWeights || isEquipmentOptionalWorkout(input) || activeInjuryText(input.Injuries) != "" ||
		hasAnyQualityTerm(normalizeQualityText(input.FitnessGoal+" "+input.WorkoutFocus), "deload", "recovery", "reduced load")

	eligible := make(map[string]struct{})
	for i := range draft.Exercises {
		exercise := &draft.Exercises[i]
		for j := range exercise.Sets {
			set := &exercise.Sets[j]
			choice, _ := draftLoadAndRepChoices(input.LoadChoices, *exercise, j)
			equipment := draftSetEquipment(input.LoadChoices, *exercise, j)
			limits := draftExerciseWeightLimits(input, exercise.Name, equipment)
			switch {
			case choice.Kind == "requested":
				if allowedDraftWeight(choice.Weight, limits.explicitCap) {
					set.Weight = choice.Weight
				}
			case choice.Kind == "blank":
				continue
			case suppress || limits.fixed:
				set.DraftBodyweight = bodyweight
			case choice.Kind == "reported":
				if allowedDraftWeight(choice.Weight, limits.inferredCap) {
					set.Weight = choice.Weight
				}
			case reader != nil && !input.SkipHistoryWeights:
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
		var progressionCap *float64
		var fixedWorkingEquipment bool
		for j := range exercise.Sets {
			set := &exercise.Sets[j]
			choice, _ := draftLoadAndRepChoices(input.LoadChoices, *exercise, j)
			equipment := draftSetEquipment(input.LoadChoices, *exercise, j)
			limits := draftExerciseWeightLimits(input, exercise.Name, equipment)
			if set.SetType == "working" {
				progressionCap = lowerDraftCap(progressionCap, limits.inferredCap)
				fixedWorkingEquipment = fixedWorkingEquipment || limits.fixed
			}
			if limits.fixed || set.Weight != nil || choice.Kind != "" {
				continue
			}
			weight := matchingDraftWeight(latest.Sets, *set)
			if allowedDraftWeight(weight, limits.inferredCap) {
				set.Weight = weight
				set.WeightSourceDate = &sourceDate
			}
		}
		if !fixedWorkingEquipment {
			exercise.DraftProgression = draftProgression(input, progressionCap, exercise.Sets, sessions, before)
		}
	}
}

func draftLoadAndRepChoices(choices []workoutDraftLoadChoice, exercise workout.ExerciseInput, index int) (load, reps workoutDraftLoadChoice) {
	var reported workoutDraftLoadChoice
	loadSpecificity, repSpecificity := -1, -1
	for _, choice := range choices {
		if !strings.EqualFold(strings.TrimSpace(choice.Exercise), exercise.Name) ||
			!draftLoadScopeMatches(choice, exercise.Sets, index) {
			continue
		}
		if choice.Kind == "reported" {
			if exercise.Sets[index].Reps == choice.Reps {
				reported = choice
			}
			continue
		}
		score := draftLoadScopeSpecificity(choice)
		if (choice.Kind == "requested" || choice.Kind == "blank") &&
			(score > loadSpecificity || (score == loadSpecificity && choice.Kind == "blank")) {
			load, loadSpecificity = choice, score
		}
		if choice.Reps > 0 && (choice.Kind == "requested" || choice.Kind == "blank" || choice.Kind == "structure") &&
			(score > repSpecificity || (score == repSpecificity && choice.Kind == "blank")) {
			reps, repSpecificity = choice, score
		}
	}
	if load.Kind == "" {
		load = reported
	}
	return load, reps
}

func draftLoadScopeSpecificity(choice workoutDraftLoadChoice) int {
	if choice.Set > 0 || choice.Last {
		return 2
	}
	if kind := draftLoadSetType(choice); kind != "" && kind != "all" {
		return 1
	}
	return 0
}

func draftLoadSetType(choice workoutDraftLoadChoice) string {
	setType := strings.ToLower(strings.TrimSpace(choice.SetType))
	if choice.Kind == "reported" && (setType == "" || setType == "all") {
		return "working"
	}
	if setType == "" && choice.Kind != "blank" && choice.Set == 0 &&
		(choice.Weight == nil || *choice.Weight != 0) {
		return "working"
	}
	return setType
}

func draftLoadScopeMatches(choice workoutDraftLoadChoice, sets []workout.SetInput, index int) bool {
	if choice.Set != 0 && choice.Set != index+1 {
		return false
	}
	setType := draftLoadSetType(choice)
	typeMatches := func(set workout.SetInput) bool {
		return setType == "" || setType == "all" || setType == set.SetType
	}
	if !typeMatches(sets[index]) {
		return false
	}
	if choice.Last {
		for _, later := range sets[index+1:] {
			if typeMatches(later) {
				return false
			}
		}
	}
	return true
}

func (choice workoutDraftLoadChoice) hasSetScope() bool {
	kind := draftLoadSetType(choice)
	return choice.Reps > 0 || choice.Set > 0 || choice.Last || kind == "warmup" || kind == "working"
}

type draftWeightLimits struct {
	explicitCap *float64
	inferredCap *float64
	fixed       bool
}

// Resolve equipment separately from the weight choice. A warm-up-only
// assignment cannot establish the equipment for a working set.
func draftSetEquipment(choices []workoutDraftLoadChoice, exercise workout.ExerciseInput, index int) string {
	var confirmed []string
	specificity := -1
	ambiguous := false
	for _, choice := range choices {
		if choice.Kind == "reported" || choice.Equipment == "" ||
			!strings.EqualFold(strings.TrimSpace(choice.Exercise), exercise.Name) ||
			!draftLoadScopeMatches(choice, exercise.Sets, index) {
			continue
		}
		score := draftLoadScopeSpecificity(choice)
		classes := draftEquipmentClasses(choice.Equipment)
		slices.Sort(classes)
		classes = slices.Compact(classes)
		if score > specificity {
			confirmed, specificity, ambiguous = classes, score, false
		} else if score == specificity && !slices.Equal(confirmed, classes) {
			ambiguous = true
		}
	}
	if ambiguous {
		return ""
	}
	return strings.Join(confirmed, " ")
}

func draftExerciseWeightLimits(input WorkoutGenerationToolInput, name, confirmedEquipment string) draftWeightLimits {
	limits := draftWeightLimits{explicitCap: input.MaxWeight, inferredCap: input.MaxWeight}
	if confirmedEquipment == "" {
		confirmedEquipment = name
	}
	exerciseEquipment := draftEquipmentClasses(confirmedEquipment)
	for _, item := range input.ExerciseWeightLimits {
		classes := draftEquipmentClasses(item.Equipment)
		confirmed := strings.EqualFold(strings.TrimSpace(item.Exercise), name)
		for _, class := range classes {
			confirmed = confirmed || slices.Contains(exerciseEquipment, class)
		}
		uncertain := strings.TrimSpace(item.Equipment) != "" && (len(exerciseEquipment) == 0 || len(classes) == 0)
		if confirmed {
			limits.explicitCap = lowerDraftCap(limits.explicitCap, item.Weight)
		}
		if confirmed || uncertain {
			limits.inferredCap = lowerDraftCap(limits.inferredCap, item.Weight)
			limits.fixed = limits.fixed || item.Fixed
		}
	}
	return limits
}

func lowerDraftCap(cap, weight *float64) *float64 {
	if weight != nil && (cap == nil || *weight < *cap) {
		return weight
	}
	return cap
}

func hasExternalDraftEquipment(equipment string) bool {
	for _, word := range strings.Fields(equipment) {
		for _, class := range draftEquipmentClasses(word) {
			if class != "bodyweight" && hasAvailableEquipmentTerm(equipment, word) {
				return true
			}
		}
	}
	return false
}

func draftEquipmentClasses(value string) []string {
	var classes []string
	words := strings.Fields(normalizeQualityText(value))
	for i, word := range words {
		class := strings.TrimSuffix(word, "s")
		if class == "machine" && i > 0 && (words[i-1] == "smith" || words[i-1] == "cable") {
			continue
		}
		switch class {
		case "db", "dumbbell":
			classes = append(classes, "dumbbell")
		case "bb", "barbell":
			classes = append(classes, "barbell")
		case "kb", "kettlebell":
			classes = append(classes, "kettlebell")
		case "cable", "smith", "machine", "band", "vest", "sandbag", "bodyweight":
			classes = append(classes, class)
		}
	}
	return classes
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

func draftProgression(input WorkoutGenerationToolInput, cap *float64, draftSets []workout.SetInput, history []DraftWeightSession, now time.Time) *recommendation.Plan {
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
	if plan == nil || !allowedDraftWeight(plan.Weight, cap) {
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
