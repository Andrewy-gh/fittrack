package workout

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestWorkoutNumericStorageBoundaries(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	q := db.New(pool)
	svc := NewService(logger, NewRepository(logger, q, pool, exercise.NewRepository(logger, q, pool)))
	h := NewHandler(logger, validator.New(), svc)
	ctx := user.WithContext(context.Background(), "test-user-a")
	body := func(weight string, reps int64, typ string) string {
		return fmt.Sprintf(`{"date":"2026-01-01T00:00:00Z","exercises":[{"name":"Boundary press","sets":[{"weight":%s,"reps":%d,"setType":"%s"}]}]}`, weight, reps, typ)
	}
	post := func(payload string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.CreateWorkout(w, httptest.NewRequest(http.MethodPost, "/api/workouts", strings.NewReader(payload)).WithContext(ctx))
		return w
	}
	require.Equal(t, http.StatusBadRequest, post(body("0", 2147483648, "working")).Code)
	require.Equal(t, http.StatusBadRequest, post(body("999999.9", 1, "working")).Code)
	mixed := `{"date":"2026-01-01T00:00:00Z","exercises":[{"name":"Before invalid","sets":[{"weight":10,"reps":2,"setType":"working"}]},{"name":"Invalid","sets":[{"weight":0,"reps":2147483648,"setType":"working"}]}]}`
	require.Equal(t, http.StatusBadRequest, post(mixed).Code)
	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM workout WHERE user_id=$1`, "test-user-a").Scan(&count))
	require.Zero(t, count)
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM exercise WHERE user_id=$1`, "test-user-a").Scan(&count))
	require.Zero(t, count)
	require.Equal(t, http.StatusOK, post(body("0", 2147483647, "working")).Code)
	var id int32
	require.NoError(t, pool.QueryRow(ctx, `SELECT id FROM workout WHERE user_id=$1`, "test-user-a").Scan(&id))
	put := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/workouts/%d", id), strings.NewReader(body("999999.9", 1, "working"))).WithContext(ctx)
	put.SetPathValue("id", fmt.Sprint(id))
	w := httptest.NewRecorder()
	h.UpdateWorkout(w, put)
	require.Equal(t, http.StatusBadRequest, w.Code)
	var reps int32
	require.NoError(t, pool.QueryRow(ctx, `SELECT reps FROM "set" WHERE workout_id=$1`, id).Scan(&reps))
	require.Equal(t, int32(2147483647), reps)

	// The accepted boundary is based on the converted tenths and on the
	// working-set PR calculation, not on a blanket repetitions limit.
	require.Equal(t, http.StatusOK, post(body("null", 2147483647, "working")).Code)
	require.Equal(t, http.StatusOK, post(body("999999999.9", 2, "warmup")).Code)
	require.Equal(t, http.StatusBadRequest, post(body("999999999.94", 2, "warmup")).Code)
	require.Equal(t, http.StatusBadRequest, post(body("999999999.96", 2, "warmup")).Code)
	require.Equal(t, http.StatusOK, post(body("909090.9", 3, "working")).Code)
	require.Equal(t, http.StatusBadRequest, post(body("909091.0", 3, "working")).Code)
	require.Equal(t, http.StatusOK, post(body("1.25", 1, "warmup")).Code)
	var rounded string
	require.NoError(t, pool.QueryRow(ctx, `SELECT weight::text FROM "set" WHERE user_id=$1 ORDER BY id DESC LIMIT 1`, "test-user-a").Scan(&rounded))
	require.Equal(t, "1.2", rounded, "the Go conversion rounds the float before PostgreSQL receives it")
}

func TestLegacyWarmupVolumeReads(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	ctx := context.Background()
	const owner = "test-user-a"
	workoutID := setupTestWorkout(t, pool, owner, "legacy")
	var exerciseID int32
	require.NoError(t, pool.QueryRow(ctx, `INSERT INTO exercise(name,user_id) VALUES ('Legacy volume',$1) RETURNING id`, owner).Scan(&exerciseID))
	_, err := pool.Exec(ctx, `INSERT INTO "set" (workout_id,exercise_id,user_id,weight,reps,set_type,exercise_order,set_order) VALUES ($1,$2,$3,999999999.9,2,'warmup',1,1)`, workoutID, exerciseID, owner)
	require.NoError(t, err)
	q := db.New(pool)
	rows, err := q.GetWorkoutWithSets(ctx, db.GetWorkoutWithSetsParams{ID: workoutID, UserID: owner})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	exerciseRows, err := q.GetExerciseWithSets(ctx, db.GetExerciseWithSetsParams{ExerciseID: exerciseID, UserID: owner})
	require.NoError(t, err)
	require.Len(t, exerciseRows, 1)
	volume, err := rows[0].Volume.Float64Value()
	require.NoError(t, err)
	require.InDelta(t, 1999999999.8, volume.Float64, 0.01)
	exerciseVolume, err := exerciseRows[0].Volume.Float64Value()
	require.NoError(t, err)
	require.InDelta(t, 1999999999.8, exerciseVolume.Float64, 0.01)
	var weight string
	require.NoError(t, pool.QueryRow(ctx, `SELECT weight::text FROM "set" WHERE workout_id=$1`, workoutID).Scan(&weight))
	require.Equal(t, "999999999.9", weight)
}

