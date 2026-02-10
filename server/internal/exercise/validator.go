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

type GetExerciseMetricsHistoryRequest struct {
	ExerciseID int32  `json:"exercise_id" validate:"required,min=1"`
	Range      string `json:"range" validate:"required,oneof=W M 6M Y"`
}

type UpdateExerciseNameRequest struct {
	Name string `json:"name" validate:"required,max=256"`
}

type UpdateExerciseHistorical1RMRequest struct {
	Mode          string   `json:"mode" validate:"omitempty,oneof=manual recompute"`
	Historical1RM *float64 `json:"historical_1rm" validate:"omitempty,gte=0,lte=999999.99"`
}
