package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

// Input structures (what the API receives)
type WorkoutRequest struct {
	Date      string          `json:"date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Notes     *string         `json:"notes,omitempty" validate:"omitempty,max=256"`
	Exercises []ExerciseInput `json:"exercises" validate:"required,min=1,dive"`
}

type ExerciseInput struct {
	Name string     `json:"name" validate:"required,min=1,max=256"`
	Sets []SetInput `json:"sets" validate:"required,min=1,dive"`
}

type SetInput struct {
	Weight  *int   `json:"weight,omitempty" validate:"omitempty,gte=0"`
	Reps    int    `json:"reps" validate:"required,gte=1"`
	SetType string `json:"setType" validate:"required,min=1,max=256"`
}

// Output structures (what matches your database schema)
type WorkoutResponse struct {
	Workout   WorkoutData    `json:"workout"`
	Exercises []ExerciseData `json:"exercises"`
	Sets      []SetData      `json:"sets"`
}

type WorkoutData struct {
	Date  time.Time `json:"date"`
	Notes *string   `json:"notes,omitempty"`
}

type ExerciseData struct {
	Name string `json:"name"`
}

type SetData struct {
	ExerciseName string `json:"exerciseName"` // Reference to exercise
	Weight       *int   `json:"weight,omitempty"`
	Reps         int    `json:"reps"`
	SetType      string `json:"setType"`
}

// Global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// validateAndTransform validates the input and transforms it to match the database schema
func validateAndTransform(req WorkoutRequest) (*WorkoutResponse, error) {
	// Validate the input
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Parse the date
	parsedDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Create workout data
	workout := WorkoutData{
		Date:  parsedDate,
		Notes: req.Notes,
	}

	// Collect unique exercises and all sets
	exerciseMap := make(map[string]bool)
	var exercises []ExerciseData
	var sets []SetData

	for _, exercise := range req.Exercises {
		// Add exercise if not already seen
		if !exerciseMap[exercise.Name] {
			exerciseMap[exercise.Name] = true
			exercises = append(exercises, ExerciseData{
				Name: exercise.Name,
			})
		}

		// Add all sets for this exercise
		for _, set := range exercise.Sets {
			sets = append(sets, SetData{
				ExerciseName: exercise.Name,
				Weight:       set.Weight,
				Reps:         set.Reps,
				SetType:      set.SetType,
			})
		}
	}

	return &WorkoutResponse{
		Workout:   workout,
		Exercises: exercises,
		Sets:      sets,
	}, nil
}

// formatValidationErrors formats validation errors in a user-friendly way
func formatValidationErrors(err error) string {
	if validationErrors, ok := err.(*validator.ValidationErrors); ok {
		var messages []string
		for _, fieldError := range *validationErrors {
			switch fieldError.Tag() {
			case "required":
				messages = append(messages, fmt.Sprintf("%s is required", fieldError.Field()))
			case "min":
				messages = append(messages, fmt.Sprintf("%s must be at least %s characters", fieldError.Field(), fieldError.Param()))
			case "max":
				messages = append(messages, fmt.Sprintf("%s must be at most %s characters", fieldError.Field(), fieldError.Param()))
			case "gte":
				messages = append(messages, fmt.Sprintf("%s must be greater than or equal to %s", fieldError.Field(), fieldError.Param()))
			case "datetime":
				messages = append(messages, fmt.Sprintf("%s must be a valid datetime in RFC3339 format", fieldError.Field()))
			default:
				messages = append(messages, fmt.Sprintf("%s failed validation (%s)", fieldError.Field(), fieldError.Tag()))
			}
		}
		return fmt.Sprintf("Validation errors: %v", messages)
	}
	return err.Error()
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("POST /api/workouts", func(w http.ResponseWriter, r *http.Request) {
		// Read the raw JSON body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		log.Printf("Received workout JSON: %s\n", string(body))

		// Parse JSON into our request structure
		var workoutReq WorkoutRequest
		if err := json.Unmarshal(body, &workoutReq); err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Validate and transform the data
		response, err := validateAndTransform(workoutReq)
		if err != nil {
			http.Error(w, formatValidationErrors(err), http.StatusBadRequest)
			return
		}

		// Convert response to JSON
		responseJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}

		log.Printf("Transformed workout data: %s\n", string(responseJSON))

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	fileServer := http.FileServer(http.Dir("./dist"))
	router.Handle("/", fileServer)

	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
