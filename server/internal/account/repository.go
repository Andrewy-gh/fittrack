package account

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

type Repository interface {
	DeleteUser(ctx context.Context, userID string) error
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

func (r *repository) DeleteUser(ctx context.Context, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := r.queries.DeleteUser(ctx, userID); err != nil {
		r.logger.Error("database error deleting user", "error", err, "user_id", userID)
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

var _ Repository = (*repository)(nil)
