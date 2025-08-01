package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	// JWKSURLTemplate is the template for Stack Auth JWKS endpoint
	JWKSURLTemplate = "https://api.stack-auth.com/api/v1/projects/%s/.well-known/jwks.json"
)

type JWKSCache struct {
	keySet jwk.Set
}

// NewJWKSCache creates a new JWKS cache with automatic refresh
func NewJWKSCache(ctx context.Context, projectID string) (*JWKSCache, error) {
	cache, err := jwk.NewCache(ctx, httprc.NewClient())
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS cache: %w", err)
	}
	jwksURL := fmt.Sprintf(JWKSURLTemplate, projectID)
	// Register the URL with refresh policy
	if err = cache.Register(ctx, jwksURL,
		jwk.WithMinInterval(15*time.Minute), // never refresh more often
		jwk.WithMaxInterval(24*time.Hour),   // refresh at least once a day
	); err != nil {
		return nil, fmt.Errorf("failed to register JWKS URL: %w", err)
	}

	// Get the cached key set
	cachedSet, err := cache.CachedSet(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached key set: %w", err)
	}

	return &JWKSCache{keySet: cachedSet}, nil
}

// ParseAndValidateToken parses and validates a JWT token using the cached JWKS
func (j *JWKSCache) ParseAndValidateToken(tokenString string) (jwt.Token, error) {
	token, err := jwt.ParseString(
		tokenString,
		jwt.WithKeySet(j.keySet),
		jwt.WithValidate(true), // Verify exp/nbf/aud etc.
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse/validate token: %w", err)
	}

	return token, nil
}

// GetUserIDFromToken extracts the user ID from a JWT token
func (j *JWKSCache) GetUserIDFromToken(tokenString string) (string, error) {
	token, err := j.ParseAndValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	userID, ok := token.Subject()
	if !ok || userID == "" {
		return "", fmt.Errorf("token missing required 'sub' claim")
	}

	return userID, nil
}
