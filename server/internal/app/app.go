package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/account"
	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/auth"
	"github.com/Andrewy-gh/fittrack/server/internal/billing"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/e2eauth"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/middleware"
	"github.com/Andrewy-gh/fittrack/server/internal/trainingprofile"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App owns the API's constructed dependencies and process lifecycle.
type App struct {
	logger        *slog.Logger
	pool          *pgxpool.Pool
	apiServer     *http.Server
	metricsServer *http.Server
}

type api struct {
	logger         *slog.Logger
	queries        *db.Queries
	pool           *pgxpool.Pool
	cfg            *config.Config
	inngestHandler http.Handler
}

// New constructs the complete API application without starting its servers.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	logger := newLogger(cfg.LogLevel)
	logger.Info("starting application",
		"environment", cfg.Environment,
		"port", cfg.Port,
		"metrics_port", cfg.MetricsPort,
		"log_level", cfg.LogLevel,
		"db_max_conns", cfg.DBMaxConns,
		"rate_limit_rpm", cfg.RateLimitRPM,
	)

	pool, err := openDatabase(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}

	apiHandler, metricsHandler, err := buildHandlers(ctx, cfg, logger, pool)
	if err != nil {
		pool.Close()
		return nil, err
	}

	return &App{
		logger:        logger,
		pool:          pool,
		apiServer:     newHTTPServer(cfg, apiHandler),
		metricsServer: newMetricsServer(cfg, metricsHandler),
	}, nil
}

func newLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}

func openDatabase(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*pgxpool.Pool, error) {
	logger.Info("connecting to database")
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	maxIdle, err := time.ParseDuration(cfg.DBMaxConnIdle)
	if err != nil {
		return nil, fmt.Errorf("parse DB_MAX_CONN_IDLE: %w", err)
	}
	maxLifetime, err := time.ParseDuration(cfg.DBMaxConnLife)
	if err != nil {
		return nil, fmt.Errorf("parse DB_MAX_CONN_LIFE: %w", err)
	}
	healthCheck, err := time.ParseDuration(cfg.DBHealthCheck)
	if err != nil {
		return nil, fmt.Errorf("parse DB_HEALTHCHECK: %w", err)
	}

	// Supabase's PgBouncer transaction pooler does not support prepared statements.
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	poolConfig.MaxConns = cfg.DBMaxConns
	poolConfig.MinConns = cfg.DBMinConns
	poolConfig.MaxConnIdleTime = maxIdle
	poolConfig.MaxConnLifetime = maxLifetime
	poolConfig.HealthCheckPeriod = healthCheck

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info("database connection successful")
	return pool, nil
}

