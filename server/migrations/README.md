# Database Migrations

This directory contains the database migration files for the FitTrack application. Migrations are managed using [Goose](https://github.com/pressly/goose).

## Migration Files

1. `00001_create_users_table.sql` - Creates the users table
2. `00002_create_base_tables.sql` - Creates the workout, exercise, and "set" tables
3. `00003_add_user_id_to_workout_exercise_table.sql` - Adds user_id foreign keys to workout and exercise tables
4. `00004_remove_exercise_name_unique_constraint.sql` - Removes the unique constraint on exercise names
5. `00005_make_user_id_not_nullable.sql` - Makes user_id columns NOT NULL
6. `00006_add_rls_policies.sql` - Adds Row Level Security (RLS) policies
7. `00007_add_user_id_to_sets.sql` - Adds user_id column to the "set" table
8. `00008_add_order_columns_to_set_table.sql` - Adds exercise_order and set_order columns to "set" table

## Set Ordering (Migration 00008)

The `00008_add_order_columns_to_set_table.sql` migration adds explicit ordering columns to control the sequence of exercises within workouts and sets within exercises.

### Columns Added

- `exercise_order INTEGER` - Position of this exercise within the workout (1, 2, 3, etc.)
- `set_order INTEGER` - Position of this set within the exercise (1, 2, 3, etc.)

### Important Notes

- **Nullable Initially**: Columns are intentionally nullable to allow safe deployment to production
- **1-indexed Convention**: Values start at 1 (not 0) for user-friendly ordering
- **No Indexes**: Pure ordering columns with no filtering, so no indexes added initially
- **Backfill Required**: Existing data needs to be populated before making columns NOT NULL

### Backfilling Data

After applying migration 00008, use the backfill script `server/scripts/backfill-set-order-columns.sql` to populate existing data:

```bash
# Run the backfill script
psql -d your_database -f server/scripts/backfill-set-order-columns.sql
```

The backfill maintains previous ordering behavior:
- Exercises ordered by name within each workout
- Sets ordered by created_at and id within each exercise

### Future Migration

A follow-up migration will:
- Make columns NOT NULL with appropriate defaults
- Add CHECK constraints ensuring values ≥ 1
- Optionally add indexes if query patterns require them

## Row Level Security (RLS)

The `00005_add_rls_policies.sql` migration adds RLS policies to ensure that users can only access their own data. The policies work as follows:

### How it works

1. A `current_user_id()` function is created that retrieves the current user ID from a PostgreSQL session variable
2. RLS is enabled on all tables
3. Policies are created for each table to restrict access based on the user ID
4. The application sets the session variable `app.current_user_id` after authentication

### Policies

- **users table**: Users can only access their own user record
- **workout table**: Users can only access workouts they created
- **exercise table**: Users can only access exercises they created
- **set table**: Users can only access sets that belong to their workouts

### Session Variable

The application automatically sets the PostgreSQL session variable `app.current_user_id` after successful authentication. This variable is used by the `current_user_id()` function in the RLS policies.

## Applying Migrations

To apply migrations, run:

```bash
goose -dir migrations postgres "user=username dbname=database_name sslmode=disable" up
```

## Rolling Back Migrations

To roll back the last migration, run:

```bash
goose -dir migrations postgres "user=username dbname=database_name sslmode=disable" down
```

## Performance Considerations

RLS policies add a WHERE clause to every query, which can impact performance. Ensure you have proper indexes on user_id columns:

- `workout(user_id)`
- `exercise(user_id)`
- `set(workout_id)` (with join to workout for user_id)

These indexes are already created in migration 00002.

## Testing

For testing with RLS enabled, you can set the session variable directly in your test setup:

```sql
SELECT set_config('app.current_user_id', 'test-user-id', false);
```

## Testing

For testing with RLS enabled, you can set the session variable directly in your test setup:

```sql
SELECT set_config('app.current_user_id', 'test-user-id', false);
```

Or use the test utility in `server/internal/testutils`:

```go
ctx = testutils.SetTestUserContext(ctx, t, db, "test-user-id")
```

## Error Handling

When RLS blocks access, PostgreSQL returns empty result sets rather than explicit permission errors. The application should implement additional checks where appropriate to provide better user experience.

A utility function `IsRowLevelSecurityError` is available in `server/internal/database/errors.go` to help with error handling, though RLS typically doesn't return specific error codes for blocked access.
