package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

const (
	defaultOutputDirName = ".codex/diagrams"
	defaultOutputName    = "fittrack-ai-chat-scenario-sweep.json"
	runTimeout           = 90 * time.Second
	maxAttempts          = 3
)

type scenario struct {
	ID          string                      `json:"id"`
	Title       string                      `json:"title"`
	Prompt      string                      `json:"prompt"`
	Expectation string                      `json:"expectation"`
	History     []aichat.RuntimeChatMessage `json:"history,omitempty"`
}

type scenarioResult struct {
	ID           string                        `json:"id"`
	Title        string                        `json:"title"`
	Prompt       string                        `json:"prompt"`
	Expectation  string                        `json:"expectation"`
	History      []historyMessage              `json:"history,omitempty"`
	Status       string                        `json:"status"`
	Text         string                        `json:"text,omitempty"`
	Error        string                        `json:"error,omitempty"`
	Model        string                        `json:"model,omitempty"`
	DurationMS   int64                         `json:"duration_ms"`
	Draft        *workout.CreateWorkoutRequest `json:"draft,omitempty"`
	DraftSummary *draftSummary                 `json:"draft_summary,omitempty"`
	Attempts     int                           `json:"attempts"`
}

type historyMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type draftSummary struct {
	ExerciseCount int      `json:"exercise_count"`
	TotalSets     int      `json:"total_sets"`
	WorkingSets   int      `json:"working_sets"`
	ExerciseNames []string `json:"exercise_names"`
}

type report struct {
	GeneratedAt   string           `json:"generated_at"`
	Model         string           `json:"model"`
	ScenarioCount int              `json:"scenario_count"`
	Results       []scenarioResult `json:"results"`
}

type stubFeatureAccessReader struct{}

var retryDelayPattern = regexp.MustCompile(`Please retry in ([0-9.]+)s`)

func (stubFeatureAccessReader) ListCurrentUserAccess(context.Context) ([]featureaccess.FeatureAccessGrant, error) {
	return nil, nil
}

func main() {
	if err := loadLocalEnv(); err != nil {
		fail("failed to load local env files: %v", err)
	}

	outPath := resolveOutputPath()
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fail("failed to create output directory: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	runtime := aichat.NewGenkitRuntime(ctx, stubFeatureAccessReader{})
	if !runtime.Available() {
		fail("ai chat runtime unavailable. Set GEMINI_API_KEY or GOOGLE_API_KEY in your shell, server/.env, or server/setenv.sh")
	}

	scenarios := scenarioSet()
	results := make([]scenarioResult, 0, len(scenarios))
	for _, item := range scenarios {
		fmt.Fprintf(os.Stderr, "Running %s: %s\n", item.ID, item.Title)
		results = append(results, runScenario(runtime, item))
	}

	payload := report{
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Model:         runtime.ModelName(),
		ScenarioCount: len(results),
		Results:       results,
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fail("failed to marshal report: %v", err)
	}
	if err := os.WriteFile(outPath, body, 0o644); err != nil {
		fail("failed to write report: %v", err)
	}

	fmt.Printf("Wrote scenario report to %s\n", outPath)
}

func runScenario(runtime *aichat.GenkitRuntime, item scenario) scenarioResult {
	started := time.Now()
	result := scenarioResult{
		ID:          item.ID,
		Title:       item.Title,
		Prompt:      item.Prompt,
		Expectation: item.Expectation,
		History:     convertHistory(item.History),
	}

	var (
		done *aichat.StreamDone
		err  error
	)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.Attempts = attempt
		done, err = runtime.StreamChat(context.Background(), item.Prompt, item.History, func(string) error {
			return nil
		})
		if err == nil {
			break
		}

		retryDelay, ok := parseRetryDelay(err)
		if !ok || attempt == maxAttempts {
			break
		}

		waitFor := time.Duration(math.Ceil(retryDelay.Seconds()+1)) * time.Second
		fmt.Fprintf(os.Stderr, "Rate limited on %s, waiting %s before retry %d/%d\n", item.ID, waitFor, attempt+1, maxAttempts)
		time.Sleep(waitFor)
	}

	result.DurationMS = time.Since(started).Milliseconds()
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}

	result.Text = strings.TrimSpace(done.Text)
	result.Model = done.Model
	result.Draft = done.WorkoutDraft
	result.DraftSummary = summarizeDraft(done.WorkoutDraft)
	switch {
	case done.WorkoutDraft != nil:
		result.Status = "structured_draft"
	case strings.Contains(result.Text, "?"):
		result.Status = "follow_up_question"
	default:
		result.Status = "text_only"
	}

	return result
}

