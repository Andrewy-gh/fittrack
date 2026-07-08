package aichateval

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

type fakeRuntime struct {
	model string
	calls []runtimeCall
	next  []runtimeResponse
}

type runtimeCall struct {
	prompt  string
	history []aichat.RuntimeChatMessage
	userID  string
	hasUser bool
}

type runtimeResponse struct {
	done *aichat.StreamDone
	err  error
}

func (f *fakeRuntime) StreamChat(ctx context.Context, prompt string, history []aichat.RuntimeChatMessage, onChunk func(string) error) (*aichat.StreamDone, error) {
	userID, hasUser := user.Current(ctx)
	f.calls = append(f.calls, runtimeCall{
		prompt:  prompt,
		history: append([]aichat.RuntimeChatMessage{}, history...),
		userID:  userID,
		hasUser: hasUser,
	})
	if len(f.next) == 0 {
		return nil, errors.New("unexpected call")
	}
	response := f.next[0]
	f.next = f.next[1:]
	return response.done, response.err
}

func (f *fakeRuntime) ModelName() string {
	if f.model == "" {
		return "fake-model"
	}
	return f.model
}

func TestRunSingleTurnClassifiesStructuredDraft(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			done: &aichat.StreamDone{
				Model:        "test-model",
				Text:         "I made a draft.",
				WorkoutDraft: testDraft(),
				ToolCalls:    []string{"generate_workout_draft"},
			},
		}},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "case-1",
		Title:           "Ready Prompt",
		Prompt:          "Beginner upper body, full gym, no injuries.",
		Expectation:     "Should generate.",
		ExpectedOutcome: ExpectedGenerateFirstTurn,
	}}, RunOptions{Mode: ModeSingleTurn})

	if report.Mode != ModeSingleTurn {
		t.Fatalf("Run() mode = %q, want %q", report.Mode, ModeSingleTurn)
	}
	if len(report.Results) != 1 {
		t.Fatalf("Run() returned %d results, want 1", len(report.Results))
	}
	if report.Summary.StructuredDraftCount != 1 {
		t.Fatalf("StructuredDraftCount = %d, want 1", report.Summary.StructuredDraftCount)
	}
	if report.Summary.StructuredOutputConversionRate != 1 {
		t.Fatalf("StructuredOutputConversionRate = %f, want 1", report.Summary.StructuredOutputConversionRate)
	}
	if report.Summary.TwoTurnConversionRate != nil {
		t.Fatalf("TwoTurnConversionRate = %v, want nil for single-turn mode", *report.Summary.TwoTurnConversionRate)
	}
	result := report.Results[0]
	if result.Status != StatusStructuredDraft {
		t.Fatalf("Result status = %q, want %q", result.Status, StatusStructuredDraft)
	}
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q, want pass", result.Passed, result.ScoreStatus)
	}
	if result.DraftSummary == nil || result.DraftSummary.WorkingSets != 2 {
		t.Fatalf("DraftSummary = %#v, want 2 working sets", result.DraftSummary)
	}
	if len(result.Turns) != 1 {
		t.Fatalf("Turns = %d, want 1", len(result.Turns))
	}
}

