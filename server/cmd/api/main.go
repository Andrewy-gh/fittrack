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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // Set minimum log level
	}))

	ctx := context.Background()

	logger.Info("connecting to database")
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
	logger.Info("database connection successful")

	// Initialize dependencies with repository pattern
	queries := db.New(pool)

	workoutRepo := workout.NewRepository(logger, queries, pool)
	exerciseRepo := exercise.NewRepository(logger, queries, pool)

	workoutService := workout.NewService(logger, workoutRepo)
	exerciseService := exercise.NewService(logger, exerciseRepo)

	workoutHandler := workout.NewHandler(logger, workoutService)
	exerciseHandler := exercise.NewHandler(logger, exerciseService)

	api := &api{
		logger:  logger,
		queries: queries,
		pool:    pool,
	}

	logger.Info("starting server", "addr", ":8080")

	err = http.ListenAndServe(":8080", api.routes(workoutHandler, exerciseHandler))
	if err != nil {
		logger.Error("server failed", "error", err.Error())
		os.Exit(1)
	}
}
