package middleware

import (
	"net/http"
	"path"
	"strings"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		written:        false,
	}
}

// WriteHeader captures the first status code written by the handler chain.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write ensures status code is captured even if WriteHeader is not called explicitly.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Flush preserves streaming support for wrapped writers such as SSE handlers.
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Unwrap lets http.ResponseController reach the original writer through this wrapper.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func routeLabel(r *http.Request) string {
	if r == nil {
		return "unknown"
	}

	if pattern := normalizePattern(r.Pattern); pattern != "" {
		return pattern
	}

	switch requestPath := r.URL.Path; {
	case requestPath == "" || requestPath == "/":
		return "/"
	case requestPath == "/health" || requestPath == "/ready" || requestPath == "/metrics" || requestPath == "/inngest":
		return requestPath
	case strings.HasPrefix(requestPath, "/api/"):
		return "/api/*"
	case path.Ext(requestPath) != "":
		return "static_asset"
	default:
		return "static_page"
	}
}

func normalizePattern(pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return ""
	}

	if method, rest, ok := strings.Cut(pattern, " "); ok && isHTTPMethod(method) {
		return strings.TrimSpace(rest)
	}

	return pattern
}

func isHTTPMethod(value string) bool {
	switch value {
	case http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace:
		return true
	default:
		return false
	}
}