func TestRunTwoTurnAnswersFollowUpAndUsesFinalStatus(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{
				done: &aichat.StreamDone{
					Model: "test-model",
					Text:  "Any injuries or equipment limits?",
				},
			},
			{
				done: &aichat.StreamDone{
					Model:        "test-model",
					Text:         "I made a draft.",
					WorkoutDraft: testDraft(),
				},
			},
		},
	}
	scenario := Scenario{
		ID:              "case-2",
		Title:           "Follow Up",
		Prompt:          "Give me a pull workout.",
		Expectation:     "Should ask, then generate.",
		ExpectedOutcome: ExpectedAskOnceThenGenerate,
		FollowUpAnswer:  "No injuries, dumbbells and bench, 45 minutes.",
	}

	report := Run(context.Background(), runtime, []Scenario{scenario}, RunOptions{Mode: ModeTwoTurn})

	result := report.Results[0]
	if result.Status != StatusStructuredDraft {
		t.Fatalf("Result status = %q, want %q", result.Status, StatusStructuredDraft)
	}
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q, want pass", result.Passed, result.ScoreStatus)
	}
	if result.Text != "I made a draft." {
		t.Fatalf("Result text = %q, want final turn text", result.Text)
	}
	if report.Summary.FollowUpQuestionCount != 0 {
		t.Fatalf("FollowUpQuestionCount = %d, want final status counts only", report.Summary.FollowUpQuestionCount)
	}
	if report.Summary.TwoTurnFollowUpCount != 1 {
		t.Fatalf("TwoTurnFollowUpCount = %d, want 1", report.Summary.TwoTurnFollowUpCount)
	}
	if report.Summary.TwoTurnAttemptCount != 1 {
		t.Fatalf("TwoTurnAttemptCount = %d, want 1", report.Summary.TwoTurnAttemptCount)
	}
	if report.Summary.TwoTurnStructuredDraftCount != 1 {
		t.Fatalf("TwoTurnStructuredDraftCount = %d, want 1", report.Summary.TwoTurnStructuredDraftCount)
	}
	if report.Summary.TwoTurnConversionRate == nil || *report.Summary.TwoTurnConversionRate != 1 {
		t.Fatalf("TwoTurnConversionRate = %v, want 1", report.Summary.TwoTurnConversionRate)
	}
	if report.Summary.PassedCount != 1 || report.Summary.PassRate != 1 {
		t.Fatalf("summary score = passed %d rate %f, want 1 and 1", report.Summary.PassedCount, report.Summary.PassRate)
	}
	if len(result.Turns) != 2 {
		t.Fatalf("Turns = %d, want 2", len(result.Turns))
	}
	if len(runtime.calls) != 2 {
		t.Fatalf("runtime calls = %d, want 2", len(runtime.calls))
	}
	second := runtime.calls[1]
	if second.prompt != scenario.FollowUpAnswer {
		t.Fatalf("second prompt = %q, want follow-up answer", second.prompt)
	}
	if len(second.history) != 2 {
		t.Fatalf("second history length = %d, want 2", len(second.history))
	}
	if second.history[0].Role != "user" || second.history[0].Text != scenario.Prompt {
		t.Fatalf("second history[0] = %#v, want original user prompt", second.history[0])
	}
	if second.history[1].Role != "assistant" || second.history[1].Text != "Any injuries or equipment limits?" {
		t.Fatalf("second history[1] = %#v, want first assistant follow-up", second.history[1])
	}
}

func TestRunSummaryCountsFinalStatuses(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "Any injuries?"}},
			{done: &aichat.StreamDone{Text: "General guidance only."}},
			{err: errors.New("provider failed")},
		},
	}

	report := Run(context.Background(), runtime, []Scenario{
		{ID: "case-1", Title: "Follow Up", Prompt: "Build legs."},
		{ID: "case-2", Title: "Text", Prompt: "Ideas."},
		{ID: "case-3", Title: "Error", Prompt: "Build push."},
	}, RunOptions{Mode: ModeSingleTurn})

	if report.Summary.TotalScenarios != 3 {
		t.Fatalf("TotalScenarios = %d, want 3", report.Summary.TotalScenarios)
	}
	if report.Summary.FollowUpQuestionCount != 1 {
		t.Fatalf("FollowUpQuestionCount = %d, want 1", report.Summary.FollowUpQuestionCount)
	}
	if report.Summary.TextOnlyCount != 1 {
		t.Fatalf("TextOnlyCount = %d, want 1", report.Summary.TextOnlyCount)
	}
	if report.Summary.ErrorCount != 1 {
		t.Fatalf("ErrorCount = %d, want 1", report.Summary.ErrorCount)
	}
	if report.Summary.StructuredOutputConversionRate != 0 {
		t.Fatalf("StructuredOutputConversionRate = %f, want 0", report.Summary.StructuredOutputConversionRate)
	}
}

