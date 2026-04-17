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

// Metrics creates a middleware that tracks HTTP request metrics for Prometheus.
// It records request count, duration, method, path, and status code.
func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture status code
			rw := newResponseWriter(w)

			// Process the request
			next.ServeHTTP(rw, r)

			// Record metrics after the request completes
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rw.statusCode)
			route := routeLabel(r)

			httpRequestsTotal.WithLabelValues(r.Method, route, status).Inc()
			httpRequestDuration.WithLabelValues(r.Method, route, status).Observe(duration)
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
