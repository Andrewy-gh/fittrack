// @title FitTrack API
// @version 1.0
// @description A fitness tracking application API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name x-stack-access-token

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

	_ "github.com/Andrewy-gh/fittrack/server/docs" // This line is necessary for go-swagger to find docs
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

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		logger.Error("missing required environment variables", "has_project_id", projectID != "")
		os.Exit(1)
	}

	jwks, err := auth.NewJWKSCache(ctx, projectID)
	if err != nil {
		logger.Error("failed to create JWKS cache", "error", err)
		os.Exit(1)
	}

	authenticator := auth.NewAuthenticator(logger, jwks, userService, pool)
	router := api.routes(workoutHandler, exerciseHandler)

	logger.Info("starting server", "addr", ":8080")
	err = http.ListenAndServe(":8080", authenticator.Middleware(router))
	if err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
