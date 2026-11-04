package aichat

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

func TestGenkitRuntime_ProviderStreamTransport(t *testing.T) {
	for _, tc := range []struct {
		name, tail string
		reject     bool
	}{
		{"complete", `data: {"candidates":[{"content":{"role":"model","parts":[{"text":" world"}]},"finishReason":"STOP"}]}` + "\n\n", false},
		{"provider error after text", `data: {"error":{"code":503,"message":"synthetic stream failure","status":"UNAVAILABLE"}}` + "\n\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := make(chan string, 8)
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r.Method + " " + r.URL.Path
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"Hello\"}]}}]}\n\n")
				w.(http.Flusher).Flush()
				fmt.Fprint(w, tc.tail)
			}))
			defer provider.Close()
			t.Setenv("GEMINI_API_KEY", "synthetic-local-test")
			t.Setenv("GOOGLE_API_KEY", "")
			t.Setenv("GOOGLE_GEMINI_BASE_URL", provider.URL)
			t.Setenv("GEMINI_MODEL", "googleai/gemini-2.5-flash")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			runtime := NewGenkitRuntime(ctx, nil)
			if !runtime.Available() {
				t.Fatal("real plugin failed to initialize")
			}
			var chunks strings.Builder
			done, err := runtime.StreamChat(ctx, "Say hello", nil, func(s string) error { chunks.WriteString(s); return nil })
			select {
			case got := <-requests:
				if got != "POST /v1beta/models/gemini-2.5-flash:streamGenerateContent" {
					t.Fatalf("unexpected provider request %q", got)
				}
			default:
				t.Fatalf("provider was never called: %v", err)
			}
			if tc.reject {
				if err == nil || done != nil {
					t.Fatalf("provider error lost: done=%v err=%v", done, err)
				}
				return
			}
			if err != nil || done == nil || done.Text != "Hello world" || chunks.String() != "Hello world" {
				t.Fatalf("done=%v chunks=%q err=%v", done, chunks.String(), err)
			}
		})
	}
}

type compatibilityRecoveryRunner struct {
	request RunRecoveryRequest
	userID  string
	calls   int
}

func (r *compatibilityRecoveryRunner) RecoverStreamingRun(ctx context.Context, req RunRecoveryRequest) error {
	r.request = req
	r.userID, _ = user.Current(ctx)
	r.calls++
	return nil
}

func TestInngestRecovery_SignedTransport(t *testing.T) {
	t.Setenv("INNGEST_DEV", "0")
	t.Setenv("INNGEST_SIGNING_KEY", "signkey-test-0123456789abcdef0123456789abcdef")
	t.Setenv("INNGEST_EVENT_KEY", "synthetic-event-key")
	t.Setenv("INNGEST_ENABLE_UNAUTHED_SYNC", "false")
	events := make(chan []map[string]any, 1)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/e/synthetic-event-key" {
			t.Errorf("unexpected event endpoint %s %s", r.Method, r.URL.Path)
		}
		var batch []map[string]any
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			t.Error(err)
		}
		events <- batch
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":200,"ids":["synthetic-event-id"]}`)
	}))
	defer provider.Close()
	t.Setenv("INNGEST_EVENT_API_BASE_URL", provider.URL)
	runner := &compatibilityRecoveryRunner{}
	recovery, err := NewInngestRecovery(slog.New(slog.NewTextHandler(io.Discard, nil)), runner)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	request := RunRecoveryRequest{ConversationID: 11, RunID: 22, UserID: "synthetic-user", Reason: "test"}
	if err := recovery.EnqueueRunRecovery(ctx, request); err != nil {
		t.Fatal(err)
	}
	batch := <-events
	if len(batch) != 1 || batch[0]["name"] != inngestRecoveryEventName || batch[0]["id"] != "ai-chat-run-recovery-22" {
		t.Fatalf("wrong recovery event: %v", batch)
	}
	data, ok := batch[0]["data"].(map[string]any)
	if !ok || data["user_id"] != request.UserID || data["run_id"] != float64(22) || data["conversation_id"] != float64(11) {
		t.Fatalf("wrong recovery data: %v", data)
	}
	// Canonical JSON and an independent HMAC signer exercise the wire contract,
	// rather than using the SDK to both manufacture and verify signatures.
	sign := func(body string, at time.Time) string {
		mac := hmac.New(sha256.New, []byte("0123456789abcdef0123456789abcdef"))
		fmt.Fprintf(mac, "%s%d", body, at.Unix())
		return fmt.Sprintf("t=%d&s=%x", at.Unix(), mac.Sum(nil))
	}
	for _, tc := range []struct {
		name, method, path, body string
		valid                    bool
	}{
		{"unsigned sync", "PUT", "/inngest", `{"url":"http://localhost/inngest"}`, false},
		{"signed sync", "PUT", "/inngest", `{"url":"http://localhost/inngest"}`, true},
		{"unsigned recovery", "POST", "/inngest?fnId=fittrack-api-ai-chat-run-recovery", `{"ctx":{"run_id":"01ARZ3NDEKTSV4RRFFQ69G5FAV"},"event":{"data":{"conversation_id":11,"reason":"test","run_id":22,"user_id":"synthetic-user"},"name":"fittrack/ai-chat.run-recovery.requested"}}`, false},
		{"signed recovery", "POST", "/inngest?fnId=fittrack-api-ai-chat-run-recovery", `{"ctx":{"run_id":"01ARZ3NDEKTSV4RRFFQ69G5FAV"},"event":{"data":{"conversation_id":11,"reason":"test","run_id":22,"user_id":"synthetic-user"},"name":"fittrack/ai-chat.run-recovery.requested"}}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)).WithContext(ctx)
			if tc.valid {
				req.Header.Set("X-Inngest-Signature", sign(tc.body, time.Now()))
			}
			response := httptest.NewRecorder()
			recovery.Handler().ServeHTTP(response, req)
			if !tc.valid {
				if response.Code != http.StatusUnauthorized {
					t.Fatalf("unsigned request status=%d body=%s", response.Code, response.Body.String())
				}
				return
			}
			if response.Code < 200 || response.Code >= 300 {
				t.Fatalf("signed request status=%d body=%s", response.Code, response.Body.String())
			}
			if tc.method == "PUT" && !strings.Contains(response.Body.String(), inngestRecoveryFunction) {
				t.Fatal("sync omitted recovery function")
			}
		})
	}
	if runner.calls != 1 || runner.request != request || runner.userID != request.UserID {
		t.Fatalf("recovery dispatch calls=%d request=%+v user=%q", runner.calls, runner.request, runner.userID)
	}
}
