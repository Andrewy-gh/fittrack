package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolPinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	logger *slog.Logger
	pool   PoolPinger
}

func NewHandler(logger *slog.Logger, pool *pgxpool.Pool) *Handler {
	return NewHandlerWithPool(logger, pool)
}

func NewHandlerWithPool(logger *slog.Logger, pool PoolPinger) *Handler {
	return &Handler{
		logger: logger,
		pool:   pool,
	}
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// Health godoc
// @Summary Health check
// @Description Returns the health status of the API
// @Tags health
// @Produce json
// @Success 200 {object} health.HealthResponse
// @Router /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode health response", "error", err)
	}
}

type ReadyResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// Ready godoc
// @Summary Readiness check
// @Description Returns the readiness status of the API including database connectivity
// @Tags health
// @Produce json
// @Success 200 {object} health.ReadyResponse
// @Failure 503 {object} health.ReadyResponse
// @Router /ready [get]
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)

	if err := h.pool.Ping(r.Context()); err != nil {
		response := ReadyResponse{
			Status:    "unhealthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Checks: map[string]string{
				"database": "failed: " + err.Error(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.Error("failed to encode ready response", "error", err)
		}
		return
	}

	checks["database"] = "ok"

	response := ReadyResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode ready response", "error", err)
	}
}