func TestRunScoresExpectedNonGenerationWhenTextIsPresent(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			done: &aichat.StreamDone{
				Model: "test-model",
				Text:  "Because of chest pain and shortness of breath, get medical care before hard conditioning.",
			},
		}},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "case-refusal",
		Title:           "Medical Refusal",
		Prompt:          "I had chest pain. Give me hard conditioning.",
		ExpectedOutcome: ExpectedDoNotGenerate,
	}}, RunOptions{Mode: ModeSingleTurn})

	result := report.Results[0]
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q reason %q, want pass", result.Passed, result.ScoreStatus, result.ScoreReason)
	}
	if report.Summary.PassedCount != 1 || report.Summary.FailedCount != 0 {
		t.Fatalf("summary scores = passed %d failed %d, want 1/0", report.Summary.PassedCount, report.Summary.FailedCount)
	}
}

func TestRunScoresUnexpectedDraftForDoNotGenerateAsFailure(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			done: &aichat.StreamDone{
				Model:        "test-model",
				Text:         "I made a draft.",
				WorkoutDraft: testDraft(),
			},
		}},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "case-bad-refusal",
		Title:           "Bad Refusal",
		Prompt:          "Build me a whole week.",
		ExpectedOutcome: ExpectedDoNotGenerate,
	}}, RunOptions{Mode: ModeSingleTurn})

	result := report.Results[0]
	if result.Passed || result.ScoreStatus != ScoreStatusFail {
		t.Fatalf("score = passed %v status %q, want fail", result.Passed, result.ScoreStatus)
	}
	if report.Summary.FailedCount != 1 || report.Summary.PassRate != 0 {
		t.Fatalf("summary scores = failed %d rate %f, want 1 and 0", report.Summary.FailedCount, report.Summary.PassRate)
	}
}

func TestRunScoresAnswerFromData(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			done: &aichat.StreamDone{
				Text:      "You last squatted on 2026-06-30.",
				ToolCalls: []string{"get_workouts"},
			},
		}},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:                "case-data",
		Title:             "Data Answer",
		Prompt:            "When did I last squat?",
		ExpectedOutcome:   ExpectedAnswerFromData,
		RequiredTextTerms: []string{"2026-06-30"},
	}}, RunOptions{Mode: ModeSingleTurn})

	result := report.Results[0]
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q reason %q, want pass", result.Passed, result.ScoreStatus, result.ScoreReason)
	}
}

func TestRunScoresAnswerWithoutTools(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			done: &aichat.StreamDone{Text: "Your last workout was 2026-07-03."},
		}},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:                "case-snapshot",
		Title:             "Snapshot Answer",
		Prompt:            "When did I last work out?",
		ExpectedOutcome:   ExpectedAnswerWithoutTools,
		RequiredTextTerms: []string{"2026-07-03"},
	}}, RunOptions{Mode: ModeSingleTurn})

	result := report.Results[0]
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q reason %q, want pass", result.Passed, result.ScoreStatus, result.ScoreReason)
	}
}

func TestRunScoresAnswerWithoutToolsAllowsListedToolCalls(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			done: &aichat.StreamDone{
				Text:      "You last worked out on July 3rd, 2026.",
				ToolCalls: []string{"get_workouts"},
			},
		}},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:                "case-snapshot-redundant-tool",
		Title:             "Snapshot Answer With Redundant Tool Call",
		Prompt:            "When did I last work out?",
		ExpectedOutcome:   ExpectedAnswerWithoutTools,
		RequiredTextTerms: []string{"2026-07-03"},
		AllowedToolCalls:  []string{"get_workouts"},
	}}, RunOptions{Mode: ModeSingleTurn})

	result := report.Results[0]
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q reason %q, want pass", result.Passed, result.ScoreStatus, result.ScoreReason)
	}
}

