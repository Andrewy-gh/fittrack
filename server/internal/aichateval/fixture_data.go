package aichateval

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
)

const FixtureUserID = "ai-chat-fixture-user"

type fixtureChatDataReader struct {
	userID   string
	workouts []aichat.ChatWorkoutView
}

func NewFixtureChatDataReader() aichat.ChatDataReader {
	return &fixtureChatDataReader{userID: FixtureUserID, workouts: fixtureWorkouts()}
}

func (r *fixtureChatDataReader) ListWorkoutsWithSets(ctx context.Context, userID string, filter aichat.WorkoutHistoryFilter) ([]aichat.ChatWorkoutView, error) {
	_ = ctx
	if userID != r.userID {
		return nil, nil
	}
	if filter.LastN <= 0 {
		filter.LastN = 5
	}
	if filter.LastN > 20 {
		filter.LastN = 20
	}

	var workouts []aichat.ChatWorkoutView
	for _, workout := range r.workouts {
		workoutDate, _ := time.Parse("2006-01-02", workout.Date)
		if filter.StartDate != nil && workoutDate.Before(startOfDay(*filter.StartDate)) {
			continue
		}
		if filter.EndDate != nil && workoutDate.After(endOfDay(*filter.EndDate)) {
			continue
		}
		if filter.WorkoutFocus != "" && !strings.Contains(strings.ToLower(workout.Focus), strings.ToLower(filter.WorkoutFocus)) {
			continue
		}
		if filter.ExerciseName != "" && !workoutHasExercise(workout, filter.ExerciseName) {
			continue
		}
		workouts = append(workouts, workout)
		if len(workouts) == filter.LastN {
			break
		}
	}
	return workouts, nil
}

func (r *fixtureChatDataReader) ResolveExerciseNames(ctx context.Context, userID string, query string) ([]string, error) {
	_ = ctx
	if userID != r.userID {
		return nil, nil
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil, nil
	}
	seen := make(map[string]bool)
	for _, workout := range r.workouts {
		for _, exercise := range workout.Exercises {
			if strings.Contains(strings.ToLower(exercise.Name), query) {
				seen[exercise.Name] = true
			}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) > 8 {
		names = names[:8]
	}
	return names, nil
}

func (r *fixtureChatDataReader) TrainingSnapshot(ctx context.Context, userID string) (*aichat.TrainingSnapshot, error) {
	_ = ctx
	if userID != r.userID {
		return nil, nil
	}
	return &aichat.TrainingSnapshot{
		LastWorkoutDate: "2026-07-03",
		WorkoutsLast30D: 7,
		TopExercises:    []string{"Bench Press", "Back Squat", "Deadlift", "Barbell Row", "Incline Bench Press"},
	}, nil
}

func fixtureWorkouts() []aichat.ChatWorkoutView {
	return []aichat.ChatWorkoutView{
		fixtureWorkout("2026-07-03", "upper body", "Top bench single felt smooth.", fixtureExercise("Bench Press", "185x5 working", "195x3 working"), fixtureExercise("Barbell Row", "155x8 working", "155x8 working")),
		fixtureWorkout("2026-06-30", "lower body", "", fixtureExercise("Back Squat", "225x5 working", "235x3 working"), fixtureExercise("Romanian Deadlift", "185x8 working")),
		fixtureWorkout("2026-06-27", "pull", "", fixtureExercise("Deadlift", "275x3 working", "285x2 working"), fixtureExercise("Barbell Row", "165x6 working")),
		fixtureWorkout("2026-06-23", "push", "", fixtureExercise("Bench Press", "180x6 working", "185x4 working"), fixtureExercise("Incline Bench Press", "145x8 working")),
		fixtureWorkout("2026-06-19", "legs", "", fixtureExercise("Front Squat", "185x5 working"), fixtureExercise("Walking Lunge", "40x10 working")),
		fixtureWorkout("2026-06-15", "pull", "", fixtureExercise("Deadlift", "265x4 working"), fixtureExercise("One-Arm Dumbbell Row", "75x10 working")),
		fixtureWorkout("2026-06-10", "upper body", "", fixtureExercise("Incline Bench Press", "140x8 working"), fixtureExercise("Chest Supported Row", "120x10 working")),
		fixtureWorkout("2026-06-05", "lower body", "", fixtureExercise("Back Squat", "215x5 working"), fixtureExercise("Leg Curl", "90x12 working")),
		fixtureWorkout("2026-05-30", "push", "", fixtureExercise("Bench Press", "175x6 working"), fixtureExercise("Dumbbell Shoulder Press", "55x8 working")),
		fixtureWorkout("2026-05-24", "full body", "", fixtureExercise("Deadlift", "255x5 working"), fixtureExercise("Back Squat", "205x6 working")),
	}
}

func fixtureWorkout(date string, focus string, notes string, exercises ...aichat.ChatExerciseView) aichat.ChatWorkoutView {
	return aichat.ChatWorkoutView{Date: date, Focus: focus, Notes: notes, Exercises: exercises}
}

func fixtureExercise(name string, sets ...string) aichat.ChatExerciseView {
	return aichat.ChatExerciseView{Name: name, Sets: sets}
}

func workoutHasExercise(workout aichat.ChatWorkoutView, name string) bool {
	for _, exercise := range workout.Exercises {
		if strings.EqualFold(exercise.Name, name) {
			return true
		}
	}
	return false
}

func startOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func endOfDay(value time.Time) time.Time {
	return startOfDay(value).Add(24*time.Hour - time.Nanosecond)
}
