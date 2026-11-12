package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestCORS(t *testing.T) {
	allowedOrigins := []string{"http://localhost:5173", "https://app.example.com"}

	tests := []struct {
		name                string
		method              string
		origin              string
		expectOriginHeader  bool
		expectedOrigin      string
		expectStatus        int
		expectMethodsHeader bool
		expectHeadersHeader bool
		expectCredsHeader   bool
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
			name: "allowed origin with PATCH", method: http.MethodPatch, origin: "https://app.example.com",
			expectOriginHeader: true, expectedOrigin: "https://app.example.com", expectStatus: http.StatusOK,
			expectMethodsHeader: true, expectHeadersHeader: true, expectCredsHeader: true,
		},
		{
			name: "disallowed origin with PATCH", method: http.MethodPatch, origin: "https://evil.com",
			expectStatus: http.StatusOK,
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

			// Create test logger
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			// Wrap with CORS middleware
			handler := CORS(allowedOrigins, logger)(nextHandler)

			// Create request
			req := httptest.NewRequest(tt.method, "/api/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.method == http.MethodOptions {
				req.Header.Set("Access-Control-Request-Method", "PATCH")
				req.Header.Set("Access-Control-Request-Headers", "x-stack-access-token")
			}

			// Record response
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if !strings.Contains(strings.Join(rr.Header().Values("Vary"), ","), "Origin") {
				t.Error("response must vary by Origin even when origin is absent or disallowed")
			}
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
				if !strings.Contains(methodsHeader, "PATCH") {
					t.Errorf("PATCH missing from Access-Control-Allow-Methods: %q", methodsHeader)
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

func TestCORS_PreservesVary(t *testing.T) {
	for _, existing := range [][]string{{"Accept-Encoding, Accept", "Cookie"}, {"Accept", "origin"}, {"*"}} {
		for _, method := range []string{"OPTIONS", "PATCH"} {
			rr := httptest.NewRecorder()
			rr.Header()["Vary"] = append([]string(nil), existing...)
			req := httptest.NewRequest(method, "/api/exercises/123", nil)
			req.Header.Set("Origin", "https://allowed.test")
			req.Header.Set("Access-Control-Request-Method", "PATCH")
			CORS([]string{"https://allowed.test"}, slog.Default())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Vary", "Accept-Language")
				w.WriteHeader(204)
			})).ServeHTTP(rr, req)
			want := append([]string(nil), existing...)
			if len(existing) == 2 && existing[1] == "Cookie" {
				want = append(want, "Origin")
			}
			if method == "PATCH" {
				want = append(want, "Accept-Language")
			}
			if !reflect.DeepEqual(rr.Result().Header.Values("Vary"), want) {
				t.Errorf("%s Vary=%v want %v", method, rr.Result().Header.Values("Vary"), want)
			}
		}
	}
}

func TestCORS_EmptyAllowedOrigins(t *testing.T) {
	// Test with empty allowed origins list
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create test logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := CORS([]string{}, logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should not set any CORS headers
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Expected no Access-Control-Allow-Origin header, got %v", got)
	}
}
