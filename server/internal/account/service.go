package account

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

type Service struct {
	logger  *slog.Logger
	repo    Repository
	billing BillingCanceler
}

type BillingCanceler interface {
	CancelCurrentSubscriptionImmediately(ctx context.Context) error
}

func NewService(logger *slog.Logger, repo Repository, billing BillingCanceler) *Service {
	return &Service{
		logger:  logger,
		repo:    repo,
		billing: billing,
	}
}

func (s *Service) DeleteCurrentUser(ctx context.Context) error {
	userID, ok := user.Current(ctx)
	if !ok || userID == "" {
		return apperrors.NewUnauthorized("account", "")
	}

	if s.billing != nil {
		if err := s.billing.CancelCurrentSubscriptionImmediately(ctx); err != nil {
			return fmt.Errorf("cancel subscription before account deletion: %w", err)
		}
	}

	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("delete current user account: %w", err)
	}

	s.logger.Info("deleted fittrack user account data", "user_id", userID)
	return nil
}
