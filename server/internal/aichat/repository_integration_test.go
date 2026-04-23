package aichat

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCompleteRun_PersistsWorkoutDraftJSONB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-complete-run-user"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)

	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "Build me a structured workout draft for today.", defaultModelName, "req-complete-run-jsonb")
	require.NoError(t, err)

	workoutFocus := "full body"
	weight := 50.0
	draft := &workout.CreateWorkoutRequest{
		Date:         "2026-04-22T12:00:00Z",
		WorkoutFocus: &workoutFocus,
		Exercises: []workout.ExerciseInput{
			{
				Name: "Goblet Squat",
				Sets: []workout.SetInput{
					{Weight: &weight, Reps: 10, SetType: "working"},
				},
			},
		},
	}

	_, run, err := repo.CompleteRun(ctx, prepared, workoutDraftSummaryMessage, draft, time.Date(2026, 4, 22, 16, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotNil(t, run)
	require.NotNil(t, run.WorkoutDraft)
	assert.Equal(t, draft.Date, run.WorkoutDraft.Date)

	storedConversation, err := repo.GetConversation(ctx, conversation.ID, userID)
	require.NoError(t, err)
	require.NotNil(t, storedConversation.LatestWorkoutDraft)
	assert.Equal(t, draft.Date, storedConversation.LatestWorkoutDraft.Date)
	require.NotNil(t, storedConversation.LatestWorkoutDraft.WorkoutFocus)
	assert.Equal(t, workoutFocus, *storedConversation.LatestWorkoutDraft.WorkoutFocus)

	expectedJSON, err := json.Marshal(draft)
	require.NoError(t, err)

	var storedRunDraft []byte
	err = pool.QueryRow(ctx, "SELECT workout_draft FROM ai_chat_run WHERE id = $1 AND user_id = $2", run.ID, userID).Scan(&storedRunDraft)
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(storedRunDraft))

	var storedLatestDraft []byte
	err = pool.QueryRow(ctx, "SELECT latest_workout_draft FROM ai_chat_conversation WHERE id = $1 AND user_id = $2", conversation.ID, userID).Scan(&storedLatestDraft)
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(storedLatestDraft))
}

func setupAIChatRepositoryTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	dbURL := getAIChatRepositoryTestDatabaseURL()
	if dbURL == "" {
		t.Skip("Skipping database-backed AI chat repository test without DATABASE_URL")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("Skipping database-backed AI chat repository test: %v", err)
	}

	tables := []string{
		"users",
		"ai_chat_conversation",
		"ai_chat_message",
		"ai_chat_run",
		"ai_chat_stream_chunk",
	}
	for _, table := range tables {
		_, err := pool.Exec(context.Background(), "ALTER TABLE "+table+" DISABLE ROW LEVEL SECURITY")
		require.NoError(t, err)
	}

	cleanup := func() {
		ctx := context.Background()
		_, err := pool.Exec(ctx, "DELETE FROM users WHERE user_id = $1", "aichat-complete-run-user")
		require.NoError(t, err)

		for _, table := range tables {
			_, err := pool.Exec(ctx, "ALTER TABLE "+table+" ENABLE ROW LEVEL SECURITY")
			require.NoError(t, err)
		}

		pool.Close()
	}

	return pool, cleanup
}

func seedAIChatRepositoryTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()

	_, err := pool.Exec(context.Background(), "INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
	require.NoError(t, err)
}

func getAIChatRepositoryTestDatabaseURL() string {
	if os.Getenv("DATABASE_URL") != "" {
		return os.Getenv("DATABASE_URL")
	}

	_ = godotenv.Load(".env", "../.env", "../../.env")
	if os.Getenv("DATABASE_URL") != "" {
		return os.Getenv("DATABASE_URL")
	}

	return "postgres://postgres:password@localhost:5432/fittrack_test?sslmode=disable"
}
