package aichateval

import (
	"context"
	"errors"
	"testing"

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
		ID:          "case-1",
		Title:       "Ready Prompt",
		Prompt:      "Beginner upper body, full gym, no injuries.",
		Expectation: "Should generate.",
	}}, RunOptions{Mode: ModeSingleTurn})

	if report.Mode != ModeSingleTurn {
		t.Fatalf("Run() mode = %q, want %q", report.Mode, ModeSingleTurn)
	}
	if len(report.Results) != 1 {
		t.Fatalf("Run() returned %d results, want 1", len(report.Results))
	}
	result := report.Results[0]
	if result.Status != StatusStructuredDraft {
		t.Fatalf("Result status = %q, want %q", result.Status, StatusStructuredDraft)
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
		ID:             "case-2",
		Title:          "Follow Up",
		Prompt:         "Give me a pull workout.",
		Expectation:    "Should ask, then generate.",
		FollowUpAnswer: "No injuries, dumbbells and bench, 45 minutes.",
	}

	report := Run(context.Background(), runtime, []Scenario{scenario}, RunOptions{Mode: ModeTwoTurn})

	result := report.Results[0]
	if result.Status != StatusStructuredDraft {
		t.Fatalf("Result status = %q, want %q", result.Status, StatusStructuredDraft)
	}
	if result.Text != "I made a draft." {
		t.Fatalf("Result text = %q, want final turn text", result.Text)
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
