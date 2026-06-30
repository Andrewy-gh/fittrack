package account

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepository struct {
	deleteErr     error
	deleteCalled  bool
	deletedUserID string
}

func (s *stubRepository) DeleteUser(ctx context.Context, userID string) error {
	s.deleteCalled = true
	s.deletedUserID = userID
	return s.deleteErr
}

type stubBillingCanceler struct {
	err    error
	called bool
}

func (s *stubBillingCanceler) CancelCurrentSubscriptionImmediately(ctx context.Context) error {
	s.called = true
	return s.err
}

func TestServiceDeleteCurrentUser_CancelsSubscriptionImmediatelyBeforeDeleting(t *testing.T) {
	repo := &stubRepository{}
	billing := &stubBillingCanceler{}
	service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, billing)
	ctx := user.WithContext(context.Background(), "user-123")

	err := service.DeleteCurrentUser(ctx)

	require.NoError(t, err)
	assert.True(t, billing.called)
	assert.Equal(t, "user-123", repo.deletedUserID)
}

func TestServiceDeleteCurrentUser_DeletesCurrentUser(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")

	err := service.DeleteCurrentUser(ctx)

	require.NoError(t, err)
	assert.Equal(t, "user-123", repo.deletedUserID)
}

func TestServiceDeleteCurrentUser_ReturnsDeleteFailure(t *testing.T) {
	expectedErr := errors.New("delete failed")
	repo := &stubRepository{deleteErr: expectedErr}
	service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")

	err := service.DeleteCurrentUser(ctx)

	require.ErrorIs(t, err, expectedErr)
}

func TestServiceDeleteCurrentUser_ReturnsZeroRowDeleteFailure(t *testing.T) {
	repo := &stubRepository{deleteErr: ErrUserNotDeleted}
	service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")

	err := service.DeleteCurrentUser(ctx)

	require.ErrorIs(t, err, ErrUserNotDeleted)
}

func TestServiceDeleteCurrentUser_DoesNotDeleteWhenSubscriptionCancellationFails(t *testing.T) {
	expectedErr := errors.New("stripe failed")
	repo := &stubRepository{}
	billing := &stubBillingCanceler{err: expectedErr}
	service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, billing)
	ctx := user.WithContext(context.Background(), "user-123")

	err := service.DeleteCurrentUser(ctx)

	require.ErrorIs(t, err, expectedErr)
	assert.True(t, billing.called)
	assert.False(t, repo.deleteCalled)
	assert.Empty(t, repo.deletedUserID)
}

func TestServiceDeleteCurrentUser_ReturnsLocalDeletionFailureAfterSubscriptionCancellation(t *testing.T) {
	expectedErr := errors.New("delete failed")
	repo := &stubRepository{deleteErr: expectedErr}
	billing := &stubBillingCanceler{}
	service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, billing)
	ctx := user.WithContext(context.Background(), "user-123")

	err := service.DeleteCurrentUser(ctx)

	require.ErrorIs(t, err, expectedErr)
	assert.ErrorContains(t, err, "delete local account data after subscription cancellation")
	assert.True(t, billing.called)
	assert.True(t, repo.deleteCalled)
	assert.Equal(t, "user-123", repo.deletedUserID)
}
