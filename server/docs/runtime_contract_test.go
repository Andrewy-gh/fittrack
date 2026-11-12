package docs_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/docs"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
)

// These fixtures use the same database values that the real handlers serialize,
// rather than constructing the documentation-only response types.
type contractExerciseRepository struct{ exercise.ExerciseRepository }

func (contractExerciseRepository) ListExercises(context.Context, string) ([]db.Exercise, error) {
	return []db.Exercise{contractExercise()}, nil
}
func (contractExerciseRepository) GetOrCreateExercise(context.Context, string, string) (db.Exercise, error) {
	return contractExercise(), nil
}
func (contractExerciseRepository) GetRecentSetsForExercise(context.Context, int32, string) ([]db.GetRecentSetsForExerciseRow, error) {
	now := pgtype.Timestamptz{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	rows := []db.GetRecentSetsForExerciseRow{}
	for range 3 {
		rows = append(rows, db.GetRecentSetsForExerciseRow{SetID: 1, WorkoutID: 2, WorkoutDate: now, CreatedAt: now, Reps: 10})
	}
	for i, value := range []string{"", "0", "12.5"} {
		if value != "" {
			if err := rows[i].Weight.Scan(value); err != nil {
				return nil, err
			}
		}
	}
	return rows, nil
}
func contractExercise() db.Exercise {
	now := pgtype.Timestamptz{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	return db.Exercise{ID: 1, Name: "Bench press", CreatedAt: now, UpdatedAt: now, UserID: "contract-user"}
}

type contractSchema struct {
	Ref        string                    `json:"$ref"`
	Type       string                    `json:"type"`
	Nullable   bool                      `json:"x-nullable"`
	Required   []string                  `json:"required"`
	Properties map[string]contractSchema `json:"properties"`
	Items      *contractSchema           `json:"items"`
}
type contractDocument struct {
	Definitions map[string]contractSchema `json:"definitions"`
	Paths       map[string]map[string]struct {
		Parameters []struct {
			In     string         `json:"in"`
			Schema contractSchema `json:"schema"`
		} `json:"parameters"`
		Responses map[string]struct {
			Schema contractSchema `json:"schema"`
		} `json:"responses"`
	} `json:"paths"`
}

func loadContract(t *testing.T) contractDocument {
	t.Helper()
	var document contractDocument
	if err := json.Unmarshal([]byte(docs.SwaggerInfo.ReadDoc()), &document); err != nil {
		t.Fatal(err)
	}
	return document
}

// Check both directions: required documented fields must be emitted, and emitted
// fields must have a documented type. This catches silently missing doc fields.
func (d contractDocument) check(schema contractSchema, value interface{}) error {
	if schema.Ref != "" {
		definition, ok := d.Definitions[strings.TrimPrefix(schema.Ref, "#/definitions/")]
		if !ok {
			return fmt.Errorf("unknown schema %s", schema.Ref)
		}
		return d.check(definition, value)
	}
	if value == nil {
		if schema.Nullable {
			return nil
		}
		return fmt.Errorf("unexpected null for %s", schema.Type)
	}
	switch schema.Type {
	case "object":
		object, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object")
		}
		for _, key := range schema.Required {
			if _, ok := object[key]; !ok {
				return fmt.Errorf("missing %s", key)
			}
		}
		for key, field := range object {
			property, ok := schema.Properties[key]
			if !ok {
				return fmt.Errorf("undocumented field %s", key)
			}
			if err := d.check(property, field); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
	case "array":
		items, ok := value.([]interface{})
		if !ok || schema.Items == nil {
			return fmt.Errorf("expected documented array")
		}
		for _, item := range items {
			if err := d.check(*schema.Items, item); err != nil {
				return err
			}
		}
	case "integer", "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean")
		}
	}
	return nil
}

func TestExerciseHandlerResponsesMatchContract(t *testing.T) {
	document := loadContract(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := exercise.NewHandler(logger, validator.New(), exercise.NewService(logger, contractExerciseRepository{}))
	for _, tc := range []struct {
		path, method, body string
		serve              http.HandlerFunc
	}{
		{"/exercises", "get", "", handler.ListExercises},
		{"/exercises", "post", `{"name":"Bench press"}`, handler.GetOrCreateExercise},
		{"/exercises/{id}/recent-sets", "get", "", handler.GetRecentSetsForExercise},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(strings.ToUpper(tc.method), "/api/exercises/1", strings.NewReader(tc.body))
			req.SetPathValue("id", "1")
			req = req.WithContext(context.WithValue(req.Context(), user.UserIDKey, "contract-user"))
			recorder := httptest.NewRecorder()
			tc.serve(recorder, req)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status %d: %s", recorder.Code, recorder.Body.String())
			}
			var value interface{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
				t.Fatal(err)
			}
			if err := document.check(document.Paths[tc.path][tc.method].Responses["200"].Schema, value); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWorkoutUpdateRequestContract(t *testing.T) {
	document := loadContract(t)
	var schema contractSchema
	for _, parameter := range document.Paths["/workouts/{id}"]["put"].Parameters {
		if parameter.In == "body" {
			schema = parameter.Schema
		}
	}
	if schema.Ref == "" {
		t.Fatal("missing workout update body schema")
	}
	for _, metadata := range []string{"", `,"notes":null,"workoutFocus":null`, `,"notes":"","workoutFocus":""`} {
		for _, weight := range []string{"null", "0", "12.5"} {
			body := `{"date":"2026-01-01T00:00:00Z","exercises":[{"name":"Bench press","sets":[{"reps":10,"setType":"working","weight":` + weight + `}]}]` + metadata + `}`
			var request workout.UpdateWorkoutRequest
			if err := json.Unmarshal([]byte(body), &request); err != nil {
				t.Fatal(err)
			}
			if err := validator.New().Struct(request); err != nil {
				t.Fatal(err)
			}
			var value interface{}
			if err := json.Unmarshal([]byte(body), &value); err != nil {
				t.Fatal(err)
			}
			if err := document.check(schema, value); err != nil {
				t.Errorf("%s: %v", body, err)
			}
		}
	}
}
