package workout

import "time"

// WorkoutResponse represents a workout response for swagger documentation
// @Description Workout response model
type WorkoutResponse struct {
	ID           int32      `json:"id" validate:"required" example:"1"`
	Date         time.Time  `json:"date" validate:"required" example:"2023-01-01T15:04:05Z"`
	Notes        *string    `json:"notes,omitempty" example:"Great workout today"`
	WorkoutFocus *string    `json:"workout_focus,omitempty" example:"Upper Body"`
	CreatedAt    time.Time  `json:"created_at" validate:"required" example:"2023-01-01T15:04:05Z"`
	UpdatedAt    time.Time  `json:"updated_at" validate:"required" example:"2023-01-01T15:04:05Z"`
	UserID       string     `json:"user_id" validate:"required" example:"user-123"`
}

// SetResponse represents a set response for swagger documentation
type SetResponse struct {
	ID            int32      `json:"id" example:"1"`
	ExerciseID    int32      `json:"exercise_id" example:"1"`
	WorkoutID     int32      `json:"workout_id" example:"1"`
	Weight        *float64   `json:"weight,omitempty" example:"225.5"`
	Reps          int32      `json:"reps" example:"10"`
	SetType       string     `json:"set_type" example:"working"`
	ExerciseOrder *int32     `json:"exercise_order,omitempty" example:"0"`
	SetOrder      *int32     `json:"set_order,omitempty" example:"1"`
	CreatedAt     time.Time  `json:"created_at" example:"2023-01-01T15:04:05Z"`
	UpdatedAt     time.Time  `json:"updated_at" example:"2023-01-01T15:04:05Z"`
}

// ExerciseResponse represents an exercise response for swagger documentation
type ExerciseResponse struct {
	ID        int32      `json:"id" validate:"required" example:"1"`
	Name      string     `json:"name" validate:"required" example:"Bench Press"`
	CreatedAt time.Time  `json:"created_at" validate:"required" example:"2023-01-01T15:04:05Z"`
	UpdatedAt time.Time  `json:"updated_at" validate:"required" example:"2023-01-01T15:04:05Z"`
	UserID    string     `json:"user_id" validate:"required" example:"user-123"`
}

// WorkoutWithSetsResponse represents a workout with sets response for swagger documentation
type WorkoutWithSetsResponse struct {
	WorkoutID     int32      `json:"workout_id" validate:"required" example:"1"`
	WorkoutDate   time.Time  `json:"workout_date" validate:"required" example:"2023-01-01T15:04:05Z"`
	WorkoutNotes  *string    `json:"workout_notes,omitempty" example:"Great workout today"`
	WorkoutFocus  *string    `json:"workout_focus,omitempty" example:"Upper Body"`
	SetID         int32      `json:"set_id" validate:"required" example:"1"`
	Weight        *float64   `json:"weight,omitempty" example:"225.5"`
	Reps          int32      `json:"reps" validate:"required" example:"10"`
	SetType       string     `json:"set_type" validate:"required" example:"working"`
	ExerciseID    int32      `json:"exercise_id" validate:"required" example:"1"`
	ExerciseName  string     `json:"exercise_name" validate:"required" example:"Bench Press"`
	ExerciseOrder *int32     `json:"exercise_order,omitempty" example:"0"`
	SetOrder      *int32     `json:"set_order,omitempty" example:"1"`
	Volume        float64    `json:"volume" validate:"required" example:"2250.5"`
}

// UpdateWorkoutRequest represents an update workout request for swagger documentation
// @Description Request model for updating existing workout metadata
type UpdateWorkoutRequestSwagger struct {
	Date         *string         `json:"date,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00" example:"2023-01-15T10:00:00Z"`
	Notes        *string         `json:"notes,omitempty" validate:"omitempty,max=256" example:"Updated workout notes"`
	WorkoutFocus *string         `json:"workout_focus,omitempty" validate:"omitempty,max=256" example:"Upper Body"`
	Exercises    []ExerciseInput `json:"exercises,omitempty" validate:"omitempty,dive" example:"[]"`
}
