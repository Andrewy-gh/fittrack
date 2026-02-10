package exercise

import (
	"context"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExerciseRepositoryForTest implements the ExerciseRepository interface for testing
type MockExerciseRepositoryForTest struct {
	mock.Mock
}

func (m *MockExerciseRepositoryForTest) ListExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Exercise), args.Error(1)
}

func (m *MockExerciseRepositoryForTest) GetExercise(ctx context.Context, id int32, userID string) (db.Exercise, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(db.Exercise), args.Error(1)
}

func (m *MockExerciseRepositoryForTest) GetOrCreateExercise(ctx context.Context, name string, userID string) (db.Exercise, error) {
	args := m.Called(ctx, name, userID)
	return args.Get(0).(db.Exercise), args.Error(1)
}

func (m *MockExerciseRepositoryForTest) GetExerciseWithSets(ctx context.Context, id int32, userID string) ([]db.GetExerciseWithSetsRow, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]db.GetExerciseWithSetsRow), args.Error(1)
}

func (m *MockExerciseRepositoryForTest) GetOrCreateExerciseTx(ctx context.Context, qtx *db.Queries, name, userID string) (db.Exercise, error) {
	args := m.Called(ctx, qtx, name, userID)
	return args.Get(0).(db.Exercise), args.Error(1)
}

func (m *MockExerciseRepositoryForTest) GetRecentSetsForExercise(ctx context.Context, id int32, userID string) ([]db.GetRecentSetsForExerciseRow, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]db.GetRecentSetsForExerciseRow), args.Error(1)
}

func (m *MockExerciseRepositoryForTest) GetExerciseMetricsHistory(ctx context.Context, req GetExerciseMetricsHistoryRequest, userID string) ([]ExerciseMetricsHistoryPoint, MetricsHistoryBucket, error) {
	args := m.Called(ctx, req, userID)
	return args.Get(0).([]ExerciseMetricsHistoryPoint), args.Get(1).(MetricsHistoryBucket), args.Error(2)
}

func TestExerciseRepository_GetRecentSetsForExercise(t *testing.T) {
	mockRepo := &MockExerciseRepositoryForTest{}

	exerciseID := int32(1)
	userID := "user-123"

	expectedSets := []db.GetRecentSetsForExerciseRow{
		{SetID: 1, Reps: 10},
		{SetID: 2, Reps: 8},
		{SetID: 3, Reps: 12},
	}

	mockRepo.On("GetRecentSetsForExercise", mock.Anything, exerciseID, userID).Return(expectedSets, nil)

	sets, err := mockRepo.GetRecentSetsForExercise(context.Background(), exerciseID, userID)

	assert.NoError(t, err)
	assert.Len(t, sets, 3)
	assert.Equal(t, expectedSets, sets)

	mockRepo.AssertExpectations(t)
}
