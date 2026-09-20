package exercise

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestCatalogLinkOwnershipAndHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewRepository(logger, db.New(pool), pool)
	service := NewService(logger, repo)
	handler := NewHandler(logger, validator.New(), service)
	ctx := user.WithContext(context.Background(), exerciseTestUserA)
	id := setupTestExercise(t, pool, exerciseTestUserA, "My custom press")
	otherID := setupTestExercise(t, pool, exerciseTestUserB, "My custom press")
	workoutID := insertMetricsHistoryWorkout(t, ctx, pool, exerciseTestUserA, time.Now().UTC())
	insertMetricsHistorySet(t, ctx, pool, exerciseTestUserA, id, workoutID, 100, 5, 500)
	rm := 125.0
	require.NoError(t, repo.SetExerciseHistorical1RM(ctx, id, &rm, &workoutID, exerciseTestUserA))
	before, err := service.GetExerciseWithSets(ctx, id)
	require.NoError(t, err)
	require.Nil(t, before.Exercise.Catalog)
	mutate := func(who, body string, target int32, status int) {
		t.Helper()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/exercises/%d/catalog", target), strings.NewReader(body))
		req.SetPathValue("id", fmt.Sprint(target))
		if who != "" {
			req = req.WithContext(user.WithContext(context.Background(), who))
		}
		response := httptest.NewRecorder()
		handler.UpdateCatalog(response, req)
		require.Equal(t, status, response.Code, response.Body.String())
	}
	mutate("", `{"catalog_id":"Pushups"}`, id, 401)
	mutate(exerciseTestUserB, `{"catalog_id":"Pushups"}`, id, 404)
	mutate(exerciseTestUserA, `{"catalog_id":"Pushups"}`, otherID, 404)
	mutate(exerciseTestUserA, `{"catalog_id":"Pushups"}`, 2147483647, 404)
	for _, body := range []string{`{}`, `{"catalog_id":null}`, `{"catalog_id":"guess"}`, `{"catalog_id":"Pushups","name":"changed"}`} {
		mutate(exerciseTestUserA, body, id, 400)
	}
	for _, catalogID := range []string{"Pushups", "Dumbbell_Bench_Press", ""} {
		body, _ := json.Marshal(map[string]string{"catalog_id": catalogID})
		mutate(exerciseTestUserA, string(body), id, 204)
		after, err := service.GetExerciseWithSets(ctx, id)
		require.NoError(t, err)
		require.Equal(t, before.Sets, after.Sets)
		require.Equal(t, before.Exercise.ID, after.Exercise.ID)
		require.Equal(t, before.Exercise.Name, after.Exercise.Name)
		require.Equal(t, before.Exercise.CreatedAt, after.Exercise.CreatedAt)
		require.Equal(t, before.Exercise.Historical1RM, after.Exercise.Historical1RM)
		require.Equal(t, before.Exercise.Historical1RMUpdatedAt, after.Exercise.Historical1RMUpdatedAt)
		require.Equal(t, before.Exercise.Historical1RMSourceWorkoutID, after.Exercise.Historical1RMSourceWorkoutID)
		require.Equal(t, before.Exercise.BestE1RM, after.Exercise.BestE1RM)
		if catalogID == "" {
			require.Nil(t, after.Exercise.Catalog)
		} else {
			require.Equal(t, catalogID, after.Exercise.Catalog.ID)
			require.Equal(t, "chest", after.Exercise.Catalog.PrimaryMuscles[0])
		}
	}
	other, err := repo.GetExerciseDetail(ctx, otherID, exerciseTestUserB)
	require.NoError(t, err)
	require.False(t, other.CatalogID.Valid)
	// Database constraint also rejects invalid IDs outside the HTTP boundary.
	_, err = pool.Exec(ctx, "UPDATE exercise SET catalog_id='unknown' WHERE id=$1", id)
	require.Error(t, err)
}
