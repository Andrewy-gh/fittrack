package aichateval

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

func ScoreResult(result *Result, mode string) {
	if result.Status == StatusError && IsOperationalError(result.Error) {
		setScore(result, ScoreStatusOperationalError, "provider or runtime issue; not scored as assistant behavior")
		return
	}
	if strings.TrimSpace(result.ExpectedOutcome) == "" {
		setScore(result, ScoreStatusUnscored, "scenario has no expected outcome")
		return
	}
	if result.Status == StatusError {
		setScore(result, ScoreStatusFail, "assistant behavior ended in an error before meeting the expected outcome")
		return
	}

	switch result.ExpectedOutcome {
	case ExpectedGenerateFirstTurn:
		scoreGenerateFirstTurn(result)
	case ExpectedAskOnceThenGenerate:
		scoreAskOnceThenGenerate(result, mode)
	case ExpectedDoNotGenerate:
		scoreDoNotGenerate(result)
	case ExpectedNarrowScopeBeforeGenerate:
		scoreNarrowScopeBeforeGenerate(result, mode)
	case ExpectedReviseWithoutRestart:
		scoreReviseWithoutRestart(result)
	case ExpectedAnswerFromData:
		scoreAnswerFromData(result)
	case ExpectedAnswerWithoutTools:
		scoreAnswerWithoutTools(result)
	case ExpectedHistoryInformedDraft:
		scoreHistoryInformedDraft(result)
	case ExpectedProfileUpdated:
		scoreProfileUpdated(result)
	default:
		setScore(result, ScoreStatusUnscored, fmt.Sprintf("unknown expected outcome %q", result.ExpectedOutcome))
	}
}

func scoreGenerateFirstTurn(result *Result) {
	if firstTurnStatus(*result) != StatusStructuredDraft {
		setScore(result, ScoreStatusFail, "expected a structured draft on the first turn")
		return
	}
	if !hasToolCall(result.ToolCalls, "generate_workout_draft") {
		setScore(result, ScoreStatusFail, "expected the workout draft tool to be called")
		return
	}
	if !passesRequiredToolCallChecks(result) {
		return
	}
	if disallowed := disallowedDataToolCalls(result); len(disallowed) > 0 {
		setScore(result, ScoreStatusFail, fmt.Sprintf("expected no data tool call for a normal draft request, got %s", strings.Join(disallowed, ", ")))
		return
	}
	if !passesTermChecks(result) {
		return
	}
	setScore(result, ScoreStatusPass, "structured draft generated on the first turn")
}

func scoreAskOnceThenGenerate(result *Result, mode string) {
	firstStatus := firstTurnStatus(*result)
	if firstStatus == StatusStructuredDraft {
		setScore(result, ScoreStatusFail, "expected one follow-up question before generating")
		return
	}
	if firstStatus != StatusFollowUpQuestion {
		setScore(result, ScoreStatusFail, "expected the first turn to ask a follow-up question")
		return
	}
	if mode != ModeTwoTurn {
		setScore(result, ScoreStatusUnscored, "first turn asked the expected follow-up; run two_turn mode to verify generation")
		return
	}
	if len(result.Turns) < 2 {
		setScore(result, ScoreStatusFail, "expected a second turn after the follow-up question")
		return
	}
	if result.Status != StatusStructuredDraft {
		setScore(result, ScoreStatusFail, "expected the second turn to generate a structured draft")
		return
	}
	if !passesTermChecks(result) {
		return
	}
	setScore(result, ScoreStatusPass, "asked one follow-up question, then generated a structured draft")
}

func scoreDoNotGenerate(result *Result) {
	if result.Draft != nil {
		setScore(result, ScoreStatusFail, "expected no structured workout draft")
		return
	}
	if strings.TrimSpace(result.Text) == "" {
		setScore(result, ScoreStatusFail, "expected explanatory text when no draft is generated")
		return
	}
	if !passesTermChecks(result) {
		return
	}
	setScore(result, ScoreStatusPass, "no draft was generated and the assistant returned text")
}

