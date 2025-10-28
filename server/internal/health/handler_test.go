package health

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPool struct {
	mock.Mock
}

func (m *MockPool) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHealth(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tests := []struct {
		name         string
		expectedCode int
		checkFields  func(*testing.T, HealthResponse)
	}{
		{
			name:         "successful health check",
			expectedCode: http.StatusOK,
			checkFields: func(t *testing.T, resp HealthResponse) {
				assert.Equal(t, "healthy", resp.Status)
				assert.Equal(t, "1.0.0", resp.Version)
				assert.NotEmpty(t, resp.Timestamp)

				_, err := time.Parse(time.RFC3339, resp.Timestamp)
				assert.NoError(t, err, "timestamp should be in RFC3339 format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool := new(MockPool)
			handler := NewHandlerWithPool(logger, mockPool)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rr := httptest.NewRecorder()

			handler.Health(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var response HealthResponse
			err := json.NewDecoder(rr.Body).Decode(&response)
			assert.NoError(t, err, "response should be valid JSON")

			if tt.checkFields != nil {
				tt.checkFields(t, response)
			}
		})
	}
}

func TestReady(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tests := []struct {
		name         string
		setupMock    func(*MockPool)
		expectedCode int
		checkFields  func(*testing.T, ReadyResponse)
	}{
		{
			name: "database connection successful",
			setupMock: func(m *MockPool) {
				m.On("Ping", mock.Anything).Return(nil)
			},
			expectedCode: http.StatusOK,
			checkFields: func(t *testing.T, resp ReadyResponse) {
				assert.Equal(t, "healthy", resp.Status)
				assert.NotEmpty(t, resp.Timestamp)
				assert.Contains(t, resp.Checks, "database")
				assert.Equal(t, "ok", resp.Checks["database"])

				_, err := time.Parse(time.RFC3339, resp.Timestamp)
				assert.NoError(t, err, "timestamp should be in RFC3339 format")
			},
		},
		{
			name: "database connection failed",
			setupMock: func(m *MockPool) {
				m.On("Ping", mock.Anything).Return(errors.New("connection refused"))
			},
			expectedCode: http.StatusServiceUnavailable,
			checkFields: func(t *testing.T, resp ReadyResponse) {
				assert.Equal(t, "unhealthy", resp.Status)
				assert.NotEmpty(t, resp.Timestamp)
				assert.Contains(t, resp.Checks, "database")
				assert.Contains(t, resp.Checks["database"], "failed:")
				assert.Contains(t, resp.Checks["database"], "connection refused")

				_, err := time.Parse(time.RFC3339, resp.Timestamp)
				assert.NoError(t, err, "timestamp should be in RFC3339 format")
			},
		},
		{
			name: "database timeout error",
			setupMock: func(m *MockPool) {
				m.On("Ping", mock.Anything).Return(errors.New("context deadline exceeded"))
			},
			expectedCode: http.StatusServiceUnavailable,
			checkFields: func(t *testing.T, resp ReadyResponse) {
				assert.Equal(t, "unhealthy", resp.Status)
				assert.Contains(t, resp.Checks, "database")
				assert.Contains(t, resp.Checks["database"], "failed:")
				assert.Contains(t, resp.Checks["database"], "context deadline exceeded")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool := new(MockPool)
			if tt.setupMock != nil {
				tt.setupMock(mockPool)
			}

			handler := NewHandlerWithPool(logger, mockPool)

			req := httptest.NewRequest(http.MethodGet, "/ready", nil)
			rr := httptest.NewRecorder()

			handler.Ready(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var response ReadyResponse
			err := json.NewDecoder(rr.Body).Decode(&response)
			assert.NoError(t, err, "response should be valid JSON")

			if tt.checkFields != nil {
				tt.checkFields(t, response)
			}

			mockPool.AssertExpectations(t)
		})
	}
}

func TestReadyResponse_Format(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockPool := new(MockPool)
	mockPool.On("Ping", mock.Anything).Return(nil)

	handler := NewHandlerWithPool(logger, mockPool)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()

	handler.Ready(rr, req)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Contains(t, response, "status")
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "checks")

	checks, ok := response["checks"].(map[string]interface{})
	assert.True(t, ok, "checks should be an object")
	assert.Contains(t, checks, "database")
}

func TestHealthResponse_Format(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockPool := new(MockPool)

	handler := NewHandlerWithPool(logger, mockPool)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.Health(rr, req)

	var response map[string]any
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Contains(t, response, "status")
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "version")

	assert.IsType(t, "", response["status"])
	assert.IsType(t, "", response["timestamp"])
	assert.IsType(t, "", response["version"])
}
