package workout

import (
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContributionData_Integration_RLS(t *testing.T) {
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
	userAID := "test-user-rls-a"
	userBID := "test-user-rls-b"

	// Create test workouts for both users on the same day
	todayStr := getDatabaseCurrentDateString(t, pool)
	workoutAID := setupTestWorkout(t, pool, userAID, "User A's workout")
	workoutBID := setupTestWorkout(t, pool, userBID, "User B's workout")

	// Add a set to each workout so they show up in contribution data
	setupTestSet(t, pool, userAID, workoutAID)
	setupTestSet(t, pool, userBID, workoutBID)

	t.Run("UserA_OnlySeesOwnWorkoutMetadata", func(t *testing.T) {
		// Set RLS context for User A
		ctx := setTestUserContext(context.Background(), t, pool, userAID)
		ctx = user.WithContext(ctx, userAID)

		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Find today's contribution data
		var todayData *ContributionDay
		for i := range result.Days {
			if result.Days[i].Date == todayStr {
				todayData = &result.Days[i]
				break
			}
		}

		require.NotNil(t, todayData, "Should have contribution data for today")
		assert.Len(t, todayData.Workouts, 1, "User A should see exactly 1 workout (their own)")
		assert.Equal(t, workoutAID, todayData.Workouts[0].ID, "User A should see their own workout ID")
		assert.NotEqual(t, workoutBID, todayData.Workouts[0].ID, "User A should NOT see User B's workout ID")
	})

	t.Run("UserB_OnlySeesOwnWorkoutMetadata", func(t *testing.T) {
		// Set RLS context for User B
		ctx := setTestUserContext(context.Background(), t, pool, userBID)
		ctx = user.WithContext(ctx, userBID)

		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Find today's contribution data
		var todayData *ContributionDay
		for i := range result.Days {
			if result.Days[i].Date == todayStr {
				todayData = &result.Days[i]
				break
			}
		}

		require.NotNil(t, todayData, "Should have contribution data for today")
		assert.Len(t, todayData.Workouts, 1, "User B should see exactly 1 workout (their own)")
		assert.Equal(t, workoutBID, todayData.Workouts[0].ID, "User B should see their own workout ID")
		assert.NotEqual(t, workoutAID, todayData.Workouts[0].ID, "User B should NOT see User A's workout ID")
	})

	t.Run("AnonymousUser_Unauthorized", func(t *testing.T) {
		// No user context set (anonymous user)
		ctx := context.Background()

		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		// Should get unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "not authorized")
	})
}

func TestContributionData_Integration(t *testing.T) {
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
	userID := "test-user-integration-contrib"

	t.Run("VerifyDateRangeFiltering", func(t *testing.T) {
		// Create workouts at different time points
		now := time.Now()
		ctx := setTestUserContext(context.Background(), t, pool, userID)

		// Workout within 52 weeks
		withinRangeDate := now.AddDate(0, 0, -200) // ~6.5 months ago
		_, err := pool.Exec(ctx,
			"INSERT INTO workout (date, notes, user_id) VALUES ($1, $2, $3)",
			withinRangeDate, "Within range", userID)
		require.NoError(t, err)

		// Workout outside 52 weeks (should not appear)
		outsideRangeDate := now.AddDate(0, 0, -400) // Over a year ago
		_, err = pool.Exec(ctx,
			"INSERT INTO workout (date, notes, user_id) VALUES ($1, $2, $3)",
			outsideRangeDate, "Outside range", userID)
		require.NoError(t, err)

		// Set user context and make request
		ctx = user.WithContext(ctx, userID)
		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Check that the within-range workout appears
		withinRangeDateStr := getDatabaseDateString(t, pool, withinRangeDate)
		var foundWithinRange bool
		for _, day := range result.Days {
			if day.Date == withinRangeDateStr {
				foundWithinRange = true
				break
			}
		}
		assert.True(t, foundWithinRange, "Workout within 52 weeks should appear in contribution data")

		// Check that the outside-range workout does NOT appear
		outsideRangeDateStr := getDatabaseDateString(t, pool, outsideRangeDate)
		var foundOutsideRange bool
		for _, day := range result.Days {
			if day.Date == outsideRangeDateStr {
				foundOutsideRange = true
				break
			}
		}
		assert.False(t, foundOutsideRange, "Workout outside 52 weeks should NOT appear in contribution data")
	})

	t.Run("VerifyMultipleWorkoutsPerDay", func(t *testing.T) {
		// Create multiple workouts on the same day
		todayStr := getDatabaseCurrentDateString(t, pool)
		ctx := setTestUserContext(context.Background(), t, pool, userID)

		// Create 3 workouts on the same day with different focuses
		// Use NOW() from database to avoid timezone issues
		var workoutIDs []int32
		focuses := []string{"Strength", "Cardio", ""}

		for i, focus := range focuses {
			var workoutID int32
			var err error
			if focus == "" {
				err = pool.QueryRow(ctx,
					"INSERT INTO workout (date, notes, user_id, workout_focus) VALUES (NOW(), $1, $2, NULL) RETURNING id",
					fmt.Sprintf("Workout %d", i+1), userID).Scan(&workoutID)
			} else {
				err = pool.QueryRow(ctx,
					"INSERT INTO workout (date, notes, user_id, workout_focus) VALUES (NOW(), $1, $2, $3) RETURNING id",
					fmt.Sprintf("Workout %d", i+1), userID, focus).Scan(&workoutID)
			}
			require.NoError(t, err)
			workoutIDs = append(workoutIDs, workoutID)

			// Add a set to each workout so they appear in contribution data
			setupTestSet(t, pool, userID, workoutID)
		}

		// Set user context and make request
		ctx = user.WithContext(ctx, userID)
		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Find today's data (use database's current date)
		var todayData *ContributionDay
		for i := range result.Days {
			if result.Days[i].Date == todayStr {
				todayData = &result.Days[i]
				break
			}
		}

		// If not found with today's date, look for the most recent day with data
		if todayData == nil && len(result.Days) > 0 {
			// Find the day with workouts that match our workout IDs
			for i := range result.Days {
				if len(result.Days[i].Workouts) >= 3 {
					todayData = &result.Days[i]
					break
				}
			}
		}

		require.NotNil(t, todayData, "Should have contribution data for workouts created")
		assert.Len(t, todayData.Workouts, 3, "Should have 3 workouts")

		// Verify workout metadata
		foundIDs := make(map[int32]bool)
		for _, workout := range todayData.Workouts {
			foundIDs[workout.ID] = true
			assert.NotEmpty(t, workout.Time, "Workout should have a time")
			assert.Greater(t, workout.Volume, 0.0, "Workout should include volume")
		}

		// Verify all workout IDs are present
		for _, wid := range workoutIDs {
			assert.True(t, foundIDs[wid], "Workout ID %d should be in contribution data", wid)
		}

		// Verify focus values
		var hasStrength, hasCardio, hasNull bool
		for _, workout := range todayData.Workouts {
			if workout.Focus != nil && *workout.Focus == "Strength" {
				hasStrength = true
			} else if workout.Focus != nil && *workout.Focus == "Cardio" {
				hasCardio = true
			} else if workout.Focus == nil {
				hasNull = true
			}
		}
		assert.True(t, hasStrength, "Should have workout with Strength focus")
		assert.True(t, hasCardio, "Should have workout with Cardio focus")
		assert.True(t, hasNull, "Should have workout with null focus")
	})

	t.Run("VerifyLevelCalculation", func(t *testing.T) {
		// Create a day with known count to verify level calculation
		now := time.Now()
		testDate := now.AddDate(0, 0, -7) // 7 days ago
		ctx := setTestUserContext(context.Background(), t, pool, userID)

		// Create a workout with exactly 15 sets (should result in count=15)
		var workoutID int32
		err := pool.QueryRow(ctx,
			"INSERT INTO workout (date, notes, user_id) VALUES ($1, $2, $3) RETURNING id",
			testDate, "Level test workout", userID).Scan(&workoutID)
		require.NoError(t, err)

		// Add 15 sets
		var exerciseID int32
		err = pool.QueryRow(ctx,
			"INSERT INTO exercise (name, user_id) VALUES ($1, $2) RETURNING id",
			"Level Test Exercise", userID).Scan(&exerciseID)
		require.NoError(t, err)

		for i := 1; i <= 15; i++ {
			_, err = pool.Exec(ctx,
				"INSERT INTO \"set\" (workout_id, exercise_id, user_id, weight, reps, set_type, exercise_order, set_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				workoutID, exerciseID, userID, 100, 10, "working", 1, i)
			require.NoError(t, err)
		}

		// Set user context and make request
		ctx = user.WithContext(ctx, userID)
		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Find the test date
		testDateStr := getDatabaseDateString(t, pool, testDate)
		var testDayData *ContributionDay
		for i := range result.Days {
			if result.Days[i].Date == testDateStr {
				testDayData = &result.Days[i]
				break
			}
		}

		require.NotNil(t, testDayData, "Should have contribution data for test date")
		assert.Equal(t, 15, testDayData.Count, "Count should be 15")
		// With static thresholds [1, 6, 11, 16], count=15 should give level 3
		// (15 >= 11 and 15 < 16)
		assert.Equal(t, 3, testDayData.Level, "Level should be 3 for count=15 with static thresholds")
	})

	t.Run("VerifyEmptyResult", func(t *testing.T) {
		// Use a different user with no workouts
		emptyUserID := "test-user-empty-contrib"
		ctx := setTestUserContext(context.Background(), t, pool, emptyUserID)
		ctx = user.WithContext(ctx, emptyUserID)

		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Should return empty days array, not nil
		assert.NotNil(t, result.Days)
		assert.Len(t, result.Days, 0, "Should have no contribution data for user with no workouts")
	})
}
