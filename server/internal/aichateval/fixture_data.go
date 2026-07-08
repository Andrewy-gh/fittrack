package aichateval

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
)

const FixtureUserID = "ai-chat-fixture-user"

type fixtureChatDataReader struct {
	userID   string
	workouts []aichat.ChatWorkoutView
	profile  *aichat.TrainingProfile
}

func NewFixtureChatDataReader() aichat.ChatDataReader {
	return &fixtureChatDataReader{
		userID:   FixtureUserID,
		workouts: fixtureWorkouts(),
		profile: &aichat.TrainingProfile{
			PrimaryGoal:                     "hypertrophy",
			ExperienceLevel:                 "intermediate",
			PreferredSessionDurationMinutes: 45,
			UsualTrainingLocation:           "home",
			AvailableEquipment:              []string{"adjustable dumbbells", "bench"},
			MovementLimitations:             nil,
		},
	}
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
		if filter.ExerciseName != "" {
			filteredExercises := matchingWorkoutExercises(workout, filter.ExerciseName)
			if len(filteredExercises) == 0 {
				continue
			}
			workout.Exercises = filteredExercises
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

func (r *fixtureChatDataReader) TrainingProfile(ctx context.Context, userID string) (*aichat.TrainingProfile, error) {
	_ = ctx
	if userID != r.userID {
		return nil, nil
	}
	return r.profile, nil
}

func (r *fixtureChatDataReader) ExerciseStats(ctx context.Context, userID string, exerciseName string, window string) (*aichat.ExerciseStatsView, error) {
	_ = ctx
	if userID != r.userID {
		return nil, nil
	}
	window = normalizeFixtureStatsWindow(window)
	stats := &aichat.ExerciseStatsView{ExerciseName: exerciseName, Window: window}
	var points []aichat.ExerciseStatsTrendPoint
	for _, workout := range r.workouts {
		exercises := matchingWorkoutExercises(workout, exerciseName)
		if len(exercises) == 0 {
			continue
		}
		best := 0.0
		volume := 0.0
		for _, set := range exercises[0].Sets {
			weight, reps, ok := parseFixtureSet(set)
			if !ok {
				continue
			}
			e1rm := weight * (1 + float64(reps)/30)
			if e1rm > best {
				best = e1rm
			}
			volume += weight * float64(reps)
		}
		if best == 0 {
			continue
		}
		if stats.BestE1RM == nil || best > stats.BestE1RM.Weight {
			stats.BestE1RM = &aichat.ExerciseBestE1RMView{Weight: best, Date: workout.Date}
		}
		points = append(points, aichat.ExerciseStatsTrendPoint{
			Date:     workout.Date,
			BestE1RM: best,
			AvgE1RM:  best,
			Volume:   volume,
		})
		if len(stats.LastSessionSets) == 0 {
			stats.LastSessionDate = workout.Date
			stats.LastSessionSets = append([]string{}, exercises[0].Sets...)
		}
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})
	stats.SessionCount = len(points)
	stats.Trend = compactFixtureTrend(points, 8)
	if stats.SessionCount == 0 {
		stats.Message = fmt.Sprintf("No working-set stats were found for %s.", exerciseName)
	}
	return stats, nil
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

func normalizeFixtureStatsWindow(window string) string {
	switch strings.ToLower(strings.TrimSpace(window)) {
	case "1y", "all":
		return strings.ToLower(strings.TrimSpace(window))
	default:
		return "3m"
	}
}

func parseFixtureSet(set string) (float64, int, bool) {
	firstField := strings.Fields(set)
	if len(firstField) == 0 {
		return 0, 0, false
	}
	parts := strings.Split(firstField[0], "x")
	if len(parts) != 2 {
		return 0, 0, false
	}
	weight, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, false
	}
	reps, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}
	return weight, reps, true
}

func compactFixtureTrend(points []aichat.ExerciseStatsTrendPoint, maxPoints int) []aichat.ExerciseStatsTrendPoint {
	if len(points) <= maxPoints {
		return points
	}
	compact := make([]aichat.ExerciseStatsTrendPoint, 0, maxPoints)
	lastIndex := len(points) - 1
	for i := 0; i < maxPoints; i++ {
		compact = append(compact, points[i*lastIndex/(maxPoints-1)])
	}
	return compact
}

func fixtureWorkout(date string, focus string, notes string, exercises ...aichat.ChatExerciseView) aichat.ChatWorkoutView {
	return aichat.ChatWorkoutView{Date: date, Focus: focus, Notes: notes, Exercises: exercises}
}

func fixtureExercise(name string, sets ...string) aichat.ChatExerciseView {
	return aichat.ChatExerciseView{Name: name, Sets: sets}
}

func matchingWorkoutExercises(workout aichat.ChatWorkoutView, name string) []aichat.ChatExerciseView {
	var matches []aichat.ChatExerciseView
	for _, exercise := range workout.Exercises {
		if strings.EqualFold(exercise.Name, name) {
			matches = append(matches, exercise)
		}
	}
	return matches
}

func startOfDay(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func endOfDay(value time.Time) time.Time {
	return startOfDay(value).Add(24*time.Hour - time.Nanosecond)
}
