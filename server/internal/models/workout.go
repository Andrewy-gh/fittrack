package models

import "time"

// Input structures (what the API receives)
type WorkoutRequest struct {
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
	SetType string `json:"setType" validate:"required,min=1,max=256"`
}

// Response structures (what the API returns)
type WorkoutResponse struct {
	WorkoutID int32              `json:"workoutId"`
	Date      time.Time          `json:"date"`
	Notes     *string            `json:"notes,omitempty"`
	Exercises []ExerciseResponse `json:"exercises"`
}

type ExerciseResponse struct {
	ExerciseID int32         `json:"exerciseId"`
	Name       string        `json:"name"`
	Sets       []SetResponse `json:"sets"`
}

type SetResponse struct {
	SetID   int32  `json:"setId"`
	Weight  *int   `json:"weight,omitempty"`
	Reps    int    `json:"reps"`
	SetType string `json:"setType"`
}
