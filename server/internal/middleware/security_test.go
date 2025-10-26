package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name               string
		path               string
		isTLS              bool
		expectedFrameOpt   string
		expectedHSTS       bool
		expectedNoSniff    string
		expectedXSSProtect string
	}{
		{
			name:               "regular path with HTTP",
			path:               "/api/users",
			isTLS:              false,
			expectedFrameOpt:   "DENY",
			expectedHSTS:       false,
			expectedNoSniff:    "nosniff",
			expectedXSSProtect: "1; mode=block",
		},
		{
			name:               "regular path with HTTPS",
			path:               "/api/users",
			isTLS:              true,
			expectedFrameOpt:   "DENY",
			expectedHSTS:       true,
			expectedNoSniff:    "nosniff",
			expectedXSSProtect: "1; mode=block",
		},
		{
			name:               "swagger path",
			path:               "/swagger/index.html",
			isTLS:              false,
			expectedFrameOpt:   "SAMEORIGIN",
			expectedHSTS:       false,
			expectedNoSniff:    "nosniff",
			expectedXSSProtect: "1; mode=block",
		},
		{
			name:               "swagger path with HTTPS",
			path:               "/swagger/docs",
			isTLS:              true,
			expectedFrameOpt:   "SAMEORIGIN",
			expectedHSTS:       true,
			expectedNoSniff:    "nosniff",
			expectedXSSProtect: "1; mode=block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that does nothing
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with security headers middleware
			handler := SecurityHeaders()(nextHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.isTLS {
				req.TLS = &tls.ConnectionState{} // Non-nil TLS indicates HTTPS
			}

			// Record response
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check X-Content-Type-Options
			if got := rr.Header().Get("X-Content-Type-Options"); got != tt.expectedNoSniff {
				t.Errorf("X-Content-Type-Options = %v, want %v", got, tt.expectedNoSniff)
			}

			// Check X-Frame-Options
			if got := rr.Header().Get("X-Frame-Options"); got != tt.expectedFrameOpt {
				t.Errorf("X-Frame-Options = %v, want %v", got, tt.expectedFrameOpt)
			}

			// Check X-XSS-Protection
			if got := rr.Header().Get("X-XSS-Protection"); got != tt.expectedXSSProtect {
				t.Errorf("X-XSS-Protection = %v, want %v", got, tt.expectedXSSProtect)
			}

			// Check Strict-Transport-Security
			hstsHeader := rr.Header().Get("Strict-Transport-Security")
			if tt.expectedHSTS {
				if hstsHeader == "" {
					t.Error("Expected HSTS header for HTTPS request, got none")
				}
				expectedHSTS := "max-age=31536000; includeSubDomains"
				if hstsHeader != expectedHSTS {
					t.Errorf("Strict-Transport-Security = %v, want %v", hstsHeader, expectedHSTS)
				}
			} else {
				if hstsHeader != "" {
					t.Errorf("Expected no HSTS header for HTTP request, got %v", hstsHeader)
				}
			}
		})
	}
}
