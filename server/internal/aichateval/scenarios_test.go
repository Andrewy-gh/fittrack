package aichateval

import "testing"

func TestDefaultScenariosIncludesMachineChestHypertrophyFollowUp(t *testing.T) {
	scenario, ok := findDefaultScenario("prompt-19")
	if !ok {
		t.Fatal("DefaultScenarios() missing prompt-19")
	}

	if scenario.ExpectedOutcome != ExpectedAskOnceThenGenerate {
		t.Fatalf("prompt-19 expected outcome = %q, want %q", scenario.ExpectedOutcome, ExpectedAskOnceThenGenerate)
	}
	if scenario.FollowUpAnswer != "Nope." {
		t.Fatalf("prompt-19 follow-up answer = %q, want observed short no-injuries answer", scenario.FollowUpAnswer)
	}
	if scenario.Prompt != "Hello I would like a chest workout for today. I have about an hour and a half with warm up included. Only machines and cables. Focusing on hypertrophy." {
		t.Fatalf("prompt-19 prompt = %q, want observed machine chest prompt", scenario.Prompt)
	}
}

func findDefaultScenario(id string) (Scenario, bool) {
	for _, scenario := range DefaultScenarios() {
		if scenario.ID == id {
			return scenario, true
		}
	}
	return Scenario{}, false
}
