package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/aichateval"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
)

const (
	defaultOutputDirName = ".codex/diagrams"
	defaultOutputName    = "fittrack-ai-chat-scenario-sweep.json"
	runTimeout           = 90 * time.Second
)

type stubFeatureAccessReader struct{}

func (stubFeatureAccessReader) ListCurrentUserAccess(context.Context) ([]featureaccess.FeatureAccessGrant, error) {
	return nil, nil
}

func main() {
	mode := flag.String("mode", aichateval.ModeSingleTurn, "eval mode: single_turn or two_turn")
	flag.Parse()
	if err := aichateval.ValidateMode(*mode); err != nil {
		fail("%v", err)
	}
	if err := loadLocalEnv(); err != nil {
		fail("failed to load local env files: %v", err)
	}

	outPath := resolveOutputPath()
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fail("failed to create output directory: %v", err)
	}

	runtimeCtx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	runtime := aichat.NewGenkitRuntime(runtimeCtx, stubFeatureAccessReader{})
	if !runtime.Available() {
		fail("ai chat runtime unavailable. Set GEMINI_API_KEY or GOOGLE_API_KEY in your shell, server/.env, or server/setenv.sh")
	}

	report := aichateval.Run(context.Background(), runtime, aichateval.DefaultScenarios(), aichateval.RunOptions{
		Mode: *mode,
		OnScenario: func(item aichateval.Scenario) {
			fmt.Fprintf(os.Stderr, "Running %s: %s\n", item.ID, item.Title)
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

	fmt.Printf("Wrote scenario report to %s\n", outPath)
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
