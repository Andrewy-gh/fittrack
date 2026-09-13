package exercise

import (
	"math"
	"slices"
	"time"
)

// Goal Check compares logged observations with general references, not outcomes.
const GoalCheckPolicy = "strength-observations-v1"

type GoalCheckWindow struct {
	Start    time.Time `json:"start" validate:"required"`
	End      time.Time `json:"end" validate:"required"`
	Timezone string    `json:"timezone" validate:"required"`
}
type GoalCheckSession struct {
	WorkoutID   int32     `json:"workout_id" validate:"required"`
	Date        time.Time `json:"date" validate:"required"`
	WorkingSets int       `json:"working_sets" validate:"required"`
}
type GoalCheckFrequency struct {
	TrainingDays int    `json:"training_days" validate:"required"`
	ReferenceMin int    `json:"reference_min" validate:"required"`
	Status       string `json:"status" validate:"required"`
}
type GoalCheckSets struct {
	Total        int      `json:"total" validate:"required"`
	Average      *float64 `json:"average" extensions:"x-nullable"`
	ReferenceMin int      `json:"reference_min" validate:"required"`
	ReferenceMax int      `json:"reference_max" validate:"required"`
	Status       string   `json:"status" validate:"required"`
}
type GoalCheckReference struct {
	Value           float64    `json:"value" validate:"required"`
	Origin          string     `json:"origin" validate:"required"`
	UpdatedAt       *time.Time `json:"updated_at" extensions:"x-nullable"`
	SourceWorkoutID *int32     `json:"source_workout_id" extensions:"x-nullable"`
}
type GoalCheckLoading struct {
	RepsMin      *int                `json:"reps_min" extensions:"x-nullable"`
	RepsMax      *int                `json:"reps_max" extensions:"x-nullable"`
	SetsWithLoad int                 `json:"sets_with_load" validate:"required"`
	Reference    *GoalCheckReference `json:"reference" extensions:"x-nullable"`
	PercentMin   *float64            `json:"percent_min" extensions:"x-nullable"`
	PercentMax   *float64            `json:"percent_max" extensions:"x-nullable"`
}
type StrengthGoalCheck struct {
	ExerciseID    int32              `json:"exercise_id" validate:"required"`
	Applicable    bool               `json:"applicable" validate:"required"`
	PolicyVersion string             `json:"policy_version" validate:"required"`
	Window        GoalCheckWindow    `json:"window" validate:"required"`
	Sessions      []GoalCheckSession `json:"sessions" validate:"required"`
	Frequency     GoalCheckFrequency `json:"frequency" validate:"required"`
	WorkingSets   GoalCheckSets      `json:"working_sets" validate:"required"`
	Loading       GoalCheckLoading   `json:"loading" validate:"required"`
}

type goalCheckSet struct {
	ID        int32
	WorkoutID int32
	Date      time.Time
	Weight    *float64
	Reps      int
}

func goalCheckWindow(now time.Time, location *time.Location) GoalCheckWindow {
	local := now.In(location)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -6)
	return GoalCheckWindow{Start: start, End: now, Timezone: location.String()}
}

// Inputs are owned working sets. Recheck boundaries to keep the policy independent
// of SQL filtering. No same-session estimated-max fallback or aggregate verdict.
func calculateGoalCheck(id int32, goals []string, window GoalCheckWindow, location *time.Location, sets []goalCheckSet, reference *GoalCheckReference) StrengthGoalCheck {
	result := StrengthGoalCheck{
		ExerciseID: id, Applicable: slices.Contains(goals, "strength"), PolicyVersion: GoalCheckPolicy,
		Window: window, Sessions: []GoalCheckSession{},
		Frequency:   GoalCheckFrequency{ReferenceMin: 2, Status: "insufficient_data"},
		WorkingSets: GoalCheckSets{ReferenceMin: 2, ReferenceMax: 3, Status: "insufficient_data"},
	}
	if reference != nil && positiveFinite(reference.Value) {
		result.Loading.Reference = reference
	}
	days := map[string]bool{}
	sessions := map[int32]int{}
	seen := map[int32]bool{}
	for _, set := range sets {
		if seen[set.ID] || set.Reps < 1 || set.Date.Before(window.Start) || set.Date.After(window.End) {
			continue
		}
		seen[set.ID] = true
		index, ok := sessions[set.WorkoutID]
		if !ok {
			index = len(result.Sessions)
			sessions[set.WorkoutID] = index
			result.Sessions = append(result.Sessions, GoalCheckSession{WorkoutID: set.WorkoutID, Date: set.Date})
		}
		result.Sessions[index].WorkingSets++
		days[set.Date.In(location).Format("2006-01-02")] = true
		result.WorkingSets.Total++
		if result.Loading.RepsMin == nil || set.Reps < *result.Loading.RepsMin {
			value := set.Reps
			result.Loading.RepsMin = &value
		}
		if result.Loading.RepsMax == nil || set.Reps > *result.Loading.RepsMax {
			value := set.Reps
			result.Loading.RepsMax = &value
		}
		if set.Weight == nil || !positiveFinite(*set.Weight) {
			continue
		}
		result.Loading.SetsWithLoad++
		if result.Loading.Reference == nil {
			continue
		}
		percent := *set.Weight / result.Loading.Reference.Value * 100
		if !positiveFinite(percent) {
			continue
		}
		if result.Loading.PercentMin == nil || percent < *result.Loading.PercentMin {
			value := percent
			result.Loading.PercentMin = &value
		}
		if result.Loading.PercentMax == nil || percent > *result.Loading.PercentMax {
			value := percent
			result.Loading.PercentMax = &value
		}
	}
	slices.SortFunc(result.Sessions, func(a, b GoalCheckSession) int {
		if order := a.Date.Compare(b.Date); order != 0 {
			return order
		}
		if a.WorkoutID < b.WorkoutID {
			return -1
		}
		if a.WorkoutID > b.WorkoutID {
			return 1
		}
		return 0
	})
	result.Frequency.TrainingDays = len(days)
	if len(result.Sessions) > 0 {
		average := float64(result.WorkingSets.Total) / float64(len(result.Sessions))
		result.WorkingSets.Average = &average
		if result.Applicable {
			result.Frequency.Status = referenceStatus(float64(len(days)), 2)
			result.WorkingSets.Status = referenceStatus(average, 2)
		}
	}
	return result
}
func positiveFinite(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
func referenceStatus(value, minimum float64) string {
	if value >= minimum {
		return "reference_met"
	}
	return "below_reference"
}
