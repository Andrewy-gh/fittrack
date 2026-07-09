package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/aichateval"
)

type sweepLogEntry struct {
	LoggedAt          string             `json:"logged_at"`
	ReportGeneratedAt string             `json:"report_generated_at"`
	Mode              string             `json:"mode"`
	Model             string             `json:"model"`
	ReportPath        string             `json:"report_path"`
	Git               gitSnapshot        `json:"git"`
	ScenarioIDs       []string           `json:"scenario_ids"`
	ScenarioCount     int                `json:"scenario_count"`
	Summary           aichateval.Summary `json:"summary"`
	Results           []sweepLogResult   `json:"results"`
}

type gitSnapshot struct {
	Branch    string `json:"branch,omitempty"`
	Commit    string `json:"commit,omitempty"`
	Dirty     bool   `json:"dirty"`
	Available bool   `json:"available"`
}

type sweepLogResult struct {
	ID                    string                              `json:"id"`
	Title                 string                              `json:"title"`
	Status                string                              `json:"status"`
	Passed                bool                                `json:"passed"`
	ScoreStatus           string                              `json:"score_status"`
	ScoreReason           string                              `json:"score_reason"`
	Error                 string                              `json:"error,omitempty"`
	DurationMS            int64                               `json:"duration_ms"`
	Attempts              int                                 `json:"attempts"`
	NarrowScopeJudge      *aichateval.NarrowScopeJudgeVerdict `json:"narrow_scope_judge,omitempty"`
	NarrowScopeJudgeError string                              `json:"narrow_scope_judge_error,omitempty"`
}

func appendSweepLog(path string, report aichateval.Report, reportPath string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	body, err := json.Marshal(buildSweepLogEntry(report, reportPath, time.Now().UTC(), currentGitSnapshot()))
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}
	body = append(body, '\n')

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(body); err != nil {
		return fmt.Errorf("write log entry: %w", err)
	}
	return nil
}

func buildSweepLogEntry(report aichateval.Report, reportPath string, loggedAt time.Time, git gitSnapshot) sweepLogEntry {
	results := make([]sweepLogResult, 0, len(report.Results))
	scenarioIDs := make([]string, 0, len(report.Results))
	for _, result := range report.Results {
		scenarioIDs = append(scenarioIDs, result.ID)
		results = append(results, sweepLogResult{
			ID:                    result.ID,
			Title:                 result.Title,
			Status:                result.Status,
			Passed:                result.Passed,
			ScoreStatus:           result.ScoreStatus,
			ScoreReason:           result.ScoreReason,
			Error:                 result.Error,
			DurationMS:            result.DurationMS,
			Attempts:              result.Attempts,
			NarrowScopeJudge:      result.NarrowScopeJudge,
			NarrowScopeJudgeError: result.NarrowScopeJudgeError,
		})
	}

	return sweepLogEntry{
		LoggedAt:          loggedAt.Format(time.RFC3339),
		ReportGeneratedAt: report.GeneratedAt,
		Mode:              report.Mode,
		Model:             report.Model,
		ReportPath:        reportPath,
		Git:               git,
		ScenarioIDs:       scenarioIDs,
		ScenarioCount:     report.ScenarioCount,
		Summary:           report.Summary,
		Results:           results,
	}
}

func currentGitSnapshot() gitSnapshot {
	branch, branchErr := gitOutput("branch", "--show-current")
	commit, commitErr := gitOutput("rev-parse", "HEAD")
	status, statusErr := gitOutput("status", "--porcelain")
	return gitSnapshot{
		Branch:    branch,
		Commit:    commit,
		Dirty:     strings.TrimSpace(status) != "",
		Available: branchErr == nil && commitErr == nil && statusErr == nil,
	}
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
