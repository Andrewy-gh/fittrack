package docs

import (
	"encoding/json"
	"testing"
)

func TestSwaggerProductContract(t *testing.T) {
	var document struct {
		Host     *string                    `json:"host"`
		BasePath string                     `json:"basePath"`
		Schemes  *[]string                  `json:"schemes"`
		Paths    map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal([]byte(SwaggerInfo.ReadDoc()), &document); err != nil {
		t.Fatalf("parse registered Swagger document: %v", err)
	}

	if document.Host != nil {
		t.Errorf("host must be omitted for origin-relative clients, got %q", *document.Host)
	}
	if document.Schemes != nil {
		t.Errorf("schemes must be omitted for origin-relative clients, got %v", *document.Schemes)
	}
	if document.BasePath != "/api" {
		t.Errorf("basePath = %q, want /api", document.BasePath)
	}
	if _, ok := document.Paths["/workouts"]; !ok {
		t.Fatal("expected product path /workouts is missing")
	}

	for path, method := range map[string]string{
		"/account":                                    "delete",
		"/billing/status":                             "get",
		"/billing/checkout-session":                   "post",
		"/billing/customer-portal-session":            "post",
		"/billing/subscription-cancel-portal-session": "post",
	} {
		var operations map[string]json.RawMessage
		if err := json.Unmarshal(document.Paths[path], &operations); err != nil || operations[method] == nil {
			t.Errorf("missing product operation %s %s", method, path)
		}
	}

	excludedPaths := []string{
		"/ai/chat/validate",
		"/ai/chat/validate/stream",
		"/health",
		"/ready",
	}
	for _, path := range excludedPaths {
		if _, ok := document.Paths[path]; ok {
			t.Errorf("excluded path %q must not appear in the product Swagger document", path)
		}
	}
}
