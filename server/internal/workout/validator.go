package workout

type GetWorkoutWithSetsRequest struct {
	WorkoutID int32 `json:"workout_id" validate:"required,min=1"`
}
