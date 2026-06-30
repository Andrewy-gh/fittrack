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

	subscriptionCancellationAttempted, err := s.cancelSubscriptionBeforeLocalAccountDeletion(ctx)
	if err != nil {
		return err
	}

	if err := s.deleteLocalAccountData(ctx, userID, subscriptionCancellationAttempted); err != nil {
		return err
	}

	s.logger.Info("deleted fittrack user account data", "user_id", userID)
	return nil
}

func (s *Service) cancelSubscriptionBeforeLocalAccountDeletion(ctx context.Context) (bool, error) {
	if s.billing != nil {
		if err := s.billing.CancelCurrentSubscriptionImmediately(ctx); err != nil {
			return true, fmt.Errorf("cancel subscription before local account deletion: %w", err)
		}
		return true, nil
	}
	return false, nil
}

func (s *Service) deleteLocalAccountData(ctx context.Context, userID string, subscriptionCancellationAttempted bool) error {
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		if subscriptionCancellationAttempted {
			return fmt.Errorf("delete local account data after subscription cancellation: %w", err)
		}
		return fmt.Errorf("delete current user account: %w", err)
	}
	return nil
}
