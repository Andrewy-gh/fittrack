package middleware

import (
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Test handler that returns 200 OK if reached
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	tests := []struct {
		name           string
		username       string
		password       string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid credentials",
			username:       "admin",
			password:       "secret",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret")),
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "invalid username",
			username:       "admin",
			password:       "secret",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("wrong:secret")),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "invalid password",
			username:       "admin",
			password:       "secret",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:wrong")),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "no auth header",
			username:       "admin",
			password:       "secret",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "malformed auth header",
			username:       "admin",
			password:       "secret",
			authHeader:     "Bearer token123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "empty credentials allows access",
			username:       "",
			password:       "",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "empty username allows access",
			username:       "",
			password:       "secret",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "empty password allows access",
			username:       "admin",
			password:       "",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := BasicAuth(tt.username, tt.password, logger)
			handler := middleware(testHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Record response
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check WWW-Authenticate header on 401
			if tt.expectedStatus == http.StatusUnauthorized {
				authHeader := rr.Header().Get("WWW-Authenticate")
				if authHeader != `Basic realm="Metrics"` {
					t.Errorf("expected WWW-Authenticate header, got: %s", authHeader)
				}
			}

			// Check body for successful requests
			if tt.expectedStatus == http.StatusOK && rr.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
