package workout

import (
	"context"
	"fmt"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

// Test user registry for tracking users created during tests
// This allows for automatic cleanup without hardcoding user IDs
var (
	testUserRegistry   = make(map[string]bool)
	testUserRegistryMu sync.Mutex
)

// registerTestUser adds a user to the cleanup registry
func registerTestUser(userID string) {
	testUserRegistryMu.Lock()
	defer testUserRegistryMu.Unlock()
	testUserRegistry[userID] = true
}

// getRegisteredTestUsers returns a slice of all registered test users
func getRegisteredTestUsers() []string {
	testUserRegistryMu.Lock()
	defer testUserRegistryMu.Unlock()
	users := make([]string, 0, len(testUserRegistry))
	for userID := range testUserRegistry {
		users = append(users, userID)
	}
	return users
}

// clearTestUserRegistry clears the registry (useful for test isolation if needed)
func clearTestUserRegistry() {
	testUserRegistryMu.Lock()
	defer testUserRegistryMu.Unlock()
	testUserRegistry = make(map[string]bool)
}

// MockWorkoutRepository implements the WorkoutRepository interface for testing
type MockWorkoutRepository struct {
	mock.Mock
}

func (m *MockWorkoutRepository) SaveWorkout(ctx context.Context, reformatted *ReformattedRequest, userID string) error {
	args := m.Called(ctx, reformatted, userID)
	return args.Error(0)
}

func (m *MockWorkoutRepository) SaveWorkoutWithID(ctx context.Context, reformatted *ReformattedRequest, userID string) (int32, error) {
	args := m.Called(ctx, reformatted, userID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *MockWorkoutRepository) GetWorkout(ctx context.Context, id int32, userID string) (db.Workout, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(db.Workout), args.Error(1)
}

func (m *MockWorkoutRepository) ListWorkouts(ctx context.Context, userID string) ([]db.Workout, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Workout), args.Error(1)
}

func (m *MockWorkoutRepository) ListWorkoutFocusTemplates(ctx context.Context, userID string) ([]db.ListWorkoutFocusTemplatesRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.ListWorkoutFocusTemplatesRow), args.Error(1)
}

func (m *MockWorkoutRepository) GetLatestWorkoutNote(ctx context.Context, userID string) (db.GetLatestWorkoutNoteRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.GetLatestWorkoutNoteRow), args.Error(1)
}

func (m *MockWorkoutRepository) GetWorkoutWithSets(ctx context.Context, id int32, userID string) ([]db.GetWorkoutWithSetsRow, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]db.GetWorkoutWithSetsRow), args.Error(1)
}

func (m *MockWorkoutRepository) UpdateWorkout(ctx context.Context, id int32, reformatted *ReformattedRequest, userID string) error {
	args := m.Called(ctx, id, reformatted, userID)
	return args.Error(0)
}

func (m *MockWorkoutRepository) DeleteWorkout(ctx context.Context, id int32, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockWorkoutRepository) ListWorkoutFocusValues(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockWorkoutRepository) GetContributionData(ctx context.Context, userID string) ([]db.GetContributionDataRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.GetContributionDataRow), args.Error(1)
}

type errorResponse struct {
	Message string `json:"message"`
}

// === HELPER FUNCTIONS FOR INTEGRATION TESTS ===

func getTestDatabaseURL() string {
	// Try to get from environment first (for CI/CD)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		return dbURL
	}

	// Load environment variables from .env file for local development
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Try again after loading .env
	dbURL = os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Final fallback to default test database
		dbURL = "postgres://postgres:password@localhost:5432/fittrack_test?sslmode=disable"
	}
	return dbURL
}

func setupTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	// Create connection pool
	pool, err := pgxpool.New(ctx, getTestDatabaseURL())
	require.NoError(t, err, "Failed to create database pool")

	// Test connectivity
	err = pool.Ping(ctx)
	require.NoError(t, err, "Failed to ping database")

	// Setup RLS policies
	setupRLS(t, pool)

	// Setup users table entries
	setupTestUsers(t, pool)

	// Backfill order columns for existing test data if they exist
	// This ensures tests work with both old and new database schemas
	backfillOrderColumnsForTests(t, pool)

	return pool, func() {
		cleanupTestData(t, pool)
		pool.Close()
	}
}

func setupRLS(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	// Check if current_user_id function exists (should exist from migration)
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'current_user_id')").Scan(&exists)
	require.NoError(t, err, "Failed to check if current_user_id function exists")
	require.True(t, exists, "current_user_id function should exist from migration")

	// Ensure RLS is enabled on tables (should be enabled from migration)
	_, err = pool.Exec(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to enable RLS on users table")

	_, err = pool.Exec(ctx, "ALTER TABLE workout ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to enable RLS on workout table")

	_, err = pool.Exec(ctx, "ALTER TABLE exercise ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to enable RLS on exercise table")

	_, err = pool.Exec(ctx, "ALTER TABLE \"set\" ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to enable RLS on set table")

	// RLS policies should already exist from migration - no need to recreate them
	// Just verify they exist
	var policyCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM pg_policies WHERE tablename IN ('users', 'workout', 'exercise', 'set')").Scan(&policyCount)
	require.NoError(t, err, "Failed to count RLS policies")
	require.Greater(t, policyCount, 0, "RLS policies should exist from migration")
}

