package aichateval

import (
	"fmt"
	"strings"
)

func DraftWeightScenarios() []Scenario {
	context := " I have 35 minutes for an upper-body workout at a full gym, no injuries, and no FitTrack history."
	weight := 115.0
	zero := 0.0
	return []Scenario{
		{ID: "weight-ordinary", Title: "Draft without weight history", Prompt: "I'm an intermediate lifter training for strength." + context, ExpectedOutcome: ExpectedGenerateFirstTurn, DraftWorkingWeights: map[string]*float64{}},
		{ID: "weight-cap", Title: "Equipment cap is not a load", Prompt: "Make a 35-minute upper-body workout at home with a bench and dumbbells up to 35 lb each, no injuries. I am a beginner and have no FitTrack history.", ExpectedOutcome: ExpectedGenerateFirstTurn, DraftWorkingWeights: map[string]*float64{}},
		{ID: "weight-reported", Title: "Reported load applies to the named exercise", Prompt: "I comfortably did Barbell Bench Press for 3 sets of 8 at 115 lb last week. Include Barbell Bench Press for sets of 8 and include other exercises." + context, ExpectedOutcome: ExpectedGenerateFirstTurn, DraftWorkingWeights: map[string]*float64{"Barbell Bench Press": &weight}},
		{ID: "weight-explicit", Title: "Explicit load, blank, and zero", Prompt: "Include Barbell Bench Press at 115 lb for the working sets. Leave Dumbbell Row weights blank and use no added weight for Pull-Ups. Include all three." + context, ExpectedOutcome: ExpectedGenerateFirstTurn, DraftWorkingWeights: map[string]*float64{"Barbell Bench Press": &weight, "Dumbbell Row": nil, "Pull-Ups": &zero}},
		{ID: "weight-unanchored", Title: "Starting weight request without a baseline", Prompt: "I am a beginner. Pick starting weights for me; I have never tracked what I lift." + context, ExpectedOutcome: ExpectedGenerateFirstTurn, DraftWorkingWeights: map[string]*float64{}},
	}
}

func passesDraftWorkingWeights(result *Result) bool {
	expected := result.DraftWorkingWeights
	if expected == nil {
		return true
	}
	if result.Draft == nil {
		setScore(result, ScoreStatusFail, "expected a draft with sets to check for weights")
		return false
	}

	seen := make(map[string]bool)
	for _, exercise := range result.Draft.Exercises {
		key := strings.TrimSpace(strings.ToLower(exercise.Name))
		var want *float64
		for name, weight := range expected {
			if strings.ToLower(name) == key {
				want = weight
				break
			}
		}
		for _, set := range exercise.Sets {
			if set.SetType == "working" {
				seen[key] = true
			}
			if set.SetType != "working" || want == nil {
				if set.Weight == nil || (set.SetType != "working" && want != nil && *want == 0 && *set.Weight == 0) {
					continue
				}
				setScore(result, ScoreStatusFail, fmt.Sprintf("unexpected weight on %s %s set", exercise.Name, set.SetType))
				return false
			}
			if set.Weight == nil || *set.Weight != *want {
				setScore(result, ScoreStatusFail, fmt.Sprintf("expected %g lb on %s working sets", *want, exercise.Name))
				return false
			}
		}
	}
	for name := range expected {
		if !seen[strings.ToLower(name)] {
			setScore(result, ScoreStatusFail, fmt.Sprintf("missing %s working sets", name))
			return false
		}
	}
	return true
}