func scoreNarrowScopeBeforeGenerate(result *Result, mode string) {
	if len(result.Turns) == 0 {
		setScore(result, ScoreStatusFail, "expected a first turn response")
		return
	}
	first := result.Turns[0]
	if first.Draft != nil {
		setScore(result, ScoreStatusFail, "expected the first turn to narrow scope before generating")
		return
	}
	if strings.TrimSpace(first.Text) == "" {
		setScore(result, ScoreStatusFail, "expected narrowing text before generating")
		return
	}
	if !narrowsToSingleWorkout(first.Text) {
		setScore(result, ScoreStatusFail, "expected the first turn to ask the user to choose one workout or session")
		return
	}
	if !passesTermChecks(result) {
		return
	}
	if mode == ModeTwoTurn {
		if len(result.Turns) < 2 {
			setScore(result, ScoreStatusFail, "expected a second turn after the user narrowed the request")
			return
		}
		if result.Status != StatusStructuredDraft {
			setScore(result, ScoreStatusFail, "expected a structured draft after the user narrowed the request")
			return
		}
	}
	setScore(result, ScoreStatusPass, "assistant narrowed scope, then generated a structured draft")
}

func narrowsToSingleWorkout(text string) bool {
	lower := strings.ToLower(text)
	if containsWholeWeekPlan(lower) {
		return false
	}
	if redirectsMealPlanToWorkoutSession(lower) {
		return true
	}
	hasSingleSessionScope := containsAny(lower, []string{
		"one workout",
		"single workout",
		"one workout at a time",
		"one structured workout draft at a time",
		"one session",
		"single session",
		"one training session",
		"single training session",
		"one at a time",
		"one day first",
		"one day to start",
		"first workout",
		"first session",
	})
	hasUserChoicePrompt := asksUserToChooseWorkout(lower)
	return hasSingleSessionScope && hasUserChoicePrompt
}

func redirectsMealPlanToWorkoutSession(text string) bool {
	refusesMealPlan := containsAny(text, []string{
		"unable to provide meal plans",
		"can't provide meal plans",
		"can't provide a meal plan",
		"can't assist with meal plans",
		"can't assist with a meal plan",
		"cannot provide meal plans",
		"cannot provide a meal plan",
		"not able to provide meal plans",
		"not able to provide a meal plan",
		"can't create meal plans",
		"can't create a meal plan",
		"cannot create meal plans",
		"cannot create a meal plan",
		"not able to create meal plans",
		"not able to create a meal plan",
		"not equipped to create meal plans",
		"not equipped to create a meal plan",
		"not equipped to build meal plans",
		"not equipped to build a meal plan",
		"not equipped to provide meal plans",
		"not equipped to provide a meal plan",
	})
	asksForWorkoutFocus := strings.Contains(text, "?") && (containsAny(text, []string{
		"what is your workout focus",
		"what's your workout focus",
		"what is your primary focus",
		"what is your primary workout focus",
		"what is your preferred workout focus",
		"what is your desired workout focus",
		"what's your primary focus",
		"what's your primary workout focus",
		"what's your preferred workout focus",
		"what's your desired workout focus",
		"what workout focus",
		"what is the workout focus",
		"what should the workout focus be",
		"what would you like to focus on today",
		"what would you like the focus to be",
		"what would you like the focus of this workout to be",
	}) || (strings.Contains(text, "workout") && strings.Contains(text, "focus")))
	asksForSessionDetails := containsAny(text, []string{
		"session to be",
		"session should be",
		"session length",
		"how long would you like the session",
		"how long do you want the session",
		"how long do you have for the session",
		"how long do you have for this session",
		"how long do you have for each session",
		"how long can you train for each session",
		"how long is your session duration",
		"how long should the session",
		"i'll plan for 45 minutes",
		"for about 45 minutes",
	})
	hasWorkoutContext := strings.Contains(text, "for your workout") || strings.Contains(text, "to build your workout")
	return refusesMealPlan && asksForWorkoutFocus && (asksForSessionDetails || hasWorkoutContext)
}

