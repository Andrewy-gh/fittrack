package aichateval

import (
	"strings"
	"testing"
)

func TestFilterScenariosCommaSeparatedIDsPreservePackOrder(t *testing.T) {
	got, err := FilterScenarios(selectionTestScenarios(), ScenarioSelection{ScenarioIDs: "prompt-04,prompt-02"})

	if err != nil {
		t.Fatalf("FilterScenarios() error = %v, want nil", err)
	}
	if gotIDs(got) != "prompt-02,prompt-04" {
		t.Fatalf("FilterScenarios() ids = %s, want default-pack order", gotIDs(got))
	}
}

func TestFilterScenariosUnknownIDFails(t *testing.T) {
	_, err := FilterScenarios(selectionTestScenarios(), ScenarioSelection{ScenarioIDs: "prompt-02,prompt-99"})

	if err == nil {
		t.Fatal("FilterScenarios() error = nil, want unknown id error")
	}
	if !strings.Contains(err.Error(), "unknown scenario id(s): prompt-99") {
		t.Fatalf("FilterScenarios() error = %q, want unknown id", err.Error())
	}
}

func TestFilterScenariosRangeReturnsInclusiveScenarios(t *testing.T) {
	got, err := FilterScenarios(selectionTestScenarios(), ScenarioSelection{FromID: "prompt-02", ToID: "prompt-04"})

	if err != nil {
		t.Fatalf("FilterScenarios() error = %v, want nil", err)
	}
	if gotIDs(got) != "prompt-02,prompt-03,prompt-04" {
		t.Fatalf("FilterScenarios() ids = %s, want inclusive range", gotIDs(got))
	}
}

func TestFilterBaseOnlyScenariosExcludesAskFirstDefaultScenariosFromFixturesMode(t *testing.T) {
	scenarios := append(DefaultScenarios(), DataFixtureScenarios()...)
	got := FilterBaseOnlyScenarios(scenarios)
	gotByID := make(map[string]Scenario, len(got))
	for _, scenario := range got {
		gotByID[scenario.ID] = scenario
	}

	if _, ok := gotByID["data-01"]; !ok {
		t.Fatal("fixtures-mode scenarios must retain non-base-only data-01")
	}

	for _, scenario := range DefaultScenarios() {
		if scenario.ExpectedOutcome != ExpectedAskOnceThenGenerate {
			continue
		}
		if !scenario.BaseOnly {
			t.Fatalf("%s has ask-first outcome but BaseOnly = false", scenario.ID)
		}
		if _, ok := gotByID[scenario.ID]; ok {
			t.Fatalf("fixtures-mode scenarios include ask-first default scenario %s", scenario.ID)
		}
	}
}

func selectionTestScenarios() []Scenario {
	return []Scenario{
		{ID: "prompt-01", Title: "One"},
		{ID: "prompt-02", Title: "Two"},
		{ID: "prompt-03", Title: "Three"},
		{ID: "prompt-04", Title: "Four"},
	}
}

func gotIDs(scenarios []Scenario) string {
	ids := make([]string, 0, len(scenarios))
	for _, scenario := range scenarios {
		ids = append(ids, scenario.ID)
	}
	return strings.Join(ids, ",")
}
