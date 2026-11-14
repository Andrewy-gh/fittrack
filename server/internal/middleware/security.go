package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders adds security-related HTTP headers to all responses.
// It sets X-Content-Type-Options, X-Frame-Options, and X-XSS-Protection headers
// to protect against common web vulnerabilities.
// Production serves public HTTPS through Fly's TLS-terminating proxy (with
// force_https enabled). Use startup configuration, never client-supplied
// Forwarded/X-Forwarded-Proto headers, to retain HSTS on its HTTP backend hop.
func SecurityHeaders(environment string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking - allow SAMEORIGIN for swagger docs, DENY for everything else
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			} else {
				w.Header().Set("X-Frame-Options", "DENY")
			}

			// Enable XSS protection in older browsers
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Direct TLS retains HSTS in every environment; development HTTP does not.
			if environment == "production" || r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}
