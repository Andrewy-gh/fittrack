package aichat

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

const (
	minNormalDurationShare = 0.5
)

type workoutDraftQualityIssue struct {
	message string
}

func validateWorkoutDraftQuality(input WorkoutGenerationToolInput, draft *workout.CreateWorkoutRequest) error {
	if draft == nil {
		return fmt.Errorf("workout draft is required")
	}

	issues := make([]workoutDraftQualityIssue, 0, 4)
	issues = append(issues, validateDraftDurationQuality(input, draft)...)
	issues = append(issues, validateDraftEquipmentQuality(input, draft)...)
	issues = append(issues, validateDraftInjuryQuality(input, draft)...)

	if len(issues) == 0 {
		return nil
	}

	messages := make([]string, 0, len(issues))
	for _, issue := range issues {
		messages = append(messages, issue.message)
	}
	return errors.New(strings.Join(messages, "; "))
}

func validateDraftDurationQuality(input WorkoutGenerationToolInput, draft *workout.CreateWorkoutRequest) []workoutDraftQualityIssue {
	if input.SessionDuration <= 0 || allowsLowerVolume(input) {
		return nil
	}
	if !isStrengthOrHypertrophy(input) {
		return nil
	}

	workingSets := countWorkingSets(draft)
	estimatedMinutes := estimateWorkoutDurationMinutes(input, draft)
	minEstimatedMinutes := float64(input.SessionDuration) * minNormalDurationShare
	minWorkingSets := minimumWorkingSets(input)

	issues := make([]workoutDraftQualityIssue, 0, 2)
	if workingSets < minWorkingSets {
		issues = append(issues, workoutDraftQualityIssue{message: fmt.Sprintf("draft has %d working sets; expected at least %d for a normal %d-minute %s request", workingSets, minWorkingSets, input.SessionDuration, strengthHypertrophyLabel(input))})
	}
	if estimatedMinutes < minEstimatedMinutes {
		issues = append(issues, workoutDraftQualityIssue{message: fmt.Sprintf("draft estimates to %.0f minutes; expected at least %.0f minutes for the requested %d-minute session", estimatedMinutes, minEstimatedMinutes, input.SessionDuration)})
	}

	return issues
}

func validateDraftEquipmentQuality(input WorkoutGenerationToolInput, draft *workout.CreateWorkoutRequest) []workoutDraftQualityIssue {
	context := normalizedWorkoutContext(input)
	if context == "" {
		return nil
	}

	disallowed := disallowedEquipmentTerms(context)
	if len(disallowed) == 0 {
		return nil
	}

	issues := make([]workoutDraftQualityIssue, 0)
	for _, exercise := range draft.Exercises {
		name := normalizeQualityText(exercise.Name)
		for _, term := range disallowed {
			if strings.Contains(name, term) {
				issues = append(issues, workoutDraftQualityIssue{message: fmt.Sprintf("%q appears to require %s, which is not available in the request", exercise.Name, term)})
				break
			}
		}
	}
	return issues
}

func validateDraftInjuryQuality(input WorkoutGenerationToolInput, draft *workout.CreateWorkoutRequest) []workoutDraftQualityIssue {
	injuries := activeInjuryText(input.Injuries)
	if injuries == "" {
		return nil
	}

	disallowed := disallowedInjuryTerms(injuries)
	if len(disallowed) == 0 {
		return nil
	}

	issues := make([]workoutDraftQualityIssue, 0)
	for _, exercise := range draft.Exercises {
		name := normalizeQualityText(exercise.Name)
		for _, term := range disallowed {
			if strings.Contains(name, term) {
				issues = append(issues, workoutDraftQualityIssue{message: fmt.Sprintf("%q conflicts with the reported injury context", exercise.Name)})
				break
			}
		}
	}
	return issues
}

func countWorkingSets(draft *workout.CreateWorkoutRequest) int {
	count := 0
	for _, exercise := range draft.Exercises {
		for _, set := range exercise.Sets {
			if set.SetType == "working" {
				count++
			}
		}
	}
	return count
}

func estimateWorkoutDurationMinutes(input WorkoutGenerationToolInput, draft *workout.CreateWorkoutRequest) float64 {
	if draft == nil || len(draft.Exercises) == 0 {
		return 0
	}

	seconds := 0
	if input.SessionDuration >= 30 && isStrengthOrHypertrophy(input) {
		seconds += 5 * 60
	}
	seconds += len(draft.Exercises) * 90

	totalSets := 0
	for _, exercise := range draft.Exercises {
		for _, set := range exercise.Sets {
			totalSets++
			seconds += 45
			if set.SetType == "warmup" {
				seconds += warmupRestSeconds(input)
			} else {
				seconds += workingRestSeconds(input)
			}
		}
	}
	if totalSets > 0 {
		seconds -= workingRestSeconds(input)
	}

	return float64(seconds) / 60
}

