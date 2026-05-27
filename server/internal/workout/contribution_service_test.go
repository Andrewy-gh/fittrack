package workout

import (
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"testing"
	"time"
)

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