func TestRunScoresAnswerWithoutToolsFailsOnDisallowedToolCall(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			done: &aichat.StreamDone{
				Text:      "You last worked out on 2026-07-03.",
				ToolCalls: []string{"generate_workout_draft"},
			},
		}},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:                "case-snapshot-disallowed-tool",
		Title:             "Snapshot Answer With Disallowed Tool Call",
		Prompt:            "When did I last work out?",
		ExpectedOutcome:   ExpectedAnswerWithoutTools,
		RequiredTextTerms: []string{"2026-07-03"},
		AllowedToolCalls:  []string{"get_workouts"},
	}}, RunOptions{Mode: ModeSingleTurn})

	result := report.Results[0]
	if result.Passed || result.ScoreStatus != ScoreStatusFail {
		t.Fatalf("score = passed %v status %q, want fail", result.Passed, result.ScoreStatus)
	}
}

func TestContainsRequiredTermDateRenderings(t *testing.T) {
	cases := []struct {
		name string
		text string
		term string
		want bool
	}{
		{"iso literal", "Your last squat was 2026-06-30.", "2026-06-30", true},
		{"month day year", "You last worked out on July 3, 2026.", "2026-07-03", true},
		{"month ordinal", "That was on July 3rd.", "2026-07-03", true},
		{"day month", "You trained on 3 July 2026.", "2026-07-03", true},
		{"day of month", "It was the 3rd of July.", "2026-07-03", true},
		{"abbreviated month", "Squats were on Jun 30.", "2026-06-30", true},
		{"day prefix does not match longer day", "That was on July 30, 2026.", "2026-07-03", false},
		{"wrong month", "That was on June 3, 2026.", "2026-07-03", false},
		{"missing date", "You worked out recently.", "2026-07-03", false},
		{"non-date term unchanged", "Bench Press felt strong.", "bench", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := containsRequiredTerm(tc.text, tc.term); got != tc.want {
				t.Fatalf("containsRequiredTerm(%q, %q) = %v, want %v", tc.text, tc.term, got, tc.want)
			}
		})
	}
}

func TestRunAppliesUserIDToEveryStreamCall(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "Any injuries?"}},
			{done: &aichat.StreamDone{Text: "I made a draft.", WorkoutDraft: testDraft()}},
		},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "case-fixture-user",
		Title:           "Fixture User",
		Prompt:          "Build pull.",
		ExpectedOutcome: ExpectedAskOnceThenGenerate,
		FollowUpAnswer:  "No injuries, full gym, 45 minutes.",
	}}, RunOptions{Mode: ModeTwoTurn, UserID: FixtureUserID})

	if !report.Results[0].Passed {
		t.Fatalf("scenario passed = false, reason %q", report.Results[0].ScoreReason)
	}
	if len(runtime.calls) != 2 {
		t.Fatalf("runtime calls = %d, want 2", len(runtime.calls))
	}
	for i, call := range runtime.calls {
		if !call.hasUser || call.userID != FixtureUserID {
			t.Fatalf("call %d user = (%q, %v), want fixture user", i+1, call.userID, call.hasUser)
		}
	}
}

func TestRunScoresNarrowScopeThenGenerateAsPass(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "I can build one workout at a time. Pick one session to start with."}},
			{done: &aichat.StreamDone{Text: "I made a focused draft.", WorkoutDraft: testDraft()}},
		},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "case-narrow-scope",
		Title:           "Weekly Split",
		Prompt:          "Build me a 4-day workout split for the whole week.",
		ExpectedOutcome: ExpectedNarrowScopeBeforeGenerate,
		FollowUpAnswer:  "Let's start with day one as an upper-body workout. No injuries, full gym, 45 minutes.",
	}}, RunOptions{Mode: ModeTwoTurn})

	result := report.Results[0]
	if !result.Passed || result.ScoreStatus != ScoreStatusPass {
		t.Fatalf("score = passed %v status %q reason %q, want pass", result.Passed, result.ScoreStatus, result.ScoreReason)
	}
	if result.Status != StatusStructuredDraft {
		t.Fatalf("Result status = %q, want %q", result.Status, StatusStructuredDraft)
	}
	if report.Summary.PassedCount != 1 || report.Summary.FailedCount != 0 {
		t.Fatalf("summary scores = passed %d failed %d, want 1/0", report.Summary.PassedCount, report.Summary.FailedCount)
	}
	if len(runtime.calls) != 2 {
		t.Fatalf("runtime calls = %d, want second turn after narrowing text", len(runtime.calls))
	}
}

