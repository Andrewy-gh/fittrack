package response

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestErrorJSON_WritesSanitizedHTTPResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req = req.WithContext(request.WithRequestID(req.Context(), "req-http-response"))
	w := httptest.NewRecorder()

	ErrorJSON(
		w,
		req,
		logger,
		http.StatusInternalServerError,
		"failed to create user",
		fmt.Errorf("insert user: %w", &pgconn.PgError{Code: "23505", Message: "duplicate key", Detail: "synthetic private value"}),
	)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
	if got, want := w.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
	if got, want := w.Body.String(), `{"message":"internal error","request_id":"req-http-response"}`; got != want {
		t.Errorf("serialized response = %s, want %s", got, want)
	}
}

func TestErrorJSON_RedactsWrappedPostgresConnectError(t *testing.T) {
	cfg, err := pgconn.ParseConfig("host=127.0.0.1 port=1 user=offline-owner dbname=offline-db sslmode=disable")
	require.NoError(t, err)
	// Exercise pgconn's error construction without opening a socket.
	cfg.DialFunc = func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("synthetic private dial detail")
	}
	_, err = pgconn.ConnectConfig(context.Background(), cfg)
	var connectErr *pgconn.ConnectError
	require.ErrorAs(t, err, &connectErr)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req = req.WithContext(request.WithRequestID(req.Context(), "connect-failure"))
	w := httptest.NewRecorder()
	ErrorJSON(w, req, logger, http.StatusInternalServerError, "read failed", fmt.Errorf("acquire connection: %w", err))
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.JSONEq(t, `{"message":"internal error","request_id":"connect-failure"}`, w.Body.String())
	var entry map[string]any
	require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
	require.Equal(t, "error response", entry["msg"])
	require.Equal(t, "database", entry["error_category"])
	require.Equal(t, "internal error", entry["response_message"])
	require.Equal(t, "connect-failure", entry["request_id"])
	require.NotContains(t, logs.String(), "offline-owner")
	require.NotContains(t, logs.String(), "offline-db")
	require.NotContains(t, logs.String(), "synthetic private dial detail")
}

func TestErrorJSON_DoesNotLeakSensitiveData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tests := []struct {
		name             string
		message          string
		err              error
		expectedResponse string
		forbidden        []string
	}{
		{
			name:    "PostgreSQL constraint details",
			message: "failed to create user",
			err: &pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint \"users_pkey\" DETAIL: Key (user_id)=(test) already exists",
			},
			expectedResponse: "internal error",
			forbidden:        []string{"duplicate key", "users_pkey", "SQLSTATE", "23505"},
		},
		{
			name:             "unclassified operation failure is not an authentication error",
			message:          "operation failed",
			err:              errors.New("could not format token count: synthetic private detail"),
			expectedResponse: "operation failed",
			forbidden:        []string{"token count", "synthetic private detail"},
		},
		{
			name:             "unclassified operation failure is not a database error",
			message:          "operation failed",
			err:              errors.New("adjustable connection option: synthetic private detail"),
			expectedResponse: "operation failed",
			forbidden:        []string{"adjustable", "synthetic private detail"},
		},
	}

	for _, code := range []string{
		"23503", "42501", "42P01", "42703", "08006", "57P01",
	} {
		tests = append(tests, struct {
			name             string
			message          string
			err              error
			expectedResponse string
			forbidden        []string
		}{
			name:    "PostgreSQL status code " + code,
			message: "operation failed",
			err: &pgconn.PgError{
				Code:    code,
				Message: "query failed",
			},
			expectedResponse: "internal error",
			forbidden:        []string{code, "SQLSTATE"},
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/test", nil)

			ErrorJSON(
				w,
				req,
				logger,
				http.StatusInternalServerError,
				tt.message,
				tt.err,
			)

			if got, want := w.Code, http.StatusInternalServerError; got != want {
				t.Errorf("status = %d, want %d", got, want)
			}

			var response Error
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if got := response.Message; got != tt.expectedResponse {
				t.Errorf("response message = %q, want %q", got, tt.expectedResponse)
			}

			for _, forbidden := range tt.forbidden {
				if strings.Contains(w.Body.String(), forbidden) {
					t.Errorf("response leaked sensitive detail %q: %s", forbidden, w.Body.String())
				}
			}
		})
	}
}

func TestErrorJSON_LogsSafeErrorSummary(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	req = req.WithContext(request.WithRequestID(req.Context(), "req-safe-logs"))
	w := httptest.NewRecorder()
	sensitiveErr := fmt.Errorf("insert user: %w", &pgconn.PgError{Code: "23505", Message: "duplicate key", ConstraintName: "users_email_key", Detail: "andy@example.com"})

	ErrorJSON(w, req, logger, http.StatusInternalServerError, "failed to create user", sensitiveErr)

	logOutput := output.String()
	for _, forbidden := range []string{
		"pq:",
		"SQLSTATE",
		"23505",
		"users_email_key",
		"andy@example.com",
		"duplicate key",
	} {
		if strings.Contains(logOutput, forbidden) {
			t.Fatalf("error log contains sensitive raw error detail %q: %s", forbidden, logOutput)
		}
	}

	var logEntry map[string]any
	if err := json.Unmarshal(output.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to decode error log output: %v", err)
	}

	if got := logEntry["msg"]; got != "error response" {
		t.Fatalf("expected safe log message, got %v", got)
	}
	if got := logEntry["status"]; got != float64(http.StatusInternalServerError) {
		t.Fatalf("expected status field, got %v", got)
	}
	if got := logEntry["request_id"]; got != "req-safe-logs" {
		t.Fatalf("expected request_id field, got %v", got)
	}
	if got := logEntry["response_message"]; got != "internal error" {
		t.Fatalf("expected sanitized response_message field, got %v", got)
	}
	if got := logEntry["error_category"]; got != errorCategoryDatabase {
		t.Fatalf("expected database error_category field, got %v", got)
	}
	if got := logEntry["error_type"]; got != "*fmt.wrapError" {
		t.Fatalf("expected stable error_type field, got %v", got)
	}
}
