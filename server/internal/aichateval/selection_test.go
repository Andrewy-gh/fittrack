package aichateval

import (
	"strings"
	"testing"
)

func TestFilterScenariosNoSelectionReturnsAll(t *testing.T) {
	scenarios := selectionTestScenarios()

	got, err := FilterScenarios(scenarios, ScenarioSelection{})

	if err != nil {
		t.Fatalf("FilterScenarios() error = %v, want nil", err)
	}
	if gotIDs(got) != "prompt-01,prompt-02,prompt-03,prompt-04" {
		t.Fatalf("FilterScenarios() ids = %s, want all ids", gotIDs(got))
	}
}

func TestFilterScenariosSingleIDReturnsOne(t *testing.T) {
	got, err := FilterScenarios(selectionTestScenarios(), ScenarioSelection{ScenarioID: "prompt-03"})

	if err != nil {
		t.Fatalf("FilterScenarios() error = %v, want nil", err)
	}
	if gotIDs(got) != "prompt-03" {
		t.Fatalf("FilterScenarios() ids = %s, want prompt-03", gotIDs(got))
	}
}

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

func TestFilterScenariosRejectsAmbiguousSelection(t *testing.T) {
	_, err := FilterScenarios(selectionTestScenarios(), ScenarioSelection{
		ScenarioID:  "prompt-01",
		ScenarioIDs: "prompt-02",
	})

	if err == nil {
		t.Fatal("FilterScenarios() error = nil, want ambiguous selection error")
	}
	if !strings.Contains(err.Error(), "use only one of scenario or scenarios") {
		t.Fatalf("FilterScenarios() error = %q, want ambiguous selection", err.Error())
	}
}

func TestFilterScenariosRejectsRangeCombinedWithIDs(t *testing.T) {
	_, err := FilterScenarios(selectionTestScenarios(), ScenarioSelection{
		ScenarioIDs: "prompt-01,prompt-02",
		FromID:      "prompt-03",
		ToID:        "prompt-04",
	})

	if err == nil {
		t.Fatal("FilterScenarios() error = nil, want range combination error")
	}
	if !strings.Contains(err.Error(), "range selection cannot be combined") {
		t.Fatalf("FilterScenarios() error = %q, want range combination error", err.Error())
	}
}

func TestFilterBaseOnlyScenariosExcludesBaseOnlyScenarios(t *testing.T) {
	scenarios := []Scenario{
		{ID: "prompt-19", Title: "Before"},
		{ID: "prompt-20", Title: "Base Only", BaseOnly: true},
		{ID: "profile-01", Title: "Fixture"},
	}

	got := FilterBaseOnlyScenarios(scenarios)

	if gotIDs(got) != "prompt-19,profile-01" {
		t.Fatalf("FilterBaseOnlyScenarios() ids = %s, want prompt-19,profile-01", gotIDs(got))
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