func TestRunScoresNarrowScopeRejectsMissingFocusedDraft(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "I can build one workout at a time. Pick one session to start with."}},
			{done: &aichat.StreamDone{Text: "I can help once you pick a session."}},
		},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "case-narrow-scope-no-draft",
		Title:           "Weekly Split",
		Prompt:          "Build me a 4-day workout split for the whole week.",
		ExpectedOutcome: ExpectedNarrowScopeBeforeGenerate,
		FollowUpAnswer:  "Let's start with day one as an upper-body workout. No injuries, full gym, 45 minutes.",
	}}, RunOptions{Mode: ModeTwoTurn})

	result := report.Results[0]
	if result.Passed || result.ScoreStatus != ScoreStatusFail {
		t.Fatalf("score = passed %v status %q, want fail", result.Passed, result.ScoreStatus)
	}
	if result.ScoreReason != "expected a structured draft after the user narrowed the request" {
		t.Fatalf("ScoreReason = %q, want missing focused draft failure", result.ScoreReason)
	}
}

func TestRunScoresNarrowScopeRejectsVagueFirstTurn(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "Sure, I can help with that."}},
			{done: &aichat.StreamDone{Text: "I made a focused draft.", WorkoutDraft: testDraft()}},
		},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "case-narrow-scope-vague",
		Title:           "Weekly Split",
		Prompt:          "Build me a 4-day workout split for the whole week.",
		ExpectedOutcome: ExpectedNarrowScopeBeforeGenerate,
		FollowUpAnswer:  "Let's start with day one as an upper-body workout. No injuries, full gym, 45 minutes.",
	}}, RunOptions{Mode: ModeTwoTurn})

	result := report.Results[0]
	if result.Passed || result.ScoreStatus != ScoreStatusFail {
		t.Fatalf("score = passed %v status %q, want fail", result.Passed, result.ScoreStatus)
	}
	if result.ScoreReason != "expected the first turn to ask the user to choose one workout or session" {
		t.Fatalf("ScoreReason = %q, want narrow-scope failure", result.ScoreReason)
	}
}

func TestRunScoresNarrowScopeRejectsTextOnlyWholeWeekPlan(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "Here is a 4-day split: Day 1 upper, Day 2 lower, Day 3 push, Day 4 pull."}},
			{done: &aichat.StreamDone{Text: "I made a focused draft.", WorkoutDraft: testDraft()}},
		},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "case-narrow-scope-week-plan",
		Title:           "Weekly Split",
		Prompt:          "Build me a 4-day workout split for the whole week.",
		ExpectedOutcome: ExpectedNarrowScopeBeforeGenerate,
		FollowUpAnswer:  "Let's start with day one as an upper-body workout. No injuries, full gym, 45 minutes.",
	}}, RunOptions{Mode: ModeTwoTurn})

	result := report.Results[0]
	if result.Passed || result.ScoreStatus != ScoreStatusFail {
		t.Fatalf("score = passed %v status %q, want fail", result.Passed, result.ScoreStatus)
	}
	if result.ScoreReason != "expected the first turn to ask the user to choose one workout or session" {
		t.Fatalf("ScoreReason = %q, want narrow-scope failure", result.ScoreReason)
	}
}

