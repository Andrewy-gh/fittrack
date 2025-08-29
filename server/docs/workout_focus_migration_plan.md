- Create a new migration file to add a new column called 'workout_focus' to the 'workout' table.
- The column should be nullable and of type VARCHAR(256).
- Run the migration
- Update `schema.sql` To include the new column in the workouts table.
- Update the `query.sql` to include the new column.
- Update pertinent structs to include the new column so we can validate it. `CreateWorkoutRequest` is an example of one them
- Update ``workout/handler.go`, `workout/service.go`, `workout/repository.go`, `workout/repository_test.go` and any other underlying tests

**UPDATES**

1.  Create the migration file with the proper naming convention
2.  Update the schema.sql file
3.  Update the query.sql file
4.  Update the workout models to include the new field
5.  Update the workout handler, service, and repository files
6.  Update the swagger documentation
7.  Run tests to ensure everything works correctly
