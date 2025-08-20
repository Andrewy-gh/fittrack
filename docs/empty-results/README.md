# Empty Results vs Not Found in REST APIs

## TL;DR

**Empty Collections**: Always return `200 OK` with an empty array `[]` when no results are found.
**Missing Resources**: Return `404 Not Found` only when the resource itself doesn't exist or isn't owned by the user.

## The Problem

Inconsistent handling of empty results creates confusion for API consumers and violates REST principles. This document establishes clear rules for when to return `200 []` vs `404` in our fitness tracking API.

## Our Policy

### ✅ Collection/Sub-collection GETs
- **Always return `200 OK` with `[]` when empty**
- Never return `404` based solely on emptiness
- Examples: `/workouts`, `/exercises`, `/exercises/{id}/recent-sets`

### ✅ Singular Resource GETs  
- **Return `404 Not Found` only when the resource is missing or not owned**
- **Return `200 OK` with object when found** (nested arrays may be empty)
- Examples: `/workouts/{id}` (as object), `/exercises/{id}` (as object)

### ⚠️ Special Case: Array-of-Rows Endpoints
Some endpoints return arrays but represent filtered views of a singular resource:
- `/exercises/{id}` (current implementation returns array of exercise-with-sets rows)
- `/workouts/{id}` (if implemented to return array of workout-with-sets rows)

**Treatment**: Treat as filtered collections → return `200 []` when empty.
**Future improvement**: Migrate to object-based responses for clearer semantics.

## Decision Matrix

| Endpoint | Empty Result | Rationale |
|----------|-------------|-----------|
| `GET /workouts` | `200 []` | Collection endpoint |
| `GET /exercises` | `200 []` | Collection endpoint |
| `GET /exercises/{id}/recent-sets` | `200 []` | Sub-collection endpoint |
| `GET /exercises/{id}` (current array shape) | `200 []` | Treated as filtered collection |
| `GET /workouts/{id}` (when returns object) | `404` if workout missing, `200 {..., "sets": []}` if found but empty sets | Singular resource |

## Examples

### ✅ Correct: Empty Collection
```http
GET /api/exercises/123/recent-sets
Authorization: Bearer token123

HTTP/1.1 200 OK
Content-Type: application/json

[]
```

### ✅ Correct: Missing Resource
```http
GET /api/workouts/999
Authorization: Bearer token123

HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "message": "workout not found"
}
```

### ✅ Correct: Found Resource with Empty Nested Data
```http
GET /api/workouts/123
Authorization: Bearer token123

HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": 123,
  "date": "2023-01-01T10:00:00Z",
  "notes": "Just created, no sets yet",
  "sets": []
}
```

### ❌ Incorrect: 404 for Empty Collection
```http
GET /api/exercises/123/recent-sets
Authorization: Bearer token123

HTTP/1.1 404 Not Found  ← WRONG!
Content-Type: application/json

{
  "message": "No sets found for this exercise"
}
```

## Rationale

### 1. **HTTP Semantics (RFC 9110)**
- `200 OK`: Request succeeded, response has payload
- `404 Not Found`: Target resource not found
- Empty collections are successful responses with payload `[]`

### 2. **Developer Experience**
- Consistent behavior reduces client-side error handling complexity
- Empty arrays are easier to iterate over than handling 404s
- Reduces false-positive alerts in monitoring systems

### 3. **Cache Behavior**
- `200` responses are cacheable by default
- `404` responses may be cached differently, causing issues when data appears later

### 4. **Industry Standards**
- GitHub API: `GET /repos/owner/repo/issues` returns `[]` when no issues
- Google APIs: Collections return empty arrays, not 404s
- REST API best practices widely recommend this pattern

## Anti-Patterns

### ❌ Don't: Use 404 for Empty Collections
```go
if len(results) == 0 {
    response.ErrorJSON(w, r, h.logger, http.StatusNotFound, "No items found", nil)
    return
}
```

### ❌ Don't: Use 204 No Content for GET Requests
```go
// Wrong - 204 is for successful operations with no response body
w.WriteHeader(http.StatusNoContent) // For empty collections
```

### ❌ Don't: Inconsistent Behavior Across Similar Endpoints
- `ListExercises` returns `200 []` when empty ✅
- `GetExerciseWithSets` returns `404` when empty ❌ ← Inconsistent!

## Implementation Guidelines

