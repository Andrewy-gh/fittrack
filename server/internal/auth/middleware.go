package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	jwksUrlTemplate = "https://api.stack-auth.com/api/v1/projects/%s/.well-known/jwks.json"
)

// JWKSProvider defines the interface for JWT validation
type JWKSProvider interface {
	GetUserIDFromToken(tokenString string) (string, error)
}

// UserServiceProvider defines the interface for user management
type UserServiceProvider interface {
	EnsureUser(ctx context.Context, userID string) (db.Users, error)
}

// Authenticator provides authentication middleware
type Authenticator struct {
	logger      *slog.Logger
	jwkCache    JWKSProvider
	userService UserServiceProvider
	dbPool      db.DBTX
}

// NewAuthenticator creates a new Authenticator
func NewAuthenticator(logger *slog.Logger, jwkCache JWKSProvider, userService UserServiceProvider, dbPool db.DBTX) *Authenticator {
	return &Authenticator{
		logger:      logger,
		jwkCache:    jwkCache,
		userService: userService,
		dbPool:      dbPool,
	}
}

// Middleware authenticates requests and ensures the user exists
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		accessToken := r.Header.Get("x-stack-access-token")
		if accessToken == "" {
			response.ErrorJSON(w, r, a.logger, http.StatusUnauthorized, "missing access token", nil)
			return
		}

		userID, err := a.jwkCache.GetUserIDFromToken(accessToken)
		if err != nil {
			response.ErrorJSON(w, r, a.logger, http.StatusUnauthorized, "invalid access token", err)
			return
		}

		dbUser, err := a.userService.EnsureUser(r.Context(), userID)
		if err != nil {
			response.ErrorJSON(w, r, a.logger, http.StatusInternalServerError, "failed to ensure user", err)
			return
		}

		// Set the current user ID as a session variable for RLS
		// This must be done for each request when using connection pooling
		// to ensure proper user context isolation
		if a.dbPool != nil {
			_, err = a.dbPool.Exec(r.Context(), "SELECT set_config('app.current_user_id', $1, false)", userID)
			if err != nil {
				response.ErrorJSON(w, r, a.logger, http.StatusInternalServerError, "failed to set user context", err)
				return
			}
		}

		ctx := user.WithContext(r.Context(), dbUser.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// JWKSCache holds the cached JWK set
type JWKSCache struct {
	keySet  jwk.Set
	cache   *jwk.Cache
	jwksURL string
}

// NewJWKSCache creates a new JWKS cache with automatic refresh
func NewJWKSCache(ctx context.Context, projectID string) (*JWKSCache, error) {
	cache, err := jwk.NewCache(ctx, httprc.NewClient())
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS cache: %w", err)
	}

	jwksURL := fmt.Sprintf(jwksUrlTemplate, projectID)
	if err := cache.Register(ctx, jwksURL,
		jwk.WithMinInterval(15*time.Minute),
		jwk.WithMaxInterval(24*time.Hour),
	); err != nil {
		return nil, fmt.Errorf("failed to register JWKS URL: %w", err)
	}

	cachedSet, err := cache.CachedSet(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached key set: %w", err)
	}

	return &JWKSCache{keySet: cachedSet, cache: cache, jwksURL: jwksURL}, nil
}

// GetUserIDFromToken extracts the user ID from a JWT token
func (j *JWKSCache) GetUserIDFromToken(tokenString string) (string, error) {
	token, err := jwt.ParseString(
		tokenString,
		jwt.WithKeySet(j.keySet),
		jwt.WithValidate(true),
	)
	if err != nil {
		return "", fmt.Errorf("failed to parse/validate token: %w", err)
	}

	userID, ok := token.Subject()
	if !ok || userID == "" {
		return "", fmt.Errorf("token missing required 'sub' claim")
	}

	return userID, nil
}