func minimumWorkingSets(input WorkoutGenerationToolInput) int {
	estimatedMinutesPerWorkingSet := 6.0
	if isHypertrophy(input) {
		estimatedMinutesPerWorkingSet = 5.0
	}
	sets := int(math.Ceil(float64(input.SessionDuration) / estimatedMinutesPerWorkingSet))
	if sets < 2 {
		return 2
	}
	return sets
}

func warmupRestSeconds(input WorkoutGenerationToolInput) int {
	if isStrength(input) {
		return 75
	}
	if isHypertrophy(input) {
		return 45
	}
	return 45
}

func workingRestSeconds(input WorkoutGenerationToolInput) int {
	switch {
	case isStrength(input):
		return 150
	case isHypertrophy(input):
		return 75
	case strings.Contains(normalizedGoalFocus(input), "endurance"):
		return 45
	default:
		return 60
	}
}

type equipmentInventory struct {
	hasFullGym    bool
	hasBench      bool
	hasBarbell    bool
	hasDumbbell   bool
	hasKettlebell bool
	hasCable      bool
	hasMachine    bool
}

func parseEquipmentInventory(context string) equipmentInventory {
	return equipmentInventory{
		hasFullGym:    hasAvailableGymContext(context),
		hasBench:      hasAvailableEquipmentTerm(context, "bench", "box", "chair", "step"),
		hasBarbell:    hasAvailableEquipmentTerm(context, "barbell", "power rack", "squat rack"),
		hasDumbbell:   hasAvailableEquipmentTerm(context, "dumbbell", "dumbbells"),
		hasKettlebell: hasAvailableEquipmentTerm(context, "kettlebell", "kettlebells"),
		hasCable:      hasAvailableEquipmentTerm(context, "cable", "cables"),
		hasMachine:    hasAvailableEquipmentTerm(context, "machine", "machines", "leg press", "lat pulldown"),
	}
}

func disallowedEquipmentTerms(context string) []string {
	inventory := parseEquipmentInventory(context)
	if inventory.hasFullGym {
		return explicitlyUnavailableGymEquipmentTerms(context)
	}

	terms := make([]string, 0, 12)
	if !inventory.hasBarbell {
		terms = append(terms, "barbell", "smith")
	}
	if !inventory.hasDumbbell {
		terms = append(terms, "dumbbell")
	}
	if !inventory.hasKettlebell {
		terms = append(terms, "kettlebell")
	}
	if !inventory.hasCable {
		terms = append(terms, "cable", "lat pulldown")
	}
	if !inventory.hasMachine {
		terms = append(terms, "machine", "leg press")
	}
	if !inventory.hasBench {
		terms = append(terms, "bench", "bench-supported")
	}

	return terms
}

func explicitlyUnavailableGymEquipmentTerms(context string) []string {
	terms := make([]string, 0, 12)
	if hasUnavailableEquipmentTerm(context, "barbell", "power rack", "squat rack") {
		terms = append(terms, "barbell", "smith")
	}
	if hasUnavailableEquipmentTerm(context, "dumbbell", "dumbbells") {
		terms = append(terms, "dumbbell")
	}
	if hasUnavailableEquipmentTerm(context, "kettlebell", "kettlebells") {
		terms = append(terms, "kettlebell")
	}
	if hasUnavailableEquipmentTerm(context, "cable", "cables") {
		terms = append(terms, "cable", "lat pulldown")
	}
	if hasUnavailableEquipmentTerm(context, "machine", "machines", "leg press", "lat pulldown") {
		terms = append(terms, "machine", "leg press")
	}
	if hasUnavailableEquipmentTerm(context, "bench", "box", "chair", "step") {
		terms = append(terms, "bench", "bench-supported")
	}
	return terms
}

func disallowedInjuryTerms(injuries string) []string {
	terms := make([]string, 0, 8)
	if hasAnyQualityTerm(injuries, "knee", "acl", "meniscus") {
		terms = append(terms, "squat", "lunge", "leg press", "jump", "step-up", "step up", "running")
	}
	if hasAnyQualityTerm(injuries, "shoulder", "rotator cuff") {
		terms = append(terms, "overhead press", "shoulder press", "upright row", "dip", "snatch")
	}
	if hasAnyQualityTerm(injuries, "back", "low back", "lower back", "spine") {
		terms = append(terms, "deadlift", "good morning", "back squat", "barbell row")
	}
	if hasAnyQualityTerm(injuries, "wrist", "elbow") {
		terms = append(terms, "dip", "skull crusher", "barbell curl")
	}
	return terms
}

