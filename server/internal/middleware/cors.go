package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// CORS creates a middleware that handles Cross-Origin Resource Sharing (CORS).
// It accepts a slice of allowed origins and validates incoming requests against them.
// Unknown origins are rejected for preflight requests.
func CORS(allowedOrigins []string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-stack-access-token")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight OPTIONS requests
			if r.Method == http.MethodOptions {
				if allowed {
					// Accept preflight from allowed origins
					w.WriteHeader(http.StatusOK)
				} else {
					// Reject preflight from unknown origins
					requestID := GetRequestID(r.Context())

					// Log the rejected CORS request
					logger.Warn("CORS preflight rejected - origin not allowed",
						"origin", origin,
						"path", r.URL.Path,
						"method", r.Method,
						"status", http.StatusForbidden,
						"request_id", requestID)

					// Return standardized error response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					resp := map[string]string{
						"message": "CORS policy: Origin not allowed",
					}
					if requestID != "" {
						resp["request_id"] = requestID
					}
					if err := json.NewEncoder(w).Encode(resp); err != nil {
						logger.Error("failed to encode CORS rejection response", "error", err, "request_id", requestID)
					}
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
