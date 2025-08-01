package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/Andrewy-gh/fittrack/server/internal/auth"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type api struct {
	logger  *slog.Logger
	queries *db.Queries
	pool    *pgxpool.Pool
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()

	logger.Info("connecting to database")
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	logger.Info("database connection successful")

	validator := validator.New()

	// Initialize database queries
	queries := db.New(pool)

	// Initialize repositories
	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	workoutRepo := workout.NewRepository(logger, queries, pool, exerciseRepo)
	userRepo := user.NewRepository(logger, queries, pool)

	// Initialize services
	workoutService := workout.NewService(logger, workoutRepo)
	exerciseService := exercise.NewService(logger, exerciseRepo)
	userService := user.NewService(logger, userRepo)

	// Initialize handlers
	workoutHandler := workout.NewHandler(logger, validator, workoutService)
	exerciseHandler := exercise.NewHandler(logger, validator, exerciseService)

	api := &api{
		logger:  logger,
		queries: queries,
		pool:    pool,
	}

	logger.Info("starting server", "addr", ":8080")

	// Initialize router with auth middleware
	router := api.routes(workoutHandler, exerciseHandler)

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		logger.Error("missing required environment variables", "has_project_id", projectID != "")
		os.Exit(1)
	}

	jwks, err := auth.NewJWKSCache(ctx, projectID)
	if err != nil {
		logger.Error("failed to create JWKS", "error", err)
		os.Exit(1)
	}

	err = http.ListenAndServe(":8080", auth.Middleware(router, logger, jwks, userService))
	if err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
