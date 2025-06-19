package workout

import "time"

type Set struct {
	Weight  *int   `json:"weight" validate:"omitempty,gte=0"`
	Reps    *int   `json:"reps" validate:"required,gte=1"`                       // Changed to pointer
	SetType string `json:"setType" validate:"required,oneof=warmup working,ne="` // ne="not equal to empty string"
}

type Exercise struct {
	Name string `json:"name" validate:"required"`
	Sets []Set  `json:"sets" validate:"required,min=1"`
}

type CreateWorkoutRequest struct {
	Date      time.Time  `json:"date" validate:"required"`
	Exercises []Exercise `json:"exercises" validate:"required,min=1"`
	Notes     *string    `json:"notes"`
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
