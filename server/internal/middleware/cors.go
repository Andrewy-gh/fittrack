package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

// CORS creates a middleware that handles Cross-Origin Resource Sharing (CORS).
// It accepts a slice of allowed origins and validates incoming requests against them.
// Unknown origins are rejected for preflight requests.
func CORS(allowedOrigins []string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Cache variants must be separated even for absent or rejected origins.
			varyOrigin := true
			for _, value := range w.Header().Values("Vary") {
				for _, field := range strings.Split(value, ",") {
					if field = strings.TrimSpace(field); field == "*" || strings.EqualFold(field, "Origin") {
						varyOrigin = false
					}
				}
			}
			if varyOrigin {
				w.Header().Add("Vary", "Origin")
			}
			origin := r.Header.Get("Origin")

			// Check if the origin is in the allowed list
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			// Set CORS headers if origin is allowed
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-stack-access-token")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight OPTIONS requests
			if r.Method == http.MethodOptions {
				if allowed {
					// Accept preflight from allowed origins
					w.WriteHeader(http.StatusOK)
				} else {
					// Reject preflight from unknown origins - return standardized error response
					response.ErrorJSON(w, r, logger, http.StatusForbidden, "origin not allowed", nil)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
