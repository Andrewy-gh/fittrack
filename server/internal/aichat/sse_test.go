package aichat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type deadlineRecorder struct {
	*httptest.ResponseRecorder
	writeDeadline time.Time
}

func (d *deadlineRecorder) SetWriteDeadline(deadline time.Time) error {
	d.writeDeadline = deadline
	return nil
}

func TestSSEWriterDisableWriteTimeout(t *testing.T) {
	recorder := &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	writer := newSSEWriter(recorder)

	if err := writer.disableWriteTimeout(); err != nil {
		t.Fatalf("disableWriteTimeout() error = %v", err)
	}

	if !recorder.writeDeadline.IsZero() {
		t.Fatalf("expected zero write deadline, got %v", recorder.writeDeadline)
	}
}

func TestSSEWriterWrite_EmitsMonotonicEventIDs(t *testing.T) {
	recorder := &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	writer := newSSEWriter(recorder)

	if err := writer.write("start", StreamEvent{Type: "start"}); err != nil {
		t.Fatalf("first write() error = %v", err)
	}
	if err := writer.write("done", StreamEvent{Type: "done"}); err != nil {
		t.Fatalf("second write() error = %v", err)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "id: 1\n") {
		t.Fatalf("expected first event id in body, got %q", body)
	}
	if !strings.Contains(body, "id: 2\n") {
		t.Fatalf("expected second event id in body, got %q", body)
	}

	var event StreamEvent
	payload := strings.Split(strings.Split(body, "data: ")[1], "\n\n")[0]
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if event.Type != "start" {
		t.Fatalf("first payload type = %q, want %q", event.Type, "start")
	}
}

var _ http.ResponseWriter = (*deadlineRecorder)(nil)
