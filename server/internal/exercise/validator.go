package exercise

type GetExerciseWithSetsRequest struct {
	ExerciseID int32 `json:"exercise_id" validate:"required,min=1"`
}

type CreateExerciseRequest struct {
	Name string `json:"name" validate:"required"`
}

type GetRecentSetsRequest struct {
	ExerciseID int32 `json:"exercise_id" validate:"required,min=1"`
}

type UpdateExerciseNameRequest struct {
	Name string `json:"name" validate:"required,max=256"`
}
