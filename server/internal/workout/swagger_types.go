package workout

import "time"

// WorkoutResponse represents a workout response for swagger documentation
type WorkoutResponse struct {
	ID        int32      `json:"id" example:"1"`
	Date      time.Time  `json:"date" example:"2023-01-01T15:04:05Z"`
	Notes     *string    `json:"notes,omitempty" example:"Great workout today"`
	CreatedAt time.Time  `json:"created_at" example:"2023-01-01T15:04:05Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2023-01-01T15:04:05Z"`
	UserID    string     `json:"user_id" example:"user-123"`
}

// SetResponse represents a set response for swagger documentation
type SetResponse struct {
	ID         int32      `json:"id" example:"1"`
	ExerciseID int32      `json:"exercise_id" example:"1"`
	WorkoutID  int32      `json:"workout_id" example:"1"`
	Weight     *int32     `json:"weight,omitempty" example:"225"`
	Reps       int32      `json:"reps" example:"10"`
	SetType    string     `json:"set_type" example:"working"`
	CreatedAt  time.Time  `json:"created_at" example:"2023-01-01T15:04:05Z"`
	UpdatedAt  time.Time  `json:"updated_at" example:"2023-01-01T15:04:05Z"`
}

// ExerciseResponse represents an exercise response for swagger documentation
type ExerciseResponse struct {
	ID        int32      `json:"id" example:"1"`
	Name      string     `json:"name" example:"Bench Press"`
	CreatedAt time.Time  `json:"created_at" example:"2023-01-01T15:04:05Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2023-01-01T15:04:05Z"`
	UserID    string     `json:"user_id" example:"user-123"`
}
