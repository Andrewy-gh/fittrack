package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// httpRequestsTotal tracks the total number of HTTP requests
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "path", "status"},
	)

	// httpRequestDuration tracks the duration of HTTP requests in seconds
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets, // Default buckets: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"method", "path", "status"},
	)

	// dbConnectionsActive tracks the number of active database connections
	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections in use",
		},
	)

	// dbConnectionsIdle tracks the number of idle database connections
	dbConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections in the pool",
		},
	)
)

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write ensures status code is captured even if WriteHeader is not called explicitly
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Metrics creates a middleware that tracks HTTP request metrics for Prometheus.
// It records request count, duration, method, path, and status code.
func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				written:        false,
			}

			// Process the request
			next.ServeHTTP(rw, r)

			// Record metrics after the request completes
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rw.statusCode)

			httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
			httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration)
		})
	}
}

// UpdateDatabaseMetrics updates the database connection pool metrics.
// This should be called periodically or before the metrics endpoint is scraped.
func UpdateDatabaseMetrics(pool *pgxpool.Pool) {
	if pool == nil {
		return
	}

	stats := pool.Stat()

	// AcquiredConns returns the number of currently acquired connections
	dbConnectionsActive.Set(float64(stats.AcquiredConns()))

	// IdleConns returns the number of idle connections
	dbConnectionsIdle.Set(float64(stats.IdleConns()))
}
