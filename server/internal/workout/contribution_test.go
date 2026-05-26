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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test for GetContributionData endpoint
func TestWorkoutHandler_GetContributionData(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		setupMock     func(*MockWorkoutRepository)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedDays  int
		expectedError string
	}{
		{
			name: "successful fetch with data",
			setupMock: func(m *MockWorkoutRepository) {
				now := time.Now()
				yesterday := now.AddDate(0, 0, -1)
				focusValue := "Strength"

				m.On("GetContributionData", mock.Anything, userID).Return([]db.GetContributionDataRow{
					{
						Date:     pgtype.Date{Time: yesterday, Valid: true},
						Count:    5,
						Workouts: []byte(`[{"id": 1, "time": "` + yesterday.Format(time.RFC3339) + `", "focus": "` + focusValue + `"}]`),
					},
					{
						Date:     pgtype.Date{Time: now, Valid: true},
						Count:    10,
						Workouts: []byte(`[{"id": 2, "time": "` + now.Format(time.RFC3339) + `", "focus": null}, {"id": 3, "time": "` + now.Format(time.RFC3339) + `", "focus": "Cardio"}]`),
					},
				}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
			expectedDays: 2,
		},
		{
			name: "successful fetch with empty result",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("GetContributionData", mock.Anything, userID).Return([]db.GetContributionDataRow{}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
			expectedDays: 0,
		},
		{
			name: "service error",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("GetContributionData", mock.Anything, userID).Return([]db.GetContributionDataRow{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get contribution data",
		},
		{
			name:          "unauthenticated user",
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectedError: "not authorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockWorkoutRepository)
			tt.setupMock(mockRepo)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			validator := validator.New()
			service := &WorkoutService{
				repo:   mockRepo,
				logger: logger,
			}
			handler := NewHandler(logger, validator, service)

			req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(tt.ctx)
			w := httptest.NewRecorder()

			// Execute
			handler.GetContributionData(w, req)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var result ContributionDataResponse
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Len(t, result.Days, tt.expectedDays)

				// Verify workout data structure if we have days
				if tt.expectedDays > 0 && tt.name == "successful fetch with data" {
					// First day should have 1 workout with focus
					assert.Len(t, result.Days[0].Workouts, 1)
					assert.Equal(t, int32(1), result.Days[0].Workouts[0].ID)
					assert.NotEmpty(t, result.Days[0].Workouts[0].Time)
					assert.NotNil(t, result.Days[0].Workouts[0].Focus)
					assert.Equal(t, "Strength", *result.Days[0].Workouts[0].Focus)

					// Second day should have 2 workouts, one with focus and one without
					assert.Len(t, result.Days[1].Workouts, 2)
					assert.Equal(t, int32(2), result.Days[1].Workouts[0].ID)
					assert.Nil(t, result.Days[1].Workouts[0].Focus)
					assert.Equal(t, int32(3), result.Days[1].Workouts[1].ID)
					assert.NotNil(t, result.Days[1].Workouts[1].Focus)
					assert.Equal(t, "Cardio", *result.Days[1].Workouts[1].Focus)
				}
			}

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test for level calculation in the service layer
func TestWorkoutService_CalculateLevelThresholds(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &WorkoutService{logger: logger}

	tests := []struct {
		name       string
		counts     []int
		wantStatic bool // true if we expect static thresholds
	}{
		{
			name:       "empty counts uses static thresholds",
			counts:     []int{},
			wantStatic: true,
		},
		{
			name:       "fewer than 30 workout days uses static thresholds",
			counts:     []int{5, 10, 3, 8, 2},
			wantStatic: true,
		},
		{
			name:       "10 workout days (below threshold) uses static thresholds",
			counts:     []int{5, 10, 3, 8, 2, 7, 4, 6, 9, 11},
			wantStatic: true,
		},
		{
			name:       "exactly 29 workout days uses static thresholds",
			counts:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29},
			wantStatic: true,
		},
		{
			name:       "exactly 30 workout days uses dynamic percentile thresholds",
			counts:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30},
			wantStatic: false,
		},
		{
			name:       "30+ workout days uses dynamic percentile thresholds",
			counts:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35},
			wantStatic: false,
		},
		{
			name:       "only zeros uses static thresholds",
			counts:     []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantStatic: true,
		},
		{
			name:       "mixed zeros and values - 30+ non-zero uses dynamic",
			counts:     []int{0, 1, 0, 2, 3, 0, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			wantStatic: false,
		},
		{
			name:       "exactly 31 workout days uses dynamic percentile thresholds",
			counts:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
			wantStatic: false,
		},
		{
			name:       "all identical counts - dynamic thresholds should handle gracefully",
			counts:     []int{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
			wantStatic: false,
		},
		{
			name:       "mostly identical with one outlier - thresholds remain valid",
			counts:     []int{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 100},
			wantStatic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thresholds := service.calculateLevelThresholds(tt.counts)

			// Static thresholds are [1, 6, 11, 16]
			staticThresholds := [4]int{1, 6, 11, 16}

			if tt.wantStatic {
				assert.Equal(t, staticThresholds, thresholds, "Expected static thresholds")
			} else {
				// Dynamic thresholds should be strictly increasing
				assert.GreaterOrEqual(t, thresholds[0], 1, "First threshold should be at least 1")
				assert.Greater(t, thresholds[1], thresholds[0], "Thresholds should be strictly increasing")
				assert.Greater(t, thresholds[2], thresholds[1], "Thresholds should be strictly increasing")
				assert.Greater(t, thresholds[3], thresholds[2], "Thresholds should be strictly increasing")
			}
		})
	}
}

// Test for level calculation based on count and thresholds
func TestWorkoutService_CalculateLevel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &WorkoutService{logger: logger}

	// Static thresholds: [1, 6, 11, 16]
	// Level 0: count == 0
	// Level 1: count >= 1 and count < 6
	// Level 2: count >= 6 and count < 11
	// Level 3: count >= 11 and count < 16
	// Level 4: count >= 16
	thresholds := [4]int{1, 6, 11, 16}

	tests := []struct {
		name          string
		count         int
		expectedLevel int
	}{
		{"zero count gives level 0", 0, 0},
		{"count 1 gives level 1", 1, 1},
		{"count 5 gives level 1", 5, 1},
		{"count 6 gives level 2", 6, 2},
		{"count 10 gives level 2", 10, 2},
		{"count 11 gives level 3", 11, 3},
		{"count 15 gives level 3", 15, 3},
		{"count 16 gives level 4", 16, 4},
		{"count 100 gives level 4", 100, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := service.calculateLevel(tt.count, thresholds)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

// Test for parseWorkouts
func TestWorkoutService_ParseWorkouts(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &WorkoutService{logger: logger}

	tests := []struct {
		name     string
		input    []byte
		expected []WorkoutSummary
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []WorkoutSummary{},
		},
		{
			name:     "empty byte slice",
			input:    []byte{},
			expected: []WorkoutSummary{},
		},
		{
			name:  "single workout with focus",
			input: []byte(`[{"id": 1, "time": "2025-01-15T10:00:00Z", "focus": "Strength", "volume": 2250}]`),
			expected: []WorkoutSummary{
				{ID: 1, Time: "2025-01-15T10:00:00Z", Focus: stringPtr("Strength"), Volume: 2250},
			},
		},
		{
			name:  "multiple workouts",
			input: []byte(`[{"id": 1, "time": "2025-01-15T10:00:00Z", "focus": "Strength", "volume": 2250}, {"id": 2, "time": "2025-01-15T14:00:00Z", "focus": null, "volume": 1800}]`),
			expected: []WorkoutSummary{
				{ID: 1, Time: "2025-01-15T10:00:00Z", Focus: stringPtr("Strength"), Volume: 2250},
				{ID: 2, Time: "2025-01-15T14:00:00Z", Focus: nil, Volume: 1800},
			},
		},
		{
			name:     "invalid JSON",
			input:    []byte(`invalid json`),
			expected: []WorkoutSummary{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.parseWorkouts(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test for convertContributionRows
func TestWorkoutService_ConvertContributionRows(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &WorkoutService{logger: logger}

	t.Run("empty rows returns empty slice", func(t *testing.T) {
		result := service.convertContributionRows([]db.GetContributionDataRow{})
		assert.Equal(t, []ContributionDay{}, result)
	})

	t.Run("converts rows with correct level calculation", func(t *testing.T) {
		testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		rows := []db.GetContributionDataRow{
			{
				Date:     pgtype.Date{Time: testDate, Valid: true},
				Count:    5,
				Workouts: []byte(`[{"id": 1, "time": "2025-01-15T10:00:00Z", "focus": "Strength", "volume": 2250}]`),
			},
		}

		result := service.convertContributionRows(rows)

		assert.Len(t, result, 1)
		assert.Equal(t, "2025-01-15", result[0].Date)
		assert.Equal(t, 5, result[0].Count)
		assert.Equal(t, 1, result[0].Level) // Static threshold: 5 < 6, so level 1
		assert.Len(t, result[0].Workouts, 1)
		assert.Equal(t, int32(1), result[0].Workouts[0].ID)
		assert.Equal(t, 2250.0, result[0].Workouts[0].Volume)
	})

	t.Run("handles multiple workouts", func(t *testing.T) {
		testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		rows := []db.GetContributionDataRow{
			{
				Date:     pgtype.Date{Time: testDate, Valid: true},
				Count:    20,
				Workouts: []byte(`[{"id": 1, "time": "2025-01-15T10:00:00Z", "focus": "Strength", "volume": 2250}, {"id": 2, "time": "2025-01-15T14:00:00Z", "focus": null, "volume": 1800}, {"id": 3, "time": "2025-01-15T18:00:00Z", "focus": "Cardio", "volume": 950}]`),
			},
		}

		result := service.convertContributionRows(rows)

		assert.Len(t, result, 1)
		assert.Equal(t, 4, result[0].Level) // 20 >= 16, so level 4
		assert.Len(t, result[0].Workouts, 3)
		assert.Equal(t, int32(1), result[0].Workouts[0].ID)
		assert.Equal(t, int32(2), result[0].Workouts[1].ID)
		assert.Equal(t, int32(3), result[0].Workouts[2].ID)
		assert.Equal(t, 950.0, result[0].Workouts[2].Volume)
	})
}

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