func hasNoActiveInjury(injuries string) bool {
	return injuries == "" ||
		injuries == "none" ||
		hasExactQualityTerm(
			injuries,
			"no injuries",
			"no injury",
			"no pain",
			"no issues",
			"no limitations",
			"no knee pain",
			"no knee issues",
			"no shoulder pain",
			"no shoulder issues",
			"no back pain",
			"no back issues",
			"no wrist pain",
			"no wrist issues",
			"no elbow pain",
			"no elbow issues",
		)
}

func activeInjuryText(value string) string {
	injuries := normalizeQualityText(value)
	if hasNoActiveInjury(injuries) {
		return ""
	}

	negatedBodyPartPhrases := []string{
		"no knee pain",
		"no knee issues",
		"no shoulder pain",
		"no shoulder issues",
		"no back pain",
		"no back issues",
		"no wrist pain",
		"no wrist issues",
		"no elbow pain",
		"no elbow issues",
	}
	for _, phrase := range negatedBodyPartPhrases {
		injuries = strings.ReplaceAll(injuries, phrase, " ")
	}

	return normalizeQualityText(injuries)
}

func allowsLowerVolume(input WorkoutGenerationToolInput) bool {
	text := normalizeQualityText(input.FitnessLevel) + " " + normalizedGoalFocus(input) + " " + normalizedWorkoutContext(input)
	return hasAnyQualityTerm(text, "minimal", "quick", "short", "beginner", "rehab", "prehab", "mobility", "warm-up", "warmup", "low volume", "easy")
}

func isStrengthOrHypertrophy(input WorkoutGenerationToolInput) bool {
	return isStrength(input) || isHypertrophy(input)
}

func isStrength(input WorkoutGenerationToolInput) bool {
	return strings.Contains(normalizedGoalFocus(input), "strength")
}

func isHypertrophy(input WorkoutGenerationToolInput) bool {
	text := normalizedGoalFocus(input)
	return strings.Contains(text, "hypertrophy") || strings.Contains(text, "muscle") || strings.Contains(text, "size")
}

func strengthHypertrophyLabel(input WorkoutGenerationToolInput) string {
	if isHypertrophy(input) {
		return "hypertrophy"
	}
	return "strength"
}

func normalizedGoalFocus(input WorkoutGenerationToolInput) string {
	return normalizeQualityText(input.FitnessGoal + " " + input.WorkoutFocus)
}

func normalizedWorkoutContext(input WorkoutGenerationToolInput) string {
	return normalizeQualityText(input.Equipment + " " + input.SpaceConstraints)
}

func hasAnyQualityTerm(text string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(text, normalizeQualityText(term)) {
			return true
		}
	}
	return false
}

func hasExactQualityTerm(text string, terms ...string) bool {
	for _, term := range terms {
		if text == normalizeQualityText(term) {
			return true
		}
	}
	return false
}

func hasAvailableEquipmentTerm(text string, terms ...string) bool {
	for _, term := range terms {
		normalizedTerm := normalizeQualityText(term)
		if !strings.Contains(text, normalizedTerm) {
			continue
		}
		if hasAnyQualityTerm(text,
			"no "+normalizedTerm,
			"without "+normalizedTerm,
			"no access to "+normalizedTerm,
		) {
			continue
		}
		return true
	}
	return false
}

func hasUnavailableEquipmentTerm(text string, terms ...string) bool {
	for _, term := range terms {
		normalizedTerm := normalizeQualityText(term)
		if hasAnyQualityTerm(text,
			"no "+normalizedTerm,
			"without "+normalizedTerm,
			"no access to "+normalizedTerm,
		) {
			return true
		}
	}
	return false
}

func hasAvailableGymContext(text string) bool {
	if hasAnyQualityTerm(
		text,
		"no gym",
		"without gym",
		"no access to gym",
		"no full gym",
		"without full gym",
		"no access to full gym",
		"no commercial gym",
		"without commercial gym",
		"no access to commercial gym",
	) {
		return false
	}

	return hasAnyQualityTerm(text, "full gym", "commercial gym") || hasQualityWord(text, "gym")
}

func hasQualityWord(text string, word string) bool {
	normalizedWord := normalizeQualityText(word)
	for _, field := range strings.Fields(text) {
		if strings.Trim(field, ".,;:!?()[]{}") == normalizedWord {
			return true
		}
	}
	return false
}

func normalizeQualityText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}
