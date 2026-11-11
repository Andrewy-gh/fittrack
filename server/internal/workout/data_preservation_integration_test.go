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

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestWorkoutHandler_DataPreservation(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	service := NewService(logger, NewRepository(logger, queries, pool, exercise.NewRepository(logger, queries, pool)))
	handler := NewHandler(logger, validator.New(), service)
	ctx := user.WithContext(context.Background(), "test-user-a")
	call := func(method, path, body string, handle http.HandlerFunc) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body)).WithContext(ctx)
		req.SetPathValue("id", path[len("/api/workouts/"):])
		response := httptest.NewRecorder()
		handle(response, req)
		return response
	}
	notes, focus := "Keep these notes", "Upper"
	zero, weighted := float64(0), float64(25)
	id, err := service.CreateWorkoutWithID(ctx, CreateWorkoutRequest{
		Date: "2026-10-08T12:00:00Z", Notes: &notes, WorkoutFocus: &focus,
		Exercises: []ExerciseInput{{Name: "Preservation movement", Sets: []SetInput{
			{Reps: 10, SetType: "working"}, {Reps: 10, Weight: &zero, SetType: "working"}, {Reps: 10, Weight: &weighted, SetType: "working"},
		}}},
	})
	require.NoError(t, err)
	path := fmt.Sprintf("/api/workouts/%d", id)
	for _, step := range []struct {
		name, fields         string
		wantNotes, wantFocus *string
	}{
		{"omitted retains", "", &notes, &focus},
		{"null retains", `,"notes":null,"workoutFocus":null`, &notes, &focus},
		{"clear notes independently", `,"notes":""`, nil, &focus},
		{"clear focus", `,"workoutFocus":""`, nil, nil},
		{"omitted after clear", "", nil, nil},
	} {
		t.Run(step.name, func(t *testing.T) {
			body := `{"date":"2026-10-09T12:00:00Z","exercises":[{"name":"Preservation movement","sets":[{"reps":11,"weight":null,"setType":"working"},{"reps":11,"weight":0,"setType":"working"},{"reps":11,"weight":25,"setType":"working"}]}]` + step.fields + `}`
			updated := call(http.MethodPut, path, body, handler.UpdateWorkout)
			require.Equal(t, http.StatusNoContent, updated.Code, updated.Body.String())
			loaded := call(http.MethodGet, path, "", handler.GetWorkoutWithSets)
			require.Equal(t, http.StatusOK, loaded.Code, loaded.Body.String())
			var result WorkoutDetailResponse
			require.NoError(t, json.Unmarshal(loaded.Body.Bytes(), &result))
			require.Len(t, result.Sets, 3)
			for _, set := range result.Sets {
				require.Equal(t, step.wantNotes, set.WorkoutNotes)
				require.Equal(t, step.wantFocus, set.WorkoutFocus)
				require.EqualValues(t, 11, set.Reps)
				require.Equal(t, "2026-10-09", set.WorkoutDate.Format("2006-01-02"))
			}
			require.Nil(t, result.Sets[0].Weight)
			require.Equal(t, &zero, result.Sets[1].Weight)
			require.Equal(t, &weighted, result.Sets[2].Weight)
		})
	}
}
