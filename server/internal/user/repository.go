package user

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgxpool.Pool
}

// NewRepository creates a new instance of UserRepository
func NewRepository(logger *slog.Logger, queries *db.Queries, conn *pgxpool.Pool) UserRepository {
	return &userRepository{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}
}

// GetUser retrieves a user by their ID
func (r *userRepository) GetUser(ctx context.Context, id string) (db.Users, error) {
	user, err := r.queries.GetUserByUserID(ctx, id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return db.Users{}, sql.ErrNoRows
		}
		r.logger.Error("database error getting user", "error", err, "user_id", id)
		return db.Users{}, fmt.Errorf("failed to get user %s: %w", id, err)
	}
	return user, nil
}

// CreateUser inserts a new user into the database
func (r *userRepository) CreateUser(ctx context.Context, userID string) (db.Users, error) {
	id, err := r.queries.CreateUser(ctx, userID)
	if err != nil {
		if db.IsUniqueConstraintError(err) {
			r.logger.Debug("user creation failed due to unique constraint", "user_id", userID)
		} else {
			r.logger.Error("database error creating user", "error", err, "user_id", userID)
		}
		return db.Users{}, fmt.Errorf("failed to create user %s: %w", userID, err)
	}

	// Since CreateUser now only returns the ID, we need to fetch the full user
	return r.queries.GetUser(ctx, id)
}

var _ UserRepository = (*userRepository)(nil)
