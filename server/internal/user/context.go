package user

import (
	"context"
)

type ctxKey struct{}

// WithContext adds a user to the given context.
func WithContext(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ctxKey{}, u)
}

// Current retrieves the user from the context.
// It returns the user and true if the user was found, otherwise it returns an empty user and false.
func Current(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(ctxKey{}).(User)
	return u, ok
}