func parseRetryDelay(err error) (time.Duration, bool) {
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

func summarizeDraft(draft *workout.CreateWorkoutRequest) *draftSummary {
	if draft == nil {
		return nil
	}

	summary := &draftSummary{
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

func convertHistory(history []aichat.RuntimeChatMessage) []historyMessage {
	if len(history) == 0 {
		return nil
	}

	items := make([]historyMessage, 0, len(history))
	for _, message := range history {
		items = append(items, historyMessage{
			Role: message.Role,
			Text: message.Text,
		})
	}
	return items
}

func scenarioSet() []scenario {
	return []scenario{
		{
			ID:          "prompt-01",
			Title:       "Home Dumbbell Pull",
			Prompt:      "I've got 45 minutes at home with just adjustable dumbbells and a bench. Give me a pull workout.",
			Expectation: "Should ideally generate a structured draft with realistic dumbbell-and-bench pull volume.",
		},
		{
			ID:          "prompt-02",
			Title:       "Commercial Gym Hypertrophy Legs",
			Prompt:      "I want a 60-minute hypertrophy leg workout at a commercial gym.",
			Expectation: "Should ask about injuries first or generate only if it can proceed after asking once; when it generates the draft should feel full enough for 60 minutes.",
		},
		{
			ID:          "prompt-03",
			Title:       "Lower Body With Knee Pain",
			Prompt:      "I have knee pain. Build me a lower-body workout.",
			Expectation: "Should gather missing readiness info first and avoid obvious knee-aggravating exercise choices if it generates.",
		},
		{
			ID:          "prompt-04",
			Title:       "Hotel Room No Equipment",
			Prompt:      "Hotel room, 25 minutes, no equipment, full body.",
			Expectation: "Should ask about injuries first or generate a concise feasible hotel-room bodyweight draft.",
		},
		{
			ID:          "prompt-05",
			Title:       "Beginner Upper Body Full Gym",
			Prompt:      "Beginner, 40 minutes, upper body, full gym, no injuries.",
			Expectation: "Should generate a structured beginner-appropriate draft without needing more info.",
		},
		{
			ID:          "prompt-06",
			Title:       "Shoulder Issue Push Workout",
			Prompt:      "I tore up my shoulder recently, but I still want a push workout with dumbbells.",
			Expectation: "Should gather missing readiness info first and avoid obvious shoulder-conflict push exercises if it generates.",
		},
		{
			ID:          "prompt-07",
			Title:       "Bodyweight Strength Full Body",
			Prompt:      "Bodyweight only, 30 minutes, strength-focused full body.",
			Expectation: "Should ask about injuries first or generate a realistic bodyweight strength-leaning draft.",
		},
		{
			ID:          "prompt-08",
			Title:       "Barbell-Only Pull Day",
			Prompt:      "No machines, no cables, just barbells and plates, 50 minutes, pull day.",
			Expectation: "Should avoid cable and machine work if it generates.",
		},
		{
			ID:          "prompt-09",
			Title:       "Mobility Rehab Back",
			Prompt:      "I only want 15 minutes of mobility and rehab work for my back.",
			Expectation: "Should allow intentionally low-volume work instead of forcing a normal lifting session.",
		},
		{
			ID:    "prompt-10",
			Title: "Elbow-Sensitive Revision Follow-Up",
			History: []aichat.RuntimeChatMessage{
				{Role: "user", Text: "Build me a 45 minute push workout with dumbbells and a bench. No injuries."},
				{Role: "assistant", Text: "I put together a structured workout draft for you."},
			},
			Prompt:      "Swap out anything that bothers my elbow.",
			Expectation: "Should ask a focused revision follow-up or adapt the next draft around the new elbow constraint.",
		},
	}
}

func loadLocalEnv() error {
	files := existingEnvFiles(".env", "setenv.sh")
	if len(files) == 0 {
		return nil
	}
	if err := godotenv.Load(files...); err != nil {
		return fmt.Errorf("load %v: %w", files, err)
	}
	return nil
}

func existingEnvFiles(paths ...string) []string {
	files := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}
	return files
}

func resolveOutputPath() string {
	if explicit := strings.TrimSpace(os.Getenv("FITTRACK_AI_CHAT_SWEEP_OUT")); explicit != "" {
		return explicit
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultOutputName
	}
	return filepath.Join(home, defaultOutputDirName, defaultOutputName)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
