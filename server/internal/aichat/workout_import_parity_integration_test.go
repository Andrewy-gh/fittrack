package aichat

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Compares the real stored-chat save with the normal create endpoint's service
// boundary using the same synthetic request. Client import is exercised separately.
func TestStoredDraftAndWorkoutCreatePreserveContext(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	defer cleanup()
	const owner = "a059-parity-owner"
	const other = "a059-parity-other"
	ctx := user.WithContext(context.Background(), owner)
	seedAIChatRepositoryTestUser(t, pool, owner)
	seedAIChatRepositoryTestUser(t, pool, other)
	defer func() {
		_, err := pool.Exec(context.Background(), "DELETE FROM users WHERE user_id IN ($1,$2)", owner, other)
		require.NoError(t, err)
	}()
	var exerciseID int32
	err := pool.QueryRow(ctx, `INSERT INTO exercise (user_id,name,catalog_id) VALUES ($1,$2,'Pushups') RETURNING id`, owner, "Parity pushups").Scan(&exerciseID)
	require.NoError(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	chatRepo := NewRepository(logger, queries, pool)
	access := new(mockFeatureAccessService)
	access.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil)
	chat := NewService(logger, access, nil, chatRepo, workout.NewTxSaver(logger, exerciseRepo))
	workouts := workout.NewService(logger, workout.NewRepository(logger, queries, pool, exerciseRepo))
	conversation, err := chatRepo.CreateConversation(ctx, owner)
	require.NoError(t, err)
	prepared, err := chatRepo.PrepareMessageStream(ctx, conversation.ID, owner, "Synthetic draft", "synthetic", "a059-parity")
	require.NoError(t, err)
	conflictingCatalog := "Dumbbell_Bench_Press"
	zero := 0.0
	snapshot := recommendation.Snapshot{ExerciseName: "Parity pushups", Recommendation: recommendation.Result{ExerciseID: exerciseID, Readiness: "normal", Explanation: "Recorded context", PolicyVersion: recommendation.PolicyVersion}}
	draft := workout.CreateWorkoutRequest{
		Date:            "2026-09-11T12:00:00-04:00",
		Exercises:       []workout.ExerciseInput{{Name: "Parity pushups", CatalogID: &conflictingCatalog, Sets: []workout.SetInput{{Reps: 5, SetType: "warmup"}, {Reps: 8, Weight: &zero, SetType: "working"}}}},
		Recommendations: []recommendation.Snapshot{snapshot},
	}
	_, run, err := chatRepo.CompleteRun(ctx, prepared, "Synthetic", &draft, time.Now())
	require.NoError(t, err)
	before, err := json.Marshal(run.WorkoutDraft)
	require.NoError(t, err)
	direct, err := chat.SaveLatestWorkoutDraft(ctx, conversation.ID)
	require.NoError(t, err)
	importedID, err := workouts.CreateWorkoutWithID(ctx, draft)
	require.NoError(t, err)
	directDetail, err := workouts.GetWorkoutWithSets(ctx, direct.WorkoutID)
	require.NoError(t, err)
	importedDetail, err := workouts.GetWorkoutWithSets(ctx, importedID)
	require.NoError(t, err)
	require.Equal(t, []recommendation.Snapshot{snapshot}, directDetail.Recommendations)
	require.Equal(t, directDetail.Recommendations, importedDetail.Recommendations)
	require.Len(t, directDetail.Sets, 2)
	require.Len(t, importedDetail.Sets, 2)
	for i := range directDetail.Sets {
		a, b := directDetail.Sets[i], importedDetail.Sets[i]
		require.Equal(t, a.ExerciseID, b.ExerciseID)
		require.Equal(t, a.Reps, b.Reps)
		require.Equal(t, a.SetType, b.SetType)
		require.Equal(t, a.Weight, b.Weight)
		require.Equal(t, a.WorkoutDate, b.WorkoutDate)
	}
	var storedCatalog string
	require.NoError(t, pool.QueryRow(ctx, `SELECT catalog_id FROM exercise WHERE id=$1 AND user_id=$2`, exerciseID, owner).Scan(&storedCatalog))
	require.Equal(t, "Pushups", storedCatalog, "an existing owned link remains authoritative")
	err = workouts.UpdateWorkout(ctx, importedID, workout.UpdateWorkoutRequest{
		Date:      draft.Date,
		Exercises: []workout.ExerciseInput{{Name: "Parity pushups", Sets: []workout.SetInput{{Reps: 12, SetType: "working"}}}},
	})
	require.NoError(t, err)
	editedDetail, err := workouts.GetWorkoutWithSets(ctx, importedID)
	require.NoError(t, err)
	require.Equal(t, []recommendation.Snapshot{snapshot}, editedDetail.Recommendations, "PUT retains historical context")
	require.Equal(t, int32(12), editedDetail.Sets[0].Reps)
	after, err := chatRepo.GetConversation(ctx, conversation.ID, owner)
	require.NoError(t, err)
	afterJSON, err := json.Marshal(after.LatestWorkoutDraft)
	require.NoError(t, err)
	require.JSONEq(t, string(before), string(afterJSON), "saving must not rewrite the original draft")
	_, err = workouts.GetWorkoutWithSets(user.WithContext(context.Background(), other), importedID)
	require.Error(t, err, "another owner must not retrieve this workout")
	var beforeCount, afterCount int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM workout WHERE user_id=$1`, owner).Scan(&beforeCount))
	invalid := draft
	invalid.Recommendations = []recommendation.Snapshot{{ExerciseName: snapshot.ExerciseName, Recommendation: recommendation.Result{ExerciseID: exerciseID + 100000, Readiness: "normal", Explanation: "Wrong owner or exercise", PolicyVersion: recommendation.PolicyVersion}}}
	_, err = workouts.CreateWorkoutWithID(ctx, invalid)
	require.ErrorIs(t, err, recommendation.ErrInvalidContext)
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM workout WHERE user_id=$1`, owner).Scan(&afterCount))
	require.Equal(t, beforeCount, afterCount, "mismatched context must roll back the workout")
}
