package aichateval

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

const (
	ModeSingleTurn = "single_turn"
	ModeTwoTurn    = "two_turn"

	StatusFollowUpQuestion = "follow_up_question"
	StatusStructuredDraft  = "structured_draft"
	StatusTextOnly         = "text_only"
	StatusError            = "error"
)

const defaultMaxAttempts = 3

type Runtime interface {
	StreamChat(ctx context.Context, prompt string, history []aichat.RuntimeChatMessage, onChunk func(string) error) (*aichat.StreamDone, error)
	ModelName() string
}

type Scenario struct {
	ID             string                      `json:"id"`
	Title          string                      `json:"title"`
	Prompt         string                      `json:"prompt"`
	Expectation    string                      `json:"expectation"`
	History        []aichat.RuntimeChatMessage `json:"history,omitempty"`
	FollowUpAnswer string                      `json:"follow_up_answer,omitempty"`
}

type RunOptions struct {
	Mode        string
	MaxAttempts int
	OnScenario  func(Scenario)
	OnRetry     func(Scenario, time.Duration, int, int)
}

type Report struct {
	GeneratedAt   string   `json:"generated_at"`
	Mode          string   `json:"mode"`
	Model         string   `json:"model"`
	ScenarioCount int      `json:"scenario_count"`
	Results       []Result `json:"results"`
}

type Result struct {
	ID             string                        `json:"id"`
	Title          string                        `json:"title"`
	Prompt         string                        `json:"prompt"`
	Expectation    string                        `json:"expectation"`
	History        []HistoryMessage              `json:"history,omitempty"`
	FollowUpAnswer string                        `json:"follow_up_answer,omitempty"`
	Status         string                        `json:"status"`
	Text           string                        `json:"text,omitempty"`
	Error          string                        `json:"error,omitempty"`
	Model          string                        `json:"model,omitempty"`
	DurationMS     int64                         `json:"duration_ms"`
	Draft          *workout.CreateWorkoutRequest `json:"draft,omitempty"`
	DraftSummary   *DraftSummary                 `json:"draft_summary,omitempty"`
	Attempts       int                           `json:"attempts"`
	Turns          []TurnResult                  `json:"turns,omitempty"`
}

type TurnResult struct {
	Turn         int                           `json:"turn"`
	Prompt       string                        `json:"prompt"`
	Status       string                        `json:"status"`
	Text         string                        `json:"text,omitempty"`
	Error        string                        `json:"error,omitempty"`
	Model        string                        `json:"model,omitempty"`
	DurationMS   int64                         `json:"duration_ms"`
	Draft        *workout.CreateWorkoutRequest `json:"draft,omitempty"`
	DraftSummary *DraftSummary                 `json:"draft_summary,omitempty"`
	Attempts     int                           `json:"attempts"`
}

type HistoryMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type DraftSummary struct {
	ExerciseCount int      `json:"exercise_count"`
	TotalSets     int      `json:"total_sets"`
	WorkingSets   int      `json:"working_sets"`
	ExerciseNames []string `json:"exercise_names"`
}

var retryDelayPattern = regexp.MustCompile(`Please retry in ([0-9.]+)s`)

func Run(ctx context.Context, runtime Runtime, scenarios []Scenario, options RunOptions) Report {
	mode := normalizeMode(options.Mode)
	results := make([]Result, 0, len(scenarios))
	for _, scenario := range scenarios {
		if options.OnScenario != nil {
			options.OnScenario(scenario)
		}
		results = append(results, runScenario(ctx, runtime, scenario, mode, options))
	}

	return Report{
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Mode:          mode,
		Model:         runtime.ModelName(),
		ScenarioCount: len(results),
		Results:       results,
	}
}

func normalizeMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case "", ModeSingleTurn:
		return ModeSingleTurn
	case ModeTwoTurn:
		return ModeTwoTurn
	default:
		return ModeSingleTurn
	}
}

