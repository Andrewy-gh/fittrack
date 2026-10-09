package workout

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
)

// The HTTP boundary owns the cross-language contract: Unicode code points,
// rather than bytes, UTF-16 units, or user-perceived graphemes.
func TestWorkoutHandler_UnicodeTextBoundaries(t *testing.T) {
	for _, alphabet := range []struct{ name, boundary string }{
		{"ASCII", strings.Repeat("a", 256)},
		{"emoji", strings.Repeat("🏋", 256)},
		{"combining", strings.Repeat("e\u0301", 128)},
	} {
		for _, field := range []string{"notes", "workoutFocus", "name"} {
			for _, method := range []string{http.MethodPost, http.MethodPut} {
				for _, over := range []bool{false, true} {
					text, suffix := alphabet.boundary, "256"
					if over {
						text += "x"
						suffix = "257"
					}
					t.Run(alphabet.name+"/"+field+"/"+method+"/"+suffix, func(t *testing.T) {
						exercise := map[string]any{"name": "Squat", "sets": []map[string]any{{"reps": 10, "setType": "working"}}}
						payload := map[string]any{"date": "2026-10-08T12:00:00Z", "exercises": []map[string]any{exercise}}
						if field == "name" {
							exercise[field] = text
						} else {
							payload[field] = text
						}
						body, err := json.Marshal(payload)
						if err != nil {
							t.Fatal(err)
						}
						repo := &MockWorkoutRepository{}
						repo.Test(t)
						if !over {
							if method == http.MethodPost {
								repo.On("SaveWorkoutWithID", mock.Anything, mock.Anything, "mobile-audit").Return(int32(1), nil)
							} else {
								repo.On("GetWorkout", mock.Anything, int32(1), "mobile-audit").Return(db.Workout{ID: 1}, nil)
								repo.On("UpdateWorkout", mock.Anything, int32(1), mock.Anything, "mobile-audit").Return(nil)
							}
						}
						logger := slog.New(slog.NewTextHandler(io.Discard, nil))
						handler := NewHandler(logger, validator.New(), &WorkoutService{repo: repo, logger: logger})
						req := httptest.NewRequest(method, "/api/workouts/1", bytes.NewReader(body))
						req = req.WithContext(context.WithValue(req.Context(), user.UserIDKey, "mobile-audit"))
						req.SetPathValue("id", "1")
						result := httptest.NewRecorder()
						if method == http.MethodPost {
							handler.CreateWorkout(result, req)
						} else {
							handler.UpdateWorkout(result, req)
						}
						want := http.StatusOK
						if method == http.MethodPut {
							want = http.StatusNoContent
						}
						if over {
							want = http.StatusBadRequest
						}
						if result.Code != want {
							t.Fatalf("status = %d, want %d: %s", result.Code, want, result.Body.String())
						}
						if over && !strings.Contains(result.Body.String(), "validation error occurred") {
							t.Fatalf("wrong rejection: %s", result.Body.String())
						}
					})
				}
			}
		}
	}
}
