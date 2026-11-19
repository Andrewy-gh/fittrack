package exercise

type CreateExerciseRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateExerciseNameRequest struct {
	Name string `json:"name" validate:"required,max=256"`
}

type UpdateExerciseHistorical1RMRequest struct {
	Mode          string   `json:"mode" validate:"omitempty,oneof=manual recompute"`
	Historical1RM *float64 `json:"historical_1rm" validate:"omitempty,gte=0,lte=999999.99"`
}
