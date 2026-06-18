package account

import (
	"context"
	"io"
	"log/slog"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

type fakeDeleteDB struct {
	commandTag pgconn.CommandTag
	err        error
}

func (f fakeDeleteDB) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return f.commandTag, f.err
}

func (f fakeDeleteDB) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	panic("query is not used by DeleteUser")
}

func (f fakeDeleteDB) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	panic("query row is not used by DeleteUser")
}

func TestRepositoryDeleteUser_ReturnsErrorWhenNoRowsDeleted(t *testing.T) {
	repo := NewRepository(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		db.New(fakeDeleteDB{commandTag: pgconn.NewCommandTag("DELETE 0")}),
	)

	err := repo.DeleteUser(context.Background(), "user-123")

	require.ErrorIs(t, err, ErrUserNotDeleted)
}

func TestRepositoryDeleteUser_SucceedsWhenUserDeleted(t *testing.T) {
	repo := NewRepository(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		db.New(fakeDeleteDB{commandTag: pgconn.NewCommandTag("DELETE 1")}),
	)

	err := repo.DeleteUser(context.Background(), "user-123")

	require.NoError(t, err)
}
