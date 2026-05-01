package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/aichateval"
)

func TestBuildSweepLogEntryIncludesRunContextAndCompactResults(t *testing.T) {
	report := aichateval.Report{
		GeneratedAt:   "2026-04-29T12:00:00Z",
		Mode:          aichateval.ModeTwoTurn,
		Model:         "googleai/gemini-2.5-flash",
		ScenarioCount: 2,
		Summary: aichateval.Summary{
			TotalScenarios: 2,
			PassedCount:    1,
			FailedCount:    1,
		},
		Results: []aichateval.Result{
			{
				ID:          "prompt-01",
				Title:       "Home Dumbbell Pull",
				Status:      aichateval.StatusStructuredDraft,
				Passed:      true,
				ScoreStatus: aichateval.ScoreStatusPass,
				ScoreReason: "generated after follow-up",
				DurationMS:  1200,
				Attempts:    2,
			},
			{
				ID:          "prompt-02",
				Title:       "Commercial Gym Hypertrophy Legs",
				Status:      aichateval.StatusError,
				ScoreStatus: aichateval.ScoreStatusOperationalError,
				ScoreReason: "provider rate limited",
				Error:       "rate limited",
				DurationMS:  3000,
				Attempts:    3,
			},
		},
	}

	entry := buildSweepLogEntry(report, "report.json", time.Date(2026, 4, 29, 12, 5, 0, 0, time.UTC), gitSnapshot{
		Branch:    "main",
		Commit:    "abc123",
		Dirty:     true,
		Available: true,
	})

	if entry.LoggedAt != "2026-04-29T12:05:00Z" {
		t.Fatalf("LoggedAt = %q", entry.LoggedAt)
	}
	if entry.Git.Branch != "main" || entry.Git.Commit != "abc123" || !entry.Git.Dirty || !entry.Git.Available {
		t.Fatalf("unexpected git snapshot: %+v", entry.Git)
	}
	if got := entry.ScenarioIDs; len(got) != 2 || got[0] != "prompt-01" || got[1] != "prompt-02" {
		t.Fatalf("ScenarioIDs = %v", got)
	}
	if len(entry.Results) != 2 || entry.Results[1].Error != "rate limited" {
		t.Fatalf("unexpected compact results: %+v", entry.Results)
	}
}

func TestAppendSweepLogWritesJsonLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs.jsonl")
	report := aichateval.Report{
		GeneratedAt:   "2026-04-29T12:00:00Z",
		Mode:          aichateval.ModeTwoTurn,
		Model:         "googleai/gemini-2.5-flash",
		ScenarioCount: 1,
		Results: []aichateval.Result{
			{ID: "prompt-01", Title: "Home Dumbbell Pull", Status: aichateval.StatusStructuredDraft, ScoreStatus: aichateval.ScoreStatusPass},
		},
	}

	if err := appendSweepLog(path, report, "report-a.json"); err != nil {
		t.Fatalf("appendSweepLog first call: %v", err)
	}
	if err := appendSweepLog(path, report, "report-b.json"); err != nil {
		t.Fatalf("appendSweepLog second call: %v", err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := splitNonEmptyLines(string(body))
	if len(lines) != 2 {
		t.Fatalf("line count = %d, body = %q", len(lines), body)
	}

	var entry sweepLogEntry
	if err := json.Unmarshal([]byte(lines[1]), &entry); err != nil {
		t.Fatalf("unmarshal second line: %v", err)
	}
	if entry.ReportPath != "report-b.json" || entry.ScenarioIDs[0] != "prompt-01" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}

func TestSelectedScenarioDelaySkipsSingleScenarioRuns(t *testing.T) {
	if got := selectedScenarioDelay(75*time.Second, 1); got != 0 {
		t.Fatalf("selectedScenarioDelay(single) = %s, want 0", got)
	}
	if got := selectedScenarioDelay(75*time.Second, 2); got != 75*time.Second {
		t.Fatalf("selectedScenarioDelay(batch) = %s, want 75s", got)
	}
}

func splitNonEmptyLines(body string) []string {
	var lines []string
	for _, line := range strings.Split(body, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
