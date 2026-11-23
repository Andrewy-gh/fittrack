package workout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestWorkoutHandler_CreateTimestampReload(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	handler := NewHandler(logger, validator.New(), NewService(logger, NewRepository(logger, queries, pool, exercise.NewRepository(logger, queries, pool))))
	ctx := user.WithContext(context.Background(), "test-user-a")

	cases := []struct{ note, submitted, want string }{
		{"fraction one", "2026-10-08T12:00:00.123456Z", "2026-10-08T12:00:00.123456Z"},
		{"fraction two", "2026-10-08T12:00:00.654321Z", "2026-10-08T12:00:00.654321Z"},
		{"offset", "2026-10-08T14:30:00.987654+02:30", "2026-10-08T12:00:00.987654Z"},
		{"whole second", "2026-10-08T12:00:00Z", "2026-10-08T12:00:00Z"},
	}
	for _, tc := range cases {
		body := fmt.Sprintf(`{"date":%q,"notes":%q,"exercises":[{"name":"Timestamp movement","sets":[{"reps":1,"setType":"working"}]}]}`, tc.submitted, tc.note)
		req := httptest.NewRequest(http.MethodPost, "/api/workouts", bytes.NewBufferString(body)).WithContext(ctx)
		response := httptest.NewRecorder()
		handler.CreateWorkout(response, req)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/api/workouts", nil).WithContext(ctx)
	response := httptest.NewRecorder()
	handler.ListWorkouts(response, req)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var workouts []db.Workout
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &workouts))
	got := make(map[string]time.Time)
	for _, workout := range workouts {
		if workout.Notes.Valid && workout.Date.Valid {
			got[workout.Notes.String] = workout.Date.Time
		}
	}
	for _, tc := range cases {
		t.Run(tc.note, func(t *testing.T) {
			want, err := time.Parse(time.RFC3339Nano, tc.want)
			require.NoError(t, err)
			actual, ok := got[tc.note]
			require.True(t, ok, "created workout missing from reload")
			require.True(t, actual.Equal(want), "got %s, want %s", actual.Format(time.RFC3339Nano), tc.want)
		})
	}
}
