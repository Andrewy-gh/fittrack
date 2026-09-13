// Package recommendation owns the versioned exercise-plan heuristic.
package recommendation

import (
	"fmt"
	"math"
)

const PolicyVersion = "exercise-plan-v2"
const MaxSets = 20

var ErrInvalidContext = fmt.Errorf("invalid recommendation context")

type Result struct {
	Plan          *Plan  `json:"plan,omitempty"`
	ExerciseID    int32  `json:"exerciseId" validate:"required"`
	Readiness     string `json:"readiness" validate:"required"`
	Explanation   string `json:"explanation" validate:"required"`
	PolicyVersion string `json:"policyVersion" validate:"required"`
}

type Snapshot struct {
	ExerciseName   string `json:"exerciseName" validate:"required"`
	Recommendation Result `json:"recommendation" validate:"required"`
}

// ValidateSnapshot checks the recommendation displayed by the client without
// recalculating it from current profile or workout history.
func ValidateSnapshot(snapshot Snapshot) error {
	result := snapshot.Recommendation
	if snapshot.ExerciseName == "" || result.ExerciseID <= 0 || result.PolicyVersion != PolicyVersion {
		return fmt.Errorf("invalid recommendation identity or policy")
	}
	if !validReadiness(result.Readiness) {
		return fmt.Errorf("invalid readiness")
	}
	if len(result.Explanation) > 512 {
		return fmt.Errorf("explanation too long")
	}
	if result.Plan == nil {
		return nil
	}
	plan := result.Plan
	if plan.Sets < 1 || plan.Sets > MaxSets || plan.Reps < 1 || plan.Reps > 100 {
		return fmt.Errorf("invalid plan")
	}
	if plan.Weight != nil && (math.IsNaN(*plan.Weight) || math.IsInf(*plan.Weight, 0) || *plan.Weight < 0 || *plan.Weight > 10000) {
		return fmt.Errorf("invalid weight")
	}
	return nil
}

func validReadiness(readiness string) bool {
	return readiness == "normal" || readiness == "great" || readiness == "sluggish"
}
