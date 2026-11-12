package exercise

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestExerciseListAndCreateReturnPersistedTimestamps(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewRepository(logger, db.New(pool), pool)
	handler := NewHandler(logger, validator.New(), NewService(logger, repo))
	ctx := user.WithContext(context.Background(), exerciseTestUserA)
	type exerciseJSON struct {
		ID        int32     `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	post := func() exerciseJSON {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/api/exercises", strings.NewReader(`{"name":"Timestamped exercise"}`)).WithContext(ctx)
		w := httptest.NewRecorder()
		handler.GetOrCreateExercise(w, req)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var result exerciseJSON
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		require.False(t, result.CreatedAt.IsZero(), "created_at must be a non-null timestamp")
		require.False(t, result.UpdatedAt.IsZero(), "updated_at must be a non-null timestamp")
		return result
	}
	checkReads := func(want exerciseJSON) {
		t.Helper()
		var storedCreated, storedUpdated time.Time
		err := pool.QueryRow(ctx, "SELECT created_at, updated_at FROM exercise WHERE id=$1", want.ID).Scan(&storedCreated, &storedUpdated)
		require.NoError(t, err)
		require.True(t, want.CreatedAt.Equal(storedCreated))
		require.True(t, want.UpdatedAt.Equal(storedUpdated))
		w := httptest.NewRecorder()
		handler.ListExercises(w, httptest.NewRequest(http.MethodGet, "/api/exercises", nil).WithContext(ctx))
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []exerciseJSON
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.Len(t, listed, 1)
		require.Equal(t, want, listed[0])
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/exercises/%d", want.ID), nil).WithContext(ctx)
		req.SetPathValue("id", fmt.Sprint(want.ID))
		w = httptest.NewRecorder()
		handler.GetExerciseWithSets(w, req)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var detail struct {
			Exercise exerciseJSON `json:"exercise"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &detail))
		require.Equal(t, want, detail.Exercise)
	}
	created := post()
	checkReads(created)

	// Distinct persisted values catch copying created_at into updated_at or
	// manufacturing a fresh update time on a get-existing request.
	createdAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := createdAt.Add(48 * time.Hour)
	_, err := pool.Exec(ctx, "UPDATE exercise SET created_at=$1, updated_at=$2 WHERE id=$3", createdAt, updatedAt, created.ID)
	require.NoError(t, err)
	existing := post()
	require.Equal(t, created.ID, existing.ID)
	require.True(t, existing.CreatedAt.Equal(createdAt))
	require.True(t, existing.UpdatedAt.Equal(updatedAt))
	checkReads(existing)

	// Workout creation has separate custom and catalog transaction paths.
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	qtx := db.New(tx)
	transactional, err := repo.GetOrCreateExerciseTx(ctx, qtx, "Custom exercise", exerciseTestUserA)
	require.NoError(t, err)
	require.True(t, transactional.CreatedAt.Valid)
	require.True(t, transactional.UpdatedAt.Valid)
	catalogID, err := qtx.GetOrCreateCatalogExercise(ctx, db.GetOrCreateCatalogExerciseParams{
		Name: "Catalog exercise", UserID: exerciseTestUserA,
		CatalogID: pgtype.Text{String: "Pushups", Valid: true},
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))
	stored, err := repo.GetExerciseDetail(ctx, transactional.ID, exerciseTestUserA)
	require.NoError(t, err)
	require.Equal(t, transactional.CreatedAt, stored.CreatedAt)
	require.Equal(t, transactional.UpdatedAt, stored.UpdatedAt)
	catalog, err := repo.GetExerciseDetail(ctx, catalogID, exerciseTestUserA)
	require.NoError(t, err)
	require.Equal(t, "Pushups", catalog.CatalogID.String)
	require.True(t, catalog.CreatedAt.Valid)
	require.True(t, catalog.UpdatedAt.Valid)
	require.False(t, catalog.CreatedAt.Time.IsZero())
	require.False(t, catalog.UpdatedAt.Time.IsZero())
}

func TestExerciseTimestampBackfillPreservesKnownHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	// Recreate the pre-backfill state transactionally; rollback restores schema
	// and fixtures even when an assertion fails.
	_, err = tx.Exec(ctx, "ALTER TABLE exercise ALTER COLUMN updated_at DROP NOT NULL")
	require.NoError(t, err)
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	updated := created.Add(48 * time.Hour)
	_, err = tx.Exec(ctx, `INSERT INTO exercise (name, user_id, created_at, updated_at)
		VALUES ('Legacy null', $1, $2, NULL), ('Known history', $1, $2, $3)`, exerciseTestUserA, created, updated)
	require.NoError(t, err)
	migration, err := os.ReadFile("../../migrations/00035_backfill_exercise_updated_at.sql")
	require.NoError(t, err)
	up := strings.SplitN(string(migration), "-- +goose Down", 2)[0]
	_, err = tx.Exec(ctx, up)
	require.NoError(t, err)
	for name, want := range map[string]time.Time{"Legacy null": created, "Known history": updated} {
		var got time.Time
		err = tx.QueryRow(ctx, "SELECT updated_at FROM exercise WHERE name=$1 AND user_id=$2", name, exerciseTestUserA).Scan(&got)
		require.NoError(t, err)
		require.True(t, want.Equal(got), name)
	}
	_, err = tx.Exec(ctx, "INSERT INTO exercise (name, user_id, updated_at) VALUES ('Explicit null', $1, NULL)", exerciseTestUserA)
	require.Error(t, err, "the persisted contract must reject explicit null timestamps")
}
