package workout

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Request/Response types
type CreateWorkoutRequest struct {
	Date         string          `json:"date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Notes        *string         `json:"notes,omitempty" validate:"omitempty,max=256"`
	WorkoutFocus *string         `json:"workoutFocus,omitempty" validate:"omitempty,max=256"`
	Exercises    []ExerciseInput `json:"exercises" validate:"required,min=1,dive"`
}

type ExerciseInput struct {
	Name string     `json:"name" validate:"required,min=1,max=256"`
	Sets []SetInput `json:"sets" validate:"required,min=1,dive"`
}

type SetInput struct {
	Weight  *float64 `json:"weight,omitempty" validate:"omitempty,gte=0,lte=999999999.9"`
	Reps    int      `json:"reps" validate:"required,gte=1"`
	SetType string   `json:"setType" validate:"required,oneof=warmup working"`
}

type UpdateExercise struct {
	Name string      `json:"name" validate:"required,min=1,max=256"`
	Sets []UpdateSet `json:"sets" validate:"required,min=1,dive"`
}

type UpdateSet struct {
	Weight  *float64 `json:"weight,omitempty" validate:"omitempty,gte=0,lte=999999999.9"`
	Reps    int      `json:"reps" validate:"required,gte=1"`
	SetType string   `json:"setType" validate:"required,oneof=warmup working"`
}

type exerciseRequestDraft struct {
	Name string
	Sets []setRequestDraft
}

type setRequestDraft struct {
	Weight  *float64
	Reps    int
	SetType string
}

type workoutRequestDraft struct {
	Date         string
	Notes        *string
	WorkoutFocus *string
	Exercises    []exerciseRequestDraft
}

// PostgreSQL-specific types

type PGWorkoutData struct {
	Date         pgtype.Timestamptz
	Notes        pgtype.Text
	WorkoutFocus pgtype.Text
}

type PGExerciseData struct {
	Name string
}

type PGSetData struct {
	ExerciseName  string
	Weight        pgtype.Numeric
	Reps          int32
	SetType       string
	ExerciseOrder int32
	SetOrder      int32
}

type PGReformattedRequest struct {
	Workout   PGWorkoutData
	Exercises []PGExerciseData
	Sets      []PGSetData
}

// Internal data structures

type WorkoutData struct {
	Date         time.Time
	Notes        *string
	WorkoutFocus *string
}

type ExerciseData struct {
	Name string
}

type SetData struct {
	ExerciseName string
	Weight       *float64
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
	Date         string           `json:"date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Notes        *string          `json:"notes,omitempty" validate:"omitempty,max=256"`
	WorkoutFocus *string          `json:"workoutFocus,omitempty" validate:"omitempty,max=256"`
	Exercises    []UpdateExercise `json:"exercises" validate:"required,min=1,dive"`
}

// Contribution Graph types for GET /api/workouts/contribution-data
type WorkoutSummary struct {
	ID     int32   `json:"id"`
	Time   string  `json:"time"`
	Focus  *string `json:"focus"`
	Volume float64 `json:"volume"`
}

type ContributionDay struct {
	Date     string           `json:"date"`
	Count    int              `json:"count"`
	Level    int              `json:"level"`
	Workouts []WorkoutSummary `json:"workouts"`
}

type ContributionDataResponse struct {
	Days []ContributionDay `json:"days"`
}

// FocusTemplateResponse identifies the newest reusable workout for one focus area.
type FocusTemplateResponse struct {
	Focus     string    `json:"focus" validate:"required" example:"Upper Body"`
	WorkoutID int32     `json:"workoutId" validate:"required" example:"1"`
	Date      time.Time `json:"date" validate:"required" example:"2023-01-01T15:04:05Z"`
}

type LatestWorkoutNoteResponse struct {
	WorkoutID int32     `json:"workoutId" validate:"required" example:"1"`
	Date      time.Time `json:"date" validate:"required" example:"2023-01-01T15:04:05Z"`
	Note      string    `json:"note" validate:"required" example:"Great workout today"`
}

type NewWorkoutContextResponse struct {
	FocusTemplates    []FocusTemplateResponse    `json:"focusTemplates"`
	LatestWorkoutNote *LatestWorkoutNoteResponse `json:"latestWorkoutNote,omitempty"`
}
