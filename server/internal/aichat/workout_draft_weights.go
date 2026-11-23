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
	"unicode"

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

func applyDraftWeights(ctx context.Context, reader draftWeightHistoryReader, input WorkoutGenerationToolInput, draft *workout.CreateWorkoutRequest, now, before time.Time) {
	equipment := normalizeQualityText(input.Equipment)
	bodyweight := input.BodyweightOnly || ((equipment == "bodyweight" ||
		hasAnyQualityTerm(equipment, "bodyweight only", "no equipment")) &&
		!hasExternalDraftEquipment(equipment))
	suppress := bodyweight || input.SuppressSuggestedWeights || isEquipmentOptionalWorkout(input) || activeInjuryText(input.Injuries) != "" ||
		hasAnyQualityTerm(normalizeQualityText(input.FitnessGoal+" "+input.WorkoutFocus), "deload", "recovery", "reduced load")

	type setPolicy struct {
		eligible bool
		limits   draftWeightLimits
	}
	type variant struct {
		exercise workout.ExerciseInput
		policies []setPolicy
	}
	var variants []variant
	eligible := make(map[string]struct{})
	for _, original := range draft.Exercises {
		variantIndex := make(map[string]int)
		for j, set := range original.Sets {
			choice, _ := draftLoadAndRepChoices(input.LoadChoices, original, j)
			assignment, unambiguous := draftSetEquipment(input.LoadChoices, original, j)
			name := draftExerciseVariant(original.Name, assignment, unambiguous)
			key := strings.ToLower(name)
			index, exists := variantIndex[key]
			if !exists {
				index = len(variants)
				variantIndex[key] = index
				variants = append(variants, variant{exercise: workout.ExerciseInput{Name: name}})
			}
			current := &variants[index]
			limits := draftExerciseWeightLimits(input, original.Name, name, assignment)
			policy := setPolicy{limits: limits}
			switch {
			case !unambiguous:
				// Never attach a load to a conflicting equipment assignment.
			case choice.Kind == "requested":
				if allowedDraftWeight(choice.Weight, limits.explicitCap) {
					set.Weight = choice.Weight
				}
			case choice.Kind == "blank":
			case suppress || limits.fixed:
				set.DraftBodyweight = bodyweight
			case choice.Kind == "reported":
				if allowedDraftWeight(choice.Weight, limits.inferredCap) {
					set.Weight = choice.Weight
				}
			case reader != nil && !input.SkipHistoryWeights:
				// If the name had to change, do not borrow history from the old
				// name, or from a name we inferred from a free-text assignment.
				// A subsequent draft naming this variant can use its own history.
				policy.eligible = current.exercise.Name == original.Name
				if policy.eligible {
					eligible[current.exercise.Name] = struct{}{}
				}
			}
			current.exercise.Sets = append(current.exercise.Sets, set)
			current.policies = append(current.policies, policy)
		}
	}

	var history map[string][]DraftWeightSession
	after := now.UTC().AddDate(0, 0, -90)
	if userID, ok := user.Current(ctx); ok && len(eligible) > 0 && !before.Before(after) {
		lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		var err error
		history, err = reader.LatestDraftExerciseSets(lookupCtx, userID, slices.Sorted(maps.Keys(eligible)), after, before)
		if err != nil {
			slog.Warn("could not look up weights for workout draft", "error", err)
			history = nil
		}
	}

	draft.Exercises = make([]workout.ExerciseInput, 0, len(variants))
	for _, pair := range variants {
		exercise := pair.exercise
		sessions := history[exercise.Name]
		if len(sessions) > 0 {
			latest := sessions[0]
			sourceDate := latest.Date.Format(time.RFC3339)
			var progressionCap *float64
			for j := range exercise.Sets {
				set, policy := &exercise.Sets[j], pair.policies[j]
				if set.SetType == "working" {
					progressionCap = lowerDraftCap(progressionCap, policy.limits.inferredCap)
				}
				if policy.eligible {
					if weight := matchingDraftWeight(latest.Sets, *set); allowedDraftWeight(weight, policy.limits.inferredCap) {
						set.Weight = weight
						set.WeightSourceDate = &sourceDate
					}
				}
			}
			exercise.DraftProgression = draftProgression(input, progressionCap, exercise.Sets, sessions, before)
		}
		draft.Exercises = append(draft.Exercises, exercise)
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
			assignment, unambiguous := draftSetEquipment(choices, exercise, index)
			if unambiguous && exercise.Sets[index].Reps == choice.Reps &&
				draftReportedEquipmentMatches(exercise.Name, choice.Equipment, assignment) {
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
	setType := draftLoadSetType(choice)
	typeMatches := func(set workout.SetInput) bool {
		return setType == "" || setType == "all" || setType == set.SetType
	}
	if !typeMatches(sets[index]) {
		return false
	}
	if choice.Set != 0 {
		position := 0
		for _, prior := range sets[:index+1] {
			if typeMatches(prior) {
				position++
			}
		}
		if position != choice.Set {
			return false
		}
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
func draftSetEquipment(choices []workoutDraftLoadChoice, exercise workout.ExerciseInput, index int) (string, bool) {
	var confirmed string
	specificity := -1
	ambiguous := false
	for _, choice := range choices {
		if choice.Kind == "reported" || choice.Equipment == "" ||
			!strings.EqualFold(strings.TrimSpace(choice.Exercise), exercise.Name) ||
			!draftLoadScopeMatches(choice, exercise.Sets, index) {
			continue
		}
		score := draftLoadScopeSpecificity(choice)
		equipment := normalizeQualityText(choice.Equipment)
		if score > specificity {
			confirmed, specificity, ambiguous = equipment, score, false
		} else if score == specificity && draftEquipmentIdentity(confirmed) != draftEquipmentIdentity(equipment) {
			ambiguous = true
		}
	}
	return confirmed, !ambiguous
}

// A class is not an identity: a selectorized machine and a plate-loaded
// machine must not compare equal just because both are machines.
func draftEquipmentIdentity(value string) string {
	switch label := draftEquipmentLabel(value); label {
	case "db":
		return "dumbbell"
	case "bb":
		return "barbell"
	case "kb":
		return "kettlebell"
	case "hex bar":
		return "trap bar"
	case "smith machine":
		return "smith"
	case "cable machine":
		return "cable"
	default:
		return label
	}
}

// Keep free-text labels comparable without losing distinctions between equipment.
func draftEquipmentLabel(value string) string {
	words := strings.FieldsFunc(normalizeQualityText(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	if len(words) > 0 {
		i := len(words) - 1
		if !strings.HasSuffix(words[i], "ss") {
			words[i] = strings.TrimSuffix(words[i], "s")
		}
	}
	return strings.Join(words, " ")
}

func draftReportedEquipmentMatches(name, reported, assignment string) bool {
	if reported == "" {
		return draftHistoryEquipmentMatches(name, assignment)
	}
	if assignment == "" {
		if len(draftEquipmentClasses(name)) == 0 {
			return true
		}
		return draftNameIdentifiesEquipment(name, reported)
	}
	return draftEquipmentIdentity(reported) == draftEquipmentIdentity(assignment)
}

func draftHistoryEquipmentMatches(exerciseName, assignment string) bool {
	return assignment == "" || draftNameIdentifiesEquipment(exerciseName, assignment)
}

func draftNameIdentifiesEquipment(name, equipment string) bool {
	label := draftEquipmentIdentity(equipment)
	if label == "" {
		return false
	}
	if label == "machine" {
		return slices.Contains(draftEquipmentClasses(name), "machine")
	}
	// Keep plural normalization at the word boundary and allow common
	// abbreviations only for supported, single-class labels.
	words := strings.Fields(draftEquipmentLabel(name))
	for i := range words {
		for j := i + 1; j <= len(words); j++ {
			if draftEquipmentIdentity(strings.Join(words[i:j], " ")) == label {
				return true
			}
		}
	}
	return false
}

// The display name also owns saved set history. Keep a customary name for
// familiar variants and use a qualified name for other explicit assignments.
func draftExerciseVariant(name, equipment string, unambiguous bool) string {
	if unambiguous && draftHistoryEquipmentMatches(name, equipment) {
		return name
	}
	base := name
	for _, prefix := range []string{"plate-loaded ", "plate loaded ", "selectorized ", "lever arm ", "trap bar ", "hex bar ", "barbell ", "dumbbell ", "kettlebell ", "smith machine ", "smith ", "cable ", "bodyweight ", "bb ", "db ", "kb "} {
		if strings.HasPrefix(strings.ToLower(base), prefix) {
			base = base[len(prefix):]
			break
		}
	}
	label := draftEquipmentIdentity(equipment)
	switch {
	case !unambiguous || label == "":
		label = "equipment unspecified"
	case label == "barbell" || label == "dumbbell" || label == "kettlebell" || label == "cable" || label == "bodyweight":
		return draftVariantName(titleDraftEquipment(label) + " " + base)
	}
	if strings.HasSuffix(strings.ToLower(base), " machine") && strings.HasSuffix(label, " machine") {
		base = strings.TrimSpace(base[:len(base)-len(" machine")])
		if strings.Contains(" "+label+" ", " "+draftEquipmentLabel(base)+" ") {
			return draftVariantName(titleDraftEquipment(label))
		}
	}
	return draftVariantName(base + " (" + titleDraftEquipment(label) + ")")
}

func titleDraftEquipment(label string) string {
	words := strings.Fields(label)
	for i, word := range words {
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func draftVariantName(name string) string {
	// The create request limits the complete persisted name to 256 characters.
	if runes := []rune(name); len(runes) > 256 {
		// Keep the distinguishing suffix when a model supplied a very long name.
		return string(runes[:120]) + "…" + string(runes[len(runes)-135:])
	}
	return name
}

func draftExerciseWeightLimits(input WorkoutGenerationToolInput, name, savedName, confirmedEquipment string) draftWeightLimits {
	limits := draftWeightLimits{explicitCap: input.MaxWeight, inferredCap: input.MaxWeight}
	hasAssignment := strings.TrimSpace(confirmedEquipment) != ""
	if !hasAssignment {
		confirmedEquipment = name
	}
	exerciseEquipment := draftEquipmentClasses(confirmedEquipment)
	for _, item := range input.ExerciseWeightLimits {
		classes := draftEquipmentClasses(item.Equipment)
		label := draftEquipmentIdentity(item.Equipment)
		target := strings.TrimSpace(item.Exercise)
		confirmed := strings.EqualFold(target, name) || strings.EqualFold(target, savedName) ||
			(label != "" && (label == draftEquipmentIdentity(confirmedEquipment) || draftNameIdentifiesEquipment(confirmedEquipment, item.Equipment) ||
				draftNameIdentifiesEquipment(savedName, item.Equipment)))
		if len(classes) == 1 && label == classes[0] {
			confirmed = confirmed || slices.Contains(exerciseEquipment, classes[0])
		}
		uncertain := strings.TrimSpace(item.Equipment) != "" && (!hasAssignment && len(exerciseEquipment) == 0)
		// A generic assignment can describe the more specific machine
		// named by the exercise. Keep caps from that name unless an
		// assignment identifies incompatible equipment.
		if !confirmed && label != "" && draftNameIdentifiesEquipment(name, item.Equipment) &&
			(!hasAssignment || slices.Contains(exerciseEquipment, "machine")) {
			confirmed = true
		}
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
	words := strings.Fields(equipment)
	for i, word := range words {
		for _, class := range draftEquipmentClasses(word) {
			if class != "bodyweight" && hasAvailableEquipmentTerm(equipment, word) {
				return true
			}
		}
		if i+1 < len(words) {
			phrase := word + " " + words[i+1]
			if len(draftEquipmentClasses(phrase)) > len(draftEquipmentClasses(word))+len(draftEquipmentClasses(words[i+1])) &&
				hasAvailableEquipmentTerm(equipment, phrase) {
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
		if (class == "trap" || class == "hex") && i+1 < len(words) && strings.TrimSuffix(words[i+1], "s") == "bar" {
			classes = append(classes, "trap bar")
		}
		if class == "lever" && i+1 < len(words) && strings.TrimSuffix(words[i+1], "s") == "arm" {
			classes = append(classes, "lever arm")
		}
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
