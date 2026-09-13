package recommendation

import (
	"math"
	"strings"
	"time"
)

// TrainingContext contains only the owner's exercise history and relevant profile facts.
type TrainingContext struct {
	Baseline       *Range
	Previous       *Session
	Sessions       [][]WorkingSet
	Goal           string
	Experience     string
	Avoided        bool
	HasLimitations bool
}

type WorkingSet struct {
	Weight *float64
	Reps   int
}

// Plan is a starting point, separate from the actual sets recorded by the user.
type Plan struct {
	Sets       int      `json:"sets" validate:"required"`
	Reps       int      `json:"reps" validate:"required"`
	Weight     *float64 `json:"weight" extensions:"x-nullable"`
	Goal       string   `json:"goal"`
	Experience string   `json:"experience"`
}

func RecommendTraining(now time.Time, context TrainingContext, readiness string) Result {
	r := Recommend(now, context.Baseline, context.Previous, readiness)
	r.PolicyVersion = "exercise-plan-v2"
	if readiness != "normal" && readiness != "great" && readiness != "sluggish" {
		return r
	}
	if context.Avoided || context.HasLimitations {
		r.Range = nil
		r.Baseline = nil
		r.Source = "none"
		r.Explanation = "Your profile lists an exercise to avoid or a movement limitation. Choose a suitable exercise or review your profile first."
		return r
	}
	low, high := 8, 12
	if context.Goal == "strength" && (context.Experience == "intermediate" || context.Experience == "advanced") {
		low, high = 5, 8
	}
	if context.Goal == "endurance" {
		low, high = 12, 15
	}
	p := &Plan{Sets: 2, Reps: low, Goal: context.Goal, Experience: context.Experience}
	if context.Experience == "intermediate" || context.Experience == "advanced" {
		p.Sets = 3
	}
	if r.Range != nil {
		p.Sets = r.Range.Min
	}
	if readiness == "sluggish" {
		// A saved range was already adjusted by Recommend; only reduce default sets here.
		if r.Range == nil {
			p.Sets = max(1, p.Sets-1)
		}
		p.Reps = max(1, p.Reps-2)
	}
	r.Plan = p
	r.Explanation = "Choose a comfortable starting weight; we don’t have recent working sets to use."
	previous := context.Previous
	if previous == nil || previous.Date.After(now) || now.Sub(previous.Date) > Recency || previous.WorkingSets < 1 || previous.WorkingSets > MaxSets || len(context.Sessions) == 0 || len(context.Sessions[0]) == 0 {
		return r
	}
	sets := context.Sessions[0]
	reps := sets[0].Reps
	weight := sets[0].Weight
	uniform := weight != nil && !math.IsNaN(*weight) && !math.IsInf(*weight, 0) && *weight >= 0 && *weight <= 10000
	for _, set := range sets {
		if set.Reps < 1 || set.Reps > 100 {
			r.Plan = nil
			r.Explanation = "Log a working set with valid reps to get a suggestion."
			return r
		}
		reps = min(reps, set.Reps)
		if set.Weight == nil || weight == nil || *set.Weight != *weight {
			uniform = false
		}
	}
	p.Reps = reps
	if uniform {
		value := *weight
		p.Weight = &value
	}
	r.Explanation = "Repeat your last working sets."
	if r.Source == "prescription" {
		r.Explanation = "Using your saved set count and last weight and reps."
	}
	if !uniform {
		r.Explanation = "Your last weights varied or weren’t recorded. Choose the weight for these reps."
	}
	if readiness == "sluggish" {
		p.Reps = max(1, reps-2)
		r.Explanation = "An easier day: fewer reps and, where possible, one fewer set."
		return r
	}
	// Require two distinct recent sessions at one load before suggesting a small increase.
	progress := uniform && len(context.Sessions) > 1 && len(sets) >= 2 && len(context.Sessions[1]) >= len(sets) && reps >= high && p.Sets <= len(sets)
	if progress {
		for _, set := range context.Sessions[1] {
			if set.Weight == nil || *set.Weight != *weight || set.Reps < high {
				progress = false
			}
		}
	}
	if progress && *weight >= 50 && *weight <= 2000 {
		value := math.Round((*weight+2.5)*10) / 10
		p.Weight, p.Reps = &value, low
		r.Explanation = "You reached the rep target in two workouts. Try a small weight increase, using the closest available weight."
	} else if readiness == "great" && uniform && reps < high && p.Sets <= len(sets) {
		p.Reps = reps + 1
		r.Explanation = "Try one extra rep at the same weight."
	} else if readiness == "great" {
		r.Explanation += " There isn’t enough evidence to increase it yet."
	}
	return r
}

func matchesExercise(names []string, name string) bool {
	for _, value := range names {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(name)) {
			return true
		}
	}
	return false
}
