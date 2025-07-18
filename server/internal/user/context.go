package user

import (
	"context"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

func WithContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func Current(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
