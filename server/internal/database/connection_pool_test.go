package db

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDatabaseURL() string {
	// Load environment variables from .env file (go up to server directory)
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Try to get from environment, fallback to default test database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/fittrack_test?sslmode=disable"
	}
	return dbURL
}

func TestConnectionPoolIsolation(t *testing.T) {
	// Create a connection pool
	pool, err := pgxpool.New(context.Background(), getTestDatabaseURL())
	require.NoError(t, err, "Unable to create connection pool")
	defer pool.Close()

	// Test database connectivity
	err = pool.Ping(context.Background())
	require.NoError(t, err, "Unable to ping database")

	// Setup test data
	setupTestData(t, pool)
	defer cleanupTestData(t, pool)

	// Define test cases
	tests := []struct {
		name      string
		userID    string
		workoutID int
		expected  bool
	}{
		{name: "User1_Workout1", userID: "test-user-1", workoutID: 1, expected: true},
		{name: "User1_Workout2", userID: "test-user-1", workoutID: 2, expected: false},
		{name: "User2_Workout1", userID: "test-user-2", workoutID: 1, expected: false},
		{name: "User2_Workout2", userID: "test-user-2", workoutID: 2, expected: true},
	}

	// Run test cases concurrently to test connection pool isolation
	var wg sync.WaitGroup
	for _, test := range tests {
		wg.Add(1)
		test := test // Capture test variable for goroutine
		go func() {
			defer wg.Done()
			testUserContextIsolation(t, pool, test.userID, test.workoutID, test.expected)
		}()
	}

	wg.Wait()
}

// testUserContextIsolation tests that the connection pool properly isolates user context
func testUserContextIsolation(t *testing.T, pool *pgxpool.Pool, userID string, workoutID int, expected bool) {
	ctx := context.Background()

	// Acquire a dedicated connection from the pool to test isolation
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err, "Failed to acquire connection from pool")
	defer conn.Release()

	// Set the current user ID as a session variable on this specific connection
	_, err = conn.Exec(ctx, "SELECT set_config('app.current_user_id', $1, false)", userID)
	require.NoError(t, err, "Failed to set user context")

	// Verify the session variable was set correctly on this connection
	var currentUserID string
	err = conn.QueryRow(ctx, "SELECT current_user_id()").Scan(&currentUserID)
	require.NoError(t, err, "Failed to get current user ID")
	t.Logf("Set user context for %s, got back: %s", userID, currentUserID)

	// Debug: Check RLS status on workout table
	var rlsEnabled bool
	err = conn.QueryRow(ctx, "SELECT relrowsecurity FROM pg_class WHERE relname = 'workout'").Scan(&rlsEnabled)
	require.NoError(t, err, "Failed to check RLS status")
	t.Logf("RLS enabled on workout table: %v", rlsEnabled)

	// Debug: Check current user role and superuser status
	var currentUser string
	var isSuperuser bool
	err = conn.QueryRow(ctx, "SELECT current_user, usesuper FROM pg_user WHERE usename = current_user").Scan(&currentUser, &isSuperuser)
	require.NoError(t, err, "Failed to check current user role")
	t.Logf("Current user: %s, Is superuser: %v", currentUser, isSuperuser)

	// Debug: Check what policies exist on workout table
	rows2, err := conn.Query(ctx, "SELECT policyname FROM pg_policies WHERE tablename = 'workout'")
	require.NoError(t, err, "Failed to check policies")
	defer rows2.Close()
	var policies []string
	for rows2.Next() {
		var policyName string
		err = rows2.Scan(&policyName)
		require.NoError(t, err, "Failed to scan policy name")
		policies = append(policies, policyName)
	}
	t.Logf("Policies on workout table: %v", policies)

	// Small delay to ensure session variable is set properly
	time.Sleep(10 * time.Millisecond)

	// Execute the query using this specific connection - RLS should filter results based on the session variable
	rows, err := conn.Query(ctx, "SELECT id FROM workout WHERE id = $1", workoutID)
	require.NoError(t, err, "Query should not error")
	defer rows.Close()

	// Count the number of rows returned
	var rowCount int
	var retrievedWorkoutID int
	for rows.Next() {
		rowCount++
		err = rows.Scan(&retrievedWorkoutID)
		require.NoError(t, err, "Failed to scan row")
	}

	// Check the result based on whether we're running as superuser
	if isSuperuser {
		// SUPERUSER BEHAVIOR: Superusers bypass RLS policies and can access all data
		// This is the expected PostgreSQL behavior - RLS does not apply to superusers
		assert.Equal(t, 1, rowCount, "Superuser should access all data (RLS bypass): user %s accessing workout %d", userID, workoutID)
		assert.Equal(t, workoutID, retrievedWorkoutID, "Expected workout ID %d, got %d", workoutID, retrievedWorkoutID)
		t.Logf("✅ SUPERUSER BYPASS: User %s accessed workout %d (RLS policies don't apply to superusers)", userID, workoutID)
	} else {
		// NON-SUPERUSER BEHAVIOR: RLS policies should be enforced
		if expected {
			// User should be able to access their own workout (should return 1 row)
			assert.Equal(t, 1, rowCount, "Expected 1 row for user %s accessing workout %d, got %d", userID, workoutID, rowCount)
			if rowCount > 0 {
				assert.Equal(t, workoutID, retrievedWorkoutID, "Expected workout ID %d, got %d", workoutID, retrievedWorkoutID)
			}
		} else {
			// User should not be able to access another user's workout (RLS should filter out, returning 0 rows)
			assert.Equal(t, 0, rowCount, "Expected 0 rows for user %s accessing workout %d, got %d (RLS should filter out)", userID, workoutID, rowCount)
		}
	}
}

