package aichat

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestTraceToggleChangesWithoutRestart(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, nil))
	for _, step := range []struct {
		value, stage string
		enabled      bool
	}{
		{"", "disabled", false}, {" YES ", "enabled", true}, {"invalid", "invalid", false}, {"on", "enabled-again", true},
	} {
		t.Setenv("AI_CHAT_TRACE_LOGS", step.value)
		logAIChatTrace(logger, step.stage)
		if got := strings.Contains(output.String(), "stage="+step.stage); got != step.enabled {
			t.Fatalf("stage %s logged = %v, want %v", step.stage, got, step.enabled)
		}
	}
}
