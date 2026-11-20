package featureaccess

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockListService struct {
	mock.Mock
}

func (m *mockListService) ListCurrentUserAccess(ctx context.Context) ([]FeatureAccessGrant, error) {
	args := m.Called(ctx)
	grants, _ := args.Get(0).([]FeatureAccessGrant)
	return grants, args.Error(1)
}

func TestHandler_ListActiveFeatureAccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("returns active grants without internal metadata", func(t *testing.T) {
		repo := new(mockRepository)
		handler := NewHandler(logger, NewService(logger, repo))
		now := time.Date(2026, 3, 25, 14, 0, 0, 0, time.UTC)
		stamp := pgtype.Timestamptz{Time: now, Valid: true}

		repo.On("ListActiveFeatureAccess", mock.Anything, "user-123").Return([]db.UserFeatureAccess{
			{
				FeatureKey:      "ai_chatbot",
				Source:          "manual",
				SourceReference: pgtype.Text{String: "demo-grant", Valid: true},
				GrantedBy:       pgtype.Text{String: "private-operator", Valid: true},
				Note:            pgtype.Text{String: "private grant note", Valid: true},
				StartsAt:        stamp,
				ExpiresAt:       pgtype.Timestamptz{Time: now.Add(24 * time.Hour), Valid: true},
				CreatedAt:       stamp,
			},
			{
				FeatureKey: "analytics_beta",
				Source:     "manual",
				StartsAt:   stamp,
				CreatedAt:  stamp,
			},
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/features/access", nil)
		req = req.WithContext(user.WithContext(req.Context(), "user-123"))
		rr := httptest.NewRecorder()

		handler.ListActiveFeatureAccess(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.JSONEq(t, `[
			{"feature_key":"ai_chatbot","source":"manual","source_reference":"demo-grant",
			 "starts_at":"2026-03-25T14:00:00Z","expires_at":"2026-03-26T14:00:00Z","created_at":"2026-03-25T14:00:00Z"},
			{"feature_key":"analytics_beta","source":"manual",
			 "starts_at":"2026-03-25T14:00:00Z","created_at":"2026-03-25T14:00:00Z"}
		]`, rr.Body.String())
		repo.AssertExpectations(t)
	})

	t.Run("returns empty array when no grants", func(t *testing.T) {
		service := new(mockListService)
		handler := NewHandler(logger, service)

		service.On("ListCurrentUserAccess", mock.Anything).Return(nil, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/features/access", nil)
		rr := httptest.NewRecorder()

		handler.ListActiveFeatureAccess(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, "[]", rr.Body.String())
		service.AssertExpectations(t)
	})

	t.Run("maps unauthorized error", func(t *testing.T) {
		service := new(mockListService)
		handler := NewHandler(logger, service)

		service.On("ListCurrentUserAccess", mock.Anything).Return(nil, apperrors.NewUnauthorized("feature access", "")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/features/access", nil)
		rr := httptest.NewRecorder()

		handler.ListActiveFeatureAccess(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "not authorized")
		service.AssertExpectations(t)
	})

	t.Run("maps internal error", func(t *testing.T) {
		service := new(mockListService)
		handler := NewHandler(logger, service)

		service.On("ListCurrentUserAccess", mock.Anything).Return(nil, errors.New("boom")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/features/access", nil)
		rr := httptest.NewRecorder()

		handler.ListActiveFeatureAccess(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "failed to list feature access")
		service.AssertExpectations(t)
	})
}