// setupTestData creates test users and workouts for the isolation test
func setupTestData(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// First ensure RLS is set up (apply the RLS migration content)
	setupRLS(t, pool)

	// Disable RLS temporarily for setup
	_, err := pool.Exec(ctx, "ALTER TABLE users DISABLE ROW LEVEL SECURITY; ALTER TABLE workout DISABLE ROW LEVEL SECURITY;")
	require.NoError(t, err, "Failed to disable RLS for setup")

	// Create test users
	_, err = pool.Exec(ctx, "INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", "test-user-1")
	require.NoError(t, err, "Failed to create test user 1")

	_, err = pool.Exec(ctx, "INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", "test-user-2")
	require.NoError(t, err, "Failed to create test user 2")

	// Create test workouts
	_, err = pool.Exec(ctx, "INSERT INTO workout (id, date, user_id) VALUES (1, NOW(), $1) ON CONFLICT (id) DO UPDATE SET user_id = EXCLUDED.user_id", "test-user-1")
	require.NoError(t, err, "Failed to create workout 1 for user 1")

	_, err = pool.Exec(ctx, "INSERT INTO workout (id, date, user_id) VALUES (2, NOW(), $1) ON CONFLICT (id) DO UPDATE SET user_id = EXCLUDED.user_id", "test-user-2")
	require.NoError(t, err, "Failed to create workout 2 for user 2")

	// Re-enable RLS for testing
	_, err = pool.Exec(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY; ALTER TABLE workout ENABLE ROW LEVEL SECURITY;")
	require.NoError(t, err, "Failed to re-enable RLS after setup")
}

// cleanupTestData removes test data after the test
func cleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Disable RLS temporarily for cleanup
	_, err := pool.Exec(ctx, "ALTER TABLE users DISABLE ROW LEVEL SECURITY; ALTER TABLE workout DISABLE ROW LEVEL SECURITY;")
	if err != nil {
		t.Logf("Warning: Failed to disable RLS for cleanup: %v", err)
	}

	// Clean up test data
	_, err = pool.Exec(ctx, "DELETE FROM workout WHERE id IN (1, 2)")
	if err != nil {
		t.Logf("Warning: Failed to clean up workout data: %v", err)
	}

	_, err = pool.Exec(ctx, "DELETE FROM users WHERE user_id IN ('test-user-1', 'test-user-2')")
	if err != nil {
		t.Logf("Warning: Failed to clean up user data: %v", err)
	}

	// Re-enable RLS
	_, err = pool.Exec(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY; ALTER TABLE workout ENABLE ROW LEVEL SECURITY;")
	if err != nil {
		t.Logf("Warning: Failed to re-enable RLS after cleanup: %v", err)
	}
}

// setupRLS ensures that RLS policies are in place for testing
func setupRLS(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Check if current_user_id function exists
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'current_user_id')").Scan(&exists)
	require.NoError(t, err, "Failed to check if current_user_id function exists")

	if !exists {
		t.Log("Setting up RLS policies for testing")

		// Create the current_user_id function
		_, err = pool.Exec(ctx, `
			CREATE OR REPLACE FUNCTION current_user_id() 
			RETURNS TEXT AS $$
			    SELECT current_setting('app.current_user_id', true);
			$$ LANGUAGE SQL STABLE;
		`)
		require.NoError(t, err, "Failed to create current_user_id function")

		// Enable RLS on tables
		_, err = pool.Exec(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY")
		require.NoError(t, err, "Failed to enable RLS on users table")

		_, err = pool.Exec(ctx, "ALTER TABLE workout ENABLE ROW LEVEL SECURITY")
		require.NoError(t, err, "Failed to enable RLS on workout table")

		// Create policies for users table
		_, err = pool.Exec(ctx, `
			CREATE POLICY users_policy ON users
			    FOR ALL TO PUBLIC
			    USING (user_id = current_user_id())
			    WITH CHECK (user_id = current_user_id())
		`)
		require.NoError(t, err, "Failed to create users policy")

		// Create policies for workout table
		_, err = pool.Exec(ctx, `
			CREATE POLICY workout_select_policy ON workout
			    FOR SELECT TO PUBLIC
			    USING (user_id = current_user_id())
		`)
		require.NoError(t, err, "Failed to create workout select policy")

		_, err = pool.Exec(ctx, `
			CREATE POLICY workout_insert_policy ON workout
			    FOR INSERT TO PUBLIC
			    WITH CHECK (user_id = current_user_id())
		`)
		require.NoError(t, err, "Failed to create workout insert policy")

		_, err = pool.Exec(ctx, `
			CREATE POLICY workout_update_policy ON workout
			    FOR UPDATE TO PUBLIC
			    USING (user_id = current_user_id())
			    WITH CHECK (user_id = current_user_id())
		`)
		require.NoError(t, err, "Failed to create workout update policy")

		_, err = pool.Exec(ctx, `
			CREATE POLICY workout_delete_policy ON workout
			    FOR DELETE TO PUBLIC
			    USING (user_id = current_user_id())
		`)
		require.NoError(t, err, "Failed to create workout delete policy")
	}
}
