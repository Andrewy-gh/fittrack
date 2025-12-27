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
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/auth"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/middleware"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/Andrewy-gh/fittrack/server/docs" // This line is necessary for go-swagger to find docs
)

type api struct {
	logger  *slog.Logger
	queries *db.Queries
	pool    *pgxpool.Pool
	cfg     *config.Config
}

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(fmt.Sprintf("invalid duration: %s", s))
	}
	return d
}

func main() {
	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	// Parse log level
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	ctx := context.Background()

	// Log startup configuration summary
	logger.Info("starting application",
		"environment", cfg.Environment,
		"port", cfg.Port,
		"log_level", cfg.LogLevel,
		"db_max_conns", cfg.DBMaxConns,
		"rate_limit_rpm", cfg.RateLimitRPM,
	)

	logger.Info("connecting to database")
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to parse database config", "error", err)
		os.Exit(1)
	}
	// Disable prepared statements (simple protocol) for PgBouncer transaction pooling / Supabase pooler
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	// Configure pool caps from config
	poolConfig.MaxConns = cfg.DBMaxConns
	poolConfig.MinConns = cfg.DBMinConns
	poolConfig.MaxConnIdleTime = mustParseDuration(cfg.DBMaxConnIdle)
	poolConfig.MaxConnLifetime = mustParseDuration(cfg.DBMaxConnLife)
	poolConfig.HealthCheckPeriod = mustParseDuration(cfg.DBHealthCheck)

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
	healthHandler := health.NewHandler(logger, pool)

	api := &api{
		logger:  logger,
		queries: queries,
		pool:    pool,
		cfg:     cfg,
	}

	jwks, err := auth.NewJWKSCache(ctx, cfg.ProjectID)
	if err != nil {
		logger.Error("failed to create JWKS cache", "error", err)
		os.Exit(1)
	}

	authenticator := auth.NewAuthenticator(logger, jwks, userService, pool)
	router := api.routes(workoutHandler, exerciseHandler, healthHandler)

	// Apply middleware in order: SecurityHeaders → CORS → RequestID → Metrics → RateLimit → Authentication
	var handler http.Handler = router
	handler = middleware.RateLimit(logger, int64(cfg.RateLimitRPM))(handler)
	handler = authenticator.Middleware(handler)
	handler = middleware.Metrics()(handler)
	handler = middleware.RequestID()(handler)
	handler = middleware.CORS(cfg.GetAllowedOrigins(), logger)(handler)
	handler = middleware.SecurityHeaders()(handler)

	// Configure HTTP server with timeouts
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	case sig := <-sigChan:
		logger.Info("shutdown signal received", "signal", sig.String())

		// Create context with 30-second timeout for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		logger.Info("draining connections")

		// Attempt graceful shutdown
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("forced shutdown", "error", err)
			pool.Close()
			os.Exit(1)
		}

		// Close database connections cleanly
		pool.Close()
		logger.Info("shutdown complete")
	}
}
