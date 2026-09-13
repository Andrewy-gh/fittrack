package recommendation_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/url"
	"os"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// Uses rollback-only fixtures and the restricted runtime role. Never run on a remote database.
func TestRecommendationPersistenceWithRLS(t *testing.T) {
	if testing.Short() {
		t.Skip("local database integration")
	}
	databaseURL := os.Getenv("RECOMMENDATION_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("RECOMMENDATION_TEST_DATABASE_URL required; use the provisioned restricted-runtime database")
	}
	parsed, err := url.Parse(databaseURL)
	require.NoError(t, err)
	require.Contains(t, []string{"localhost", "127.0.0.1"}, parsed.Hostname(), "requires a verified local database")
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	require.NoError(t, err)
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, "SET LOCAL ROLE fittrack_app")
	require.NoError(t, err, "provision the restricted runtime role first")

	owner := "recommendation-integration-owner"
	other := "recommendation-integration-other"
	setOwner := func(id string) {
		_, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id',$1,true)", id)
		require.NoError(t, err)
	}
	setOwner(owner)
	queries := db.New(tx)
	_, err = queries.CreateUser(ctx, owner)
	require.NoError(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	press, err := exerciseRepo.GetOrCreateExerciseTx(ctx, queries, "Recommendation press", owner)
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Second)
	saver := workout.NewTxSaver(logger, exerciseRepo)
	save := func(date time.Time, sets []workout.SetInput, snapshots []recommendation.Snapshot) int32 {
		id, err := saver.SaveWorkoutTx(ctx, queries, workout.CreateWorkoutRequest{
			Date:            date.Format(time.RFC3339),
			Exercises:       []workout.ExerciseInput{{Name: press.Name, Sets: sets}},
			Recommendations: snapshots,
		}, owner)
		require.NoError(t, err)
		return id
	}
	warmup := workout.SetInput{Reps: 5, SetType: "warmup"}
	weight := 100.0
	loaded := workout.SetInput{Weight: &weight, Reps: 8, SetType: "working"}
	_, err = tx.Exec(ctx, `INSERT INTO user_training_profile(user_id,primary_goal,goals,experience_level) VALUES($1,'strength',ARRAY['strength'],'intermediate')`, owner)
	require.NoError(t, err)
	save(now.Add(-2*time.Hour), []workout.SetInput{warmup, loaded, loaded}, nil)
	save(now.Add(-time.Hour), []workout.SetInput{loaded, loaded}, nil)

	service := recommendation.NewService(recommendation.NewRepository(queries), func() time.Time { return now })
	ownerContext := user.WithContext(ctx, owner)
	result, err := service.Get(ownerContext, press.ID, "normal")
	require.NoError(t, err)
	require.Equal(t, 2, result.Plan.Sets)
	require.Equal(t, 102.5, *result.Plan.Weight)
	require.Equal(t, 5, result.Plan.Reps)

	snapshot := recommendation.Snapshot{ExerciseName: press.Name, Recommendation: result}
	workoutID := save(now, []workout.SetInput{loaded}, []recommendation.Snapshot{snapshot})
	raw, err := queries.GetWorkoutRecommendationContext(ctx, db.GetWorkoutRecommendationContextParams{ID: workoutID, UserID: owner})
	require.NoError(t, err)
	var recorded []recommendation.Snapshot
	require.NoError(t, json.Unmarshal(raw, &recorded))
	require.Equal(t, []recommendation.Snapshot{snapshot}, recorded)
	sets, err := queries.GetWorkoutWithSets(ctx, db.GetWorkoutWithSetsParams{ID: workoutID, UserID: owner})
	require.NoError(t, err)
	require.Len(t, sets, 1, "guidance must never insert actual sets")

	setOwner(other)
	_, err = queries.CreateUser(ctx, other)
	require.NoError(t, err)
	_, err = service.Get(user.WithContext(ctx, other), press.ID, "normal")
	require.ErrorIs(t, err, recommendation.ErrNotFound)
	_, err = queries.GetWorkoutRecommendationContext(ctx, db.GetWorkoutRecommendationContextParams{ID: workoutID, UserID: other})
	require.Error(t, err)
}
