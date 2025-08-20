# Object-Based API Response Migration Guide

## Overview

This document explains how to migrate from "array-of-rows" API responses to semantically clearer "object-based" responses in our REST API.

## The Problem: Array-of-Rows Pattern

### Current Implementation Issues

Our current `GET /exercises/{id}` endpoint returns an array of rows directly:

```json
// Current response from GET /exercises/123
[
  {
    "workout_id": 1,
    "workout_date": "2023-01-01T10:00:00Z",
    "workout_notes": "Morning workout",
    "set_id": 1,
    "weight": 225,
    "reps": 10,
    "set_type": "working",
    "exercise_id": 123,
    "exercise_name": "Bench Press",
    "volume": 2250
  },
  {
    "workout_id": 1,
    "workout_date": "2023-01-01T10:00:00Z",
    "workout_notes": "Morning workout", 
    "set_id": 2,
    "weight": 225,
    "reps": 8,
    "set_type": "working",
    "exercise_id": 123,
    "exercise_name": "Bench Press",
    "volume": 1800
  }
]
```

### Problems with This Approach

1. **Semantic Confusion**: URL path suggests singular resource (`/exercises/{id}`) but returns array
2. **Data Duplication**: Exercise information repeated in every row
3. **Empty Result Ambiguity**: Empty array `[]` could mean:
   - "Exercise not found" 
   - "Exercise exists but has no sets"
4. **Scalability Issues**: More exercise metadata = more duplication
5. **Type Safety**: Harder to model in strongly-typed clients

## Solution: Object-Based Response

### Option 1: Nested Object Structure

#### Response Types

```go
// New response type in swagger_types.go
type ExerciseWithSetsObjectResponse struct {
    // Exercise information (singular)
    ID          int32     `json:"id" validate:"required" example:"123"`
    Name        string    `json:"name" validate:"required" example:"Bench Press"`
    CreatedAt   time.Time `json:"created_at" validate:"required" example:"2023-01-01T15:04:05Z"`
    UpdatedAt   time.Time `json:"updated_at" validate:"required" example:"2023-01-01T15:04:05Z"`
    UserID      string    `json:"user_id" validate:"required" example:"user-123"`
    
    // Sets information (collection within the object)
    Sets        []ExerciseSetDetail `json:"sets" validate:"required"`
    
    // Future extensibility
    PersonalRecords *PersonalRecords `json:"personal_records,omitempty"`
    Notes          *string          `json:"notes,omitempty"`
}

type ExerciseSetDetail struct {
    SetID       int32     `json:"set_id" validate:"required" example:"1"`
    WorkoutID   int32     `json:"workout_id" validate:"required" example:"1"`
    WorkoutDate time.Time `json:"workout_date" validate:"required" example:"2023-01-01T10:00:00Z"`
    WorkoutNotes *string  `json:"workout_notes,omitempty" example:"Morning workout"`
    Weight      *int32    `json:"weight,omitempty" example:"225"`
    Reps        int32     `json:"reps" validate:"required" example:"10"`
    SetType     string    `json:"set_type" validate:"required" example:"working"`
    Volume      int32     `json:"volume" validate:"required" example:"2250"`
}

type PersonalRecords struct {
    MaxWeight    *int32 `json:"max_weight,omitempty" example:"315"`
    MaxReps      *int32 `json:"max_reps,omitempty" example:"15"`
    MaxVolume    *int32 `json:"max_volume,omitempty" example:"5000"`
}
```

#### JSON Response Examples