func buildHandlers(ctx context.Context, cfg *config.Config, logger *slog.Logger, pool *pgxpool.Pool) (http.Handler, http.Handler, error) {
	queries := db.New(pool)
	validate := validator.New()

	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	featureAccessRepo := featureaccess.NewRepository(logger, queries)
	accountRepo := account.NewRepository(logger, queries)
	billingRepo := billing.NewRepository(logger, queries, pool)
	trainingProfileRepo := trainingprofile.NewRepository(logger, queries, pool)
	workoutRepo := workout.NewRepository(logger, queries, pool, exerciseRepo)
	userRepo := user.NewRepository(logger, queries, pool)
	workoutTxSaver := workout.NewTxSaver(logger, exerciseRepo)

	workoutService := workout.NewService(logger, workoutRepo)
	exerciseService := exercise.NewService(logger, exerciseRepo)
	featureAccessService := featureaccess.NewService(logger, featureAccessRepo)
	billingService := billing.NewService(
		logger,
		billingRepo,
		cfg.StripeSecretKey,
		cfg.StripeWebhookSecret,
		cfg.StripePremiumPriceID,
		cfg.AppBaseURL,
		cfg.AIChatTrialPromptCap,
	)
	trainingProfileService := trainingprofile.NewService(logger, trainingProfileRepo)
	accountService := account.NewService(logger, accountRepo, billingService)
	userService := user.NewService(logger, userRepo)
	aiChatRepo := aichat.NewRepository(logger, queries, pool, cfg.AIChatTrialPromptCap)
	aiChatRuntime := aichat.NewGenkitRuntime(ctx, aiChatRepo)
	aiChatService := aichat.NewService(logger, featureAccessService, aiChatRuntime, aiChatRepo, workoutTxSaver)

	var inngestRecovery *aichat.InngestRecovery
	var err error
	switch {
	case cfg.AIChatRecoveryConfigured():
		inngestRecovery, err = aichat.NewInngestRecovery(logger, aiChatService)
		if err != nil {
			return nil, nil, fmt.Errorf("initialize Inngest recovery: %w", err)
		}
		aiChatService.SetRecoveryDispatcher(inngestRecovery)
	case strings.TrimSpace(cfg.InngestEventKey) != "" || strings.TrimSpace(cfg.InngestSigningKey) != "":
		logger.Warn("ai chat recovery disabled because both INNGEST_EVENT_KEY and INNGEST_SIGNING_KEY are required")
	default:
		logger.Info("ai chat recovery disabled because Inngest is not configured")
	}

	workoutHandler := workout.NewHandler(logger, validate, workoutService)
	exerciseHandler := exercise.NewHandler(logger, validate, exerciseService)
	featureAccessHandler := featureaccess.NewHandler(logger, featureAccessService)
	accountHandler := account.NewHandler(logger, accountService)
	billingHandler := billing.NewHandler(logger, billingService)
	trainingProfileHandler := trainingprofile.NewHandler(logger, trainingProfileService)
	healthHandler := health.NewHandler(logger, pool)
	aiChatHandler := aichat.NewHandler(logger, aiChatService)

	var e2eAuthHandler *e2eauth.Handler
	if cfg.LocalE2EAuthConfigured() {
		e2eAuthHandler = e2eauth.NewHandler(logger, e2eauth.NewService(
			logger,
			queries,
			pool,
			userService,
			cfg.LocalE2EAuthUserID,
			cfg.LocalE2EAuthEmail,
			cfg.LocalE2EAuthDisplayName,
		))
	}

	api := &api{
		logger:  logger,
		queries: queries,
		pool:    pool,
		cfg:     cfg,
	}
	if inngestRecovery != nil {
		api.inngestHandler = inngestRecovery.Handler()
	}

	jwks, err := auth.NewJWKSCache(ctx, cfg.ProjectID)
	if err != nil {
		return nil, nil, fmt.Errorf("create JWKS cache: %w", err)
	}
	authenticator := auth.NewAuthenticator(logger, jwks, userService, pool)
	if cfg.LocalE2EAuthConfigured() {
		authenticator.WithLocalE2EAuth(auth.LocalE2EAuthConfig{
			Enabled: true,
			UserID:  cfg.LocalE2EAuthUserID,
		})
	}

	router := api.routes(
		workoutHandler,
		exerciseHandler,
		featureAccessHandler,
		healthHandler,
		aiChatHandler,
		billingHandler,
		trainingProfileHandler,
		accountHandler,
		e2eAuthHandler,
	)

	var handler http.Handler = router
	handler = middleware.RateLimit(logger, int64(cfg.RateLimitRPM))(handler)
	handler = authenticator.Middleware(handler)
	handler = middleware.Metrics()(handler)
	handler = middleware.RequestLog(logger)(handler)
	handler = middleware.RequestID()(handler)
	handler = middleware.CORS(cfg.GetAllowedOrigins(), logger)(handler)
	handler = middleware.SecurityHeaders()(handler)

	return handler, api.metricsHandler(), nil
}

func newHTTPServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func newMetricsServer(cfg *config.Config, handler http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", handler)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// Run starts all application servers and blocks until cancellation or failure.
func (a *App) Run(ctx context.Context) error {
	serverErrors := make(chan error, 2)

	go func() {
		a.logger.Info("starting server", "addr", a.apiServer.Addr)
		serverErrors <- a.apiServer.ListenAndServe()
	}()
	go func() {
		a.logger.Info("starting metrics server", "addr", a.metricsServer.Addr)
		serverErrors <- a.metricsServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
		return a.shutdown()
	case err := <-serverErrors:
		shutdownErr := a.shutdown()
		if errors.Is(err, http.ErrServerClosed) {
			return shutdownErr
		}
		return errors.Join(fmt.Errorf("server failed: %w", err), shutdownErr)
	}
}

func (a *App) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a.logger.Info("draining connections")
	apiErr := a.apiServer.Shutdown(ctx)
	metricsErr := a.metricsServer.Shutdown(ctx)
	if apiErr == nil && metricsErr == nil {
		a.logger.Info("shutdown complete")
	}

	return errors.Join(apiErr, metricsErr)
}

// Close releases resources owned by the application.
func (a *App) Close() {
	a.pool.Close()
}
