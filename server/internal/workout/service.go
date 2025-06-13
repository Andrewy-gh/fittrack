package workout

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
)

// WorkoutService handles workout business logic
type WorkoutService struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgx.Conn
}

func NewService(logger *slog.Logger, queries *db.Queries, conn *pgx.Conn) *WorkoutService {
	return &WorkoutService{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}
}

func (ws *WorkoutService) ListWorkouts(ctx context.Context) ([]db.Workout, error) {
	workouts, err := ws.queries.ListWorkouts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}
	return workouts, nil
}

func (ws *WorkoutService) GetWorkoutWithSets(ctx context.Context, id int32) ([]db.GetWorkoutWithSetsRow, error) {
	workouts, err := ws.queries.GetWorkoutWithSets(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout with sets: %w", err)
	}
	return workouts, nil
}
