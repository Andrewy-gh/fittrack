package aichateval

import (
	"fmt"
	"strings"
)

// ScenarioSelection describes a narrow subset of the default scenario pack.
type ScenarioSelection struct {
	ScenarioID  string
	ScenarioIDs string
	FromID      string
	ToID        string
}

// FilterScenarios returns scenarios matching one selection mode while preserving pack order.
func FilterScenarios(scenarios []Scenario, selection ScenarioSelection) ([]Scenario, error) {
	singleID, err := normalizedID(selection.ScenarioID, "scenario")
	if err != nil {
		return nil, err
	}
	listIDs, err := normalizedIDs(selection.ScenarioIDs)
	if err != nil {
		return nil, err
	}
	fromID, err := normalizedID(selection.FromID, "from")
	if err != nil {
		return nil, err
	}
	toID, err := normalizedID(selection.ToID, "to")
	if err != nil {
		return nil, err
	}

	hasSingle := singleID != ""
	hasList := len(listIDs) > 0
	hasRange := fromID != "" || toID != ""
	if hasSingle && hasList {
		return nil, fmt.Errorf("use only one of scenario or scenarios")
	}
	if hasRange && (hasSingle || hasList) {
		return nil, fmt.Errorf("range selection cannot be combined with scenario or scenarios")
	}
	if hasRange {
		return filterScenarioRange(scenarios, fromID, toID)
	}
	if hasSingle {
		listIDs = []string{singleID}
	}
	if len(listIDs) == 0 {
		return append([]Scenario(nil), scenarios...), nil
	}
	return filterScenarioIDs(scenarios, listIDs)
}

func filterScenarioIDs(scenarios []Scenario, ids []string) ([]Scenario, error) {
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}

	filtered := make([]Scenario, 0, len(ids))
	for _, scenario := range scenarios {
		if wanted[scenario.ID] {
			filtered = append(filtered, scenario)
			delete(wanted, scenario.ID)
		}
	}
	if len(wanted) > 0 {
		return nil, fmt.Errorf("unknown scenario id(s): %s", joinUnknownIDs(ids, wanted))
	}
	return filtered, nil
}

func filterScenarioRange(scenarios []Scenario, fromID string, toID string) ([]Scenario, error) {
	if fromID == "" || toID == "" {
		return nil, fmt.Errorf("from and to must both be provided for range selection")
	}

	start := -1
	end := -1
	for index, scenario := range scenarios {
		if scenario.ID == fromID {
			start = index
		}
		if scenario.ID == toID {
			end = index
		}
	}

	var unknown []string
	if start == -1 {
		unknown = append(unknown, fromID)
	}
	if end == -1 {
		unknown = append(unknown, toID)
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("unknown scenario id(s): %s", strings.Join(unknown, ", "))
	}
	if start > end {
		return nil, fmt.Errorf("from scenario %q must not come after to scenario %q", fromID, toID)
	}
	return append([]Scenario(nil), scenarios[start:end+1]...), nil
}

func normalizedID(id string, flagName string) (string, error) {
	trimmed := strings.TrimSpace(id)
	if id != "" && trimmed == "" {
		return "", fmt.Errorf("%s must include a scenario id", flagName)
	}
	return trimmed, nil
}

func normalizedIDs(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("scenarios must include at least one scenario id")
	}

	parts := strings.Split(raw, ",")
	ids := make([]string, 0, len(parts))
	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id == "" {
			return nil, fmt.Errorf("scenarios must not include empty scenario ids")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func joinUnknownIDs(ids []string, unknown map[string]bool) string {
	ordered := make([]string, 0, len(unknown))
	for _, id := range ids {
		if unknown[id] {
			ordered = append(ordered, id)
			delete(unknown, id)
		}
	}
	return strings.Join(ordered, ", ")
}
