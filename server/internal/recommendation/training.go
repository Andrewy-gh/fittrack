package recommendation

import (
	"math"
	"strings"
	"time"
)

const Recency = 28 * 24 * time.Hour

// TrainingContext contains only the owner's recent exercise sessions and relevant profile facts.
type TrainingContext struct {
	Sessions       []Session
	Goal           string
	Experience     string
	Avoided        bool
	HasLimitations bool
}

type Session struct {
	Date time.Time
	Sets []WorkingSet
}

type WorkingSet struct {
	Weight *float64
	Reps   int
}

// Plan is a starting point, separate from the actual sets recorded by the user.
type Plan struct {
	Sets   int      `json:"sets" validate:"required"`
	Reps   int      `json:"reps" validate:"required"`
	Weight *float64 `json:"weight" extensions:"x-nullable"`
}

func RecommendTraining(now time.Time, context TrainingContext, readiness string) Result {
	result := Result{
		Readiness:     readiness,
		Explanation:   "Choose how you feel today to get a suggestion.",
		PolicyVersion: PolicyVersion,
	}
	if !validReadiness(readiness) {
		return result
	}
	if context.Avoided || context.HasLimitations {
		result.Explanation = "Your profile lists an exercise to avoid or a movement limitation. Choose a suitable exercise or review your profile first."
		return result
	}

	low, high := 8, 12
	if context.Goal == "strength" && (context.Experience == "intermediate" || context.Experience == "advanced") {
		low, high = 5, 8
	} else if context.Goal == "endurance" {
		low, high = 12, 15
	}
	plan := &Plan{Sets: 2, Reps: low}
	if context.Experience == "intermediate" || context.Experience == "advanced" {
		plan.Sets = 3
	}

	latest, hasRecentHistory := recentSession(now, context.Sessions)
	if hasRecentHistory {
		plan.Sets = len(latest.Sets)
	}
	if readiness == "sluggish" {
		plan.Sets = max(1, plan.Sets-1)
		plan.Reps = max(1, plan.Reps-2)
	}
	result.Plan = plan
	result.Explanation = "Choose a comfortable starting weight; we don’t have recent working sets to use."
	if !hasRecentHistory {
		return result
	}

	sets := latest.Sets
	reps := sets[0].Reps
	weight := sets[0].Weight
	uniform := validWeight(weight)
	for _, set := range sets {
		if set.Reps < 1 || set.Reps > 100 {
			result.Plan = nil
			result.Explanation = "Log a working set with valid reps to get a suggestion."
			return result
		}
		reps = min(reps, set.Reps)
		if set.Weight == nil || weight == nil || *set.Weight != *weight {
			uniform = false
		}
	}
	plan.Reps = reps
	if uniform {
		value := *weight
		plan.Weight = &value
	}
	result.Explanation = "Repeat your last working sets."
	if !uniform {
		result.Explanation = "Your last weights varied or weren’t recorded. Choose the weight for these reps."
	}
	if readiness == "sluggish" {
		plan.Reps = max(1, reps-2)
		result.Explanation = "An easier day: fewer reps and, where possible, one fewer set."
		return result
	}

	progress := uniform && reps >= high && reachedTarget(context.Sessions, len(sets), high)
	if progress && *weight >= 50 && *weight <= 2000 {
		value := math.Round((*weight+2.5)*10) / 10
		plan.Weight, plan.Reps = &value, low
		result.Explanation = "You reached the rep target in two workouts. Try a small weight increase, using the closest available weight."
	} else if readiness == "great" && uniform && reps < high {
		plan.Reps = reps + 1
		result.Explanation = "Try one extra rep at the same weight."
	} else if readiness == "great" {
		result.Explanation += " There isn’t enough evidence to increase it yet."
	}
	return result
}

func recentSession(now time.Time, sessions []Session) (Session, bool) {
	if len(sessions) == 0 {
		return Session{}, false
	}
	latest := sessions[0]
	if latest.Date.After(now) || now.Sub(latest.Date) > Recency || len(latest.Sets) < 1 || len(latest.Sets) > MaxSets {
		return Session{}, false
	}
	return latest, true
}

func reachedTarget(sessions []Session, latestSetCount, target int) bool {
	if len(sessions) < 2 || latestSetCount < 2 || len(sessions[1].Sets) < latestSetCount {
		return false
	}
	if !sessions[1].Date.Before(sessions[0].Date) {
		return false
	}
	weight := sessions[0].Sets[0].Weight
	for _, set := range sessions[1].Sets {
		if set.Weight == nil || weight == nil || *set.Weight != *weight || set.Reps < target {
			return false
		}
	}
	return true
}

func validWeight(weight *float64) bool {
	return weight != nil && !math.IsNaN(*weight) && !math.IsInf(*weight, 0) && *weight >= 0 && *weight <= 10000
}

func matchesExercise(names []string, name string) bool {
	for _, value := range names {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(name)) {
			return true
		}
	}
	return false
}
