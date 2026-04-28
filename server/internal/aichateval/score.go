package aichateval

import (
	"context"
	"fmt"
	"strings"
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
	default:
		setScore(result, ScoreStatusUnscored, fmt.Sprintf("unknown expected outcome %q", result.ExpectedOutcome))
	}
}

func scoreGenerateFirstTurn(result *Result) {
	if firstTurnStatus(*result) != StatusStructuredDraft {
		setScore(result, ScoreStatusFail, "expected a structured draft on the first turn")
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

func passesTermChecks(result *Result) bool {
	text := allResponseText(*result)
	for _, term := range result.RequiredTextTerms {
		if !containsFold(text, term) {
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

func containsFold(text string, term string) bool {
	normalizedTerm := strings.TrimSpace(term)
	if normalizedTerm == "" {
		return true
	}
	return strings.Contains(strings.ToLower(text), strings.ToLower(normalizedTerm))
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
