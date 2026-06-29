package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/e2eauth"
	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	jwksUrlTemplate           = "https://api.stack-auth.com/api/v1/projects/%s/.well-known/jwks.json"
	setUserIDQuery            = "SELECT set_config('app.current_user_id', $1, false) WHERE $1 IS NOT NULL"
	stackAccessTokenClockSkew = time.Minute
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
	localE2E    *LocalE2EAuthConfig
}

type LocalE2EAuthConfig struct {
	Enabled bool
	UserID  string
}

func NewAuthenticator(logger *slog.Logger, jwkCache JWKSProvider, userService UserServiceProvider, dbPool db.DBTX) *Authenticator {
	return &Authenticator{
		logger:      logger,
		jwkCache:    jwkCache,
		userService: userService,
		dbPool:      dbPool,
	}
}

func (a *Authenticator) WithLocalE2EAuth(config LocalE2EAuthConfig) *Authenticator {
	a.localE2E = &LocalE2EAuthConfig{
		Enabled: config.Enabled,
		UserID:  strings.TrimSpace(config.UserID),
	}
	return a
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
				"userID", userID,
				"error_category", "rls_context",
				"error_present", true,
				"error_type", fmt.Sprintf("%T", err))
		} else {
			a.logger.Error("failed to set session variable",
				"userID", userID,
				"error_category", "database",
				"error_present", true,
				"error_type", fmt.Sprintf("%T", err))
		}
		return fmt.Errorf("failed to set user context: %w", err)
	}

	a.logger.Debug("RLS context set successfully",
		"userID", userID)
	return nil
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		if userID, ok, err := a.resolveLocalE2EUserID(r); err != nil {
			a.logger.Warn("invalid local e2e auth header", "path", r.URL.Path, "request_id", request.GetRequestID(r.Context()))
			response.ErrorJSON(w, r, a.logger, http.StatusUnauthorized, "invalid local e2e auth header", err)
			return
		} else if ok {
			if !a.authenticateUser(w, r, next, userID) {
				return
			}
			return
		}

		accessToken := r.Header.Get("x-stack-access-token")
		if accessToken == "" {
			a.logger.Warn("missing access token", "path", r.URL.Path, "method", r.Method, "request_id", request.GetRequestID(r.Context()))
			response.ErrorJSON(w, r, a.logger, http.StatusUnauthorized, "missing access token", nil)
			return
		}

		userID, err := a.jwkCache.GetUserIDFromToken(accessToken)
		if err != nil {
			a.logger.Error("invalid access token",
				"path", r.URL.Path,
				"method", r.Method,
				"status", http.StatusUnauthorized,
				"request_id", request.GetRequestID(r.Context()),
				"error_category", "jwt",
				"error_present", true,
				"error_type", fmt.Sprintf("%T", err))
			response.ErrorJSON(w, r, a.logger, http.StatusUnauthorized, "invalid access token", err)
			return
		}

		a.authenticateUser(w, r, next, userID)
	})
}

func (a *Authenticator) authenticateUser(w http.ResponseWriter, r *http.Request, next http.Handler, userID string) bool {
	dbUser, err := a.userService.EnsureUser(r.Context(), userID)
	if err != nil {
		a.logger.Error("failed to ensure user",
			"userID", userID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", http.StatusInternalServerError,
			"request_id", request.GetRequestID(r.Context()),
			"error_category", "database",
			"error_present", true,
			"error_type", fmt.Sprintf("%T", err))
		response.ErrorJSON(w, r, a.logger, http.StatusInternalServerError, "failed to ensure user", err)
		return false
	}

	// Set the current user ID as a session variable for RLS
	if a.dbPool != nil {
		if err := a.setSessionUserID(r.Context(), userID); err != nil {
			a.logger.Error("failed to set user context",
				"userID", userID,
				"path", r.URL.Path,
				"method", r.Method,
				"status", http.StatusInternalServerError,
				"request_id", request.GetRequestID(r.Context()),
				"error_category", "database",
				"error_present", true,
				"error_type", fmt.Sprintf("%T", err))
			response.ErrorJSON(w, r, a.logger,
				http.StatusInternalServerError,
				"failed to set user context",
				err)
			return false
		}
	}

	ctx := user.WithContext(r.Context(), dbUser.UserID)
	next.ServeHTTP(w, r.WithContext(ctx))
	return true
}

func (a *Authenticator) resolveLocalE2EUserID(r *http.Request) (string, bool, error) {
	headerValue := strings.TrimSpace(r.Header.Get(e2eauth.DevAuthHeaderName))
	if headerValue == "" {
		return "", false, nil
	}
	if a.localE2E == nil || !a.localE2E.Enabled || a.localE2E.UserID == "" {
		return "", false, fmt.Errorf("local e2e auth is disabled")
	}
	if headerValue != a.localE2E.UserID {
		return "", false, fmt.Errorf("unexpected local e2e user")
	}
	return a.localE2E.UserID, true, nil
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
		jwt.WithAcceptableSkew(stackAccessTokenClockSkew),
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