func asksUserToChooseWorkout(text string) bool {
	if sentenceContainsQuestionPhrase(text, []string{
		"which day",
		"which workout",
		"which session",
		"which one",
		"what day",
		"what workout",
		"what session",
		"what would be the focus for your first workout",
		"what is the focus for your first workout",
		"what should the focus be for your first workout",
		"what would be the focus for your first session",
		"what is the focus for your first session",
		"what should the focus be for your first session",
	}) {
		return true
	}
	return sentenceStartsWithAny(text, []string{"pick", "choose", "select"})
}

func sentenceContainsQuestionPhrase(text string, phrases []string) bool {
	sentenceStart := 0
	for i, char := range text {
		if char != '.' && char != '?' && char != '!' && char != '\n' {
			continue
		}
		if sentenceHasQuestionPhrase(text[sentenceStart:i+1], phrases) {
			return true
		}
		sentenceStart = i + 1
	}
	return sentenceHasQuestionPhrase(text[sentenceStart:], phrases)
}

func sentenceHasQuestionPhrase(sentence string, phrases []string) bool {
	return strings.Contains(sentence, "?") && containsAny(sentence, phrases)
}

func sentenceStartsWithAny(text string, verbs []string) bool {
	trimmed := strings.TrimSpace(text)
	for _, verb := range verbs {
		for _, prefix := range []string{"", "please "} {
			start := prefix + verb + " "
			if strings.HasPrefix(trimmed, start) || containsAny(trimmed, []string{
				". " + start,
				"? " + start,
				"! " + start,
				"\n" + start,
			}) {
				return true
			}
		}
	}
	return false
}

func containsWholeWeekPlan(text string) bool {
	dayNumberPlan := strings.Contains(text, "day 1") && strings.Contains(text, "day 2")
	dayWordPlan := strings.Contains(text, "day one") && strings.Contains(text, "day two")
	return dayNumberPlan || dayWordPlan
}

func scoreReviseWithoutRestart(result *Result) {
	if firstTurnStatus(*result) == StatusFollowUpQuestion {
		setScore(result, ScoreStatusFail, "expected revision without restarting discovery")
		return
	}
	if result.Status != StatusStructuredDraft {
		setScore(result, ScoreStatusFail, "expected a revised structured draft")
		return
	}
	if !passesTermChecks(result) {
		return
	}
	setScore(result, ScoreStatusPass, "revised the draft without restarting discovery")
}

func scoreHistoryInformedDraft(result *Result) {
	if firstTurnStatus(*result) != StatusStructuredDraft {
		setScore(result, ScoreStatusFail, "expected a structured draft on the first turn")
		return
	}
	if !hasToolCall(result.ToolCalls, "generate_workout_draft") {
		setScore(result, ScoreStatusFail, "expected the workout draft tool to be called")
		return
	}
	if !hasAnyToolCall(result.ToolCalls, []string{"get_workouts", "get_exercise_stats"}) {
		setScore(result, ScoreStatusFail, "expected a data tool to be called before the workout draft")
		return
	}
	if !passesTermChecks(result) {
		return
	}
	setScore(result, ScoreStatusPass, "used workout history before generating a structured draft")
}

func scoreAnswerFromData(result *Result) {
	if result.Draft != nil {
		setScore(result, ScoreStatusFail, "expected no structured workout draft for a data answer")
		return
	}
	if hasToolCall(result.ToolCalls, "generate_workout_draft") {
		setScore(result, ScoreStatusFail, "expected no workout draft tool call for a data answer")
		return
	}
	if len(result.RequiredToolCalls) == 0 && !hasToolCall(result.ToolCalls, "get_workouts") {
		setScore(result, ScoreStatusFail, "expected get_workouts to be called")
		return
	}
	if !passesRequiredToolCallChecks(result) {
		return
	}
	if len(result.RequiredToolCalls) > 0 {
		if disallowed := disallowedDataToolCalls(result); len(disallowed) > 0 {
			setScore(result, ScoreStatusFail, fmt.Sprintf("expected no data tool call beyond %s, got %s", strings.Join(result.AllowedToolCalls, ", "), strings.Join(disallowed, ", ")))
			return
		}
	}
	if strings.TrimSpace(result.Text) == "" {
		setScore(result, ScoreStatusFail, "expected answer text from workout data")
		return
	}
	if !passesTermChecks(result) {
		return
	}
	setScore(result, ScoreStatusPass, "answered from workout data with get_workouts")
}

