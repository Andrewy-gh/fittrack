// Package recommendation owns the versioned working-set heuristic. It does not prescribe load or reps.
package recommendation

import (
	"fmt"
	"math"
	"time"
)

const PolicyVersion = "working-sets-v1"
const MaxSets = 20
const Recency = 28 * 24 * time.Hour

var ErrInvalidContext = fmt.Errorf("invalid recommendation context")

type Range struct {
	Min int `json:"min" validate:"required"`
	Max int `json:"max" validate:"required"`
}

func (r Range) Valid() bool { return r.Min >= 1 && r.Max >= r.Min && r.Max <= MaxSets }

type Session struct {
	WorkoutID   int32     `json:"workoutId" validate:"required"`
	Date        time.Time `json:"date" validate:"required"`
	WorkingSets int       `json:"workingSets" validate:"required"`
}

type Result struct {
	Plan          *Plan    `json:"plan,omitempty"`
	ExerciseID    int32    `json:"exerciseId" validate:"required"`
	Readiness     string   `json:"readiness" validate:"required"`
	Range         *Range   `json:"range" extensions:"x-nullable"`
	Baseline      *Range   `json:"baseline" extensions:"x-nullable"`
	Source        string   `json:"source" validate:"required"`
	Previous      *Session `json:"previous" extensions:"x-nullable"`
	Explanation   string   `json:"explanation" validate:"required"`
	PolicyVersion string   `json:"policyVersion" validate:"required"`
}

type Snapshot struct {
	ExerciseName   string `json:"exerciseName" validate:"required"`
	Recommendation Result `json:"recommendation" validate:"required"`
	Feedback       string `json:"feedback,omitempty"`
}

// Recommend uses only an explicit prescription or the latest eligible exact-exercise session.
func Recommend(now time.Time, prescription *Range, previous *Session, readiness string) Result {
	result := Result{Readiness: readiness, Previous: previous, Source: "none", PolicyVersion: PolicyVersion}
	if readiness != "normal" && readiness != "great" && readiness != "sluggish" {
		result.Explanation = "Choose how you feel today to get a suggestion."
		return result
	}
	if prescription != nil {
		if !prescription.Valid() {
			result.Explanation = "Your saved range must be between 1 and 20 working sets."
			return result
		}
		baseline := *prescription
		result.Baseline, result.Source = &baseline, "prescription"
	} else if previous != nil && !previous.Date.After(now) && now.Sub(previous.Date) <= Recency && previous.WorkingSets >= 1 && previous.WorkingSets <= MaxSets {
		result.Baseline = &Range{Min: previous.WorkingSets, Max: previous.WorkingSets}
		result.Source = "history"
	}
	if result.Baseline == nil {
		result.Explanation = "We don’t have recent sets we can use. Save your own range or log sets without a suggestion."
		return result
	}
	proposed := *result.Baseline
	result.Explanation = "Stick with the same number of sets today."
	if readiness == "sluggish" {
		proposed.Min = max(1, proposed.Min-1)
		proposed.Max = max(1, proposed.Max-1)
		result.Explanation = "It’s okay to do less today, or rest if you need to."
	}
	result.Range = &proposed
	return result
}

// ValidateSnapshot validates the recorded display without consulting today's prescription or history.
func ValidateSnapshot(s Snapshot) error {
	r := s.Recommendation
	if s.ExerciseName == "" || r.ExerciseID <= 0 || (r.PolicyVersion != PolicyVersion && r.PolicyVersion != "exercise-plan-v2") {
		return fmt.Errorf("invalid recommendation identity or policy")
	}
	if p := r.Plan; p != nil {
		if r.PolicyVersion != "exercise-plan-v2" || p.Sets < 1 || p.Sets > MaxSets || p.Reps < 1 || p.Reps > 100 || len(p.Goal) > 64 || len(p.Experience) > 64 {
			return fmt.Errorf("invalid plan")
		}
		if p.Weight != nil && (math.IsNaN(*p.Weight) || math.IsInf(*p.Weight, 0) || *p.Weight < 0 || *p.Weight > 10000) {
			return fmt.Errorf("invalid weight")
		}
		if r.Range != nil && p.Sets != r.Range.Min {
			return fmt.Errorf("plan does not match suggested set count")
		}
	}
	if s.Feedback != "" && s.Feedback != "too_little" && s.Feedback != "about_right" && s.Feedback != "too_much" {
		return fmt.Errorf("invalid exercise feedback")
	}
	if r.Readiness != "normal" && r.Readiness != "great" && r.Readiness != "sluggish" {
		return fmt.Errorf("invalid readiness")
	}
	if len(r.Explanation) > 512 {
		return fmt.Errorf("explanation too long")
	}
	if r.Source == "none" {
		if r.Range != nil || r.Baseline != nil {
			return fmt.Errorf("fallback cannot contain a range")
		}
		return nil
	}
	if r.Source != "prescription" && r.Source != "history" {
		return fmt.Errorf("invalid baseline source")
	}
	if r.Baseline == nil || !r.Baseline.Valid() || r.Range == nil || !r.Range.Valid() {
		return fmt.Errorf("invalid working-set range")
	}
	expected := *r.Baseline
	if r.Readiness == "sluggish" {
		expected.Min = max(1, expected.Min-1)
		expected.Max = max(1, expected.Max-1)
	}
	if expected != *r.Range {
		return fmt.Errorf("range does not match recorded policy")
	}
	if r.Source == "history" && (r.Previous == nil || r.Previous.WorkoutID <= 0 || r.Previous.WorkingSets != r.Baseline.Min || r.Baseline.Min != r.Baseline.Max) {
		return fmt.Errorf("missing history reference")
	}
	return nil
}
