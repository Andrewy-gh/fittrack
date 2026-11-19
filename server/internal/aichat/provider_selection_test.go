package aichat

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"
)

type stalledSelectionRepository struct {
	Repository
	selections []string
	failure    error
}

func (r *stalledSelectionRepository) SetRunModel(ctx context.Context, prepared *PreparedMessageStream, model string) error {
	r.selections = append(r.selections, model)
	<-ctx.Done()
	return ctx.Err()
}

func (r *stalledSelectionRepository) FailRun(_ context.Context, _ *PreparedMessageStream, _ string, failure error, _ time.Time) error {
	r.failure = failure
	return nil
}

func TestDeadlineWhileRecordingFallbackModel(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "fixture")
	t.Setenv("GEMINI_API_KEY", "fixture")
	requests := 0
	old := http.DefaultClient.Transport
	http.DefaultClient.Transport = responseTransport(func(req *http.Request) (*http.Response, error) {
		requests++
		return &http.Response{StatusCode: 503, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader("{}")), Request: req}, nil
	})
	t.Cleanup(func() { http.DefaultClient.Transport = old })
	primary := newModelRuntime(context.Background(), nil, lunaModel)
	standby := newModelRuntime(context.Background(), nil, defaultModelName)
	primary.fallback = standby
	repo := &stalledSelectionRepository{}
	service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, primary, repo, nil)
	prepared := &PreparedMessageStream{
		Conversation:     &Conversation{ID: 1},
		Run:              &ChatRun{ID: 2, Model: lunaModel},
		AssistantMessage: &ChatMessage{ID: 3},
		Prompt:           "Show me my workouts",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done, err := service.StreamMessage(ctx, prepared, nil)

	if len(repo.selections) != 1 || repo.selections[0] != standby.ModelName() || requests != 1 {
		t.Fatalf("fixture did not time out recording Gemini before standby generation: selections=%v requests=%d", repo.selections, requests)
	}
	if done != nil || !errors.Is(err, ErrGenerationTimeout) || !errors.Is(repo.failure, ErrGenerationTimeout) {
		t.Fatalf("expected FitTrack generation timeout from service and in run failure, got done=%+v err=%v persisted=%v", done, err, repo.failure)
	}
}
