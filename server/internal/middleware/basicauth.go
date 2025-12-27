package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
)

// BasicAuth returns a middleware that requires HTTP Basic Authentication
// If username or password are empty strings, the middleware allows all requests through
func BasicAuth(username, password string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If credentials are not configured, allow access
			if username == "" || password == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract credentials from request
			user, pass, ok := r.BasicAuth()

			// Check if credentials were provided and are correct
			// Use constant-time comparison to prevent timing attacks
			if !ok ||
				subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {

				// Get request ID from context
				requestID := GetRequestID(r.Context())

				// Log the unauthorized attempt
				logger.Warn("unauthorized metrics access attempt",
					"path", r.URL.Path,
					"method", r.Method,
					"status", http.StatusUnauthorized,
					"request_id", requestID,
				)

				// Set headers
				w.Header().Set("WWW-Authenticate", `Basic realm="Metrics"`)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)

				// Return standardized error response
				resp := map[string]string{
					"message": "unauthorized",
				}
				if requestID != "" {
					resp["request_id"] = requestID
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					logger.Error("failed to encode basic auth response", "error", err, "request_id", requestID)
				}
				return
			}

			// Credentials are valid, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}
