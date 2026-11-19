package aichat

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

func TestLunaProviderSwitch(t *testing.T) {
	const textEvent = "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Hello\"}\n\n"
	const completion = "data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"output\":[{\"type\":\"message\",\"content\":[{\"type\":\"output_text\",\"text\":\"Hello\"}]}]}}\n\n"
	const toolEvent = `data: {"type":"response.completed","response":{"status":"completed","output":[{"type":"function_call","call_id":"call_one","name":"generate_workout_draft","arguments":"{\"sessionDuration\":30,\"workoutFocus\":\"full body\",\"equipment\":\"bodyweight\",\"injuries\":\"none\"}"}]}}` + "\n\n"
	for _, tc := range []struct {
		name               string
		status             int
		stream             string
		wantModel, visible string
		failFallback       bool
	}{
		{"primary success", 200, textEvent + completion, lunaModel, "Hello", false},
		{"tool started then unavailable", 200, toolEvent, "", "", false},
		{"user cancellation", 0, "", "", "", false},
		{"first response timeout", -1, "", "standby", "", false},
		{"empty completion", 200, "data: {\"type\":\"response.completed\",\"response\":{\"output\":[]}}\n\n", "standby", "", false},
		{"unavailable", 503, "", "standby", "", false},
		{"rate limited", 429, "", "standby", "", false},
		{"truncated denial", 403, "broken body", "", "", false},
		{"truncated event", 200, "data: {\"type\":\n\n", "standby", "", false},
		{"empty standby", 503, "empty standby", "", "", false},
		{"access denied", 403, "", "", "", false},
		{"both unavailable", 503, "", "", "", true},
		{"disconnect after text", 200, textEvent, "", "Hello", false},
		{"disconnect before text", 200, "", "standby", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			t.Setenv("OPENAI_API_KEY", "sample")
			t.Setenv("AI_CHAT_PRIMARY_MODEL", lunaModel)

			requests := 0
			old := http.DefaultClient.Transport
			http.DefaultClient.Transport = responseTransport(func(req *http.Request) (*http.Response, error) {
				requests++
				if req.URL.Path != "/v1/responses" {
					t.Fatalf("unexpected path %s", req.URL.Path)
				}
				header := http.Header{}
				body := tc.stream
				status := tc.status
				if status == 0 {
					cancel()
					return nil, context.Canceled
				}
				if status == -1 {
					<-req.Context().Done()
					return nil, req.Context().Err()
				}
				if tc.stream == toolEvent && requests > 1 {
					status = 503
				}
				if status == 200 {
					header.Set("Content-Type", "text/event-stream")
				} else {
					header.Set("Content-Type", "application/json")
					body = `{"error":{"message":"test failure"}}`
				}
				var reader io.Reader = strings.NewReader(body)
				if tc.stream == "broken body" {
					reader = iotest.ErrReader(io.ErrUnexpectedEOF)
				}
				return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(reader), Request: req}, nil
			})
			defer func() { http.DefaultClient.Transport = old }()
			primary := newModelRuntime(ctx, nil, lunaModel)
			standbyCalls := 0
			g := genkit.Init(ctx)
			genkit.DefineModel(g, "standby", &ai.ModelOptions{Supports: &ai.ModelSupports{Tools: true, Multiturn: true, SystemRole: true}}, func(ctx context.Context, req *ai.ModelRequest, cb ai.ModelStreamCallback) (*ai.ModelResponse, error) {
				standbyCalls++
				if tc.stream == "empty standby" {
					return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewTextPart("")), FinishReason: ai.FinishReasonStop}, nil
				}
				if tc.failFallback {
					return nil, fmt.Errorf("standby unavailable")
				}
				return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewTextPart("Recovered")), FinishReason: ai.FinishReasonStop}, nil
			})
			primary.fallback = &GenkitRuntime{available: true, modelName: "standby", g: g, workoutDraftTool: defineWorkoutDraftTool(g, "standby")}
			var visible strings.Builder
			done, err := primary.StreamChat(ctx, "hello", nil, func(s string) error { visible.WriteString(s); return nil })
			if (err == nil) != (tc.wantModel != "") || (done != nil && done.Model != tc.wantModel) {
				t.Fatalf("done=%v error=%v", done, err)
			}
			expectedStandby := 0
			if tc.wantModel == "standby" || tc.failFallback {
				expectedStandby = 1
			}
			if tc.stream == "empty standby" {
				expectedStandby = chatEmptyResponseRetryLimit + 1
			}
			expectedRequests := 1
			if tc.stream == toolEvent {
				expectedRequests = 2
			}
			if requests != expectedRequests || standbyCalls != expectedStandby || visible.String() != tc.visible {
				t.Fatalf("primary=%d standby=%d streamed=%q", requests, standbyCalls, visible.String())
			}
		})
	}
}
