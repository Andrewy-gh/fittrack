package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
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

				// Set WWW-Authenticate header before sending error response
				w.Header().Set("WWW-Authenticate", `Basic realm="Metrics"`)

				// Return standardized error response
				response.ErrorJSON(w, r, logger, http.StatusUnauthorized, "unauthorized", nil)
				return
			}

			// Credentials are valid, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}