func TestNarrowsToSingleWorkoutAcceptsNaturalNarrowingText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "existing no question mark wording",
			text: "I can build one workout at a time. Pick one session to start with.",
			want: true,
		},
		{
			name: "one at a time and which day",
			text: "I can draft workouts one at a time. Which day should we build first?",
			want: true,
		},
		{
			name: "focus on one day first",
			text: "Let's focus on one day first. Which session should we start with?",
			want: true,
		},
		{
			name: "which one prompt",
			text: "I can build one workout at a time. Which one should we do first?",
			want: true,
		},
		{
			name: "please choose prompt",
			text: "I can build one session at a time. Please choose one session to start.",
			want: true,
		},
		{
			name: "which workout question",
			text: "I can build one workout at a time. Which workout should we build first?",
			want: true,
		},
		{
			name: "first workout details question",
			text: "I can help you build individual workout drafts. What would be the focus for your first workout, how long will it be, and what equipment do you have available? Do you have any injuries I should be aware of?",
			want: true,
		},
		{
			name: "meal plan refusal redirects to workout session details",
			text: "I can help you create a workout plan, but I'm unable to provide meal plans.\n\nTo create your workout, I need a few more details:\n* What is your workout focus (e.g., full body, upper body, legs)?\n* How long would you like the session to be?\n* What equipment do you have available, and do you have any injuries?",
			want: true,
		},
		{
			name: "meal plan refusal asks how long session should be",
			text: "I can help you create a workout plan, but I'm unable to provide meal plans. What is your workout focus? How long should the session be? What equipment do you have available?",
			want: true,
		},
		{
			name: "vague text",
			text: "Sure, I can help with that.",
			want: false,
		},
		{
			name: "whole week text plan",
			text: "Here is a 4-day split: Day 1 upper, Day 2 lower, Day 3 push, Day 4 pull.",
			want: false,
		},
		{
			name: "single scope without user choice",
			text: "I can build workouts one at a time.",
			want: false,
		},
		{
			name: "focus statement without user choice",
			text: "Let's focus on one day first.",
			want: false,
		},
		{
			name: "narrow statement without user choice",
			text: "We should narrow to one workout.",
			want: false,
		},
		{
			name: "start with statement without user choice",
			text: "Let's start with one workout.",
			want: false,
		},
		{
			name: "assistant picks workout",
			text: "I'll pick one workout to start.",
			want: false,
		},
		{
			name: "assistant chooses session",
			text: "I'll choose one session for you.",
			want: false,
		},
		{
			name: "assistant selects session",
			text: "I'll select one session for you.",
			want: false,
		},
		{
			name: "which workout statement",
			text: "I know which workout to build: one session.",
			want: false,
		},
		{
			name: "first workout without user choice",
			text: "I can build your first workout after this.",
			want: false,
		},
		{
			name: "meal plan refusal without session details",
			text: "I can help with workouts, but I'm unable to provide meal plans.",
			want: false,
		},
		{
			name: "meal plan refusal asks for workout plan details",
			text: "I can help you create a workout plan, but I'm unable to provide meal plans. What is your workout focus? How long should the plan be? What equipment do you have available?",
			want: false,
		},
		{
			name: "meal plan refusal asks broad plan duration",
			text: "I cannot create meal plans, but I can help with your workout plan. What is your workout focus, how long should the plan run, and what equipment do you have?",
			want: false,
		},
		{
			name: "which session without question mark",
			text: "I can build one workout at a time. Which session should we do first",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := narrowsToSingleWorkout(tt.text); got != tt.want {
				t.Fatalf("narrowsToSingleWorkout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunScoresAskOnceThenGenerateRejectsImmediateDraft(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			done: &aichat.StreamDone{
				Text:         "I revised the draft.",
				WorkoutDraft: testDraft(),
			},
		}},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:              "case-elbow-revision",
		Title:           "Elbow Revision",
		Prompt:          "Swap out anything that bothers my elbow.",
		ExpectedOutcome: ExpectedAskOnceThenGenerate,
		FollowUpAnswer:  "It's mild elbow irritation during deep pressing and skull crushers.",
	}}, RunOptions{Mode: ModeTwoTurn})

	result := report.Results[0]
	if result.Passed || result.ScoreStatus != ScoreStatusFail {
		t.Fatalf("score = passed %v status %q, want fail", result.Passed, result.ScoreStatus)
	}
	if result.ScoreReason != "expected one follow-up question before generating" {
		t.Fatalf("ScoreReason = %q, want immediate draft failure", result.ScoreReason)
	}
}

