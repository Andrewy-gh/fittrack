package account

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

type Repository interface {
	DeleteUser(ctx context.Context, userID string) error
}

var ErrUserNotDeleted = errors.New("user account was not deleted")

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

func (r *repository) DeleteUser(ctx context.Context, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rowsAffected, err := r.queries.DeleteUser(ctx, userID)
	if err != nil {
		r.logger.Error("database error deleting user", "error", err, "user_id", userID)
		return fmt.Errorf("delete user: %w", err)
	}
	if rowsAffected == 0 {
		r.logger.Error("database delete affected zero users", "user_id", userID)
		return ErrUserNotDeleted
	}
	return nil
}

var _ Repository = (*repository)(nil)
