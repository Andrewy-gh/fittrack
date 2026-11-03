package workout

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test for ListWorkoutFocusValues endpoint
func TestWorkoutHandler_ListWorkoutFocusValues(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		setupMock     func(*MockWorkoutRepository)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedBody  []string
		expectedError string
	}{
		{
			name: "successful fetch with values",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusValues", mock.Anything, userID).Return([]string{"Chest", "Back", "Legs"}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
			expectedBody: []string{"Chest", "Back", "Legs"},
		},
		{
			name: "successful fetch with empty result",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusValues", mock.Anything, userID).Return([]string{}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
			expectedBody: []string{},
		},
		{
			name: "service error",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusValues", mock.Anything, userID).Return([]string{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to list workout focus values",
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

			req := httptest.NewRequest("GET", "/api/workouts/focus-values", nil).WithContext(tt.ctx)
			w := httptest.NewRecorder()

			// Execute
			handler.ListWorkoutFocusValues(w, req)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var result []string
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				// Ensure we always return an empty slice, not nil
				if tt.expectedBody == nil {
					tt.expectedBody = []string{}
				}
				assert.Equal(t, tt.expectedBody, result)
			}

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutHandler_GetNewWorkoutContext(t *testing.T) {
	userID := "test-user-id"
	templateDate := time.Date(2026, 5, 1, 14, 0, 0, 0, time.UTC)
	noteDate := time.Date(2026, 5, 2, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		setupMock     func(*MockWorkoutRepository)
		ctx           context.Context
		expectedCode  int
		expectedError string
		assertBody    func(*testing.T, NewWorkoutContextResponse)
	}{
		{
			name: "successful fetch with templates and latest note",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusTemplates", mock.Anything, userID).Return([]db.ListWorkoutFocusTemplatesRow{
					{
						WorkoutID:    11,
						Date:         pgtype.Timestamptz{Time: templateDate, Valid: true},
						WorkoutFocus: "Push",
					},
				}, nil)
				m.On("GetLatestWorkoutNote", mock.Anything, userID).Return(db.GetLatestWorkoutNoteRow{
					WorkoutID: 12,
					Date:      pgtype.Timestamptz{Time: noteDate, Valid: true},
					Notes:     pgtype.Text{String: "Keep shoulders warm.", Valid: true},
				}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			assertBody: func(t *testing.T, result NewWorkoutContextResponse) {
				assert.Equal(t, []FocusTemplateResponse{{
					Focus:     "Push",
					WorkoutID: 11,
					Date:      templateDate,
				}}, result.FocusTemplates)
				if assert.NotNil(t, result.LatestWorkoutNote) {
					assert.Equal(t, int32(12), result.LatestWorkoutNote.WorkoutID)
					assert.Equal(t, "Keep shoulders warm.", result.LatestWorkoutNote.Note)
				}
			},
		},
		{
			name: "missing note is optional",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusTemplates", mock.Anything, userID).Return([]db.ListWorkoutFocusTemplatesRow{}, nil)
				m.On("GetLatestWorkoutNote", mock.Anything, userID).Return(db.GetLatestWorkoutNoteRow{}, pgx.ErrNoRows)
			},
			ctx:          user.WithContext(context.Background(), userID),
			expectedCode: http.StatusOK,
			assertBody: func(t *testing.T, result NewWorkoutContextResponse) {
				assert.Empty(t, result.FocusTemplates)
				assert.Nil(t, result.LatestWorkoutNote)
			},
		},
		{
			name: "template failure keeps priority when both reads fail",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusTemplates", mock.Anything, userID).Return([]db.ListWorkoutFocusTemplatesRow{}, errors.New("templates unavailable"))
				m.On("GetLatestWorkoutNote", mock.Anything, userID).Return(db.GetLatestWorkoutNoteRow{}, errors.New("note unavailable"))
			},
			ctx:           user.WithContext(context.Background(), userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get new workout context",
		},
		{
			name: "note failure is reported",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusTemplates", mock.Anything, userID).Return([]db.ListWorkoutFocusTemplatesRow{}, nil)
				m.On("GetLatestWorkoutNote", mock.Anything, userID).Return(db.GetLatestWorkoutNoteRow{}, errors.New("note unavailable"))
			},
			ctx:           user.WithContext(context.Background(), userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get new workout context",
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
			mockRepo := new(MockWorkoutRepository)
			tt.setupMock(mockRepo)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			service := &WorkoutService{
				repo:   mockRepo,
				logger: logger,
			}
			handler := NewHandler(logger, validator.New(), service)

			req := httptest.NewRequest("GET", "/api/workouts/new-workout-context", nil).WithContext(tt.ctx)
			w := httptest.NewRecorder()

			handler.GetNewWorkoutContext(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.assertBody != nil {
				var result NewWorkoutContextResponse
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				tt.assertBody(t, result)
			}

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// The repository boundary blocks each read until both have started. This guards
// independent I/O overlap without using elapsed-time performance assertions.
func TestNewWorkoutContextOverlapsReadsAndJoinsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(user.WithContext(context.Background(), "owner"), 2*time.Second)
	defer cancel()
	started := make(chan string, 2)
	repo := new(MockWorkoutRepository)
	repo.On("ListWorkoutFocusTemplates", mock.Anything, "owner").Run(func(args mock.Arguments) {
		started <- "templates"
		<-args.Get(0).(context.Context).Done()
	}).Return([]db.ListWorkoutFocusTemplatesRow{}, context.Canceled)
	repo.On("GetLatestWorkoutNote", mock.Anything, "owner").Run(func(args mock.Arguments) {
		started <- "note"
		<-args.Get(0).(context.Context).Done()
	}).Return(db.GetLatestWorkoutNoteRow{}, context.Canceled)
	done := make(chan error, 1)
	go func() {
		_, err := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo).GetNewWorkoutContext(ctx)
		done <- err
	}()
	for range 2 {
		select {
		case <-started:
		case <-ctx.Done():
			<-done
			t.Fatal("both reads must start before either finishes")
		}
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation = %v", err)
	}
	repo.AssertExpectations(t)
}

func TestNewWorkoutContextPreservesErrorPriority(t *testing.T) {
	repo := new(MockWorkoutRepository)
	templateErr, noteErr := errors.New("templates unavailable"), errors.New("note unavailable")
	repo.On("ListWorkoutFocusTemplates", mock.Anything, "owner").Return([]db.ListWorkoutFocusTemplatesRow{}, templateErr)
	repo.On("GetLatestWorkoutNote", mock.Anything, "owner").Return(db.GetLatestWorkoutNoteRow{}, noteErr)
	_, err := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo).GetNewWorkoutContext(user.WithContext(context.Background(), "owner"))
	if !errors.Is(err, templateErr) {
		t.Fatalf("error priority = %v", err)
	}
	repo.AssertExpectations(t)
}