func TestRunScoresProviderQuotaAndContextErrorsAsOperational(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{err: errors.New("provider failed")},
			{err: errors.New("rate limited: Please retry in 30s")},
			{err: context.DeadlineExceeded},
		},
	}

	report := Run(context.Background(), runtime, []Scenario{
		{ID: "case-provider", Title: "Provider", Prompt: "Build push.", ExpectedOutcome: ExpectedGenerateFirstTurn},
		{ID: "case-quota", Title: "Quota", Prompt: "Build pull.", ExpectedOutcome: ExpectedGenerateFirstTurn},
		{ID: "case-context", Title: "Context", Prompt: "Build legs.", ExpectedOutcome: ExpectedGenerateFirstTurn},
	}, RunOptions{Mode: ModeSingleTurn, MaxAttempts: 1})

	if report.Summary.OperationalErrorCount != 3 {
		t.Fatalf("OperationalErrorCount = %d, want 3", report.Summary.OperationalErrorCount)
	}
	if report.Summary.FailedCount != 0 || report.Summary.PassRate != 0 {
		t.Fatalf("summary scores = failed %d rate %f, want no behavior failures and zero denominator rate", report.Summary.FailedCount, report.Summary.PassRate)
	}
	for _, result := range report.Results {
		if result.ScoreStatus != ScoreStatusOperationalError {
			t.Fatalf("%s score status = %q, want operational_error", result.ID, result.ScoreStatus)
		}
	}
}

func TestRunSelectedScenarioSummaryUsesOnlySelectedScenarios(t *testing.T) {
	scenarios := []Scenario{
		{ID: "case-1", Title: "Skipped", Prompt: "Build push.", ExpectedOutcome: ExpectedGenerateFirstTurn},
		{ID: "case-2", Title: "Selected Pass", Prompt: "Ready upper.", ExpectedOutcome: ExpectedGenerateFirstTurn},
		{ID: "case-3", Title: "Selected Fail", Prompt: "Do not generate.", ExpectedOutcome: ExpectedDoNotGenerate},
	}
	selected, err := FilterScenarios(scenarios, ScenarioSelection{ScenarioIDs: "case-2,case-3"})
	if err != nil {
		t.Fatalf("FilterScenarios() error = %v, want nil", err)
	}
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "I made a draft.", WorkoutDraft: testDraft(), ToolCalls: []string{"generate_workout_draft"}}},
			{done: &aichat.StreamDone{Text: "I made a draft.", WorkoutDraft: testDraft()}},
		},
	}

	report := Run(context.Background(), runtime, selected, RunOptions{Mode: ModeSingleTurn})

	if report.ScenarioCount != 2 || report.Summary.TotalScenarios != 2 {
		t.Fatalf("selected counts = scenario_count %d total %d, want 2/2", report.ScenarioCount, report.Summary.TotalScenarios)
	}
	if report.Summary.PassedCount != 1 || report.Summary.FailedCount != 1 {
		t.Fatalf("score counts = passed %d failed %d, want 1/1", report.Summary.PassedCount, report.Summary.FailedCount)
	}
	if len(runtime.calls) != 2 {
		t.Fatalf("runtime calls = %d, want selected scenarios only", len(runtime.calls))
	}
	if runtime.calls[0].prompt != "Ready upper." || runtime.calls[1].prompt != "Do not generate." {
		t.Fatalf("runtime prompts = %#v, want selected prompts only", runtime.calls)
	}
}