func scoreProfileUpdated(result *Result) {
	if result.Draft != nil {
		setScore(result, ScoreStatusFail, "expected no structured workout draft for a profile update")
		return
	}
	if !hasToolCall(result.ToolCalls, "update_training_profile") {
		setScore(result, ScoreStatusFail, "expected update_training_profile to be called")
		return
	}
	if !passesRequiredToolCallChecks(result) {
		return
	}
	if disallowed := disallowedToolCalls(result.ToolCalls, result.AllowedToolCalls); len(disallowed) > 0 {
		setScore(result, ScoreStatusFail, fmt.Sprintf("expected no tool calls beyond %s, got %s", strings.Join(result.AllowedToolCalls, ", "), strings.Join(disallowed, ", ")))
		return
	}
	if strings.TrimSpace(result.Text) == "" {
		setScore(result, ScoreStatusFail, "expected profile update announcement text")
		return
	}
	if !passesProfileUpdateTermChecks(result) {
		return
	}
	setScore(result, ScoreStatusPass, "updated the training profile and announced the saved facts")
}

func scoreAnswerWithoutTools(result *Result) {
	if result.Draft != nil {
		setScore(result, ScoreStatusFail, "expected no structured workout draft")
		return
	}
	if disallowed := disallowedToolCalls(result.ToolCalls, result.AllowedToolCalls); len(disallowed) > 0 {
		if len(result.AllowedToolCalls) == 0 {
			setScore(result, ScoreStatusFail, fmt.Sprintf("expected zero tool calls, got %s", strings.Join(disallowed, ", ")))
			return
		}
		setScore(result, ScoreStatusFail, fmt.Sprintf("expected no tool calls beyond %s, got %s", strings.Join(result.AllowedToolCalls, ", "), strings.Join(disallowed, ", ")))
		return
	}
	if strings.TrimSpace(result.Text) == "" {
		setScore(result, ScoreStatusFail, "expected answer text without tools")
		return
	}
	if !passesTermChecks(result) {
		return
	}
	if len(result.ToolCalls) > 0 {
		setScore(result, ScoreStatusPass, fmt.Sprintf("answered correctly with tolerated tool calls: %s", strings.Join(result.ToolCalls, ", ")))
		return
	}
	setScore(result, ScoreStatusPass, "answered without tools")
}

func passesTermChecks(result *Result) bool {
	text := allResponseText(*result)
	for _, term := range result.RequiredTextTerms {
		if !containsRequiredTerm(text, term) {
			setScore(result, ScoreStatusFail, fmt.Sprintf("response text is missing required term %q", term))
			return false
		}
	}
	for _, term := range result.ForbiddenTextTerms {
		if containsFold(text, term) {
			setScore(result, ScoreStatusFail, fmt.Sprintf("response text includes forbidden term %q", term))
			return false
		}
	}

	exercises := exerciseText(result)
	for _, term := range result.ForbiddenExerciseTerms {
		if containsFold(exercises, term) {
			setScore(result, ScoreStatusFail, fmt.Sprintf("draft includes forbidden exercise term %q", term))
			return false
		}
	}
	return true
}

func passesProfileUpdateTermChecks(result *Result) bool {
	text := allResponseText(*result)
	for _, term := range result.RequiredTextTerms {
		if term == "remember" && announcesProfileUpdate(text) {
			continue
		}
		if !containsRequiredTerm(text, term) {
			setScore(result, ScoreStatusFail, fmt.Sprintf("response text is missing required term %q", term))
			return false
		}
	}
	for _, term := range result.ForbiddenTextTerms {
		if containsFold(text, term) {
			setScore(result, ScoreStatusFail, fmt.Sprintf("response text includes forbidden term %q", term))
			return false
		}
	}
	return true
}

