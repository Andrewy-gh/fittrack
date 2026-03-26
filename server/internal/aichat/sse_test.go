package aichat

import (
	"net/http"
	"net/http/httptest"
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

var _ http.ResponseWriter = (*deadlineRecorder)(nil)
