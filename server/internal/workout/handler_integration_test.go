package workout

import (
	"context"
	"encoding/json"
	"fmt"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestWorkoutHandler_Integration_ListWorkouts_RLS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup real database connection
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Initialize components with real database
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()
	queries := db.New(pool)

	// Initialize repositories
	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	workoutRepo := NewRepository(logger, queries, pool, exerciseRepo)
	workoutService := NewService(logger, workoutRepo)
	handler := NewHandler(logger, validator, workoutService)

	// Test data
	userAID := "test-user-a"
	userBID := "test-user-b"

	// Create test data with proper RLS context
	workoutAID := setupTestWorkout(t, pool, userAID, "User A's workout")
	workoutBID := setupTestWorkout(t, pool, userBID, "User B's workout")

	t.Run("Scenario1_UserA_CanRetrieveOwnWorkout", func(t *testing.T) {
		// Set RLS context for User A
		ctx := setTestUserContext(context.Background(), t, pool, userAID)
		ctx = user.WithContext(ctx, userAID)

		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListWorkouts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var workouts []db.Workout
		err := json.Unmarshal(w.Body.Bytes(), &workouts)
		assert.NoError(t, err)

		// User A should see only their own workout
		assert.Len(t, workouts, 1)
		assert.Equal(t, workoutAID, workouts[0].ID)
		assert.Equal(t, userAID, workouts[0].UserID)
	})

	t.Run("Scenario2_UserB_CannotRetrieveUserAWorkout", func(t *testing.T) {
		// Set RLS context for User B
		ctx := setTestUserContext(context.Background(), t, pool, userBID)
		ctx = user.WithContext(ctx, userBID)

		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListWorkouts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var workouts []db.Workout
		err := json.Unmarshal(w.Body.Bytes(), &workouts)
		assert.NoError(t, err)

		// User B should see only their own workout, not User A's
		assert.Len(t, workouts, 1)
		assert.Equal(t, workoutBID, workouts[0].ID)
		assert.Equal(t, userBID, workouts[0].UserID)
		assert.NotEqual(t, workoutAID, workouts[0].ID, "User B should not see User A's workout")
	})

	t.Run("Scenario3_AnonymousUser_CannotAccessWorkouts", func(t *testing.T) {
		// No user context set (anonymous user)
		ctx := context.Background()

		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListWorkouts(w, req)

		// Should get unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "not authorized")
	})

	t.Run("Scenario4_GetSpecificWorkout_UserB_CannotAccessUserAWorkout", func(t *testing.T) {
		// User B tries to access User A's specific workout
		ctx := setTestUserContext(context.Background(), t, pool, userBID)
		ctx = user.WithContext(ctx, userBID)

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/workouts/%d", workoutAID), nil).WithContext(ctx)
		req.SetPathValue("id", fmt.Sprintf("%d", workoutAID))
		w := httptest.NewRecorder()

		handler.GetWorkoutWithSets(w, req)

		// Should succeed but return empty results due to RLS
		assert.Equal(t, http.StatusOK, w.Code)
		var result []db.GetWorkoutWithSetsRow
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Empty(t, result, "User B should not see User A's workout data")
	})

	t.Run("Scenario5_ConcurrentRequests_ProperIsolation", func(t *testing.T) {
		// Test concurrent requests from different users
		var wg sync.WaitGroup
		results := make(map[string][]db.Workout)
		mu := sync.Mutex{}

		// User A request
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := setTestUserContext(context.Background(), t, pool, userAID)
			ctx = user.WithContext(ctx, userAID)

			req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ListWorkouts(w, req)

			if w.Code == http.StatusOK {
				var workouts []db.Workout
				json.Unmarshal(w.Body.Bytes(), &workouts)
				mu.Lock()
				results[userAID] = workouts
				mu.Unlock()
			}
		}()

		// User B request
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := setTestUserContext(context.Background(), t, pool, userBID)
			ctx = user.WithContext(ctx, userBID)

			req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ListWorkouts(w, req)

			if w.Code == http.StatusOK {
				var workouts []db.Workout
				json.Unmarshal(w.Body.Bytes(), &workouts)
				mu.Lock()
				results[userBID] = workouts
				mu.Unlock()
			}
		}()

		wg.Wait()

		// Verify isolation - each user should only see their own workouts
		assert.Len(t, results[userAID], 1)
		assert.Equal(t, workoutAID, results[userAID][0].ID)
		assert.Equal(t, userAID, results[userAID][0].UserID)

		assert.Len(t, results[userBID], 1)
		assert.Equal(t, workoutBID, results[userBID][0].ID)
		assert.Equal(t, userBID, results[userBID][0].UserID)
	})
}
