   1. Add SQL Query: I'll add the new SQL query to server/query.sql.
   2. Update Database Queries: Run sqlc to regenerate the Go database query code in
      server/internal/database/query.sql.go.
   3. Update Workout Repository: Add a method to the WorkoutRepository interface and its implementation in
      workout/repository.go to execute the new query.
   4. Update Workout Service: Add a method to the WorkoutService in workout/service.go to call the repository
      method, handling user authentication and errors.
   5. Create Handler: Add a new handler method ListWorkoutFocusValues in workout/handler.go with proper Swagger
       annotations.
   6. Update Routes: Add the new route GET /workout-focus-values in cmd/api/routes.go.
   7. Add Tests: Create or update tests for the new functionality.