**Exercise with sets:**
```json
{
  "id": 123,
  "name": "Bench Press", 
  "created_at": "2023-01-01T15:04:05Z",
  "updated_at": "2023-01-01T15:04:05Z",
  "user_id": "user-123",
  "sets": [
    {
      "set_id": 1,
      "workout_id": 1,
      "workout_date": "2023-01-01T10:00:00Z",
      "workout_notes": "Morning workout",
      "weight": 225,
      "reps": 10,
      "set_type": "working",
      "volume": 2250
    },
    {
      "set_id": 2,
      "workout_id": 1,
      "workout_date": "2023-01-01T10:00:00Z", 
      "workout_notes": "Morning workout",
      "weight": 225,
      "reps": 8,
      "set_type": "working",
      "volume": 1800
    }
  ],
  "personal_records": {
    "max_weight": 315,
    "max_reps": 12,
    "max_volume": 3500
  },
  "notes": "Focus on proper form"
}
```

**Exercise with no sets (clear semantics):**
```json
{
  "id": 123,
  "name": "Bench Press",
  "created_at": "2023-01-01T15:04:05Z",
  "updated_at": "2023-01-01T15:04:05Z", 
  "user_id": "user-123",
  "sets": [],  // Clear: exercise exists, just no sets yet
  "personal_records": null,
  "notes": null
}
```

**Exercise not found:**
```json
// HTTP 404 Not Found
{
  "message": "Exercise not found"
}
```

#### Handler Implementation

```go
// GetExerciseWithSets godoc
// @Summary Get exercise with sets (object response)
// @Description Get a specific exercise with all its sets. Returns 404 if exercise not found, 200 with empty sets array if no sets exist.
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Success 200 {object} exercise.ExerciseWithSetsObjectResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Exercise Not Found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /exercises/{id} [get]
func (h *ExerciseHandler) GetExerciseWithSets(w http.ResponseWriter, r *http.Request) {
    exerciseID := r.PathValue("id")
    if exerciseID == "" {
        response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing exercise ID", nil)
        return
    }

    exerciseIDInt, err := strconv.Atoi(exerciseID)
    if err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID", err)
        return
    }

    req := GetExerciseWithSetsRequest{
        ExerciseID: int32(exerciseIDInt),
    }

    if err := h.validator.Struct(req); err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID: must be positive", err)
        return
    }

    // First, get the exercise itself to verify it exists
    exercise, err := h.exerciseService.GetExercise(r.Context(), req.ExerciseID)
    if err != nil {
        var errUnauthorized *ErrUnauthorized
        if errors.As(err, &errUnauthorized) {
            response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
        } else {
            // Proper 404 - exercise doesn't exist or not owned by user
            response.ErrorJSON(w, r, h.logger, http.StatusNotFound, "Exercise not found", err)
        }
        return
    }

    // Then get its sets (can be empty array)
    sets, err := h.exerciseService.GetExerciseWithSets(r.Context(), req.ExerciseID)
    if err != nil {
        var errUnauthorized *ErrUnauthorized
        if errors.As(err, &errUnauthorized) {
            response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
        } else {
            response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to get exercise sets", err)
        }
        return
    }

    // Build the object response
    exerciseResponse := ExerciseWithSetsObjectResponse{
        ID:        exercise.ID,
        Name:      exercise.Name,
        CreatedAt: exercise.CreatedAt,
        UpdatedAt: exercise.UpdatedAt,
        UserID:    exercise.UserID,
        Sets:      transformSetsToDetails(sets), // Helper function to transform
    }

    // Add personal records if available
    if personalRecords := calculatePersonalRecords(sets); personalRecords != nil {
        exerciseResponse.PersonalRecords = personalRecords
    }

    if err := response.JSON(w, http.StatusOK, exerciseResponse); err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
        return
    }
}

// Helper function to transform database rows to response objects
func transformSetsToDetails(dbSets []db.GetExerciseWithSetsRow) []ExerciseSetDetail {
    details := make([]ExerciseSetDetail, len(dbSets))
    for i, set := range dbSets {
        var weight *int32
        if set.Weight.Valid {
            weight = &set.Weight.Int32
        }
        
        var notes *string
        if set.WorkoutNotes.Valid {
            notes = &set.WorkoutNotes.String
        }

        details[i] = ExerciseSetDetail{
            SetID:        set.SetID,
            WorkoutID:    set.WorkoutID,
            WorkoutDate:  set.WorkoutDate.Time,
            WorkoutNotes: notes,
            Weight:       weight,
            Reps:         set.Reps,
            SetType:      set.SetType,
            Volume:       set.Volume,
        }
    }
    return details
}

// Helper function to calculate personal records
func calculatePersonalRecords(sets []ExerciseSetDetail) *PersonalRecords {
    if len(sets) == 0 {
        return nil
    }

    var maxWeight, maxReps, maxVolume int32
    for _, set := range sets {
        if set.Weight != nil && *set.Weight > maxWeight {
            maxWeight = *set.Weight
        }
        if set.Reps > maxReps {
            maxReps = set.Reps
        }
        if set.Volume > maxVolume {
            maxVolume = set.Volume
        }
    }

    return &PersonalRecords{
        MaxWeight: &maxWeight,
        MaxReps:   &maxReps,
        MaxVolume: &maxVolume,
    }
}
```

