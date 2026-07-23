package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// AgentSwaggerContentType identifies the filtered Swagger 2 discovery artifact.
const AgentSwaggerContentType = "application/vnd.oai.openapi+json;version=2.0"

type agentOperationPolicy struct {
	Method      string
	Path        string
	OperationID string
}

// agentOperationAllowlist is the complete public agent contract. Additions require
// an explicit method, path, stable operation ID, and contract test update.
var agentOperationAllowlist = []agentOperationPolicy{
	{Method: "get", Path: "/workouts", OperationID: "listWorkouts"},
	{Method: "get", Path: "/workouts/new-workout-context", OperationID: "getNewWorkoutContext"},
	{Method: "get", Path: "/workouts/focus-values", OperationID: "listWorkoutFocusValues"},
	{Method: "get", Path: "/workouts/contribution-data", OperationID: "getWorkoutContributionData"},
	{Method: "get", Path: "/exercises", OperationID: "listExercises"},
	{Method: "get", Path: "/exercises/{id}", OperationID: "getExercise"},
	{Method: "get", Path: "/exercises/{id}/metrics-history", OperationID: "getExerciseMetricsHistory"},
	{Method: "get", Path: "/features/access", OperationID: "listFeatureAccess"},
	{Method: "get", Path: "/training-profile", OperationID: "getTrainingProfile"},
}

var agentSwaggerJSON = mustBuildAgentSwaggerJSON([]byte(swaggerJSON))

// AgentSwaggerJSON returns the deterministic read-only agent contract derived
// from the canonical embedded Swagger document.
func AgentSwaggerJSON() []byte {
	return bytes.Clone(agentSwaggerJSON)
}

func mustBuildAgentSwaggerJSON(canonical []byte) []byte {
	filtered, err := buildAgentSwaggerJSON(canonical)
	if err != nil {
		panic(fmt.Sprintf("build agent Swagger document: %v", err))
	}
	return filtered
}

func buildAgentSwaggerJSON(canonical []byte) ([]byte, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &document); err != nil {
		return nil, fmt.Errorf("decode canonical Swagger document: %w", err)
	}
	if err := requireStringField(document, "swagger", "2.0"); err != nil {
		return nil, err
	}
	if err := requireStringField(document, "basePath", "/api"); err != nil {
		return nil, err
	}
	if err := validateStackAuthDefinition(document["securityDefinitions"]); err != nil {
		return nil, err
	}

	var canonicalPaths map[string]map[string]json.RawMessage
	if err := json.Unmarshal(document["paths"], &canonicalPaths); err != nil {
		return nil, fmt.Errorf("decode canonical paths: %w", err)
	}

	filteredPaths := make(map[string]map[string]json.RawMessage, len(agentOperationAllowlist))
	operationIDs := make(map[string]struct{}, len(agentOperationAllowlist))
	for _, policy := range agentOperationAllowlist {
		pathOperations, ok := canonicalPaths[policy.Path]
		if !ok {
			return nil, fmt.Errorf("allowlisted path %q is missing", policy.Path)
		}
		operation, ok := pathOperations[policy.Method]
		if !ok {
			return nil, fmt.Errorf("allowlisted operation %s %s is missing", strings.ToUpper(policy.Method), policy.Path)
		}
		if _, duplicate := operationIDs[policy.OperationID]; duplicate {
			return nil, fmt.Errorf("duplicate agent operation ID %q", policy.OperationID)
		}
		operationIDs[policy.OperationID] = struct{}{}

		var operationObject map[string]json.RawMessage
		if err := json.Unmarshal(operation, &operationObject); err != nil {
			return nil, fmt.Errorf("decode %s %s: %w", strings.ToUpper(policy.Method), policy.Path, err)
		}
		if err := validateOperationSecurity(operationObject["security"]); err != nil {
			return nil, fmt.Errorf("%s %s: %w", strings.ToUpper(policy.Method), policy.Path, err)
		}
		operationObject["operationId"] = mustMarshalRaw(policy.OperationID)
		filteredOperation, err := json.Marshal(operationObject)
		if err != nil {
			return nil, fmt.Errorf("encode %s %s: %w", strings.ToUpper(policy.Method), policy.Path, err)
		}
		if filteredPaths[policy.Path] == nil {
			filteredPaths[policy.Path] = make(map[string]json.RawMessage)
		}
		filteredPaths[policy.Path][policy.Method] = filteredOperation
	}

	encodedPaths, err := json.Marshal(filteredPaths)
	if err != nil {
		return nil, fmt.Errorf("encode filtered paths: %w", err)
	}
	document["paths"] = encodedPaths
	delete(document, "host")
	delete(document, "schemes")

	if err := pruneDefinitions(document, encodedPaths); err != nil {
		return nil, err
	}
	filtered, err := json.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("encode filtered Swagger document: %w", err)
	}
	return filtered, nil
}

