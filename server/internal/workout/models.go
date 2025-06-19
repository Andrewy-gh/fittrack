package workout

import "time"

// Incoming request structure
type IncomingRequest struct {
	Date      time.Time `json:"date"`
	Exercises []struct {
		Name string `json:"name"`
		Sets []struct {
			Weight  *int   `json:"weight"` // Pointer to handle null values
			Reps    int    `json:"reps"`
			SetType string `json:"setType"`
		} `json:"sets"`
	} `json:"exercises"`
	Notes *string `json:"notes"` // Optional field
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
