package exercise

import "time"

// ExerciseResponse represents an exercise response for swagger documentation
type ExerciseResponse struct {
	ID        int32     `json:"id" example:"1"`
	Name      string    `json:"name" example:"Bench Press"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T15:04:05Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T15:04:05Z"`
	UserID    string    `json:"user_id" example:"user-123"`
}

// ExerciseWithSetsResponse represents an exercise with sets response for swagger documentation
type ExerciseWithSetsResponse struct {
	WorkoutID    int32     `json:"workout_id" example:"1"`
	WorkoutDate  time.Time `json:"workout_date" example:"2023-01-01T15:04:05Z"`
	WorkoutNotes *string   `json:"workout_notes,omitempty" example:"Great workout today"`
	SetID        int32     `json:"set_id" example:"1"`
	Weight       *int32    `json:"weight,omitempty" example:"225"`
	Reps         int32     `json:"reps" example:"10"`
	SetType      string    `json:"set_type" example:"working"`
	ExerciseID   int32     `json:"exercise_id" example:"1"`
	ExerciseName string    `json:"exercise_name" example:"Bench Press"`
	Volume       int32     `json:"volume" example:"2250"`
}

// CreateExerciseResponse represents the response when creating/getting an exercise
type CreateExerciseResponse struct {
	ID        int32     `json:"id" example:"1"`
	Name      string    `json:"name" example:"Bench Press"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T15:04:05Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T15:04:05Z"`
	UserID    string    `json:"user_id" example:"user-123"`
}