### Handler Pattern
```go
// ✅ Correct pattern for collection endpoints
func (h *ExerciseHandler) GetRecentSetsForExercise(w http.ResponseWriter, r *http.Request) {
    sets, err := h.exerciseService.GetRecentSetsForExercise(r.Context(), exerciseID)
    if err != nil {
        // Handle service errors (unauthorized, internal error)
        // ...
        return
    }

    // Always return 200 with array (may be empty)
    if err := response.JSON(w, http.StatusOK, sets); err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
        return
    }
}

// ✅ Correct pattern for singular resource endpoints  
func (h *WorkoutHandler) GetWorkout(w http.ResponseWriter, r *http.Request) {
    workout, err := h.workoutService.GetWorkout(r.Context(), workoutID)
    if err != nil {
        var errNotFound *ErrNotFound
        if errors.As(err, &errNotFound) {
            response.ErrorJSON(w, r, h.logger, http.StatusNotFound, errNotFound.Message, nil)
            return
        }
        // Handle other errors...
    }

    if err := response.JSON(w, http.StatusOK, workout); err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
        return
    }
}
```

### Service Layer Pattern
```go
// Service layer should NOT determine HTTP status codes
// Return empty slices normally, use custom errors for missing resources
func (s *ExerciseService) GetRecentSetsForExercise(ctx context.Context, exerciseID int32) ([]db.GetRecentSetsForExerciseRow, error) {
    userID, ok := user.Current(ctx)
    if !ok {
        return nil, &ErrUnauthorized{Message: "user not authenticated"}
    }

    sets, err := s.repo.GetRecentSetsForExercise(ctx, exerciseID, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get recent sets: %w", err)
    }

    // Return empty slice normally - handler decides HTTP status
    return sets, nil
}
```

### Swagger Documentation
```go
// GetRecentSetsForExercise godoc
// @Summary Get recent sets for exercise
// @Description Get the 3 most recent sets for a specific exercise. Returns empty array when exercise has no sets.
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Success 200 {array} exercise.RecentSetsResponse "Success (may be empty array)"
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /exercises/{id}/recent-sets [get]
```

## Migration Considerations

### Current Inconsistency: `GetExerciseWithSets`
**Issue**: Returns 404 when `len(exerciseWithSets) == 0`
**Fix**: Remove empty check, always return 200 with array

```go
// REMOVE this block:
if len(exerciseWithSets) == 0 {
    response.ErrorJSON(w, r, h.logger, http.StatusNotFound, "No sets found for this exercise", nil)
    return
}

// Always return the array (may be empty)
if err := response.JSON(w, http.StatusOK, exerciseWithSets); err != nil {
    response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
    return
}
```

### Breaking Change Notice
Clients previously expecting 404 for empty `GetExerciseWithSets` must be updated to handle 200 with empty array.

### Future Improvements
Consider migrating array-returning singular resource endpoints to object-based responses:

```go
// Current: GET /exercises/{id} returns array
[
  {
    "set_id": 1,
    "exercise_id": 123,
    "exercise_name": "Bench Press",
    // ...
  }
]

// Better: GET /exercises/{id} returns object
{
  "id": 123,
  "name": "Bench Press",
  "created_at": "2023-01-01T10:00:00Z",
  "sets": [
    {
      "id": 1,
      "weight": 225,
      "reps": 10
    }
  ]
}

// Or: New sub-collection endpoint
// GET /exercises/{id}/sets returns array
```

## Testing

### Required Test Cases
Every collection endpoint must test:
```go
func TestHandler_EmptyCollection(t *testing.T) {
    // Given: valid auth and parameters but no data
    // When: calling collection endpoint
    // Then: expect 200 OK with empty array []
}
```

Every singular resource endpoint must test:
```go
func TestHandler_ResourceNotFound(t *testing.T) {
    // Given: valid auth but resource doesn't exist/isn't owned
    // When: calling singular resource endpoint  
    // Then: expect 404 Not Found with error message
}
```

## Monitoring

After implementing changes, monitor for:
- Increased 4xx errors from clients expecting 404s
- Client application errors in logs
- User-reported issues with "missing data" scenarios

## Further Reading

- [RFC 9110: HTTP Semantics](https://tools.ietf.org/rfc/rfc9110.txt)
- [RESTful API Guidelines](https://github.com/microsoft/api-guidelines)
- [HTTP Status Codes Decision Tree](https://httpstatuses.com/)
