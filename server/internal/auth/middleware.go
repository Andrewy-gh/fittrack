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
	setUserIDQuery  = "SELECT set_config('app.current_user_id', $1, false) WHERE $1 IS NOT NULL"
)

type JWKSProvider interface {
	GetUserIDFromToken(tokenString string) (string, error)
}

type UserServiceProvider interface {
	EnsureUser(ctx context.Context, userID string) (db.Users, error)
}

type Authenticator struct {
	logger      *slog.Logger
	jwkCache    JWKSProvider
	userService UserServiceProvider
	dbPool      db.DBTX
}

func NewAuthenticator(logger *slog.Logger, jwkCache JWKSProvider, userService UserServiceProvider, dbPool db.DBTX) *Authenticator {
	return &Authenticator{
		logger:      logger,
		jwkCache:    jwkCache,
		userService: userService,
		dbPool:      dbPool,
	}
}

func (a *Authenticator) setSessionUserID(ctx context.Context, userID string) error {
	if a.dbPool == nil {
		return nil
	}
	_, err := a.dbPool.Exec(ctx,
		setUserIDQuery,
		userID)
	if err != nil {
		// Check if this is an RLS context error
		if db.IsRLSContextError(err) {
			a.logger.Error("RLS context setup failed",
				"error", err,
				"userID", userID,
				"error_type", "rls_context")
			a.logger.Debug("raw RLS context error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "userID", userID)
		} else {
			a.logger.Error("failed to set session variable",
				"error", err,
				"userID", userID)
			a.logger.Debug("raw session variable error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "userID", userID)
		}
		return fmt.Errorf("failed to set user context: %w", err)
	}

	a.logger.Debug("RLS context set successfully",
		"userID", userID)
	return nil
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers for all API requests
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-stack-access-token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// Handle CORS preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		accessToken := r.Header.Get("x-stack-access-token")
		if accessToken == "" {
			a.logger.Warn("missing access token", "path", r.URL.Path, "method", r.Method)
			response.ErrorJSON(w, r, a.logger, http.StatusUnauthorized, "missing access token", nil)
			return
		}

		userID, err := a.jwkCache.GetUserIDFromToken(accessToken)
		if err != nil {
			a.logger.Error("invalid access token", "error", err, "path", r.URL.Path)
			response.ErrorJSON(w, r, a.logger, http.StatusUnauthorized, "invalid access token", err)
			return
		}

		dbUser, err := a.userService.EnsureUser(r.Context(), userID)
		if err != nil {
			a.logger.Error("failed to ensure user", "error", err, "userID", userID)
			response.ErrorJSON(w, r, a.logger, http.StatusInternalServerError, "failed to ensure user", err)
			return
		}

		// Set the current user ID as a session variable for RLS
		if a.dbPool != nil {
			if err := a.setSessionUserID(r.Context(), userID); err != nil {
				a.logger.Error("failed to set user context",
					"error", err,
					"userID", userID,
					"path", r.URL.Path)
				response.ErrorJSON(w, r, a.logger,
					http.StatusInternalServerError,
					"failed to set user context",
					err)
				return
			}
		}

		ctx := user.WithContext(r.Context(), dbUser.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type JWKSCache struct {
	keySet  jwk.Set
	cache   *jwk.Cache
	jwksURL string
}

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

var (
	_ JWKSProvider        = (*JWKSCache)(nil)
	_ UserServiceProvider = (*user.Service)(nil)
)