### Option 2: Separate Sub-Collection Endpoint

Even better would be to create dedicated sub-collection endpoints:

#### Route Structure
```go
// New routes in routes.go
mux.HandleFunc("GET /api/exercises/{id}", eh.GetExercise)           // Returns exercise object
mux.HandleFunc("GET /api/exercises/{id}/sets", eh.GetExerciseSets) // Returns sets array
mux.HandleFunc("GET /api/exercises/{id}/stats", eh.GetExerciseStats) // Returns statistics object
```

#### Handler Implementation
```go
// GetExercise returns just the exercise information
func (h *ExerciseHandler) GetExercise(w http.ResponseWriter, r *http.Request) {
    // ... validation logic ...
    
    exercise, err := h.exerciseService.GetExercise(r.Context(), req.ExerciseID)
    if err != nil {
        // Handle errors - 404 if not found
    }

    if err := response.JSON(w, http.StatusOK, exercise); err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
        return
    }
}

// GetExerciseSets returns just the sets for an exercise
func (h *ExerciseHandler) GetExerciseSets(w http.ResponseWriter, r *http.Request) {
    // ... validation logic ...
    
    sets, err := h.exerciseService.GetExerciseSets(r.Context(), req.ExerciseID)
    if err != nil {
        // Handle errors
    }

    // This can return empty array [] when no sets - semantically correct
    if err := response.JSON(w, http.StatusOK, sets); err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
        return
    }
}
```

#### URL Structure Becomes:
- `GET /exercises/{id}` → Returns exercise object (404 if not found)
- `GET /exercises/{id}/sets` → Returns array of sets (200 [] if empty)
- `GET /exercises/{id}/stats` → Returns statistics object (200 with null fields if no data)

## Benefits of Object-Based Response

### 1. Clear Semantics
- `GET /exercises/123` returns an exercise object or 404
- Empty sets array means "exercise exists, no sets yet"
- 404 means "exercise doesn't exist or not owned by user"

### 2. No Data Duplication
- Exercise info appears once at the top level
- Sets array contains only set-specific data
- Bandwidth savings on large datasets

### 3. Better Client Experience
```typescript
// TypeScript client code becomes more intuitive
interface Exercise {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
  user_id: string;
  sets: ExerciseSet[];
  personal_records?: PersonalRecords;
  notes?: string;
}

// Easy to work with
const exercise = await api.getExercise(123);
if (exercise.sets.length === 0) {
  showMessage("No sets recorded for this exercise yet");
}

// Access exercise info without array indexing
document.title = `${exercise.name} - Fitness Tracker`;
```

### 4. Future Extensibility
```json
{
  "id": 123,
  "name": "Bench Press",
  "sets": [...],
  "personal_records": {        // Easy to add
    "max_weight": 315,
    "max_reps": 12,
    "max_volume": 3500
  },
  "progression_data": {        // Easy to add
    "trend": "increasing",
    "weekly_volume": 12500
  },
  "notes": "Focus on form",    // Easy to add
  "tags": ["chest", "push"],   // Easy to add
  "difficulty_rating": 7       // Easy to add
}
```

