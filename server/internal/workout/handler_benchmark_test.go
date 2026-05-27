package workout

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
)

func BenchmarkWorkoutHandler_ListWorkouts(b *testing.B) {
	userID := "test-user-id"
	mockRepo := &MockWorkoutRepository{}
	mockRepo.On("ListWorkouts", mock.Anything, userID).Return([]db.Workout{
		{ID: 1, Date: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
	}, nil)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()
	service := &WorkoutService{
		repo:   mockRepo,
		logger: logger,
	}
	handler := NewHandler(logger, validator, service)

	ctx := context.WithValue(context.Background(), user.UserIDKey, userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ListWorkouts(w, req)
	}
}
