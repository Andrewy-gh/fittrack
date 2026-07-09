package trainingprofile

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryGetReturnsEmptyProfileForFirstTimeUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed training profile repository test in short mode")
	}

	pool, cleanup := setupTrainingProfileRepositoryTestDatabase(t)
	defer cleanup()

	const userID = "training-profile-empty-user"
	seedTrainingProfileRepositoryTestUser(t, pool, userID)

	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	profile, err := repo.Get(context.Background(), userID)

	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.Nil(t, profile.PrimaryGoal)
	assert.Empty(t, profile.AvailableEquipment)
	assert.Nil(t, profile.MovementLimitations)
}

func TestRepositoryUpsertFullReplacementAndMovementTriState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed training profile repository test in short mode")
	}

	pool, cleanup := setupTrainingProfileRepositoryTestDatabase(t)
	defer cleanup()

	const userID = "training-profile-upsert-user"
	seedTrainingProfileRepositoryTestUser(t, pool, userID)

	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	goal := "hypertrophy"
	experience := "intermediate"
	duration := int32(60)
	location := "gym"
	limitations := []string{"no deep squats"}

	profile, err := repo.Upsert(context.Background(), userID, UpdateProfileRequest{
		PrimaryGoal:                     &goal,
		ExperienceLevel:                 &experience,
		PreferredSessionDurationMinutes: &duration,
		UsualTrainingLocation:           &location,
		AvailableEquipment:              []string{"barbell", "rack"},
		AvoidedExercises:                []string{"burpees"},
		MovementLimitations:             &limitations,
	})
	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.Equal(t, "hypertrophy", *profile.PrimaryGoal)
	assert.Equal(t, []string{"barbell", "rack"}, profile.AvailableEquipment)
	require.NotNil(t, profile.MovementLimitations)
	assert.Equal(t, []string{"no deep squats"}, *profile.MovementLimitations)

	none := []string{}
	profile, err = repo.Upsert(context.Background(), userID, UpdateProfileRequest{
		AvailableEquipment:  []string{"dumbbells"},
		AvoidedExercises:    []string{},
		MovementLimitations: &none,
	})
	require.NoError(t, err)
	assert.Nil(t, profile.PrimaryGoal)
	assert.Nil(t, profile.ExperienceLevel)
	assert.Nil(t, profile.PreferredSessionDurationMinutes)
	assert.Nil(t, profile.UsualTrainingLocation)
	assert.Equal(t, []string{"dumbbells"}, profile.AvailableEquipment)
	assert.Empty(t, profile.AvoidedExercises)
	require.NotNil(t, profile.MovementLimitations)
	assert.Empty(t, *profile.MovementLimitations)

	profile, err = repo.Upsert(context.Background(), userID, UpdateProfileRequest{
		AvailableEquipment:  []string{},
		AvoidedExercises:    []string{},
		MovementLimitations: nil,
	})
	require.NoError(t, err)
	assert.Nil(t, profile.MovementLimitations)
}

func setupTrainingProfileRepositoryTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	dbURL := trainingProfileRepositoryTestDatabaseURL()
	if dbURL == "" {
		t.Skip("Skipping database-backed training profile repository test without DATABASE_URL")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("Skipping database-backed training profile repository test: %v", err)
	}

	for _, table := range []string{"users", "user_training_profile"} {
		_, err := pool.Exec(context.Background(), "ALTER TABLE "+table+" DISABLE ROW LEVEL SECURITY")
		require.NoError(t, err)
	}

	cleanupTrainingProfileRepositoryTestUsers(t, pool)

	cleanup := func() {
		cleanupTrainingProfileRepositoryTestUsers(t, pool)
		for _, table := range []string{"user_training_profile", "users"} {
			_, err := pool.Exec(context.Background(), "ALTER TABLE "+table+" ENABLE ROW LEVEL SECURITY")
			require.NoError(t, err)
		}
		pool.Close()
	}

	return pool, cleanup
}

func seedTrainingProfileRepositoryTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()

	_, err := pool.Exec(context.Background(), "INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
	require.NoError(t, err)
}

func cleanupTrainingProfileRepositoryTestUsers(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	for _, userID := range []string{"training-profile-empty-user", "training-profile-upsert-user"} {
		_, err := pool.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", userID)
		require.NoError(t, err)
	}
}

func trainingProfileRepositoryTestDatabaseURL() string {
	if os.Getenv("DATABASE_URL") != "" {
		return os.Getenv("DATABASE_URL")
	}
	_ = godotenv.Load(".env", "../.env", "../../.env")
	if os.Getenv("DATABASE_URL") != "" {
		return os.Getenv("DATABASE_URL")
	}
	return "postgres://postgres:password@localhost:5432/fittrack_test?sslmode=disable"
}
