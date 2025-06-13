package exercise

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

// ExerciseService handles exercise business logic
type ExerciseService struct {
	logger  *slog.Logger
	queries *db.Queries
}

func NewService(logger *slog.Logger, queries *db.Queries) *ExerciseService {
	return &ExerciseService{
		logger:  logger,
		queries: queries,
	}
}

func (es *ExerciseService) ListExercises(ctx context.Context) ([]db.Exercise, error) {
	exercises, err := es.queries.ListExercises(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}
	return exercises, nil
}
