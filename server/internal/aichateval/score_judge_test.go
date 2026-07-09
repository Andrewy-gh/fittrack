package aichateval

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
)

type fakeNarrowScopeJudge struct {
	verdict *NarrowScopeJudgeVerdict
	err     error
	inputs  []NarrowScopeJudgeInput
}

func (f *fakeNarrowScopeJudge) JudgeNarrowScope(_ context.Context, input NarrowScopeJudgeInput) (*NarrowScopeJudgeVerdict, error) {
	f.inputs = append(f.inputs, input)
	if f.err != nil {
		return nil, f.err
	}
	return f.verdict, nil
}

func TestRunScoresNarrowScopeWithJudgeVerdict(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "FitTrack builds one workout draft at a time. Which session should we start with?"}},
			{done: &aichat.StreamDone{Text: "I made a focused draft.", WorkoutDraft: testDraft()}},
		},
	}
	judge := &fakeNarrowScopeJudge{verdict: &NarrowScopeJudgeVerdict{
		NarrowsToSingleWorkout: true,
		AsksUserToChoose:       true,
		Rationale:              "Asks which single session to start with.",
	}}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "prompt-14",
		Title:           "Weekly Split Request",
		Prompt:          "Build me a 4-day workout split for the whole week.",
		ExpectedOutcome: ExpectedNarrowScopeBeforeGenerate,
		FollowUpAnswer:  "Let's start with day one as an upper-body workout. No injuries, full gym, 45 minutes.",
	}}, RunOptions{Mode: ModeTwoTurn, NarrowScopeJudge: judge})

	result := report.Results[0]
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q reason %q, want pass", result.Passed, result.ScoreStatus, result.ScoreReason)
	}
	if result.NarrowScopeJudge == nil || result.NarrowScopeJudge.Rationale != judge.verdict.Rationale {
		t.Fatalf("NarrowScopeJudge = %+v, want recorded verdict", result.NarrowScopeJudge)
	}
	if len(judge.inputs) != 1 || judge.inputs[0].ScenarioID != "prompt-14" || !strings.Contains(judge.inputs[0].ResponseText, "Which session") {
		t.Fatalf("judge inputs = %+v, want scenario and first-turn text", judge.inputs)
	}
	if strings.Contains(result.ScoreReason, "term-list fallback") {
		t.Fatalf("ScoreReason = %q, did not want fallback marker when judge passed", result.ScoreReason)
	}
}

func TestRunScoresNarrowScopeWithMealPlanJudgeVerdict(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "I cannot create meal plans, but I can help with one workout session. What workout focus do you want?"}},
			{done: &aichat.StreamDone{Text: "I made a focused draft.", WorkoutDraft: testDraft()}},
		},
	}
	judge := &fakeNarrowScopeJudge{verdict: &NarrowScopeJudgeVerdict{
		RefusesMealPlan:        true,
		AsksWorkoutFocus:       true,
		NarrowsToSingleWorkout: true,
		Rationale:              "Refuses meal planning and redirects to one workout.",
	}}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "prompt-18",
		Title:           "Workout And Meal Plan Request",
		Prompt:          "Make me a workout and meal plan for fat loss.",
		ExpectedOutcome: ExpectedNarrowScopeBeforeGenerate,
		FollowUpAnswer:  "Focus on the workout only: beginner, no injuries, 35 minutes, full gym.",
	}}, RunOptions{Mode: ModeTwoTurn, NarrowScopeJudge: judge})

	result := report.Results[0]
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q reason %q, want pass", result.Passed, result.ScoreStatus, result.ScoreReason)
	}
}

func TestRunScoresNarrowScopeRejectsJudgeFailureVerdict(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "Sure, I can help with that."}},
			{done: &aichat.StreamDone{Text: "I made a focused draft.", WorkoutDraft: testDraft()}},
		},
	}
	judge := &fakeNarrowScopeJudge{verdict: &NarrowScopeJudgeVerdict{
		Rationale: "Does not ask the user to choose a single workout.",
	}}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "prompt-14",
		Title:           "Weekly Split Request",
		Prompt:          "Build me a 4-day workout split for the whole week.",
		ExpectedOutcome: ExpectedNarrowScopeBeforeGenerate,
		FollowUpAnswer:  "Let's start with day one as an upper-body workout. No injuries, full gym, 45 minutes.",
	}}, RunOptions{Mode: ModeTwoTurn, NarrowScopeJudge: judge})

	result := report.Results[0]
	if result.Passed || result.ScoreStatus != ScoreStatusFail {
		t.Fatalf("score = passed %v status %q, want fail", result.Passed, result.ScoreStatus)
	}
	if !strings.Contains(result.ScoreReason, "judge rejected narrowing response") || !strings.Contains(result.ScoreReason, judge.verdict.Rationale) {
		t.Fatalf("ScoreReason = %q, want judge rationale", result.ScoreReason)
	}
	if result.NarrowScopeJudge == nil {
		t.Fatal("NarrowScopeJudge = nil, want recorded failing verdict")
	}
}

func TestRunScoresNarrowScopeFallsBackToTermListWhenJudgeFails(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "I can build one workout at a time. Pick one session to start with."}},
			{done: &aichat.StreamDone{Text: "I made a focused draft.", WorkoutDraft: testDraft()}},
		},
	}
	judge := &fakeNarrowScopeJudge{err: errors.New("judge unavailable")}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "prompt-14",
		Title:           "Weekly Split Request",
		Prompt:          "Build me a 4-day workout split for the whole week.",
		ExpectedOutcome: ExpectedNarrowScopeBeforeGenerate,
		FollowUpAnswer:  "Let's start with day one as an upper-body workout. No injuries, full gym, 45 minutes.",
	}}, RunOptions{Mode: ModeTwoTurn, NarrowScopeJudge: judge})

	result := report.Results[0]
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q reason %q, want fallback pass", result.Passed, result.ScoreStatus, result.ScoreReason)
	}
	if !strings.Contains(result.ScoreReason, "term-list fallback") {
		t.Fatalf("ScoreReason = %q, want fallback marker", result.ScoreReason)
	}
	if result.NarrowScopeJudge != nil {
		t.Fatalf("NarrowScopeJudge = %+v, want nil when judge failed", result.NarrowScopeJudge)
	}
}