func TestLegacyWorkingPRRecomputationFailsWithoutDataLoss(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	const owner = "test-user-a"
	ctx := user.WithContext(context.Background(), owner)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	q := db.New(pool)
	svc := NewService(logger, NewRepository(logger, q, pool, exercise.NewRepository(logger, q, pool)))
	weight := 100.0
	id, err := svc.CreateWorkoutWithID(ctx, CreateWorkoutRequest{Date: "2026-01-01T00:00:00Z", Exercises: []ExerciseInput{{Name: "Legacy PR", Sets: []SetInput{{Weight: &weight, Reps: 1, SetType: "working"}}}}})
	require.NoError(t, err)
	var exerciseID int32
	require.NoError(t, pool.QueryRow(ctx, `SELECT id FROM exercise WHERE user_id=$1 AND name='Legacy PR'`, owner).Scan(&exerciseID))
	legacyID := setupTestWorkout(t, pool, owner, "legacy PR")
	_, err = pool.Exec(ctx, `INSERT INTO "set" (workout_id,exercise_id,user_id,weight,reps,set_type,exercise_order,set_order) VALUES ($1,$2,$3,999999999.9,2,'working',1,1)`, legacyID, exerciseID, owner)
	require.NoError(t, err)
	// Detail and aggregate reads must not narrow the legacy value.
	_, err = q.GetExerciseDetail(ctx, db.GetExerciseDetailParams{ID: exerciseID, UserID: owner})
	require.NoError(t, err)
	_, err = q.GetExerciseMetricsHistoryRawAll(ctx, db.GetExerciseMetricsHistoryRawAllParams{ExerciseID: exerciseID, UserID: owner})
	require.NoError(t, err)
	_, err = svc.GetContributionData(ctx)
	require.NoError(t, err)
	// Exercise recomputation and workout replacement/deletion must report the
	// same conflict without changing the existing PR, workout, or sets.
	exerciseHandler := exercise.NewHandler(logger, validator.New(), exercise.NewService(logger, exercise.NewRepository(logger, q, pool)))
	recompute := httptest.NewRequest(http.MethodPatch, "/api/exercises/1/historical-1rm", strings.NewReader(`{"mode":"recompute"}`)).WithContext(ctx)
	recompute.SetPathValue("id", fmt.Sprint(exerciseID))
	response := httptest.NewRecorder()
	exerciseHandler.UpdateExerciseHistorical1RM(response, recompute)
	require.Equal(t, http.StatusConflict, response.Code)
	workoutHandler := NewHandler(logger, validator.New(), svc)
	update := httptest.NewRequest(http.MethodPut, "/api/workouts/1", strings.NewReader(`{"date":"2026-01-03T00:00:00Z","notes":"replacement","exercises":[{"name":"Legacy PR","sets":[{"weight":50,"reps":1,"setType":"working"}]}]}`)).WithContext(ctx)
	update.SetPathValue("id", fmt.Sprint(id))
	response = httptest.NewRecorder()
	workoutHandler.UpdateWorkout(response, update)
	require.Equal(t, http.StatusConflict, response.Code)
	var originalWeight string
	require.NoError(t, pool.QueryRow(ctx, `SELECT weight::text FROM "set" WHERE workout_id=$1`, id).Scan(&originalWeight))
	require.Equal(t, "100.0", originalWeight)
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/workouts/1", nil).WithContext(ctx)
	deleteReq.SetPathValue("id", fmt.Sprint(id))
	response = httptest.NewRecorder()
	workoutHandler.DeleteWorkout(response, deleteReq)
	require.Equal(t, http.StatusConflict, response.Code)
	err = svc.DeleteWorkout(ctx, id)
	var conflict *apperrors.Historical1RMConflict
	require.ErrorAs(t, err, &conflict)
	var remaining int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM workout WHERE user_id=$1 AND id IN ($2,$3)`, owner, id, legacyID).Scan(&remaining))
	require.Equal(t, 2, remaining)
	var pr float64
	var source int32
	require.NoError(t, pool.QueryRow(ctx, `SELECT historical_1rm, historical_1rm_source_workout_id FROM exercise WHERE id=$1`, exerciseID).Scan(&pr, &source))
	require.InDelta(t, 103.33, pr, 0.01)
	require.Equal(t, id, source)
}