func setupTestUsers(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	// Temporarily disable RLS for setup
	_, err := pool.Exec(ctx, "ALTER TABLE users DISABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to disable RLS for setup")

	// Create test users and register them for cleanup
	// Include all users needed by integration tests
	userIDs := []string{
		"test-user-a",
		"test-user-b",
		"test-user-rls-a",
		"test-user-rls-b",
		"test-user-integration-contrib",
		"test-user-empty-contrib",
	}
	for _, userID := range userIDs {
		_, err = pool.Exec(ctx, "INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
		require.NoError(t, err, "Failed to create test user: %s", userID)
		registerTestUser(userID)
	}

	// Re-enable RLS
	_, err = pool.Exec(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to re-enable RLS")
}

func setupTestWorkout(t *testing.T, pool *pgxpool.Pool, userID, notes string) int32 {
	t.Helper()
	ctx := context.Background()

	// Set user context for RLS
	ctx = testutils.SetTestUserContext(ctx, t, pool, userID)

	// Create workout
	var workoutID int32
	err := pool.QueryRow(ctx,
		"INSERT INTO workout (date, notes, user_id) VALUES (NOW(), $1, $2) RETURNING id",
		notes, userID).Scan(&workoutID)
	require.NoError(t, err, "Failed to create test workout for user %s", userID)

	return workoutID
}

func setupTestSet(t *testing.T, pool *pgxpool.Pool, userID string, workoutID int32) int32 {
	t.Helper()
	ctx := context.Background()

	// Set user context for RLS
	ctx = testutils.SetTestUserContext(ctx, t, pool, userID)

	// Get or create a test exercise
	var exerciseID int32
	err := pool.QueryRow(ctx,
		"INSERT INTO exercise (name, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id",
		"Test Exercise", userID).Scan(&exerciseID)
	if err != nil {
		// If conflict, query for existing
		err = pool.QueryRow(ctx,
			"SELECT id FROM exercise WHERE name = $1 AND user_id = $2 LIMIT 1",
			"Test Exercise", userID).Scan(&exerciseID)
		require.NoError(t, err, "Failed to get test exercise for user %s", userID)
	}

	// Create a working set
	var setID int32
	err = pool.QueryRow(ctx,
		"INSERT INTO \"set\" (workout_id, exercise_id, user_id, weight, reps, set_type, exercise_order, set_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		workoutID, exerciseID, userID, 100, 10, "working", 1, 1).Scan(&setID)
	require.NoError(t, err, "Failed to create test set for user %s", userID)

	return setID
}

func setTestUserContext(ctx context.Context, t *testing.T, pool *pgxpool.Pool, userID string) context.Context {
	t.Helper()
	return testutils.SetTestUserContext(ctx, t, pool, userID)
}

func getDatabaseCurrentDateString(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()

	var date string
	err := pool.QueryRow(context.Background(), "SELECT CURRENT_DATE::TEXT").Scan(&date)
	require.NoError(t, err, "Failed to get current database date")
	return date
}

func getDatabaseDateString(t *testing.T, pool *pgxpool.Pool, date time.Time) string {
	t.Helper()

	var dateString string
	err := pool.QueryRow(
		context.Background(),
		"SELECT DATE_TRUNC('day', $1::timestamptz)::DATE::TEXT",
		date,
	).Scan(&dateString)
	require.NoError(t, err, "Failed to normalize date with database timezone")
	return dateString
}

func cleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Disable RLS temporarily for cleanup
	tables := []string{"users", "workout", "exercise", `"set"`}
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("ALTER TABLE %s DISABLE ROW LEVEL SECURITY", table))
		if err != nil {
			t.Logf("Warning: Failed to disable RLS on %s for cleanup: %v", table, err)
		}
	}

	// Get all registered test users from the registry
	// This automatically includes any users created during test setup
	testUserIDs := getRegisteredTestUsers()

	if len(testUserIDs) == 0 {
		t.Logf("Warning: No test users registered for cleanup")
	}

	// Clean up data for each registered test user
	for _, userID := range testUserIDs {
		// Clean up dependent data first (sets â†’ exercises â†’ workouts â†’ users)
		// exercise.historical_1rm_source_workout_id references workout(id), so exercises must be removed before workouts.
		_, err := pool.Exec(ctx, "DELETE FROM \"set\" WHERE user_id = $1", userID)
		if err != nil {
			t.Logf("Warning: Failed to clean up set data for user %s: %v", userID, err)
		}

		_, err = pool.Exec(ctx, "DELETE FROM exercise WHERE user_id = $1", userID)
		if err != nil {
			t.Logf("Warning: Failed to clean up exercise data for user %s: %v", userID, err)
		}

		_, err = pool.Exec(ctx, "DELETE FROM workout WHERE user_id = $1", userID)
		if err != nil {
			t.Logf("Warning: Failed to clean up workout data for user %s: %v", userID, err)
		}

		_, err = pool.Exec(ctx, "DELETE FROM users WHERE user_id = $1", userID)
		if err != nil {
			t.Logf("Warning: Failed to clean up user data for user %s: %v", userID, err)
		}
	}

	// Re-enable RLS
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", table))
		if err != nil {
			t.Logf("Warning: Failed to re-enable RLS on %s after cleanup: %v", table, err)
		}
	}

	t.Logf("Test cleanup complete for %d registered users: %v", len(testUserIDs), testUserIDs)
}

func backfillOrderColumnsForTests(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	// Backfill order columns for all test users
	// This ensures tests work whether or not the migration has been applied
	testUsers := []string{"test-user-a", "test-user-b"}

	for _, userID := range testUsers {
		// Set user context for RLS
		ctxUser := testutils.SetTestUserContext(ctx, t, pool, userID)

		// Backfill order columns for this user
		testutils.BackfillSetOrderColumns(ctxUser, t, pool, userID)
	}
}
