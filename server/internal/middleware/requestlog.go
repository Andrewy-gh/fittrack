package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

// RequestLog emits one structured completion log for operational request triage.
func RequestLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			_, hasUser := user.Current(r.Context())
			duration := time.Since(start)
			logAttrs := []any{
				"method", r.Method,
				"route", routeLabel(r),
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
				"request_id", request.GetRequestID(r.Context()),
				"user_present", hasUser,
			}

			switch {
			case rw.statusCode >= http.StatusInternalServerError:
				logger.Error("request completed", logAttrs...)
			case rw.statusCode >= http.StatusBadRequest:
				logger.Warn("request completed", logAttrs...)
			default:
				logger.Info("request completed", logAttrs...)
			}
		})
	}
}
