package featureaccess

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	logger *slog.Logger
	repo   Repository
}

func NewService(logger *slog.Logger, repo Repository) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) ListCurrentUserAccess(ctx context.Context) ([]FeatureAccessGrant, error) {
	userID, ok := user.Current(ctx)
	if !ok || userID == "" {
		return nil, apperrors.NewUnauthorized("feature access", "")
	}

	rows, err := s.repo.ListActiveFeatureAccess(ctx, userID)
	if err != nil {
		s.logger.Error("failed to list active feature access", "error", err, "user_id", userID)
		return nil, fmt.Errorf("list active feature access: %w", err)
	}

	return convertRows(rows)
}

func (s *Service) HasCurrentUserFeatureAccess(ctx context.Context, featureKey string) (bool, error) {
	userID, ok := user.Current(ctx)
	if !ok || userID == "" {
		return false, apperrors.NewUnauthorized("feature access", "")
	}

	trimmedFeatureKey := strings.TrimSpace(featureKey)
	if trimmedFeatureKey == "" {
		return false, nil
	}

	hasAccess, err := s.repo.HasActiveFeatureAccess(ctx, userID, trimmedFeatureKey)
	if err != nil {
		s.logger.Error("failed to check feature access", "error", err, "user_id", userID, "feature_key", trimmedFeatureKey)
		return false, fmt.Errorf("check feature access: %w", err)
	}

	return hasAccess, nil
}

func convertRows(rows []db.UserFeatureAccess) ([]FeatureAccessGrant, error) {
	grants := make([]FeatureAccessGrant, 0, len(rows))
	for _, row := range rows {
		startsAt, err := timeFromTimestamptz(row.StartsAt)
		if err != nil {
			return nil, err
		}

		createdAt, err := timeFromTimestamptz(row.CreatedAt)
		if err != nil {
			return nil, err
		}

		expiresAt, err := timePtrFromTimestamptz(row.ExpiresAt)
		if err != nil {
			return nil, err
		}

		grants = append(grants, FeatureAccessGrant{
			FeatureKey:      row.FeatureKey,
			Source:          row.Source,
			SourceReference: textPtr(row.SourceReference),
			GrantedBy:       textPtr(row.GrantedBy),
			Note:            textPtr(row.Note),
			StartsAt:        startsAt,
			ExpiresAt:       expiresAt,
			CreatedAt:       createdAt,
		})
	}

	return grants, nil
}

func timeFromTimestamptz(ts pgtype.Timestamptz) (time.Time, error) {
	if !ts.Valid {
		return time.Time{}, fmt.Errorf("invalid timestamptz value")
	}

	return ts.Time.UTC(), nil
}

func timePtrFromTimestamptz(ts pgtype.Timestamptz) (*time.Time, error) {
	if !ts.Valid {
		return nil, nil
	}

	utc := ts.Time.UTC()
	return &utc, nil
}

func textPtr(txt pgtype.Text) *string {
	if !txt.Valid {
		return nil
	}

	value := txt.String
	return &value
}