func announcesProfileUpdate(text string) bool {
	lower := strings.ToLower(text)
	return containsAny(lower, []string{
		"updated your training profile",
		"updated your profile",
		"saved to your training profile",
		"added to your training profile",
		"saved your usual training setup",
	})
}

func passesRequiredToolCallChecks(result *Result) bool {
	for _, required := range result.RequiredToolCalls {
		if !hasToolCall(result.ToolCalls, required) {
			setScore(result, ScoreStatusFail, fmt.Sprintf("expected %s to be called", required))
			return false
		}
	}
	return true
}

func disallowedDataToolCalls(result *Result) []string {
	dataTools := []string{"get_workouts", "get_exercise_stats", "update_training_profile"}
	var disallowed []string
	for _, tool := range dataTools {
		if hasToolCall(result.ToolCalls, tool) && !hasToolCall(result.AllowedToolCalls, tool) {
			disallowed = append(disallowed, tool)
		}
	}
	return disallowed
}

func allResponseText(result Result) string {
	if len(result.Turns) == 0 {
		return result.Text
	}

	parts := make([]string, 0, len(result.Turns))
	for _, turn := range result.Turns {
		parts = append(parts, turn.Text)
	}
	return strings.Join(parts, "\n")
}

func exerciseText(result *Result) string {
	if result.Draft == nil {
		return ""
	}

	names := make([]string, 0, len(result.Draft.Exercises))
	for _, exercise := range result.Draft.Exercises {
		names = append(names, exercise.Name)
	}
	return strings.Join(names, "\n")
}

var isoDateTermPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// containsRequiredTerm matches like containsFold, except an ISO-date term
// (e.g. "2026-06-30") also accepts prose renderings such as "June 30th, 2026"
// or "30 June".
func containsRequiredTerm(text string, term string) bool {
	trimmed := strings.TrimSpace(term)
	if !isoDateTermPattern.MatchString(trimmed) {
		return containsFold(text, term)
	}
	if containsFold(text, trimmed) {
		return true
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return false
	}
	month := parsed.Month().String()
	monthAlternatives := fmt.Sprintf(`(?:%s|%s\.?)`, month, month[:3])
	day := fmt.Sprintf(`0?%d(?:st|nd|rd|th)?`, parsed.Day())
	pattern := fmt.Sprintf(`(?i)\b(?:%[1]s\s+%[2]s|%[2]s\s+(?:of\s+)?%[1]s)\b`, monthAlternatives, day)
	return regexp.MustCompile(pattern).MatchString(text)
}

func disallowedToolCalls(calls []string, allowed []string) []string {
	var disallowed []string
	for _, call := range calls {
		if !hasToolCall(allowed, call) {
			disallowed = append(disallowed, call)
		}
	}
	return disallowed
}

func containsFold(text string, term string) bool {
	normalizedTerm := strings.TrimSpace(term)
	if normalizedTerm == "" {
		return true
	}
	return strings.Contains(strings.ToLower(text), strings.ToLower(normalizedTerm))
}

func containsAny(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

func hasToolCall(calls []string, name string) bool {
	for _, call := range calls {
		if call == name {
			return true
		}
	}
	return false
}

func hasAnyToolCall(calls []string, names []string) bool {
	for _, name := range names {
		if hasToolCall(calls, name) {
			return true
		}
	}
	return false
}

func setScore(result *Result, status string, reason string) {
	result.ScoreStatus = status
	result.ScoreReason = reason
	result.Passed = status == ScoreStatusPass
}

func IsOperationalError(message string) bool {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return false
	}

	if trimmed == context.Canceled.Error() || trimmed == context.DeadlineExceeded.Error() {
		return true
	}

	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "validate workout draft quality") || strings.Contains(lower, "exhausted quality repair attempts") {
		return false
	}

	operationalTerms := []string{
		"context canceled",
		"context deadline exceeded",
		"deadline exceeded",
		"timed out",
		"timeout",
		"quota",
		"rate limit",
		"rate-limited",
		"resource_exhausted",
		"please retry in",
		"429",
		"provider",
		"gemini",
		"googleapi",
		"genkit",
		"unavailable",
	}
	for _, term := range operationalTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}
