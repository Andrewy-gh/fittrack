package workout

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

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

// structs for db insertion
type PGWorkoutData struct {
	Date  pgtype.Timestamptz
	Notes pgtype.Text
}

type PGExerciseData struct {
	Name string // Already a string, no conversion needed
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

// Reformatted structures for efficient DB operations
type WorkoutData struct {
	Date  time.Time
	Notes *string
}

type ExerciseData struct {
	Name string
}

type SetData struct {
	ExerciseName string // We'll use this to link to exercise after insertion
	Weight       *int
	Reps         int
	SetType      string
}

type ReformattedRequest struct {
	Workout   WorkoutData
	Exercises []ExerciseData // Unique exercises only
	Sets      []SetData      // All sets with exercise references
}
