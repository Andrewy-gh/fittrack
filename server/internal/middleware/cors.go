package middleware

import "net/http"

// CORS creates a middleware that handles Cross-Origin Resource Sharing (CORS).
// It accepts a slice of allowed origins and validates incoming requests against them.
// Unknown origins are rejected for preflight requests.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
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
					w.WriteHeader(http.StatusForbidden)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
