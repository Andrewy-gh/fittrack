package workout

func toCreateWorkoutRequest(reformatted *ReformattedRequest) (CreateWorkoutRequest, error) {
	request := CreateWorkoutRequest{
		Date:      reformatted.Workout.Date.UTC().Format("2006-01-02T15:04:05Z07:00"),
		Exercises: make([]ExerciseInput, 0, len(reformatted.Exercises)),
	}

	if reformatted.Workout.Notes != nil {
		notes := *reformatted.Workout.Notes
		request.Notes = &notes
	}

	if reformatted.Workout.WorkoutFocus != nil {
		workoutFocus := *reformatted.Workout.WorkoutFocus
		request.WorkoutFocus = &workoutFocus
	}

	setsByExercise := make(map[string][]SetInput, len(reformatted.Exercises))
	for _, set := range reformatted.Sets {
		var weight *float64
		if set.Weight != nil {
			value := *set.Weight
			weight = &value
		}
		setsByExercise[set.ExerciseName] = append(setsByExercise[set.ExerciseName], SetInput{
			Weight:  weight,
			Reps:    set.Reps,
			SetType: set.SetType,
		})
	}

	for _, exercise := range reformatted.Exercises {
		request.Exercises = append(request.Exercises, ExerciseInput{
			Name: exercise.Name,
			Sets: setsByExercise[exercise.Name],
		})
	}

	return request, nil
}