func requireStringField(document map[string]json.RawMessage, field, expected string) error {
	var actual string
	if err := json.Unmarshal(document[field], &actual); err != nil || actual != expected {
		return fmt.Errorf("canonical %s must be %q", field, expected)
	}
	return nil
}

func validateStackAuthDefinition(raw json.RawMessage) error {
	var definitions map[string]json.RawMessage
	if err := json.Unmarshal(raw, &definitions); err != nil {
		return fmt.Errorf("decode security definitions: %w", err)
	}
	stackAuth, ok := definitions["StackAuth"]
	if !ok {
		return fmt.Errorf("StackAuth security definition is missing")
	}
	var definition map[string]any
	if err := json.Unmarshal(stackAuth, &definition); err != nil {
		return fmt.Errorf("decode StackAuth security definition: %w", err)
	}
	if len(definition) != 3 || definition["type"] != "apiKey" || definition["name"] != "x-stack-access-token" || definition["in"] != "header" {
		return fmt.Errorf("StackAuth security definition is not the expected header apiKey")
	}
	return nil
}

func validateOperationSecurity(raw json.RawMessage) error {
	var requirements []map[string][]string
	if err := json.Unmarshal(raw, &requirements); err != nil {
		return fmt.Errorf("decode security requirement: %w", err)
	}
	if len(requirements) != 1 || len(requirements[0]) != 1 {
		return fmt.Errorf("security must contain exactly one StackAuth requirement")
	}
	scopes, ok := requirements[0]["StackAuth"]
	if !ok || len(scopes) != 0 {
		return fmt.Errorf("security must be exactly StackAuth with no scopes")
	}
	return nil
}

func pruneDefinitions(document map[string]json.RawMessage, paths json.RawMessage) error {
	var canonicalDefinitions map[string]json.RawMessage
	if err := json.Unmarshal(document["definitions"], &canonicalDefinitions); err != nil {
		return fmt.Errorf("decode canonical definitions: %w", err)
	}

	needed := collectDefinitionReferences(paths)
	filteredDefinitions := make(map[string]json.RawMessage)
	for len(needed) > 0 {
		var name string
		for candidate := range needed {
			name = candidate
			break
		}
		delete(needed, name)
		if _, included := filteredDefinitions[name]; included {
			continue
		}
		definition, ok := canonicalDefinitions[name]
		if !ok {
			return fmt.Errorf("referenced definition %q is missing", name)
		}
		filteredDefinitions[name] = definition
		for dependency := range collectDefinitionReferences(definition) {
			if _, included := filteredDefinitions[dependency]; !included {
				needed[dependency] = struct{}{}
			}
		}
	}

	encodedDefinitions, err := json.Marshal(filteredDefinitions)
	if err != nil {
		return fmt.Errorf("encode filtered definitions: %w", err)
	}
	document["definitions"] = encodedDefinitions
	return nil
}

func collectDefinitionReferences(raw json.RawMessage) map[string]struct{} {
	const prefix = "#/definitions/"
	references := make(map[string]struct{})
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return references
	}
	var visit func(any)
	visit = func(current any) {
		switch typed := current.(type) {
		case map[string]any:
			for key, child := range typed {
				if key == "$ref" {
					if reference, ok := child.(string); ok && strings.HasPrefix(reference, prefix) {
						references[strings.TrimPrefix(reference, prefix)] = struct{}{}
					}
				}
				visit(child)
			}
		case []any:
			for _, child := range typed {
				visit(child)
			}
		}
	}
	visit(value)
	return references
}

func mustMarshalRaw(value any) json.RawMessage {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return encoded
}
