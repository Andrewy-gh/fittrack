package workout

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

// WorkoutService handles workout business logic
type WorkoutService struct {
	logger  *slog.Logger
	queries *db.Queries
}

func NewService(logger *slog.Logger, queries *db.Queries) *WorkoutService {
	return &WorkoutService{
		logger:  logger,
		queries: queries,
	}
}

func (ws *WorkoutService) ListWorkouts(ctx context.Context) ([]db.Workout, error) {
	workouts, err := ws.queries.ListWorkouts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}
	return workouts, nil
}
