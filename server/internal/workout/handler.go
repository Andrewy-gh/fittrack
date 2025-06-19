package workout

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
)

type WorkoutHandler struct {
	workoutService *WorkoutService
}

func NewHandler(workoutService *WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		workoutService: workoutService,
	}
}

func (h *WorkoutHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts, err := h.workoutService.ListWorkouts(r.Context())
	if err != nil {
		h.workoutService.logger.Error("failed to list workouts", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(workouts)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func (h *WorkoutHandler) GetWorkoutWithSets(w http.ResponseWriter, r *http.Request) {
	workoutID := r.PathValue("id")
	if workoutID == "" {
		http.Error(w, "Missing workout ID", http.StatusBadRequest)
		return
	}

	workoutIDInt, err := strconv.ParseInt(workoutID, 10, 32)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}

	workoutWithSets, err := h.workoutService.GetWorkoutWithSets(r.Context(), int32(workoutIDInt))
	if err != nil {
		h.workoutService.logger.Error("failed to get workout with sets", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(workoutWithSets)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// func printValidationErrors(err error) {
// 	fmt.Println("\n=== VALIDATION ERRORS ===")
// 	if errs, ok := err.(validator.ValidationErrors); ok {
// 		for _, e := range errs {
// 			fmt.Printf("Field: %s\n", e.Namespace())
// 			fmt.Printf("Tag: %s\n", e.Tag())
// 			fmt.Printf("Type: %v\n", e.Type())
// 			fmt.Printf("Value: %v\n", e.Value())
// 			fmt.Printf("Param: %s\n\n", e.Param())
// 		}
// 	} else {
// 		fmt.Printf("Non-validation error: %v\n", err)
// 	}
// 	fmt.Println("========================\n")
// }

func FormatValidationErrors(err error) string {
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

func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	var request CreateWorkoutRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		fmt.Printf("JSON decode error: %v\n", err)
		return
	}
	fmt.Println("Decoded JSON:", request)

	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		fmt.Println("Validation error occurred")
		http.Error(w, FormatValidationErrors(err), http.StatusBadRequest)
		return
	}

	// Call service with validated struct
	if err := h.workoutService.CreateWorkout(r.Context(), request); err != nil {
		// You can add error type checking here for different HTTP status codes
		http.Error(w, "Failed to create workout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
// 	// 1. Read the body
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "Failed to read body", http.StatusBadRequest)
// 		return
// 	}
// 	defer r.Body.Close()

// 	// 2. Print the raw body
// 	fmt.Println("Raw body:", string(body))

// 	// 3. Parse the body as JSON (into a generic map)
// 	var parsedBody map[string]interface{}
// 	if err := json.Unmarshal(body, &parsedBody); err != nil {
// 		http.Error(w, "Invalid JSON", http.StatusBadRequest)
// 		return
// 	}

// 	// 4. Print the parsed JSON
// 	fmt.Printf("Parsed JSON: %+v\n", parsedBody)

// 	// 5. Respond with {"success": true}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]bool{"success": true})
// }
