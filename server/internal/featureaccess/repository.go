package featureaccess

import (
	"context"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

type Repository interface {
	ListActiveFeatureAccess(ctx context.Context, userID string) ([]db.UserFeatureAccess, error)
	HasActiveFeatureAccess(ctx context.Context, userID string, featureKey string) (bool, error)
}

type repository struct {
	logger  *slog.Logger
	queries *db.Queries
}

func NewRepository(logger *slog.Logger, queries *db.Queries) Repository {
	return &repository{
		logger:  logger,
		queries: queries,
	}
}

func (r *repository) ListActiveFeatureAccess(ctx context.Context, userID string) ([]db.UserFeatureAccess, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return r.queries.ListActiveFeatureAccess(ctx, userID)
}

func (r *repository) HasActiveFeatureAccess(ctx context.Context, userID string, featureKey string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return r.queries.HasActiveFeatureAccess(ctx, db.HasActiveFeatureAccessParams{
		UserID:     userID,
		FeatureKey: featureKey,
	})
}

var _ Repository = (*repository)(nil)
