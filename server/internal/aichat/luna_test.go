package aichat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type responseTransport func(*http.Request) (*http.Response, error)

func (f responseTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestLunaContinuesStatelessToolCall(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test")
	var calls int
	old := http.DefaultClient.Transport
	http.DefaultClient.Transport = responseTransport(func(req *http.Request) (*http.Response, error) {
		var body struct {
			Store   bool
			Include []string
			Input   []struct {
				Type      string
				CallID    string `json:"call_id"`
				Encrypted string `json:"encrypted_content"`
				Output    string
			}
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Store || len(body.Include) != 1 || body.Include[0] != "reasoning.encrypted_content" {
			t.Fatal("expected encrypted stateless continuity")
		}
		calls++
		var out string
		switch calls {
		case 1:
			out = `[{"type":"reasoning","id":"rs_one","encrypted_content":"opaque-example","summary":[]},{"type":"function_call","call_id":"call_one","name":"lookup","arguments":"{}"}]`
		case 2:
			items := body.Input
			if len(items) < 3 {
				t.Fatal("missing tool continuation")
			}
			items = items[len(items)-3:]
			if items[0].Type != "reasoning" || items[0].Encrypted != "opaque-example" || items[1].Type != "function_call" || items[1].CallID != "call_one" || items[2].Type != "function_call_output" || items[2].CallID != "call_one" || !strings.Contains(items[2].Output, "found") {
				t.Fatalf("invalid tool continuation: %+v", items)
			}
			out = `[{"type":"message","content":[{"type":"output_text","text":"Found it."}]}]`
		default:
			t.Fatalf("unexpected request %d", calls)
		}
		stream := "data: " + `{"type":"response.completed","response":{"status":"completed","output":` + out + "}}\n\n"
		if calls == 1 {
			stream = "data: " + `{"type":"response.output_text.delta","delta":"Checking"}` + "\n\ndata: " + `{"type":"response.output_text.delta","delta":" "}` + "\n\n" + stream
		}
		if calls == 2 {
			stream = "data: " + `{"type":"response.output_text.delta","delta":"Found"}` + "\n\ndata: " + `{"type":"response.output_text.delta","delta":" "}` + "\n\ndata: " + `{"type":"response.output_text.delta","delta":"it."}` + "\n\n" + stream
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(stream)), Request: req}, nil
	})
	defer func() { http.DefaultClient.Transport = old }()
	t.Setenv("AI_CHAT_PRIMARY_MODEL", lunaModel)
	t.Setenv("AI_CHAT_FALLBACK", "")
	r := NewGenkitRuntime(context.Background(), nil, config.LoadAIChat())
	r.workoutDraftTool = genkit.DefineTool(r.g, "lookup", "lookup", func(ctx *ai.ToolContext, input struct{}) (string, error) { return "found", nil })
	var streamed strings.Builder
	done, err := r.StreamChat(context.Background(), "Look it up.", nil, func(delta string) error {
		if strings.TrimSpace(delta) == "" {
			t.Fatalf("unpersistable chunk %q", delta)
		}
		streamed.WriteString(delta)
		return nil
	})
	if err != nil || done.Text != "Found it." || calls != 2 || streamed.String() != "Checking Found it." {
		t.Fatalf("done=%v calls=%d err=%v", done, calls, err)
	}
}
