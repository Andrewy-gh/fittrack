package user

import (
	"context"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

// Service handles user business logic
type Service struct {
	logger *slog.Logger
	repo   UserRepository
}

// NewService creates a new user service
func NewService(logger *slog.Logger, repo UserRepository) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

// EnsureUser retrieves a user by ID, creating them if they don't exist
func (s *Service) EnsureUser(ctx context.Context, userID string) (db.Users, error) {
	user, err := s.repo.CreateUser(ctx, userID)
	if err != nil {
		if db.IsUniqueConstraintError(err) {
			s.logger.Debug("user already exists, fetching", "user_id", userID)
			return s.repo.GetUser(ctx, userID)
		}
		s.logger.Error("failed to create user", "error", err, "user_id", userID)
		return db.Users{}, err
	}
	s.logger.Info("created new user", "user_id", userID)
	return user, nil
}
