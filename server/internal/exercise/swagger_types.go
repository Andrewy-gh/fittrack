package exercise

import "time"

// ExerciseResponse represents an exercise response for swagger documentation
type ExerciseResponse struct {
	ID        int32     `json:"id" validate:"required" example:"1"`
	Name      string    `json:"name" validate:"required" example:"Bench Press"`
	CreatedAt time.Time `json:"created_at" validate:"required" example:"2023-01-01T15:04:05Z"`
	UpdatedAt time.Time `json:"updated_at" validate:"required" example:"2023-01-01T15:04:05Z"`
	UserID    string    `json:"user_id" validate:"required" example:"user-123"`
}

// ExerciseWithSetsResponse represents an exercise with sets response for swagger documentation
type ExerciseWithSetsResponse struct {
	WorkoutID     int32     `json:"workout_id" validate:"required" example:"1"`
	WorkoutDate   time.Time `json:"workout_date" validate:"required" example:"2023-01-01T15:04:05Z"`
	WorkoutNotes  *string   `json:"workout_notes,omitempty" example:"Great workout today"`
	SetID         int32     `json:"set_id" validate:"required" example:"1"`
	Weight        *int32    `json:"weight,omitempty" example:"225"`
	Reps          int32     `json:"reps" validate:"required" example:"10"`
	SetType       string    `json:"set_type" validate:"required" example:"working"`
	ExerciseID    int32     `json:"exercise_id" validate:"required" example:"1"`
	ExerciseName  string    `json:"exercise_name" validate:"required" example:"Bench Press"`
	ExerciseOrder *int32    `json:"exercise_order,omitempty" example:"0"`
	SetOrder      *int32    `json:"set_order,omitempty" example:"1"`
	Volume        int32     `json:"volume" validate:"required" example:"2250"`
}

// CreateExerciseResponse represents the response when creating/getting an exercise
type CreateExerciseResponse struct {
	ID        int32     `json:"id" validate:"required" example:"1"`
	Name      string    `json:"name" validate:"required" example:"Bench Press"`
	CreatedAt time.Time `json:"created_at" validate:"required" example:"2023-01-01T15:04:05Z"`
	UpdatedAt time.Time `json:"updated_at" validate:"required" example:"2023-01-01T15:04:05Z"`
	UserID    string    `json:"user_id" validate:"required" example:"user-123"`
}

// RecentSetsResponse represents recent sets response for swagger documentation
type RecentSetsResponse struct {
	SetID       int32     `json:"set_id" validate:"required" example:"1"`
	WorkoutDate time.Time `json:"workout_date" validate:"required" example:"2023-01-01T15:04:05Z"`
	Weight      *int32    `json:"weight,omitempty" example:"225"`
	Reps        int32     `json:"reps" validate:"required" example:"10"`
	CreatedAt   time.Time `json:"created_at" validate:"required" example:"2023-01-01T15:04:05Z"`
}
