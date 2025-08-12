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

// @securityDefinitions.apikey StackAuth
// @in header
// @name x-stack-access-token

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/auth"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/Andrewy-gh/fittrack/server/docs" // This line is necessary for go-swagger to find docs
)

const (
	defaultMaxConns       = int32(15)
	defaultMinConns       = int32(2)
)

var (
	defaultMaxConnIdle     = 30 * time.Second
	defaultMaxConnLifetime = 30 * time.Minute
	defaultHealthCheck     = 30 * time.Second
)

type api struct {
	logger  *slog.Logger
	queries *db.Queries
	pool    *pgxpool.Pool
}

func envInt32(name string, def int32) int32 {
	if v := os.Getenv(name); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			return int32(i)
		}
	}
	return def
}

func envDuration(name string, def time.Duration) time.Duration {
	if v := os.Getenv(name); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()

	logger.Info("connecting to database")
	dbURL := os.Getenv("DATABASE_URL")
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logger.Error("failed to parse database config", "error", err)
		os.Exit(1)
	}
	// Disable prepared statements (simple protocol) for PgBouncer transaction pooling / Supabase pooler
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	// Configure pool caps with env overrides
	poolConfig.MaxConns = envInt32("DB_MAX_CONNS", defaultMaxConns)
	poolConfig.MinConns = envInt32("DB_MIN_CONNS", defaultMinConns)
	poolConfig.MaxConnIdleTime = envDuration("DB_MAX_CONN_IDLE", defaultMaxConnIdle)
	poolConfig.MaxConnLifetime = envDuration("DB_MAX_CONN_LIFE", defaultMaxConnLifetime)
	poolConfig.HealthCheckPeriod = envDuration("DB_HEALTHCHECK", defaultHealthCheck)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
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
