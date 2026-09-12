package recommendation_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/require"
)

type ownedRepository struct{ baseline *recommendation.Range }

func (r *ownedRepository) Load(_ context.Context, owner string, id int32, _ time.Time) (*recommendation.Range, *recommendation.Session, error) {
	if owner != "owner" || id != 1 {
		return nil, nil, recommendation.ErrNotFound
	}
	return r.baseline, nil, nil
}
func (r *ownedRepository) SavePrescription(_ context.Context, owner string, id int32, baseline *recommendation.Range) error {
	if owner != "owner" || id != 1 {
		return recommendation.ErrNotFound
	}
	r.baseline = baseline
	return nil
}
func (r *ownedRepository) Saved(_ context.Context, owner string, _ int32) ([]recommendation.Snapshot, error) {
	if owner != "owner" {
		return nil, recommendation.ErrNotFound
	}
	return []recommendation.Snapshot{}, nil
}

func TestHTTPAuthenticationPrescriptionAndFallback(t *testing.T) {
	handler := recommendation.NewHandler(recommendation.NewService(&ownedRepository{}, time.Now), slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /exercises/{id}/recommendation", handler.Get)
	mux.HandleFunc("PUT /exercises/{id}/prescription", handler.Prescribe)
	mux.HandleFunc("GET /workouts/{id}/recommendations", handler.Saved)
	request := func(method, path, body, owner string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if owner != "" {
			req = req.WithContext(user.WithContext(req.Context(), owner))
		}
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, req)
		return recorder
	}
	require.Equal(t, 401, request("GET", "/exercises/1/recommendation", "", "").Code)
	require.Equal(t, 401, request("PUT", "/exercises/1/prescription", `{"baseline":{"min":2,"max":3}}`, "").Code)
	require.Equal(t, 401, request("GET", "/workouts/1/recommendations", "", "").Code)
	require.Equal(t, 404, request("GET", "/exercises/1/recommendation", "", "other").Code)
	require.Equal(t, 404, request("PUT", "/exercises/1/prescription", `{"baseline":{"min":2,"max":3}}`, "other").Code)
	require.Equal(t, 400, request("PUT", "/exercises/1/prescription", `{"baseline":{"min":3,"max":2}}`, "owner").Code)
	require.Equal(t, 400, request("PUT", "/exercises/1/prescription", `{"baseline":{"min":2,"max":3},"extra":true}`, "owner").Code)
	require.Equal(t, 400, request("PUT", "/exercises/1/prescription", `{"baseline":null} {}`, "owner").Code)
	require.Equal(t, 400, request("PUT", "/exercises/1/prescription", `{}`, "owner").Code)
	require.Equal(t, 400, request("PUT", "/exercises/1/prescription", `null`, "owner").Code)
	require.Equal(t, 400, request("PUT", "/exercises/1/prescription", `{"baseline":{"min":1,"max":2,"extra":true}}`, "owner").Code)
	require.Equal(t, 204, request("PUT", "/exercises/1/prescription", `{"baseline":{"min":2,"max":3}}`, "owner").Code)
	response := request("GET", "/exercises/1/recommendation?readiness=normal", "", "owner")
	require.Equal(t, 200, response.Code)
	require.Contains(t, response.Body.String(), `"source":"prescription"`)
	require.Equal(t, 204, request("PUT", "/exercises/1/prescription", `{"baseline":null}`, "owner").Code)
	response = request("GET", "/exercises/1/recommendation?readiness=normal", "", "owner")
	require.Contains(t, response.Body.String(), `"range":null`)
}
