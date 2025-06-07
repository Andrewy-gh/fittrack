package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Andrewy-gh/fittrack/internal/db"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

// Response structure for the API
type WorkoutResponse struct {
	WorkoutID int32              `json:"workoutId"`
	Date      time.Time          `json:"date"`
	Notes     *string            `json:"notes,omitempty"`
	Exercises []ExerciseResponse `json:"exercises"`
}

type ExerciseResponse struct {
	ExerciseID int32         `json:"exerciseId"`
	Name       string        `json:"name"`
	Sets       []SetResponse `json:"sets"`
}

type SetResponse struct {
	SetID   int32  `json:"setId"`
	Weight  *int   `json:"weight,omitempty"`
	Reps    int    `json:"reps"`
	SetType string `json:"setType"`
}

// Global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Database service struct
type WorkoutService struct {
	queries *db.Queries
}

// saveWorkoutToDB saves the validated workout data to the database
func (ws *WorkoutService) saveWorkoutToDB(ctx context.Context, req WorkoutRequest) (*WorkoutResponse, error) {
	// Parse the date
	parsedDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Convert to pgtype.Timestamptz
	pgDate := pgtype.Timestamptz{
		Time:  parsedDate,
		Valid: true,
	}

	// Convert notes to pgtype.Text
	pgNotes := pgtype.Text{}
	if req.Notes != nil {
		pgNotes = pgtype.Text{
			String: *req.Notes,
			Valid:  true,
		}
	}

	// Create the workout
	workout, err := ws.queries.CreateWorkout(ctx, db.CreateWorkoutParams{
		Date:  pgDate,
		Notes: pgNotes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create workout: %w", err)
	}

	// Process exercises and sets
	var exerciseResponses []ExerciseResponse

	for _, exerciseInput := range req.Exercises {
		// Get or create the exercise
		exercise, err := ws.queries.GetOrCreateExercise(ctx, exerciseInput.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create exercise %s: %w", exerciseInput.Name, err)
		}

		// Create sets for this exercise
		var setResponses []SetResponse
		for _, setInput := range exerciseInput.Sets {
			// Convert weight to pgtype.Int4
			var pgWeight pgtype.Int4
			if setInput.Weight != nil {
				pgWeight = pgtype.Int4{
					Int32: int32(*setInput.Weight),
					Valid: true,
				}
			}

			// Create the set
			set, err := ws.queries.CreateSet(ctx, db.CreateSetParams{
				ExerciseID: exercise.ID,
				WorkoutID:  workout.ID,
				Weight:     pgWeight,
				Reps:       int32(setInput.Reps),
				SetType:    setInput.SetType,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create set: %w", err)
			}

			// Convert pgtype back to regular types for response
			var weight *int
			if set.Weight.Valid {
				w := int(set.Weight.Int32)
				weight = &w
			}

			setResponses = append(setResponses, SetResponse{
				SetID:   set.ID,
				Weight:  weight,
				Reps:    int(set.Reps),
				SetType: set.SetType,
			})
		}

		exerciseResponses = append(exerciseResponses, ExerciseResponse{
			ExerciseID: exercise.ID,
			Name:       exercise.Name,
			Sets:       setResponses,
		})
	}

	// Convert notes back for response
	var notes *string
	if workout.Notes.Valid {
		notes = &workout.Notes.String
	}

	return &WorkoutResponse{
		WorkoutID: workout.ID,
		Date:      workout.Date.Time,
		Notes:     notes,
		Exercises: exerciseResponses,
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
	// Database connection - replace with your actual connection string
	dbURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close(context.Background())

	// Create queries instance
	queries := db.New(conn)
	workoutService := &WorkoutService{queries: queries}

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

		// Validate the input
		if err := validate.Struct(workoutReq); err != nil {
			http.Error(w, formatValidationErrors(err), http.StatusBadRequest)
			return
		}

		// Save to database
		response, err := workoutService.saveWorkoutToDB(r.Context(), workoutReq)
		if err != nil {
			log.Printf("Database error: %v", err)
			http.Error(w, "Failed to save workout", http.StatusInternalServerError)
			return
		}

		// Convert response to JSON
		responseJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}

		log.Printf("Saved workout with ID: %d", response.WorkoutID)

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(responseJSON)
	})

	fileServer := http.FileServer(http.Dir("./dist"))
	router.Handle("/", fileServer)

	log.Println("Starting server on port 8080...")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
