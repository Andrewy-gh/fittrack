package exercise

import (
	"context"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/mock"
)



func TestExerciseService_GetRecentSetsForExercise(t *testing.T) {
	mockRepo := new(MockExerciseRepository)
	service := NewService(nil, mockRepo)

	exerciseID := int32(1)
	userID := "user-123"

	expectedSets := []db.GetRecentSetsForExerciseRow{
		{SetID: 1, Reps: 10},
		{SetID: 2, Reps: 8},
		{SetID: 3, Reps: 12},
	}

	mockRepo.On("GetRecentSetsForExercise", mock.Anything, exerciseID, userID).Return(expectedSets, nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, user.UserIDKey, userID)

	sets, err := service.GetRecentSetsForExercise(ctx, exerciseID)
	if err != nil {
		t.Errorf("error was not expected while getting recent sets: %s", err)
	}

	if len(sets) != 3 {
		t.Errorf("expected 3 sets, but got %d", len(sets))
	}

	mockRepo.AssertExpectations(t)
}

