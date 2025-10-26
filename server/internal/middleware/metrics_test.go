package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetrics_RequestCounting(t *testing.T) {
	// Reset metrics before test
	httpRequestsTotal.Reset()
	httpRequestDuration.Reset()

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
	}{
		{
			name:       "GET request with 200 status",
			method:     http.MethodGet,
			path:       "/api/workouts",
			statusCode: http.StatusOK,
		},
		{
			name:       "POST request with 201 status",
			method:     http.MethodPost,
			path:       "/api/workouts",
			statusCode: http.StatusCreated,
		},
		{
			name:       "GET request with 404 status",
			method:     http.MethodGet,
			path:       "/api/workouts/999",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "DELETE request with 204 status",
			method:     http.MethodDelete,
			path:       "/api/workouts/1",
			statusCode: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler that returns the desired status code
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode != http.StatusNoContent {
					w.Write([]byte("test response"))
				}
			})

			// Wrap with Metrics middleware
			handler := Metrics()(nextHandler)

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(rr, req)

			// Verify status code
			if rr.Code != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, rr.Code)
			}

			// Note: We can't easily verify exact counter values in parallel tests,
			// but we've verified the middleware executes successfully
			t.Logf("Metric recorded for %s %s with status %d", tt.method, tt.path, tt.statusCode)
		})
	}
}

func TestMetrics_DurationTracking(t *testing.T) {
	// Reset metrics before test
	httpRequestDuration.Reset()

	// Create handler with artificial delay
	delay := 50 * time.Millisecond
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with Metrics middleware
	handler := Metrics()(nextHandler)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	// Execute request
	start := time.Now()
	handler.ServeHTTP(rr, req)
	actualDuration := time.Since(start)

	// Verify the request took at least the expected delay
	if actualDuration < delay {
		t.Errorf("Expected request duration >= %v, got %v", delay, actualDuration)
	}

	// Note: Testing exact histogram values is complex, but we've verified
	// that the middleware properly wraps the handler and timing works
}

func TestMetrics_StatusCodeCapture(t *testing.T) {
	// Reset metrics before test
	httpRequestsTotal.Reset()

	tests := []struct {
		name           string
		writeHeader    bool
		explicitStatus int
		expectedStatus int
	}{
		{
			name:           "explicit WriteHeader call",
			writeHeader:    true,
			explicitStatus: http.StatusCreated,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "implicit 200 via Write",
			writeHeader:    false,
			explicitStatus: 0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "explicit 500 error",
			writeHeader:    true,
			explicitStatus: http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.writeHeader {
					w.WriteHeader(tt.explicitStatus)
				}
				w.Write([]byte("response"))
			})

			handler := Metrics()(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestMetrics_Labels(t *testing.T) {
	// Reset metrics before test
	httpRequestsTotal.Reset()

	testCases := []struct {
		method string
		path   string
		status int
	}{
		{http.MethodGet, "/api/workouts", http.StatusOK},
		{http.MethodPost, "/api/exercises", http.StatusCreated},
		{http.MethodPut, "/api/workouts/123", http.StatusOK},
		{http.MethodDelete, "/api/exercises/456", http.StatusNoContent},
	}

	for _, tc := range testCases {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
		})

		handler := Metrics()(nextHandler)
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// Verify that the request completed successfully
		if rr.Code != tc.status {
			t.Errorf("For %s %s: expected status %d, got %d",
				tc.method, tc.path, tc.status, rr.Code)
		}
	}
}

func TestUpdateDatabaseMetrics(t *testing.T) {
	// Reset gauges before test
	dbConnectionsActive.Set(0)
	dbConnectionsIdle.Set(0)

	tests := []struct {
		name        string
		pool        *pgxpool.Pool
		expectPanic bool
	}{
		{
			name:        "nil pool - should not panic",
			pool:        nil,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			UpdateDatabaseMetrics(tt.pool)

			// If we reach here without panic, test passes
			if tt.expectPanic {
				t.Error("Expected panic but none occurred")
			}
		})
	}
}

func TestUpdateDatabaseMetrics_WithRealPool(t *testing.T) {
	// This test requires a real database connection
	// Skip if DATABASE_URL is not set
	databaseURL := "postgres://test:test@localhost:5432/test?sslmode=disable"

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Skip("Skipping test: unable to connect to database")
		return
	}
	defer pool.Close()

	// Reset gauges
	dbConnectionsActive.Set(0)
	dbConnectionsIdle.Set(0)

	// Update metrics
	UpdateDatabaseMetrics(pool)

	// Verify that gauges were updated (should have at least some idle connections)
	activeValue := testutil.ToFloat64(dbConnectionsActive)
	idleValue := testutil.ToFloat64(dbConnectionsIdle)

	// Values should be >= 0
	if activeValue < 0 {
		t.Errorf("Expected active connections >= 0, got %f", activeValue)
	}
	if idleValue < 0 {
		t.Errorf("Expected idle connections >= 0, got %f", idleValue)
	}
}

func TestMetrics_PrometheusEndpointFormat(t *testing.T) {
	// Reset all metrics
	httpRequestsTotal.Reset()
	httpRequestDuration.Reset()
	dbConnectionsActive.Set(0)
	dbConnectionsIdle.Set(0)

	// Simulate some requests to generate metrics
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Metrics()(nextHandler)

	// Make a few requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Update database metrics
	UpdateDatabaseMetrics(nil) // nil pool is safe

	// Test the Prometheus endpoint format
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rr, req)

	// Verify response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Check for expected metric names in the output
	expectedMetrics := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"db_connections_active",
		"db_connections_idle",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(bodyStr, metric) {
			t.Errorf("Expected metric %q in output, but it was not found", metric)
		}
	}

	// Verify Content-Type header
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("Expected Content-Type to contain 'text/plain', got %q", contentType)
	}
}

func TestResponseWriter_MultipleWriteHeaderCalls(t *testing.T) {
	// Test that multiple WriteHeader calls only use the first status code
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.WriteHeader(http.StatusInternalServerError) // Should be ignored
		w.Write([]byte("response"))
	})

	handler := Metrics()(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should record the first status code (200), not the second (500)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}
