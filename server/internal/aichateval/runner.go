package aichateval

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
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

const (
	ExpectedGenerateFirstTurn         = "generate_first_turn"
	ExpectedAskOnceThenGenerate       = "ask_once_then_generate"
	ExpectedDoNotGenerate             = "do_not_generate"
	ExpectedNarrowScopeBeforeGenerate = "narrow_scope_before_generating"
	ExpectedReviseWithoutRestart      = "revise_without_restart"
	ExpectedAnswerFromData            = "answer_from_data"
	ExpectedAnswerWithoutTools        = "answer_without_tools"
	ExpectedHistoryInformedDraft      = "history_informed_draft"
	ExpectedProfileUpdated            = "profile_updated"
)

const (
	ScoreStatusPass             = "pass"
	ScoreStatusFail             = "fail"
	ScoreStatusOperationalError = "operational_error"
	ScoreStatusUnscored         = "unscored"
)

const defaultMaxAttempts = 3

type Runtime interface {
	StreamChat(ctx context.Context, prompt string, history []aichat.RuntimeChatMessage, onChunk func(string) error) (*aichat.StreamDone, error)
	ModelName() string
}

type Scenario struct {
	ID                     string                      `json:"id"`
	Title                  string                      `json:"title"`
	Prompt                 string                      `json:"prompt"`
	Expectation            string                      `json:"expectation"`
	ExpectedOutcome        string                      `json:"expected_outcome,omitempty"`
	RequiredTextTerms      []string                    `json:"required_text_terms,omitempty"`
	RequiredToolCalls      []string                    `json:"required_tool_calls,omitempty"`
	ForbiddenTextTerms     []string                    `json:"forbidden_text_terms,omitempty"`
	ForbiddenExerciseTerms []string                    `json:"forbidden_exercise_terms,omitempty"`
	AllowedToolCalls       []string                    `json:"allowed_tool_calls,omitempty"`
	History                []aichat.RuntimeChatMessage `json:"history,omitempty"`
	FollowUpAnswer         string                      `json:"follow_up_answer,omitempty"`
	BaseOnly               bool                        `json:"base_only,omitempty"`
}

type RunOptions struct {
	Mode               string
	MaxAttempts        int
	InterScenarioDelay time.Duration
	UserID             string
	OnScenario         func(Scenario)
	OnScenarioDelay    func(time.Duration, Scenario)
	OnRetry            func(Scenario, time.Duration, int, int)
}

type Report struct {
	GeneratedAt   string   `json:"generated_at"`
	Mode          string   `json:"mode"`
	Model         string   `json:"model"`
	ScenarioCount int      `json:"scenario_count"`
	Summary       Summary  `json:"summary"`
	Results       []Result `json:"results"`
}

type Summary struct {
	TotalScenarios                 int      `json:"total_scenarios"`
	StructuredDraftCount           int      `json:"structured_draft_count"`
	FollowUpQuestionCount          int      `json:"follow_up_question_count"`
	TextOnlyCount                  int      `json:"text_only_count"`
	ErrorCount                     int      `json:"error_count"`
	PassedCount                    int      `json:"passed_count"`
	FailedCount                    int      `json:"failed_count"`
	OperationalErrorCount          int      `json:"operational_error_count"`
	UnscoredCount                  int      `json:"unscored_count"`
	PassRate                       float64  `json:"pass_rate"`
	StructuredOutputConversionRate float64  `json:"structured_output_conversion_rate"`
	TwoTurnFollowUpCount           int      `json:"two_turn_follow_up_count"`
	TwoTurnAttemptCount            int      `json:"two_turn_attempt_count"`
	TwoTurnStructuredDraftCount    int      `json:"two_turn_structured_draft_count"`
	TwoTurnConversionRate          *float64 `json:"two_turn_conversion_rate"`
}

type Result struct {
	ID                     string                        `json:"id"`
	Title                  string                        `json:"title"`
	Prompt                 string                        `json:"prompt"`
	Expectation            string                        `json:"expectation"`
	ExpectedOutcome        string                        `json:"expected_outcome,omitempty"`
	RequiredTextTerms      []string                      `json:"required_text_terms,omitempty"`
	RequiredToolCalls      []string                      `json:"required_tool_calls,omitempty"`
	ForbiddenTextTerms     []string                      `json:"forbidden_text_terms,omitempty"`
	ForbiddenExerciseTerms []string                      `json:"forbidden_exercise_terms,omitempty"`
	AllowedToolCalls       []string                      `json:"allowed_tool_calls,omitempty"`
	History                []HistoryMessage              `json:"history,omitempty"`
	FollowUpAnswer         string                        `json:"follow_up_answer,omitempty"`
	Status                 string                        `json:"status"`
	Passed                 bool                          `json:"passed"`
	ScoreStatus            string                        `json:"score_status"`
	ScoreReason            string                        `json:"score_reason"`
	Text                   string                        `json:"text,omitempty"`
	Error                  string                        `json:"error,omitempty"`
	Model                  string                        `json:"model,omitempty"`
	DurationMS             int64                         `json:"duration_ms"`
	Draft                  *workout.CreateWorkoutRequest `json:"draft,omitempty"`
	DraftSummary           *DraftSummary                 `json:"draft_summary,omitempty"`
	ToolCalls              []string                      `json:"tool_calls,omitempty"`
	Attempts               int                           `json:"attempts"`
	Turns                  []TurnResult                  `json:"turns,omitempty"`
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
	ToolCalls    []string                      `json:"tool_calls,omitempty"`
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
	for index, scenario := range scenarios {
		if index > 0 && options.InterScenarioDelay > 0 && ctx.Err() == nil {
			if options.OnScenarioDelay != nil {
				options.OnScenarioDelay(options.InterScenarioDelay, scenario)
			}
			_ = waitForRetry(ctx, options.InterScenarioDelay)
		}
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
		Summary:       BuildSummary(mode, results),
		Results:       results,
	}
}

