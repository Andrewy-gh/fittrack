package aichateval

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const narrowScopeJudgeMaxAttempts = 2

type NarrowScopeJudge interface {
	JudgeNarrowScope(ctx context.Context, input NarrowScopeJudgeInput) (*NarrowScopeJudgeVerdict, error)
}

type NarrowScopeJudgeInput struct {
	ScenarioID    string
	ScenarioTitle string
	UserPrompt    string
	ResponseText  string
}

type NarrowScopeJudgeVerdict struct {
	RefusesMealPlan        bool   `json:"refuses_meal_plan"`
	AsksWorkoutFocus       bool   `json:"asks_workout_focus"`
	NarrowsToSingleWorkout bool   `json:"narrows_to_single_workout"`
	AsksUserToChoose       bool   `json:"asks_user_to_choose"`
	Rationale              string `json:"rationale"`
}

type LLMNarrowScopeJudge struct {
	Generate func(ctx context.Context, prompt string) (string, error)
}

func (j LLMNarrowScopeJudge) JudgeNarrowScope(ctx context.Context, input NarrowScopeJudgeInput) (*NarrowScopeJudgeVerdict, error) {
	if j.Generate == nil {
		return nil, errors.New("narrow-scope judge generator is not configured")
	}

	prompt := buildNarrowScopeJudgePrompt(input)
	var lastErr error
	for attempt := 1; attempt <= narrowScopeJudgeMaxAttempts; attempt++ {
		raw, err := j.Generate(ctx, prompt)
		if err != nil {
			lastErr = err
			continue
		}
		verdict, err := ParseNarrowScopeJudgeVerdict(raw)
		if err == nil {
			return verdict, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("narrow-scope judge failed after %d attempts: %w", narrowScopeJudgeMaxAttempts, lastErr)
}

func ParseNarrowScopeJudgeVerdict(raw string) (*NarrowScopeJudgeVerdict, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, errors.New("empty judge response")
	}

	var parsed struct {
		RefusesMealPlan        *bool   `json:"refuses_meal_plan"`
		AsksWorkoutFocus       *bool   `json:"asks_workout_focus"`
		NarrowsToSingleWorkout *bool   `json:"narrows_to_single_workout"`
		AsksUserToChoose       *bool   `json:"asks_user_to_choose"`
		Rationale              *string `json:"rationale"`
	}
	decoder := json.NewDecoder(bytes.NewBufferString(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&parsed); err != nil {
		return nil, fmt.Errorf("parse judge JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, errors.New("judge response included trailing JSON values")
	}
	if parsed.RefusesMealPlan == nil ||
		parsed.AsksWorkoutFocus == nil ||
		parsed.NarrowsToSingleWorkout == nil ||
		parsed.AsksUserToChoose == nil ||
		parsed.Rationale == nil {
		return nil, errors.New("judge response missing required field")
	}

	rationale := strings.Join(strings.Fields(*parsed.Rationale), " ")
	if rationale == "" {
		return nil, errors.New("judge response missing rationale")
	}

	return &NarrowScopeJudgeVerdict{
		RefusesMealPlan:        *parsed.RefusesMealPlan,
		AsksWorkoutFocus:       *parsed.AsksWorkoutFocus,
		NarrowsToSingleWorkout: *parsed.NarrowsToSingleWorkout,
		AsksUserToChoose:       *parsed.AsksUserToChoose,
		Rationale:              rationale,
	}, nil
}

func buildNarrowScopeJudgePrompt(input NarrowScopeJudgeInput) string {
	return fmt.Sprintf(`You are judging one FitTrack AI chat eval response.

Return STRICT JSON only. Do not use markdown. Do not include any keys beyond:
{"refuses_meal_plan":bool,"asks_workout_focus":bool,"narrows_to_single_workout":bool,"asks_user_to_choose":bool,"rationale":"one short line"}

Definitions:
- refuses_meal_plan: true only when the assistant clearly declines to create meal plans or nutrition plans.
- asks_workout_focus: true only when the assistant asks what workout focus, body area, training goal, or supported workout type the user wants.
- narrows_to_single_workout: true only when the assistant keeps FitTrack scoped to one workout/session/draft now, not a full multi-day plan, weekly split, bundled plan, or meal plan.
- asks_user_to_choose: true only when the assistant asks the user to choose, pick, select, or tell which one day/workout/session to build first. This can be a question or a direct instruction.
- rationale: one short line explaining the verdict.

Judge the assistant response against the user's prompt. Do not require exact words.

Scenario: %s - %s
User prompt:
%s

Assistant response:
%s`, input.ScenarioID, input.ScenarioTitle, input.UserPrompt, input.ResponseText)
}

func narrowScopeJudgePass(result *Result, verdict NarrowScopeJudgeVerdict) bool {
	if isMealPlanPrompt(result.Prompt) {
		return verdict.RefusesMealPlan && verdict.AsksWorkoutFocus && verdict.NarrowsToSingleWorkout
	}
	return verdict.NarrowsToSingleWorkout && verdict.AsksUserToChoose
}

func isMealPlanPrompt(prompt string) bool {
	lower := strings.ToLower(prompt)
	return strings.Contains(lower, "meal plan") || strings.Contains(lower, "meal plans")
}
