package workout

import (
	"context"
	"encoding/json"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
