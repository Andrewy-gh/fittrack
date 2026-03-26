package aichat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type sseWriter struct {
	w           http.ResponseWriter
	controller  *http.ResponseController
	nextEventID int
}

func newSSEWriter(w http.ResponseWriter) *sseWriter {
	return &sseWriter{
		w:          w,
		controller: http.NewResponseController(w),
	}
}

func (s *sseWriter) prepareHeaders() {
	headers := s.w.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	headers.Set("X-Accel-Buffering", "no")
}

func (s *sseWriter) disableWriteTimeout() error {
	return s.controller.SetWriteDeadline(time.Time{})
}

func (s *sseWriter) write(event string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	s.nextEventID++
	if _, err := fmt.Fprintf(s.w, "id: %d\n", s.nextEventID); err != nil {
		return err
	}
	if event != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", event); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", body); err != nil {
		return err
	}

	return s.controller.Flush()
}
