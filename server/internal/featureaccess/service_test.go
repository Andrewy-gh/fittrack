package featureaccess

import (
	"context"
	"errors"
	"io"
	"log/slog"
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

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) ListActiveFeatureAccess(ctx context.Context, userID string) ([]db.UserFeatureAccess, error) {
	args := m.Called(ctx, userID)
	rows, _ := args.Get(0).([]db.UserFeatureAccess)
	return rows, args.Error(1)
}

func (m *mockRepository) HasActiveFeatureAccess(ctx context.Context, userID string, featureKey string) (bool, error) {
	args := m.Called(ctx, userID, featureKey)
	return args.Bool(0), args.Error(1)
}

func TestService_ListCurrentUserAccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("unauthorized without user context", func(t *testing.T) {
		repo := new(mockRepository)
		service := NewService(logger, repo)

		grants, err := service.ListCurrentUserAccess(context.Background())

		require.Error(t, err)
		assert.Nil(t, grants)
		var errUnauthorized *apperrors.Unauthorized
		assert.ErrorAs(t, err, &errUnauthorized)
		repo.AssertExpectations(t)
	})

	t.Run("returns mapped active grants", func(t *testing.T) {
		repo := new(mockRepository)
		service := NewService(logger, repo)
		ctx := user.WithContext(context.Background(), "user-123")
		now := time.Date(2026, 3, 25, 13, 30, 0, 0, time.UTC)
		expiresAt := now.Add(24 * time.Hour)

		repo.On("ListActiveFeatureAccess", mock.Anything, "user-123").Return([]db.UserFeatureAccess{
			{
				FeatureKey:      "ai_chatbot",
				Source:          "manual",
				SourceReference: pgtype.Text{String: "demo-grant", Valid: true},
				GrantedBy:       pgtype.Text{String: "andy", Valid: true},
				Note:            pgtype.Text{String: "dev access", Valid: true},
				StartsAt:        pgtype.Timestamptz{Time: now, Valid: true},
				ExpiresAt:       pgtype.Timestamptz{Time: expiresAt, Valid: true},
				CreatedAt:       pgtype.Timestamptz{Time: now, Valid: true},
			},
		}, nil).Once()

		grants, err := service.ListCurrentUserAccess(ctx)

		require.NoError(t, err)
		require.Len(t, grants, 1)
		assert.Equal(t, "ai_chatbot", grants[0].FeatureKey)
		assert.Equal(t, "manual", grants[0].Source)
		assert.Equal(t, "demo-grant", *grants[0].SourceReference)
		assert.Equal(t, "andy", *grants[0].GrantedBy)
		assert.Equal(t, "dev access", *grants[0].Note)
		assert.Equal(t, now, grants[0].StartsAt)
		require.NotNil(t, grants[0].ExpiresAt)
		assert.Equal(t, expiresAt, *grants[0].ExpiresAt)
		repo.AssertExpectations(t)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := new(mockRepository)
		service := NewService(logger, repo)
		ctx := user.WithContext(context.Background(), "user-123")

		repo.On("ListActiveFeatureAccess", mock.Anything, "user-123").Return(nil, errors.New("boom")).Once()

		grants, err := service.ListCurrentUserAccess(ctx)

		require.Error(t, err)
		assert.Nil(t, grants)
		assert.Contains(t, err.Error(), "list active feature access")
		repo.AssertExpectations(t)
	})
}

func TestService_HasCurrentUserFeatureAccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("unauthorized without user context", func(t *testing.T) {
		repo := new(mockRepository)
		service := NewService(logger, repo)

		hasAccess, err := service.HasCurrentUserFeatureAccess(context.Background(), "ai_chatbot")

		require.Error(t, err)
		assert.False(t, hasAccess)
		var errUnauthorized *apperrors.Unauthorized
		assert.ErrorAs(t, err, &errUnauthorized)
		repo.AssertExpectations(t)
	})

	t.Run("blank feature key returns false without repository call", func(t *testing.T) {
		repo := new(mockRepository)
		service := NewService(logger, repo)
		ctx := user.WithContext(context.Background(), "user-123")

		hasAccess, err := service.HasCurrentUserFeatureAccess(ctx, "   ")

		require.NoError(t, err)
		assert.False(t, hasAccess)
		repo.AssertExpectations(t)
	})

	t.Run("checks repository with trimmed feature key", func(t *testing.T) {
		repo := new(mockRepository)
		service := NewService(logger, repo)
		ctx := user.WithContext(context.Background(), "user-123")

		repo.On("HasActiveFeatureAccess", mock.Anything, "user-123", "ai_chatbot").Return(true, nil).Once()

		hasAccess, err := service.HasCurrentUserFeatureAccess(ctx, " ai_chatbot ")

		require.NoError(t, err)
		assert.True(t, hasAccess)
		repo.AssertExpectations(t)
	})
}