### 5. Better HTTP Semantics
- **200**: Exercise found (sets may be empty)
- **404**: Exercise not found or not owned
- **401**: User not authenticated
- **400**: Invalid exercise ID

## Comparison: Current vs Object-Based

| Aspect | Current (Array-of-Rows) | Object-Based |
|--------|-------------------------|--------------|
| **Semantic Clarity** | Confusing (singular path, array response) | Clear (singular path, object response) |
| **Data Duplication** | High (exercise info repeated) | None (exercise info once) |
| **Empty Results** | Ambiguous (`[]` could mean anything) | Clear (exercise exists, `sets: []`) |
| **404 Usage** | Inconsistent (used for empty results) | Correct (only for missing resource) |
| **Extensibility** | Difficult (would duplicate more data) | Easy (add fields to object) |
| **Client Code** | Complex (check array[0] for exercise info) | Simple (direct object access) |
| **Type Safety** | Poor (array could be empty) | Good (object always has structure) |
| **Bandwidth** | Wasteful (repeated data) | Efficient (no duplication) |

## Migration Strategy

Since this would be a breaking change, here are migration approaches:

### Option 1: API Versioning
```go
// Keep current endpoint for backward compatibility
mux.HandleFunc("GET /api/v1/exercises/{id}", eh.GetExerciseWithSetsLegacy)

// Add new object-based endpoint  
mux.HandleFunc("GET /api/v2/exercises/{id}", eh.GetExerciseWithSets)
```

### Option 2: New Endpoint Names
```go
// Keep current endpoint (mark as deprecated)
mux.HandleFunc("GET /api/exercises/{id}", eh.GetExerciseWithSetsLegacy)

// Add new endpoints with clear names
mux.HandleFunc("GET /api/exercises/{id}/details", eh.GetExerciseWithSets)
mux.HandleFunc("GET /api/exercises/{id}/sets", eh.GetExerciseSets)
```

### Option 3: Query Parameter Flag
```go
// Single endpoint with behavior flag
func (h *ExerciseHandler) GetExerciseWithSets(w http.ResponseWriter, r *http.Request) {
    if r.URL.Query().Get("format") == "object" {
        // Return object-based response
    } else {
        // Return legacy array response
    }
}
```

### Migration Timeline
1. **Phase 1**: Add new object-based endpoints alongside existing ones
2. **Phase 2**: Update client applications to use new endpoints
3. **Phase 3**: Mark old endpoints as deprecated
4. **Phase 4**: Remove old endpoints after sufficient adoption period

## Implementation Checklist

### Backend Changes
- [ ] Create new response type structs
- [ ] Add transformation helper functions
- [ ] Implement new handler methods
- [ ] Add proper Swagger documentation
- [ ] Update route registrations
- [ ] Add comprehensive tests
- [ ] Update service layer if needed

### Documentation Updates
- [ ] Update API documentation
- [ ] Create migration guide for clients
- [ ] Add examples to Swagger/OpenAPI
- [ ] Update postman collections

### Client Updates
- [ ] Update TypeScript/JavaScript types
- [ ] Modify API client libraries
- [ ] Update frontend components
- [ ] Test all integration points

### Testing
- [ ] Unit tests for new handlers
- [ ] Integration tests for new endpoints  
- [ ] Backward compatibility tests
- [ ] Performance testing (bandwidth/response time)

## Conclusion

Object-based responses provide better semantic clarity, eliminate data duplication, and offer cleaner client integration. While the current array-of-rows approach works functionally, migrating to object-based responses would improve API usability and maintainability in the long term.

The key insight is that URL paths should match response structure:
- Singular paths (`/exercises/{id}`) should return objects
- Collection paths (`/exercises`) should return arrays
- Sub-collection paths (`/exercises/{id}/sets`) should return arrays

This creates intuitive, predictable API behavior that follows REST principles and makes client development more straightforward.
