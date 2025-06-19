package workout

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkoutService handles workout business logic
type WorkoutService struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgxpool.Pool
}

func NewService(logger *slog.Logger, queries *db.Queries, conn *pgxpool.Pool) *WorkoutService {
	return &WorkoutService{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}
}

// func ReformatWorkoutRequest(requestBody []byte) (*ReformattedRequest, error) {
// 	var incoming IncomingRequest
// 	if err := json.Unmarshal(requestBody, &incoming); err != nil {
// 		return nil, err
// 	}

// 	// Create the workout data
// 	workout := WorkoutData{
// 		Date:  incoming.Date,
// 		Notes: incoming.Notes, // Will be nil if not provided
// 	}

// 	// Collect unique exercises and all sets
// 	exerciseMap := make(map[string]bool)
// 	var exercises []ExerciseData
// 	var sets []SetData

// 	for _, exercise := range incoming.Exercises {
// 		// Add unique exercises
// 		if !exerciseMap[exercise.Name] {
// 			exerciseMap[exercise.Name] = true
// 			exercises = append(exercises, ExerciseData{
// 				Name: exercise.Name,
// 			})
// 		}

// 		// Add all sets for this exercise
// 		for _, set := range exercise.Sets {
// 			sets = append(sets, SetData{
// 				ExerciseName: exercise.Name,
// 				Weight:       set.Weight,
// 				Reps:         set.Reps,
// 				SetType:      set.SetType,
// 			})
// 		}
// 	}

// 	return &ReformattedRequest{
// 		Workout:   workout,
// 		Exercises: exercises,
// 		Sets:      sets,
// 	}, nil
// }

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

func (ws *WorkoutService) CreateWorkout(ctx context.Context, requestBody []byte) error {
	return nil
}
