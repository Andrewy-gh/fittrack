package aichateval

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
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
}

type runtimeResponse struct {
	done *aichat.StreamDone
	err  error
}

func (f *fakeRuntime) StreamChat(ctx context.Context, prompt string, history []aichat.RuntimeChatMessage, onChunk func(string) error) (*aichat.StreamDone, error) {
	f.calls = append(f.calls, runtimeCall{
		prompt:  prompt,
		history: append([]aichat.RuntimeChatMessage{}, history...),
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
			{done: &aichat.StreamDone{Text: "I made a draft.", WorkoutDraft: testDraft()}},
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
