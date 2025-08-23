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
	ExerciseName  string
	Weight        pgtype.Int4
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

// Interfaces for generic transformation
type WorkoutRequestTransformable interface {
	GetDate() *string
	GetNotes() *string
	GetExercises() []ExerciseTransformable
}

type ExerciseTransformable interface {
	GetName() string
	GetSets() []SetTransformable
}

type SetTransformable interface {
	GetWeight() *int
	GetReps() int
	GetSetType() string
}

// UPDATE endpoint types for PUT /api/workouts/{id}
// Returns 204 No Content on success
type UpdateWorkoutRequest struct {
	Date      string           `json:"date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Notes     *string          `json:"notes,omitempty" validate:"omitempty,max=256"`
	Exercises []UpdateExercise `json:"exercises" validate:"required,min=1,dive"`
}

type UpdateExercise struct {
	Name string      `json:"name" validate:"required,min=1,max=256"`
	Sets []UpdateSet `json:"sets" validate:"required,min=1,dive"`
}

type UpdateSet struct {
	Weight  *int   `json:"weight,omitempty" validate:"omitempty,gte=0"`
	Reps    int    `json:"reps" validate:"required,gte=1"`
	SetType string `json:"setType" validate:"required,oneof=warmup working"`
}

// Interface implementations for CreateWorkoutRequest
func (c CreateWorkoutRequest) GetDate() *string {
	return &c.Date
}

func (c CreateWorkoutRequest) GetNotes() *string {
	return c.Notes
}

func (c CreateWorkoutRequest) GetExercises() []ExerciseTransformable {
	result := make([]ExerciseTransformable, len(c.Exercises))
	for i, exercise := range c.Exercises {
		result[i] = exercise
	}
	return result
}

// Interface implementations for UpdateWorkoutRequest
func (u UpdateWorkoutRequest) GetDate() *string {
	return &u.Date
}

func (u UpdateWorkoutRequest) GetNotes() *string {
	return u.Notes
}

func (u UpdateWorkoutRequest) GetExercises() []ExerciseTransformable {
	result := make([]ExerciseTransformable, len(u.Exercises))
	for i, exercise := range u.Exercises {
		result[i] = exercise
	}
	return result
}

// Interface implementations for ExerciseInput
func (e ExerciseInput) GetName() string {
	return e.Name
}

func (e ExerciseInput) GetSets() []SetTransformable {
	result := make([]SetTransformable, len(e.Sets))
	for i, set := range e.Sets {
		result[i] = set
	}
	return result
}

// Interface implementations for UpdateExercise
func (u UpdateExercise) GetName() string {
	return u.Name
}

func (u UpdateExercise) GetSets() []SetTransformable {
	result := make([]SetTransformable, len(u.Sets))
	for i, set := range u.Sets {
		result[i] = set
	}
	return result
}

// Interface implementations for SetInput
func (s SetInput) GetWeight() *int {
	return s.Weight
}

func (s SetInput) GetReps() int {
	return s.Reps
}

func (s SetInput) GetSetType() string {
	return s.SetType
}

// Interface implementations for UpdateSet
func (u UpdateSet) GetWeight() *int {
	return u.Weight
}

func (u UpdateSet) GetReps() int {
	return u.Reps
}

func (u UpdateSet) GetSetType() string {
	return u.SetType
}
