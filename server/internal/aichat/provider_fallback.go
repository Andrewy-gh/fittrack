package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
)

type providerSelectedKey struct{}
type fallbackAttemptKey struct{}
type fallbackAttempt struct {
	phase     atomic.Int32
	transient atomic.Bool
	stop      func() bool
}

const (
	waitingForOutput int32 = iota
	providerCommitted
	providerTimedOut
)

var errPrimarySlow = errors.New("primary did not start responding in time")

// Once the model emits text or hands a tool call to Genkit, replay could
// duplicate user-visible output or an already-applied profile update.
func commitProvider(ctx context.Context) bool {
	attempt, ok := ctx.Value(fallbackAttemptKey{}).(*fallbackAttempt)
	if !ok {
		return true
	}
	if attempt.phase.CompareAndSwap(waitingForOutput, providerCommitted) {
		attempt.stop()
		return true
	}
	return attempt.phase.Load() == providerCommitted
}

func markTransientProviderFailure(ctx context.Context) {
	if a, ok := ctx.Value(fallbackAttemptKey{}).(*fallbackAttempt); ok {
		a.transient.Store(true)
	}
}

func markProviderError(ctx context.Context, response *http.Response, err error) {
	attempt, ok := ctx.Value(fallbackAttemptKey{}).(*fallbackAttempt)
	if !ok {
		return
	}
	// Use the received status even when the SDK cannot read its error body.
	if response != nil && response.StatusCode >= 400 {
		status := response.StatusCode
		attempt.transient.Store(status == 408 || status == 429 || status >= 500)
		return
	}
	var networkErr net.Error
	var malformedEvent *json.SyntaxError
	if errors.As(err, &networkErr) || errors.As(err, &malformedEvent) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		attempt.transient.Store(true)
	}

}

func (r *GenkitRuntime) streamWithFallback(ctx context.Context, prompt string, history []RuntimeChatMessage, onChunk func(string) error) (*StreamDone, error) {
	turn, cancel := context.WithTimeout(ctx, chatGenerationTimeout)
	defer cancel()
	primary, stopPrimary := context.WithCancelCause(turn)
	defer stopPrimary(nil)
	attempt := &fallbackAttempt{}
	timer := time.AfterFunc(20*time.Second, func() {
		if attempt.phase.CompareAndSwap(waitingForOutput, providerTimedOut) {
			stopPrimary(errPrimarySlow)
		}
	})
	attempt.stop = timer.Stop
	defer timer.Stop()
	primary = context.WithValue(primary, fallbackAttemptKey{}, attempt)
	result, err := r.streamModel(primary, prompt, history, onChunk)
	slow := attempt.phase.Load() == providerTimedOut
	if turn.Err() != nil {
		return nil, turn.Err()
	}
	if attempt.phase.Load() == providerCommitted || (!slow && (err == nil || !attempt.transient.Load())) {
		return result, err
	}
	timer.Stop()
	slog.InfoContext(ctx, "switching AI chat provider", "primary", r.modelName, "fallback", r.fallback.modelName, "first_response_timeout", slow, "request_id", request.GetRequestID(ctx))
	return r.fallback.streamModel(turn, prompt, history, onChunk)
}
