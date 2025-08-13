package workout

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Request/Response types
type CreateWorkoutRequest struct {
	Date      string          `json:"date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Notes     *string         `json:"notes,omitempty" validate:"omitempty,max=256"`
	Exercises []ExerciseInput `json:"exercises" validate:"required,min=1,dive"`
}

type ExerciseInput struct {
	Name string     `json:"name" validate:"required,min=1,max=256"`
	Sets []SetInput `json:"sets" validate:"required,min=1,dive"`
}

type SetInput struct {
	Weight  *int   `json:"weight,omitempty" validate:"omitempty,gte=0"`
	Reps    int    `json:"reps" validate:"required,gte=1"`
	SetType string `json:"setType" validate:"required,oneof=warmup working"`
}

// PostgreSQL-specific types

type PGWorkoutData struct {
	Date  pgtype.Timestamptz
	Notes pgtype.Text
}

type PGExerciseData struct {
	Name string
}

type PGSetData struct {
	ExerciseName string
	Weight       pgtype.Int4
	Reps         int32
	SetType      string
}

type PGReformattedRequest struct {
	Workout   PGWorkoutData
	Exercises []PGExerciseData
	Sets      []PGSetData
}

// Internal data structures

type WorkoutData struct {
	Date  time.Time
	Notes *string
}

type ExerciseData struct {
	Name string
}

type SetData struct {
	ExerciseName string
	Weight       *int
	Reps         int
	SetType      string
}
type ReformattedRequest struct {
	Workout   WorkoutData
	Exercises []ExerciseData
	Sets      []SetData
}

// UPDATE endpoint types for PUT /api/workouts/{id}
// Returns 204 No Content on success
type UpdateWorkoutRequest struct {
	Date      *string         `json:"date,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Notes     *string         `json:"notes,omitempty" validate:"omitempty,max=256"`
	Exercises []ExerciseInput `json:"exercises,omitempty" validate:"omitempty,dive"`
}
