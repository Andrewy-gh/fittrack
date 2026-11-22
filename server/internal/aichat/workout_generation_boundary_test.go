package aichat

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const ordinaryDraftJSON = `{"date":"2026-01-01T12:00:00Z","notes":"Rest briefly","workoutFocus":"push","exercises":[{"name":"Pushups","sets":[{"weight":5,"reps":6,"setType":"warmup"},{"reps":8,"setType":"working"}]}]}`

func syntheticDraftModel(ctx context.Context, reply func(*ai.ModelRequest) string) *genkit.Genkit {
	g := genkit.Init(ctx)
	genkit.DefineModel(g, "synthetic", &ai.ModelOptions{Supports: &ai.ModelSupports{Output: []string{"json"}, Constrained: ai.ConstrainedSupportAll, Multiturn: true, SystemRole: true}}, func(_ context.Context, req *ai.ModelRequest, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
		return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewTextPart(reply(req))), FinishReason: ai.FinishReasonStop}, nil
	})
	return g
}

func TestGeneratedDraftModelBoundary(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name, payload, forbidden string
	}{
		{"ordinary", ordinaryDraftJSON, ""},
		{"fabricated recommendation", `{"date":"2026-01-01T12:00:00Z","exercises":[{"name":"Pushups","sets":[{"reps":8,"setType":"working"}]}],"recommendations":[{"exerciseName":"Pushups","recommendation":{"exerciseId":2147483647,"readiness":"normal","explanation":"synthetic","policyVersion":"exercise-plan-v2"}}]}`, "recommendations"},
		{"fabricated catalog ID", `{"date":"2026-01-01T12:00:00Z","exercises":[{"name":"Pushups","catalog_id":"invented-id","sets":[{"reps":8,"setType":"working"}]}]}`, "catalog_id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests []*ai.ModelRequest
			g := syntheticDraftModel(ctx, func(req *ai.ModelRequest) string {
				requests = append(requests, req)
				return tc.payload
			})
			draft, err := generateWorkoutDraft(ctx, g, "synthetic", WorkoutGenerationToolInput{
				Equipment: "bodyweight", SessionDuration: 10, WorkoutFocus: "push", Injuries: "none",
			}, time.Now())
			require.Len(t, requests, 1, "schema errors do not trigger quality repair")
			require.NotNil(t, requests[0].Output)
			schema, marshalErr := json.Marshal(requests[0].Output.Schema)
			require.NoError(t, marshalErr)
			for _, property := range []string{"date", "notes", "workoutFocus", "exercises", "name", "sets", "weight", "reps", "setType"} {
				require.Contains(t, string(schema), strconv.Quote(property))
			}
			require.NotContains(t, string(schema), `"recommendations"`)
			require.NotContains(t, string(schema), `"catalog_id"`)
			if tc.forbidden != "" {
				require.ErrorContains(t, err, tc.forbidden)
				require.Nil(t, draft)
			} else {
				require.NoError(t, err)
				raw, err := json.Marshal(draft)
				require.NoError(t, err)
				require.JSONEq(t, tc.payload, string(raw))
			}
		})
	}
}

