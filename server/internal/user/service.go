// service.go
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

type UserRepository interface {
	GetUser(ctx context.Context, id string) (db.Users, error)
	CreateUser(ctx context.Context, userID string) (db.Users, error)
}

type Service struct {
	logger *slog.Logger
	repo   UserRepository
}

func NewService(logger *slog.Logger, repo UserRepository) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) EnsureUser(ctx context.Context, userID string) (db.Users, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err == nil {
		s.logger.Debug("user found", "user_id", userID)
		return user, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("failed to get user", "error", err, "user_id", userID)
		s.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "user_id", userID)
		return db.Users{}, err
	}

	s.logger.Debug("user not found, creating", "user_id", userID)
	user, err = s.repo.CreateUser(ctx, userID)
	if err != nil {
		if db.IsUniqueConstraintError(err) {
			s.logger.Debug("user created by another process, fetching", "user_id", userID)
			return s.repo.GetUser(ctx, userID)
		}
		s.logger.Error("failed to create user", "error", err, "user_id", userID)
		s.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "user_id", userID)
		return db.Users{}, err
	}

	s.logger.Info("created new user", "user_id", userID)
	return user, nil
}
