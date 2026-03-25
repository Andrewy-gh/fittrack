package featureaccess

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
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

	t.Run("returns active grants", func(t *testing.T) {
		service := new(mockListService)
		handler := NewHandler(logger, service)
		now := time.Date(2026, 3, 25, 14, 0, 0, 0, time.UTC)

		service.On("ListCurrentUserAccess", mock.Anything).Return([]FeatureAccessGrant{
			{
				FeatureKey: "ai_chatbot",
				Source:     "manual",
				StartsAt:   now,
				CreatedAt:  now,
			},
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/features/access", nil)
		rr := httptest.NewRecorder()

		handler.ListActiveFeatureAccess(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var grants []FeatureAccessGrant
		err := json.NewDecoder(rr.Body).Decode(&grants)
		require.NoError(t, err)
		require.Len(t, grants, 1)
		assert.Equal(t, "ai_chatbot", grants[0].FeatureKey)
		service.AssertExpectations(t)
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
