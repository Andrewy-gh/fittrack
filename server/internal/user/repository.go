package user

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// GetUser retrieves a user by their ID
	GetUser(ctx context.Context, id string) (db.Users, error)
	// CreateUser inserts a new user into the database
	CreateUser(ctx context.Context, userID string) (db.Users, error)
}

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
		r.logger.Error("failed to get user", "error", err, "user_id", id)
		return db.Users{}, err
	}
	return user, nil
}

// CreateUser inserts a new user into the database
func (r *userRepository) CreateUser(ctx context.Context, userID string) (db.Users, error) {
	user, err := r.queries.CreateUser(ctx, userID)
	if err != nil {
		r.logger.Error("failed to create user", "error", err, "user_id", userID)
		return db.Users{}, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// Ensure the userRepository implements the UserRepository interface
var _ UserRepository = (*userRepository)(nil)