func TestGeneratedAndLegacyDraftSaveBoundary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed draft save boundary in short mode")
	}
	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	defer cleanup()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool)
	saver := workout.NewTxSaver(logger, exercise.NewRepository(logger, queries, pool))

	for i, tc := range []struct {
		name, payload, wantLoggedError string
		wantCatalog                    *string
	}{
		{name: "ordinary generated"},
		{name: "legacy valid catalog link", payload: `{"date":"2026-01-01T12:00:00Z","exercises":[{"name":"Pushups","catalog_id":"Pushups","sets":[{"reps":8,"setType":"working"}]}]}`, wantCatalog: new("Pushups")},
		{name: "legacy fabricated recommendation", payload: `{"date":"2026-01-01T12:00:00Z","exercises":[{"name":"Pushups","sets":[{"reps":8,"setType":"working"}]}],"recommendations":[{"exerciseName":"Pushups","recommendation":{"exerciseId":2147483647,"readiness":"normal","explanation":"synthetic","policyVersion":"exercise-plan-v2"}}]}`, wantLoggedError: "recommendation does not match logged exercise"},
		{name: "legacy fabricated catalog ID", payload: `{"date":"2026-01-01T12:00:00Z","exercises":[{"name":"Pushups","catalog_id":"invented-id","sets":[{"reps":8,"setType":"working"}]}]}`, wantLoggedError: "invalid or conflicting exercise catalog selection"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			userID := "aichat-draft-boundary-" + strconv.Itoa(i)
			ctx := user.WithContext(context.Background(), userID)
			seedAIChatRepositoryTestUser(t, pool, userID)
			t.Cleanup(func() {
				_, err := pool.Exec(ctx, "DELETE FROM users WHERE user_id = $1", userID)
				require.NoError(t, err)
			})

			var draft *workout.CreateWorkoutRequest
			var err error
			payload := tc.payload
			if payload == "" {
				payload = ordinaryDraftJSON
				g := syntheticDraftModel(ctx, func(*ai.ModelRequest) string { return ordinaryDraftJSON })
				draft, err = generateWorkoutDraft(ctx, g, "synthetic", WorkoutGenerationToolInput{
					Equipment: "bodyweight", SessionDuration: 10, WorkoutFocus: "push", Injuries: "none",
				}, time.Now())
			} else {
				// Simulate a draft written before the model schema was narrowed.
				draft = new(workout.CreateWorkoutRequest)
				err = json.Unmarshal([]byte(payload), draft)
			}
			require.NoError(t, err)

			conversation, err := repo.CreateConversation(ctx, userID)
			require.NoError(t, err)
			prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "Draft a workout", "synthetic", "draft-boundary")
			require.NoError(t, err)
			_, _, err = repo.CompleteRun(ctx, prepared, workoutDraftSummaryMessage, draft, time.Now())
			require.NoError(t, err)

			access := new(mockFeatureAccessService)
			access.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil)
			service := NewService(logger, access, nil, repo, saver)
			detail, err := service.GetConversation(ctx, conversation.ID)
			require.NoError(t, err)
			stored, err := json.Marshal(detail.Conversation.LatestWorkoutDraft)
			require.NoError(t, err)
			require.JSONEq(t, payload, string(stored))

			req := httptest.NewRequest(http.MethodPost, "/ai/conversations/"+strconv.Itoa(int(conversation.ID))+"/latest-workout-draft/save", nil).WithContext(ctx)
			req.SetPathValue("id", strconv.Itoa(int(conversation.ID)))
			rec := httptest.NewRecorder()
			NewHandler(logger, service).SaveLatestWorkoutDraft(rec, req)
			if tc.wantLoggedError != "" {
				require.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
				_, err := service.SaveLatestWorkoutDraft(ctx, conversation.ID)
				require.ErrorContains(t, err, tc.wantLoggedError)
				var workouts, exercises, sets int
				require.NoError(t, pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM workout WHERE user_id = $1), (SELECT count(*) FROM exercise WHERE user_id = $1), (SELECT count(*) FROM "set" WHERE user_id = $1)`, userID).Scan(&workouts, &exercises, &sets))
				require.Zero(t, workouts)
				require.Zero(t, exercises)
				require.Zero(t, sets)
				detail, err := service.GetConversation(ctx, conversation.ID)
				require.NoError(t, err)
				require.False(t, detail.Conversation.LatestWorkoutDraftStatus.IsSaved)
				return
			}

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var saved SaveLatestWorkoutDraftResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &saved))
			require.Positive(t, saved.WorkoutID)
			require.True(t, saved.Conversation.LatestWorkoutDraftStatus.IsSaved)
			var catalogID *string
			var reps int
			require.NoError(t, pool.QueryRow(ctx, `SELECT e.catalog_id, s.reps FROM workout w JOIN "set" s ON s.workout_id = w.id JOIN exercise e ON e.id = s.exercise_id WHERE w.id = $1 AND w.user_id = $2 AND s.set_type = 'working'`, saved.WorkoutID, userID).Scan(&catalogID, &reps))
			require.Equal(t, tc.wantCatalog, catalogID)
			require.Equal(t, 8, reps)
		})
	}
}
