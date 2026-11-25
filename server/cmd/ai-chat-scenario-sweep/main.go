package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/joho/godotenv"
	"google.golang.org/genai"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/aichateval"
)

const (
	defaultOutputDirName = "tmp/ai-chat-scenario-sweeps"
	defaultOutputName    = "fittrack-ai-chat-scenario-sweep.json"
	defaultLogName       = "fittrack-ai-chat-scenario-sweep-runs.jsonl"
	defaultRunTimeout    = 15 * time.Minute
	defaultScenarioDelay = 75 * time.Second
	baseTimeoutHeadroom  = 5 * time.Minute
	scenarioHeadroom     = 90 * time.Second
	geminiAPIKeyEnvVar   = "GEMINI_API_KEY"
	googleAPIKeyEnvVar   = "GOOGLE_API_KEY"
)

func main() {
	mode := flag.String("mode", aichateval.ModeSingleTurn, "eval mode: single_turn or two_turn")
	timeout := flag.Duration("timeout", defaultRunTimeout, "maximum wall-clock runtime for the full scenario sweep")
	scenarioDelay := flag.Duration("scenario-delay", defaultScenarioDelay, "pause between selected scenarios to reduce provider per-minute rate limits")
	scenarioID := flag.String("scenario", "", "run one scenario id from the default pack, for example prompt-03")
	scenarioIDs := flag.String("scenarios", "", "run comma-separated scenario ids from the default pack, for example prompt-03,prompt-04")
	fromID := flag.String("from", "", "run an inclusive scenario id range starting at this id")
	toID := flag.String("to", "", "run an inclusive scenario id range ending at this id")
	withDataFixtures := flag.Bool("with-data-fixtures", false, "include AI chat data Q&A scenarios backed by in-memory workout fixtures")
	weightScenarios := flag.Bool("draft-weights", false, "run the no-history workout draft weight scenarios")
	explicitModel := flag.String("model", "", "select luna or gemini independently, with fallback disabled")
	forcedFallback := flag.Bool("force-luna-fallback", false, "inject a local Luna 503 and verify the weight draft comes from Gemini")
	flag.Parse()
	if *forcedFallback && (!*weightScenarios || *explicitModel != "") {
		fail("force-luna-fallback requires draft-weights and no model override")
	}
	if *explicitModel != "" && *explicitModel != "luna" && *explicitModel != "gemini" {
		fail("model must be luna or gemini")
	}
	if *weightScenarios && *withDataFixtures {
		fail("draft-weights scenarios require empty workout history")
	}
	timeoutExplicit := flagWasSet("timeout")
	if err := aichateval.ValidateMode(*mode); err != nil {
		fail("%v", err)
	}
	if timeoutExplicit && *timeout <= 0 {
		fail("timeout must be greater than zero")
	}
	if *scenarioDelay < 0 {
		fail("scenario-delay must be zero or greater")
	}
	scenarios := aichateval.DefaultScenarios()
	if *weightScenarios {
		scenarios = aichateval.DraftWeightScenarios()
	}
	var dataReader aichat.ChatDataReader
	evalUserID := ""
	if *withDataFixtures {
		dataReader = aichateval.NewFixtureChatDataReader()
		evalUserID = aichateval.FixtureUserID
		scenarios = append(scenarios, aichateval.DataFixtureScenarios()...)
		scenarios = aichateval.FilterBaseOnlyScenarios(scenarios)
	}
	scenarios, err := aichateval.FilterScenarios(scenarios, aichateval.ScenarioSelection{
		ScenarioID:  *scenarioID,
		ScenarioIDs: *scenarioIDs,
		FromID:      *fromID,
		ToID:        *toID,
	})
	if err != nil {
		fail("invalid scenario selection: %v", err)
	}
	runTimeout := selectedRunTimeout(*timeout, timeoutExplicit, len(scenarios), *scenarioDelay, *mode)
	if runTimeout <= 0 {
		fail("timeout must be greater than zero")
	}
	fmt.Fprintf(os.Stderr, "Using sweep timeout %s (source=%s, scenarios=%d, mode=%s, scenario_delay=%s)\n", runTimeout, timeoutSource(timeoutExplicit), len(scenarios), *mode, *scenarioDelay)
	if err := loadLocalEnv(); err != nil {
		fail("failed to load local env files: %v", err)
	}

	outPath := resolveOutputPath()
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fail("failed to create output directory: %v", err)
	}

	runtimeCtx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	if *explicitModel != "" {
		model := ""
		if *explicitModel == "luna" {
			model = "openai/gpt-6-luna"
		}
		if err := os.Setenv("AI_CHAT_PRIMARY_MODEL", model); err != nil {
			fail("select eval model: %v", err)
		}
		if err := os.Unsetenv("AI_CHAT_FALLBACK"); err != nil {
			fail("disable fallback for independent eval: %v", err)
		}
	}
	lunaFailures := &lunaUnavailableTransport{base: http.DefaultTransport}
	if *forcedFallback {
		if err := os.Setenv("AI_CHAT_PRIMARY_MODEL", "openai/gpt-6-luna"); err != nil {
			fail("select Luna for fallback eval: %v", err)
		}
		if err := os.Setenv("AI_CHAT_FALLBACK", "gemini"); err != nil {
			fail("enable Gemini fallback: %v", err)
		}
		http.DefaultClient.Transport = lunaFailures
	}
	runtime := aichat.NewGenkitRuntime(runtimeCtx, dataReader)
	if !runtime.Available() {
		fail("ai chat runtime unavailable. Set the API key for the selected chat model")
	}
	judge := newNarrowScopeJudge(runtimeCtx, aichat.GeminiModelName())

	report := aichateval.Run(runtimeCtx, runtime, scenarios, aichateval.RunOptions{
		Mode:               *mode,
		InterScenarioDelay: selectedScenarioDelay(*scenarioDelay, len(scenarios)),
		UserID:             evalUserID,
		NarrowScopeJudge:   judge,
		OnScenario: func(item aichateval.Scenario) {
			fmt.Fprintf(os.Stderr, "Running %s: %s\n", item.ID, item.Title)
		},
		OnScenarioDelay: func(waitFor time.Duration, next aichateval.Scenario) {
			fmt.Fprintf(os.Stderr, "Waiting %s before %s to reduce provider rate-limit noise\n", waitFor, next.ID)
		},
		OnRetry: func(item aichateval.Scenario, waitFor time.Duration, nextAttempt int, maxAttempts int) {
			fmt.Fprintf(os.Stderr, "Rate limited on %s, waiting %s before retry %d/%d\n", item.ID, waitFor, nextAttempt, maxAttempts)
		},
	})

	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fail("failed to marshal report: %v", err)
	}
	if err := os.WriteFile(outPath, body, 0o644); err != nil {
		fail("failed to write report: %v", err)
	}

	logPath := resolveLogPath()
	if err := appendSweepLog(logPath, report, outPath); err != nil {
		fail("failed to append sweep log: %v", err)
	}

	fmt.Printf("Wrote scenario report to %s\n", outPath)
	fmt.Printf("Appended sweep log to %s\n", logPath)
	if *forcedFallback {
		if lunaFailures.calls.Load() == 0 {
			fail("fallback eval did not attempt Luna")
		}
		for _, result := range report.Results {
			if result.Model != aichat.GeminiModelName() || !result.Passed {
				fail("fallback eval %s did not produce a passing Gemini draft", result.ID)
			}
		}
	}
	if *weightScenarios && report.Summary.PassedCount != len(scenarios) {
		fail("draft weight eval failed: %d/%d passing", report.Summary.PassedCount, len(scenarios))
	}
}

