# Move Exercise Calculations to Server

## Context
`exercise-detail.tsx` doing heavy client-side calculations (stats aggregations, grouping). Moving to server for performance.

## Decisions
- **Endpoint:** Modify existing `GET /api/exercises/{id}` (breaking change, coordinated deployment)
- **Approach:** SQL for stats aggregation, Go for grouping/transformation
- **Testing:** Unit tests for grouping logic only
- **Scope:** This endpoint only (keep client sorting utils for other components)

## Implementation Steps

### Server

1. **Add stats SQL query** → `server/query.sql`
   - `GetExerciseStats :one` - COUNT, AVG, MAX for sets/workouts/weight/volume
   - Takes `exercise_id`, `user_id`

2. **Add unit tests FIRST (TDD)** → `server/internal/exercise/service_test.go`
   - Test `groupSetsByWorkout()` with multiple workouts
   - Test empty sets, single workout, multiple workouts
   - Tests will fail initially (function doesn't exist yet)

3. **Add grouping logic** → `server/internal/exercise/service.go`
   - `WorkoutGroup` struct (workout_id, date, notes, sets[], total_reps, total_volume)
   - `ExerciseDetailResponse` struct (exercise_name, stats, workout_groups[])
   - `groupSetsByWorkout()` pure function
   - `GetExerciseDetail()` service method
   - Implement until tests pass

4. **Run sqlc** → `cd server && sqlc generate`

5. **Update handler** → `server/internal/exercise/handler.go`
   - `GetExerciseWithSets()` calls new service method
   - Returns new response shape

6. **Update Swagger** → `handler.go`
   - Document new response schema

### Client

7. **Regenerate OpenAPI client** → `cd client && npm run generate:client`

8. **Update component** → `client/src/components/exercises/exercise-detail.tsx`
   - Remove: stats calculations (lines 33-47), sorting (line 49), grouping (lines 52-68)
   - Use: `data.stats.*`, `data.workout_groups`

## New API Response Shape
```json
{
  "exercise_name": "Bench Press",
  "stats": {
    "total_sets": 50,
    "unique_workouts": 12,
    "avg_weight": 135,
    "max_weight": 225,
    "avg_volume": 2700,
    "max_volume": 4500
  },
  "workout_groups": [
    {
      "workout_id": 10,
      "date": "2025-01-15T10:00:00Z",
      "notes": "Heavy day",
      "total_reps": 25,
      "total_volume": 3375,
      "sets": [...]
    }
  ]
}
```

## Files Modified
- `server/query.sql`
- `server/internal/exercise/service.go`
- `server/internal/exercise/service_test.go`
- `server/internal/exercise/handler.go`
- `client/src/components/exercises/exercise-detail.tsx`
