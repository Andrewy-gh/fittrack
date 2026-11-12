package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

func TestStartupConfiguresAIRuntimeLogging(t *testing.T) {
	// No parallel global logger/env mutations. The provider is exclusively local.
	oldLogger, oldWriter, oldFlags := slog.Default(), log.Writer(), log.Flags()
	oldStdout := os.Stdout
	t.Cleanup(func() {
		slog.SetDefault(oldLogger)
		log.SetOutput(oldWriter)
		log.SetFlags(oldFlags)
		os.Stdout = oldStdout
	})
	var providerCalls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerCalls.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"Safe response\"}]},\"finishReason\":\"STOP\"}]}\n\n")
	}))
	defer provider.Close()
	t.Setenv("GEMINI_API_KEY", "private-provider-secret")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GOOGLE_GEMINI_BASE_URL", provider.URL)
	t.Setenv("GEMINI_MODEL", "googleai/gemini-2.5-flash")
	t.Setenv("AI_CHAT_TRACE_LOGS", "true")
	for _, level := range []string{"debug", "info", "warn", "error"} {
		t.Run(level, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "startup-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			os.Stdout = output
			// Fail closed in startup before opening any DB or auth connection. Logging
			// is a process-startup concern and must be configured before dependencies.
			_, err = New(context.Background(), &config.Config{LogLevel: level, Environment: "production", DatabaseURL: "postgres://unused:unused@localhost:1/unused"})
			if err == nil || !strings.Contains(err.Error(), "session pooler") {
				t.Fatalf("expected pre-network startup rejection: %v", err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			runtime := aichat.NewGenkitRuntime(ctx, nil)
			if !runtime.Available() {
				t.Fatal("local provider runtime unavailable")
			}
			done, err := runtime.StreamChat(user.WithContext(ctx, "private-user-id"), "private-prompt-do-not-log", nil, func(string) error { return nil })
			if err != nil || done == nil || done.Text != "Safe response" {
				t.Fatalf("local provider failed: done=%v err=%v", done, err)
			}
			if _, err := output.Seek(0, 0); err != nil {
				t.Fatal(err)
			}
			traces := 0
			scanner := bufio.NewScanner(output)
			for scanner.Scan() {
				var entry map[string]any
				if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
					t.Fatalf("non-JSON startup/AI log: %s", scanner.Text())
				}
				if strings.Contains(scanner.Text(), "private-") {
					t.Fatalf("sensitive input in log: %s", scanner.Text())
				}
				if entry["msg"] == "ai chat stream trace" {
					traces++
				}
				if (level == "error" && entry["level"] != "ERROR") || (level == "warn" && entry["level"] != "WARN" && entry["level"] != "ERROR") {
					t.Errorf("LOG_LEVEL=%s emitted %v", level, entry)
				}
			}
			if err := scanner.Err(); err != nil {
				t.Fatal(err)
			}
			if (level == "info" || level == "debug") && traces == 0 {
				t.Error("AI runtime trace missing from startup-configured JSON output")
			}
			if (level == "error" || level == "warn") && traces != 0 {
				t.Error("AI trace ignored configured level")
			}
		})
	}
	if providerCalls.Load() != 4 {
		t.Fatalf("local provider calls=%d want 4", providerCalls.Load())
	}
}