type lunaUnavailableTransport struct {
	base  http.RoundTripper
	calls atomic.Int64
}

func (transport *lunaUnavailableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "api.openai.com" && req.URL.Path == "/v1/responses" {
		transport.calls.Add(1)
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"simulated Luna outage"}}`)),
			Request:    req,
		}, nil
	}
	return transport.base.RoundTrip(req)
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
	return defaultSweepArtifactPath(defaultOutputName)
}

func resolveLogPath() string {
	if explicit := strings.TrimSpace(os.Getenv("FITTRACK_AI_CHAT_SWEEP_LOG")); explicit != "" {
		return explicit
	}
	return defaultSweepArtifactPath(defaultLogName)
}

func defaultSweepArtifactPath(name string) string {
	return filepath.Join(defaultOutputDirName, name)
}

func selectedScenarioDelay(delay time.Duration, scenarioCount int) time.Duration {
	if scenarioCount <= 1 {
		return 0
	}
	return delay
}

func selectedRunTimeout(explicit time.Duration, explicitSet bool, scenarioCount int, delay time.Duration, mode string) time.Duration {
	if explicitSet {
		return explicit
	}

	headroom := scenarioHeadroom
	if mode == aichateval.ModeTwoTurn {
		headroom *= 2
	}
	scaled := time.Duration(scenarioCount)*(delay+headroom) + baseTimeoutHeadroom
	if scaled < defaultRunTimeout {
		return defaultRunTimeout
	}
	return scaled
}

func timeoutSource(explicitSet bool) string {
	if explicitSet {
		return "explicit"
	}
	return "computed_default"
}

func flagWasSet(name string) bool {
	wasSet := false
	flag.Visit(func(item *flag.Flag) {
		if item.Name == name {
			wasSet = true
		}
	})
	return wasSet
}

func newNarrowScopeJudge(ctx context.Context, modelName string) aichateval.NarrowScopeJudge {
	if configuredJudgeAPIKeyEnvVar() == "" {
		fmt.Fprintln(os.Stderr, "Narrow-scope judge unavailable: GEMINI_API_KEY or GOOGLE_API_KEY is not set; term-list fallback will be used")
		return nil
	}

	var g *genkit.Genkit
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				fmt.Fprintf(os.Stderr, "Narrow-scope judge unavailable after Genkit panic: %v; term-list fallback will be used\n", recovered)
				g = nil
			}
		}()
		g = genkit.Init(ctx,
			genkit.WithPlugins(&googlegenai.GoogleAI{}),
			genkit.WithDefaultModel(modelName),
		)
	}()
	if g == nil {
		return nil
	}

	return aichateval.LLMNarrowScopeJudge{
		Generate: func(callCtx context.Context, prompt string) (string, error) {
			output, _, err := genkit.GenerateData[aichateval.NarrowScopeJudgeVerdict](callCtx, g,
				ai.WithModelName(modelName),
				ai.WithConfig(genai.GenerateContentConfig{Temperature: genai.Ptr[float32](0)}),
				ai.WithPrompt(prompt),
			)
			if err != nil {
				return "", err
			}
			body, err := json.Marshal(output)
			if err != nil {
				return "", err
			}
			return string(body), nil
		},
	}
}

func configuredJudgeAPIKeyEnvVar() string {
	if strings.TrimSpace(os.Getenv(geminiAPIKeyEnvVar)) != "" {
		return geminiAPIKeyEnvVar
	}
	if strings.TrimSpace(os.Getenv(googleAPIKeyEnvVar)) != "" {
		return googleAPIKeyEnvVar
	}
	return ""
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
