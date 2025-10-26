package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	allowedOrigins := []string{"http://localhost:5173", "https://app.example.com"}

	tests := []struct {
		name                 string
		method               string
		origin               string
		expectOriginHeader   bool
		expectedOrigin       string
		expectStatus         int
		expectMethodsHeader  bool
		expectHeadersHeader  bool
		expectCredsHeader    bool
	}{
		{
			name:                "allowed origin with GET",
			method:              http.MethodGet,
			origin:              "http://localhost:5173",
			expectOriginHeader:  true,
			expectedOrigin:      "http://localhost:5173",
			expectStatus:        http.StatusOK,
			expectMethodsHeader: true,
			expectHeadersHeader: true,
			expectCredsHeader:   true,
		},
		{
			name:                "allowed origin with OPTIONS preflight",
			method:              http.MethodOptions,
			origin:              "https://app.example.com",
			expectOriginHeader:  true,
			expectedOrigin:      "https://app.example.com",
			expectStatus:        http.StatusOK,
			expectMethodsHeader: true,
			expectHeadersHeader: true,
			expectCredsHeader:   true,
		},
		{
			name:                "unknown origin with GET",
			method:              http.MethodGet,
			origin:              "https://evil.com",
			expectOriginHeader:  false,
			expectStatus:        http.StatusOK,
			expectMethodsHeader: false,
			expectHeadersHeader: false,
			expectCredsHeader:   false,
		},
		{
			name:                "unknown origin with OPTIONS preflight - should reject",
			method:              http.MethodOptions,
			origin:              "https://evil.com",
			expectOriginHeader:  false,
			expectStatus:        http.StatusForbidden,
			expectMethodsHeader: false,
			expectHeadersHeader: false,
			expectCredsHeader:   false,
		},
		{
			name:                "no origin header",
			method:              http.MethodGet,
			origin:              "",
			expectOriginHeader:  false,
			expectStatus:        http.StatusOK,
			expectMethodsHeader: false,
			expectHeadersHeader: false,
			expectCredsHeader:   false,
		},
		{
			name:                "OPTIONS with no origin - should reject",
			method:              http.MethodOptions,
			origin:              "",
			expectOriginHeader:  false,
			expectStatus:        http.StatusForbidden,
			expectMethodsHeader: false,
			expectHeadersHeader: false,
			expectCredsHeader:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with CORS middleware
			handler := CORS(allowedOrigins)(nextHandler)

			// Create request
			req := httptest.NewRequest(tt.method, "/api/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			// Record response
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectStatus {
				t.Errorf("status code = %v, want %v", rr.Code, tt.expectStatus)
			}

			// Check Access-Control-Allow-Origin header
			originHeader := rr.Header().Get("Access-Control-Allow-Origin")
			if tt.expectOriginHeader {
				if originHeader != tt.expectedOrigin {
					t.Errorf("Access-Control-Allow-Origin = %v, want %v", originHeader, tt.expectedOrigin)
				}
			} else {
				if originHeader != "" {
					t.Errorf("Expected no Access-Control-Allow-Origin header, got %v", originHeader)
				}
			}

			// Check Access-Control-Allow-Methods header
			methodsHeader := rr.Header().Get("Access-Control-Allow-Methods")
			if tt.expectMethodsHeader {
				if methodsHeader == "" {
					t.Error("Expected Access-Control-Allow-Methods header, got none")
				}
			} else {
				if methodsHeader != "" {
					t.Errorf("Expected no Access-Control-Allow-Methods header, got %v", methodsHeader)
				}
			}

			// Check Access-Control-Allow-Headers header
			headersHeader := rr.Header().Get("Access-Control-Allow-Headers")
			if tt.expectHeadersHeader {
				if headersHeader == "" {
					t.Error("Expected Access-Control-Allow-Headers header, got none")
				}
			} else {
				if headersHeader != "" {
					t.Errorf("Expected no Access-Control-Allow-Headers header, got %v", headersHeader)
				}
			}

			// Check Access-Control-Allow-Credentials header
			credsHeader := rr.Header().Get("Access-Control-Allow-Credentials")
			if tt.expectCredsHeader {
				if credsHeader != "true" {
					t.Errorf("Access-Control-Allow-Credentials = %v, want true", credsHeader)
				}
			} else {
				if credsHeader != "" {
					t.Errorf("Expected no Access-Control-Allow-Credentials header, got %v", credsHeader)
				}
			}
		})
	}
}

func TestCORS_EmptyAllowedOrigins(t *testing.T) {
	// Test with empty allowed origins list
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS([]string{})(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should not set any CORS headers
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Expected no Access-Control-Allow-Origin header, got %v", got)
	}
}
