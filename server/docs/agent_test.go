package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestAgentSwaggerJSONHasExactReadOnlyContract(t *testing.T) {
	first := AgentSwaggerJSON()
	second := AgentSwaggerJSON()
	if !bytes.Equal(first, second) {
		t.Fatal("agent Swagger output is not deterministic")
	}

	var document map[string]json.RawMessage
	if err := json.Unmarshal(first, &document); err != nil {
		t.Fatalf("decode agent Swagger: %v", err)
	}
	if _, present := document["host"]; present {
		t.Fatal("agent Swagger must not contain host")
	}
	if _, present := document["schemes"]; present {
		t.Fatal("agent Swagger must not contain schemes")
	}
	if err := requireStringField(document, "swagger", "2.0"); err != nil {
		t.Fatal(err)
	}
	if err := requireStringField(document, "basePath", "/api"); err != nil {
		t.Fatal(err)
	}

	var paths map[string]map[string]map[string]json.RawMessage
	if err := json.Unmarshal(document["paths"], &paths); err != nil {
		t.Fatalf("decode agent paths: %v", err)
	}
	if len(paths) != len(agentOperationAllowlist) {
		t.Fatalf("path count = %d, want %d; paths = %v", len(paths), len(agentOperationAllowlist), paths)
	}

	operationIDs := make(map[string]struct{}, len(agentOperationAllowlist))
	for _, policy := range agentOperationAllowlist {
		methods, ok := paths[policy.Path]
		if !ok {
			t.Errorf("missing path %s", policy.Path)
			continue
		}
		if len(methods) != 1 {
			t.Errorf("%s methods = %v, want only GET", policy.Path, methods)
			continue
		}
		operation, ok := methods[policy.Method]
		if !ok {
			t.Errorf("missing %s %s", strings.ToUpper(policy.Method), policy.Path)
			continue
		}
		var operationID string
		if err := json.Unmarshal(operation["operationId"], &operationID); err != nil {
			t.Errorf("decode operationId for %s: %v", policy.Path, err)
		} else if operationID != policy.OperationID {
			t.Errorf("operationId for %s = %q, want %q", policy.Path, operationID, policy.OperationID)
		}
		if _, duplicate := operationIDs[operationID]; duplicate {
			t.Errorf("duplicate operationId %q", operationID)
		}
		operationIDs[operationID] = struct{}{}
		if err := validateOperationSecurity(operation["security"]); err != nil {
			t.Errorf("%s %s: %v", strings.ToUpper(policy.Method), policy.Path, err)
		}
	}

	var definitions map[string]json.RawMessage
	if err := json.Unmarshal(document["definitions"], &definitions); err != nil {
		t.Fatalf("decode definitions: %v", err)
	}
	referenced := collectDefinitionReferences(document["paths"])
	for {
		before := len(referenced)
		for name := range referenced {
			for dependency := range collectDefinitionReferences(definitions[name]) {
				referenced[dependency] = struct{}{}
			}
		}
		if len(referenced) == before {
			break
		}
	}
	if len(definitions) != len(referenced) {
		t.Fatalf("retained definitions = %d, transitively referenced definitions = %d", len(definitions), len(referenced))
	}
	for name := range definitions {
		if _, ok := referenced[name]; !ok {
			t.Errorf("unreferenced definition leaked: %s", name)
		}
	}
}

func TestBuildAgentSwaggerJSONRejectsMissingAllowlistedOperation(t *testing.T) {
	canonical := decodeCanonicalForMutation(t)
	var paths map[string]map[string]json.RawMessage
	if err := json.Unmarshal(canonical["paths"], &paths); err != nil {
		t.Fatal(err)
	}
	delete(paths["/training-profile"], "get")
	canonical["paths"] = marshalForTest(t, paths)

	_, err := buildAgentSwaggerJSON(marshalForTest(t, canonical))
	if err == nil || !strings.Contains(err.Error(), "GET /training-profile is missing") {
		t.Fatalf("error = %v, want missing allowlisted operation", err)
	}
}

func TestBuildAgentSwaggerJSONRejectsWrongAllowlistedSecurity(t *testing.T) {
	tests := []struct {
		name     string
		security any
	}{
		{name: "absent", security: nil},
		{name: "empty", security: []any{}},
		{name: "wrong scheme", security: []map[string][]string{{"OtherAuth": {}}}},
		{name: "unexpected scope", security: []map[string][]string{{"StackAuth": {"read"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canonical := decodeCanonicalForMutation(t)
			var paths map[string]map[string]map[string]json.RawMessage
			if err := json.Unmarshal(canonical["paths"], &paths); err != nil {
				t.Fatal(err)
			}
			operation := paths["/workouts"]["get"]
			if tt.security == nil {
				delete(operation, "security")
			} else {
				operation["security"] = marshalForTest(t, tt.security)
			}
			canonical["paths"] = marshalForTest(t, paths)

			_, err := buildAgentSwaggerJSON(marshalForTest(t, canonical))
			if err == nil || !strings.Contains(err.Error(), "GET /workouts:") || !strings.Contains(err.Error(), "security") {
				t.Fatalf("error = %v, want exact StackAuth rejection", err)
			}
		})
	}
}

func TestBuildAgentSwaggerJSONStripsCanonicalAuthority(t *testing.T) {
	canonical := decodeCanonicalForMutation(t)
	canonical["host"] = marshalForTest(t, "fixed.example")
	canonical["schemes"] = marshalForTest(t, []string{"https"})

	filtered, err := buildAgentSwaggerJSON(marshalForTest(t, canonical))
	if err != nil {
		t.Fatalf("build filtered contract: %v", err)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(filtered, &document); err != nil {
		t.Fatal(err)
	}
	if _, present := document["host"]; present {
		t.Fatal("host leaked from canonical document")
	}
	if _, present := document["schemes"]; present {
		t.Fatal("schemes leaked from canonical document")
	}
}

func decodeCanonicalForMutation(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	var canonical map[string]json.RawMessage
	if err := json.Unmarshal([]byte(swaggerJSON), &canonical); err != nil {
		t.Fatalf("decode canonical Swagger: %v", err)
	}
	return canonical
}

func marshalForTest(t *testing.T, value any) json.RawMessage {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal test value: %v", err)
	}
	return encoded
}