func runScenario(ctx context.Context, runtime Runtime, scenario Scenario, mode string, options RunOptions) Result {
	result := Result{
		ID:             scenario.ID,
		Title:          scenario.Title,
		Prompt:         scenario.Prompt,
		Expectation:    scenario.Expectation,
		History:        ConvertHistory(scenario.History),
		FollowUpAnswer: scenario.FollowUpAnswer,
	}

	first := runTurn(ctx, runtime, scenario, scenario.Prompt, scenario.History, options)
	applyTurnToResult(&result, first)
	if mode != ModeTwoTurn || first.Status != StatusFollowUpQuestion || strings.TrimSpace(scenario.FollowUpAnswer) == "" {
		return result
	}

	history := append([]aichat.RuntimeChatMessage{}, scenario.History...)
	history = append(history,
		aichat.RuntimeChatMessage{Role: "user", Text: scenario.Prompt},
		aichat.RuntimeChatMessage{Role: "assistant", Text: first.Text},
	)
	second := runTurn(ctx, runtime, scenario, scenario.FollowUpAnswer, history, options)
	second.Turn = 2
	result.Turns = append(result.Turns, second)
	applyTurnToResult(&result, second)

	return result
}

func applyTurnToResult(result *Result, turn TurnResult) {
	result.Status = turn.Status
	result.Text = turn.Text
	result.Error = turn.Error
	result.Model = turn.Model
	result.DurationMS += turn.DurationMS
	result.Draft = turn.Draft
	result.DraftSummary = turn.DraftSummary
	result.Attempts += turn.Attempts
	if len(result.Turns) == 0 {
		turn.Turn = 1
		result.Turns = append(result.Turns, turn)
		return
	}
	result.Turns[len(result.Turns)-1] = turn
}

func runTurn(ctx context.Context, runtime Runtime, scenario Scenario, prompt string, history []aichat.RuntimeChatMessage, options RunOptions) TurnResult {
	started := time.Now()
	result := TurnResult{
		Turn:   1,
		Prompt: prompt,
	}

	var (
		done *aichat.StreamDone
		err  error
	)
	maxAttempts := options.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.Attempts = attempt
		done, err = runtime.StreamChat(ctx, prompt, history, func(string) error {
			return nil
		})
		if err == nil {
			break
		}

		retryDelay, ok := ParseRetryDelay(err)
		if !ok || attempt == maxAttempts {
			break
		}

		waitFor := time.Duration(math.Ceil(retryDelay.Seconds()+1)) * time.Second
		if options.OnRetry != nil {
			options.OnRetry(scenario, waitFor, attempt+1, maxAttempts)
		}
		time.Sleep(waitFor)
	}

	result.DurationMS = time.Since(started).Milliseconds()
	if err != nil {
		result.Status = StatusError
		result.Error = err.Error()
		return result
	}

	result.Text = strings.TrimSpace(done.Text)
	result.Model = done.Model
	result.Draft = done.WorkoutDraft
	result.DraftSummary = SummarizeDraft(done.WorkoutDraft)
	result.Status = Classify(done)
	return result
}

func Classify(done *aichat.StreamDone) string {
	if done == nil {
		return StatusError
	}
	if done.WorkoutDraft != nil {
		return StatusStructuredDraft
	}
	if strings.Contains(strings.TrimSpace(done.Text), "?") {
		return StatusFollowUpQuestion
	}
	return StatusTextOnly
}

func ParseRetryDelay(err error) (time.Duration, bool) {
	if err == nil {
		return 0, false
	}
	matches := retryDelayPattern.FindStringSubmatch(err.Error())
	if len(matches) != 2 {
		return 0, false
	}

	delay, parseErr := time.ParseDuration(matches[1] + "s")
	if parseErr != nil {
		return 0, false
	}
	return delay, true
}

func SummarizeDraft(draft *workout.CreateWorkoutRequest) *DraftSummary {
	if draft == nil {
		return nil
	}

	summary := &DraftSummary{
		ExerciseCount: len(draft.Exercises),
		ExerciseNames: make([]string, 0, len(draft.Exercises)),
	}
	for _, exercise := range draft.Exercises {
		summary.ExerciseNames = append(summary.ExerciseNames, exercise.Name)
		for _, set := range exercise.Sets {
			summary.TotalSets++
			if set.SetType == "working" {
				summary.WorkingSets++
			}
		}
	}
	return summary
}

func ConvertHistory(history []aichat.RuntimeChatMessage) []HistoryMessage {
	if len(history) == 0 {
		return nil
	}

	items := make([]HistoryMessage, 0, len(history))
	for _, message := range history {
		items = append(items, HistoryMessage{
			Role: message.Role,
			Text: message.Text,
		})
	}
	return items
}

func ValidateMode(mode string) error {
	switch mode {
	case ModeSingleTurn, ModeTwoTurn:
		return nil
	default:
		return fmt.Errorf("unsupported mode %q; use %q or %q", mode, ModeSingleTurn, ModeTwoTurn)
	}
}
