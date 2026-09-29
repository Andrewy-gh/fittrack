package exercise

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExerciseService_GetExerciseWithSets_LookupErrorClassification(t *testing.T) {
	userID := "user-123"
	exerciseID := int32(1)

	tests := []struct {
		name         string
		lookupError  error
		wantNotFound bool
	}{
		{
			name:         "no rows returns not found",
			lookupError:  pgx.ErrNoRows,
			wantNotFound: true,
		},
		{
			name:        "database failure is wrapped",
			lookupError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockExerciseRepository)
			mockRepo.On("GetExerciseDetail", mock.Anything, exerciseID, userID).Return(db.GetExerciseDetailRow{}, tt.lookupError)

			service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockRepo)
			ctx := context.WithValue(context.Background(), user.UserIDKey, userID)

			_, err := service.GetExerciseWithSets(ctx, exerciseID)
			if err == nil {
				t.Fatal("expected error but got none")
			}

			var notFoundErr *apperrors.NotFound
			if tt.wantNotFound {
				if !errors.As(err, &notFoundErr) {
					t.Errorf("expected NotFound error but got %T", err)
				}
			} else {
				if errors.As(err, &notFoundErr) {
					t.Errorf("expected non-NotFound error but got %T", err)
				}
				if !errors.Is(err, tt.lookupError) {
					t.Errorf("expected wrapped lookup error but got %v", err)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
