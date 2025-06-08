package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/db"
	"github.com/Andrewy-gh/fittrack/server/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

// WorkoutService handles workout business logic
type WorkoutService struct {
	queries *db.Queries
}

// NewWorkoutService creates a new workout service
func NewWorkoutService(queries *db.Queries) *WorkoutService {
	return &WorkoutService{
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

// CreateWorkout creates a new workout with exercises and sets
func (ws *WorkoutService) CreateWorkout(ctx context.Context, req models.WorkoutRequest) (*models.WorkoutResponse, error) {
	// Parse the date
	parsedDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Convert to pgtype.Timestamptz
	pgDate := pgtype.Timestamptz{
		Time:  parsedDate,
		Valid: true,
	}

	// Convert notes to pgtype.Text
	var pgNotes pgtype.Text
	if req.Notes != nil {
		pgNotes = pgtype.Text{
			String: *req.Notes,
			Valid:  true,
		}
	}

	// Create the workout
	workout, err := ws.queries.CreateWorkout(ctx, db.CreateWorkoutParams{
		Date:  pgDate,
		Notes: pgNotes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create workout: %w", err)
	}

	// Process exercises and sets
	var exerciseResponses []models.ExerciseResponse

	for _, exerciseInput := range req.Exercises {
		// Get or create the exercise
		exercise, err := ws.queries.GetOrCreateExercise(ctx, exerciseInput.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create exercise %s: %w", exerciseInput.Name, err)
		}

		// Create sets for this exercise
		var setResponses []models.SetResponse
		for _, setInput := range exerciseInput.Sets {
			// Convert weight to pgtype.Int4
			var pgWeight pgtype.Int4
			if setInput.Weight != nil {
				pgWeight = pgtype.Int4{
					Int32: int32(*setInput.Weight),
					Valid: true,
				}
			}

			// Create the set
			set, err := ws.queries.CreateSet(ctx, db.CreateSetParams{
				ExerciseID: exercise.ID,
				WorkoutID:  workout.ID,
				Weight:     pgWeight,
				Reps:       int32(setInput.Reps),
				SetType:    setInput.SetType,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create set: %w", err)
			}

			// Convert pgtype back to regular types for response
			var weight *int
			if set.Weight.Valid {
				w := int(set.Weight.Int32)
				weight = &w
			}

			setResponses = append(setResponses, models.SetResponse{
				SetID:   set.ID,
				Weight:  weight,
				Reps:    int(set.Reps),
				SetType: set.SetType,
			})
		}

		exerciseResponses = append(exerciseResponses, models.ExerciseResponse{
			ExerciseID: exercise.ID,
			Name:       exercise.Name,
			Sets:       setResponses,
		})
	}

	// Convert notes back for response
	var notes *string
	if workout.Notes.Valid {
		notes = &workout.Notes.String
	}

	return &models.WorkoutResponse{
		WorkoutID: workout.ID,
		Date:      workout.Date.Time,
		Notes:     notes,
		Exercises: exerciseResponses,
	}, nil
}
