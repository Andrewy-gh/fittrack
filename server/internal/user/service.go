package user

import (
	"context"
	"database/sql"
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
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Info("user not found, creating new user", "user_id", userID)
			return s.repo.CreateUser(ctx, userID)
		}
		s.logger.Error("failed to get user", "error", err, "user_id", userID)
		return db.Users{}, err
	}
	return user, nil
}
