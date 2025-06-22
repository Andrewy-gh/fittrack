package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5/pgxpool"
)

type api struct {
	logger  *slog.Logger
	queries *db.Queries
	pool    *pgxpool.Pool
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("failed to connect to database", "error", err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	// Test the connection pool
	if err := pool.Ping(ctx); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	// Initialize dependencies with repository pattern
	queries := db.New(pool)

	// Create repositories
	workoutRepo := workout.NewRepository(queries, pool)
	exerciseRepo := exercise.NewRepository(queries, pool)

	// Create services with repositories
	workoutService := workout.NewService(logger, workoutRepo)
	workoutHandler := workout.NewHandler(workoutService)

	exerciseService := exercise.NewService(logger, exerciseRepo)
	exerciseHandler := exercise.NewHandler(exerciseService)

	api := &api{
		logger:  logger,
		queries: queries,
		pool:    pool,
	}

	logger.Info("starting server", "addr", ":8080")

	err = http.ListenAndServe(":8080", api.routes(workoutHandler, exerciseHandler))
	logger.Error(err.Error())
	os.Exit(1)
}
