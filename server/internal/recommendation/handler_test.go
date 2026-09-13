package recommendation_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/require"
)

type ownedRepository struct{}

func (*ownedRepository) Load(_ context.Context, owner string, id int32, now time.Time) (recommendation.TrainingContext, error) {
	if owner != "owner" || id != 1 {
		return recommendation.TrainingContext{}, recommendation.ErrNotFound
	}
	return recommendation.TrainingContext{Sessions: []recommendation.Session{{
		Date: now.Add(-time.Hour),
		Sets: []recommendation.WorkingSet{{Reps: 8}, {Reps: 8}},
	}}}, nil
}

func TestHTTPRecommendation(t *testing.T) {
	handler := recommendation.NewHandler(recommendation.NewService(&ownedRepository{}, time.Now), slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /exercises/{id}/recommendation", handler.Get)
	request := func(path, owner string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if owner != "" {
			req = req.WithContext(user.WithContext(req.Context(), owner))
		}
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, req)
		return recorder
	}

	require.Equal(t, http.StatusUnauthorized, request("/exercises/1/recommendation", "").Code)
	require.Equal(t, http.StatusNotFound, request("/exercises/1/recommendation", "other").Code)
	require.Equal(t, http.StatusBadRequest, request("/exercises/1/recommendation?readiness=excellent", "owner").Code)
	response := request("/exercises/1/recommendation?readiness=normal", "owner")
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `"sets":2`)
	require.NotContains(t, response.Body.String(), `"range"`)
}
