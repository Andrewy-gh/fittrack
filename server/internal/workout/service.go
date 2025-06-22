package workout

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

// WorkoutService handles workout business logic
type WorkoutService struct {
	logger *slog.Logger
	repo   WorkoutRepository
}

func NewService(logger *slog.Logger, repo WorkoutRepository) *WorkoutService {
	return &WorkoutService{
		logger: logger,
		repo:   repo,
	}
}

// Update other methods to use the repository similarly
func (ws *WorkoutService) ListWorkouts(ctx context.Context) ([]db.Workout, error) {
	return ws.repo.ListWorkouts(ctx)
}

func (ws *WorkoutService) GetWorkoutWithSets(ctx context.Context, id int32) ([]db.GetWorkoutWithSetsRow, error) {
	return ws.repo.GetWorkoutWithSets(ctx, id)
}

func (ws *WorkoutService) CreateWorkout(ctx context.Context, requestBody CreateWorkoutRequest) error {
	// Transform the request to our internal format
	reformatted, err := transformRequest(requestBody)
	if err != nil {
		return fmt.Errorf("failed to transform request: %w", err)
	}

	// Convert to PG types
	pgData, err := convertToPGTypes(reformatted)
	if err != nil {
		return fmt.Errorf("failed to convert to PG types: %w", err)
	}

	// Use repository to save the workout
	if err := ws.repo.SaveWorkout(ctx, pgData); err != nil {
		return fmt.Errorf("failed to save workout: %w", err)
	}

	return nil
}

func transformRequest(request CreateWorkoutRequest) (*ReformattedRequest, error) {
	// Parse date
	parsedDate, err := time.Parse("2006-01-02T15:04:05Z07:00", request.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Create workout data
	workout := WorkoutData{
		Date:  parsedDate,
		Notes: request.Notes,
	}

	// Process exercises and sets (your existing logic)
	exerciseMap := make(map[string]bool)
	var exercises []ExerciseData
	var sets []SetData

	for _, exercise := range request.Exercises {
		if !exerciseMap[exercise.Name] {
			exerciseMap[exercise.Name] = true
			exercises = append(exercises, ExerciseData{
				Name: exercise.Name,
			})
		}

		for _, set := range exercise.Sets {
			sets = append(sets, SetData{
				ExerciseName: exercise.Name,
				Weight:       set.Weight,
				Reps:         set.Reps,
				SetType:      set.SetType,
			})
		}
	}

	return &ReformattedRequest{
		Workout:   workout,
		Exercises: exercises,
		Sets:      sets,
	}, nil
}

func convertToPGTypes(reformatted *ReformattedRequest) (*PGReformattedRequest, error) {
	// Convert workout
	pgWorkout := PGWorkoutData{
		Date: pgtype.Timestamptz{
			Time:  reformatted.Workout.Date,
			Valid: true,
		},
		Notes: pgtype.Text{
			String: "",
			Valid:  false,
		},
	}

	if reformatted.Workout.Notes != nil {
		pgWorkout.Notes = pgtype.Text{
			String: *reformatted.Workout.Notes,
			Valid:  true,
		}
	}

	// Convert exercises
	var pgExercises []PGExerciseData
	for _, exercise := range reformatted.Exercises {
		pgExercises = append(pgExercises, PGExerciseData(exercise))
	}

	// Convert sets
	var pgSets []PGSetData
	for _, set := range reformatted.Sets {
		pgSet := PGSetData{
			ExerciseName: set.ExerciseName,
			Weight: pgtype.Int4{
				Int32: 0,
				Valid: false,
			},
			Reps:    int32(set.Reps),
			SetType: set.SetType,
		}

		if set.Weight != nil {
			pgSet.Weight = pgtype.Int4{
				Int32: int32(*set.Weight),
				Valid: true,
			}
		}

		pgSets = append(pgSets, pgSet)
	}

	return &PGReformattedRequest{
		Workout:   pgWorkout,
		Exercises: pgExercises,
		Sets:      pgSets,
	}, nil
}