func BuildSummary(mode string, results []Result) Summary {
	summary := Summary{
		TotalScenarios: len(results),
	}
	for _, result := range results {
		switch result.Status {
		case StatusStructuredDraft:
			summary.StructuredDraftCount++
		case StatusFollowUpQuestion:
			summary.FollowUpQuestionCount++
		case StatusTextOnly:
			summary.TextOnlyCount++
		case StatusError:
			summary.ErrorCount++
		}

		switch result.ScoreStatus {
		case ScoreStatusPass:
			summary.PassedCount++
		case ScoreStatusFail:
			summary.FailedCount++
		case ScoreStatusOperationalError:
			summary.OperationalErrorCount++
		case ScoreStatusUnscored:
			summary.UnscoredCount++
		}

		if mode == ModeTwoTurn && firstTurnStatus(result) == StatusFollowUpQuestion {
			summary.TwoTurnFollowUpCount++
			if len(result.Turns) > 1 {
				summary.TwoTurnAttemptCount++
				if result.Status == StatusStructuredDraft {
					summary.TwoTurnStructuredDraftCount++
				}
			}
		}
	}

	summary.PassRate = rate(summary.PassedCount, summary.PassedCount+summary.FailedCount)
	summary.StructuredOutputConversionRate = rate(summary.StructuredDraftCount, summary.TotalScenarios)
	if mode == ModeTwoTurn {
		twoTurnRate := rate(summary.TwoTurnStructuredDraftCount, summary.TwoTurnAttemptCount)
		summary.TwoTurnConversionRate = &twoTurnRate
	}

	return summary
}

func firstTurnStatus(result Result) string {
	if len(result.Turns) == 0 {
		return ""
	}
	return result.Turns[0].Status
}

func rate(count int, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total)
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
		ID:                     scenario.ID,
		Title:                  scenario.Title,
		Prompt:                 scenario.Prompt,
		Expectation:            scenario.Expectation,
		ExpectedOutcome:        scenario.ExpectedOutcome,
		RequiredTextTerms:      append([]string{}, scenario.RequiredTextTerms...),
		RequiredToolCalls:      append([]string{}, scenario.RequiredToolCalls...),
		ForbiddenTextTerms:     append([]string{}, scenario.ForbiddenTextTerms...),
		ForbiddenExerciseTerms: append([]string{}, scenario.ForbiddenExerciseTerms...),
		AllowedToolCalls:       append([]string{}, scenario.AllowedToolCalls...),
		History:                ConvertHistory(scenario.History),
		FollowUpAnswer:         scenario.FollowUpAnswer,
	}

	first := runTurn(ctx, runtime, scenario, scenario.Prompt, scenario.History, options)
	applyTurnToResult(&result, first)
	if !shouldRunFollowUpTurn(scenario, mode, first) {
		ScoreResult(&result, mode)
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

	ScoreResult(&result, mode)
	return result
}

func shouldRunFollowUpTurn(scenario Scenario, mode string, first TurnResult) bool {
	if mode != ModeTwoTurn || strings.TrimSpace(scenario.FollowUpAnswer) == "" {
		return false
	}
	if first.Status == StatusFollowUpQuestion {
		return true
	}
	if scenario.ExpectedOutcome != ExpectedNarrowScopeBeforeGenerate {
		return false
	}
	return first.Draft == nil && first.Status != StatusError && strings.TrimSpace(first.Text) != ""
}

func applyTurnToResult(result *Result, turn TurnResult) {
	result.Status = turn.Status
	result.Text = turn.Text
	result.Error = turn.Error
	result.Model = turn.Model
	result.DurationMS += turn.DurationMS
	result.Draft = turn.Draft
	result.DraftSummary = turn.DraftSummary
	result.ToolCalls = append(result.ToolCalls, turn.ToolCalls...)
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
		if ctxErr := ctx.Err(); ctxErr != nil {
			err = ctxErr
			break
		}
		result.Attempts = attempt
		callCtx := ctx
		if strings.TrimSpace(options.UserID) != "" {
			callCtx = user.WithContext(ctx, strings.TrimSpace(options.UserID))
		}
		done, err = runtime.StreamChat(callCtx, prompt, history, func(string) error {
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
		if waitErr := waitForRetry(ctx, waitFor); waitErr != nil {
			err = waitErr
			break
		}
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
	result.ToolCalls = append([]string(nil), done.ToolCalls...)
	result.Status = Classify(done)
	return result
}

func waitForRetry(ctx context.Context, waitFor time.Duration) error {
	timer := time.NewTimer(waitFor)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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