func TestRunWaitsBetweenSelectedScenarios(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{
			{done: &aichat.StreamDone{Text: "First response."}},
			{done: &aichat.StreamDone{Text: "Second response."}},
		},
	}

	var delayedBefore []string
	report := Run(context.Background(), runtime, []Scenario{
		{ID: "case-1", Title: "First", Prompt: "Build push."},
		{ID: "case-2", Title: "Second", Prompt: "Build pull."},
	}, RunOptions{
		Mode:               ModeSingleTurn,
		InterScenarioDelay: time.Millisecond,
		OnScenarioDelay: func(_ time.Duration, next Scenario) {
			delayedBefore = append(delayedBefore, next.ID)
		},
	})

	if report.ScenarioCount != 2 {
		t.Fatalf("ScenarioCount = %d, want 2", report.ScenarioCount)
	}
	if len(delayedBefore) != 1 || delayedBefore[0] != "case-2" {
		t.Fatalf("delayedBefore = %v, want [case-2]", delayedBefore)
	}
}

func TestRunTwoTurnDoesNotAnswerWhenFirstTurnIsTextOnly(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			done: &aichat.StreamDone{
				Model: "test-model",
				Text:  "Here are some general ideas.",
			},
		}},
	}

	report := Run(context.Background(), runtime, []Scenario{{
		ID:             "case-3",
		Title:          "Text Only",
		Prompt:         "Give me ideas.",
		Expectation:    "Should be tracked as text only.",
		FollowUpAnswer: "No injuries.",
	}}, RunOptions{Mode: ModeTwoTurn})

	if report.Results[0].Status != StatusTextOnly {
		t.Fatalf("Result status = %q, want %q", report.Results[0].Status, StatusTextOnly)
	}
	if len(runtime.calls) != 1 {
		t.Fatalf("runtime calls = %d, want 1", len(runtime.calls))
	}
}

func TestClassifyErrorAndRetryDelay(t *testing.T) {
	if got := Classify(nil); got != StatusError {
		t.Fatalf("Classify(nil) = %q, want %q", got, StatusError)
	}

	delay, ok := ParseRetryDelay(errors.New("Please retry in 2.4s"))
	if !ok {
		t.Fatal("ParseRetryDelay() ok = false, want true")
	}
	if delay.String() != "2.4s" {
		t.Fatalf("ParseRetryDelay() = %s, want 2.4s", delay)
	}
}

func TestRunRetryWaitStopsWhenContextCanceled(t *testing.T) {
	runtime := &fakeRuntime{
		next: []runtimeResponse{{
			err: errors.New("provider rate limited: Please retry in 300s"),
		}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := time.Now()
	result := runTurn(ctx, runtime, Scenario{ID: "case-retry"}, "Build legs.", nil, RunOptions{
		OnRetry: func(Scenario, time.Duration, int, int) {
			cancel()
		},
	})

	if result.Status != StatusError {
		t.Fatalf("result.Status = %q, want %q", result.Status, StatusError)
	}
	if result.Error != context.Canceled.Error() {
		t.Fatalf("result.Error = %q, want %q", result.Error, context.Canceled.Error())
	}
	if result.Attempts != 1 {
		t.Fatalf("result.Attempts = %d, want 1", result.Attempts)
	}
	if len(runtime.calls) != 1 {
		t.Fatalf("runtime calls = %d, want 1", len(runtime.calls))
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("runTurn waited %s, want context cancellation before retry sleep completes", elapsed)
	}
}

func testDraft() *workout.CreateWorkoutRequest {
	return &workout.CreateWorkoutRequest{
		Date: "2026-04-25T12:00:00Z",
		Exercises: []workout.ExerciseInput{
			{
				Name: "Dumbbell Row",
				Sets: []workout.SetInput{
					{Reps: 10, SetType: "working"},
					{Reps: 10, SetType: "working"},
				},
			},
		},
	}
}
