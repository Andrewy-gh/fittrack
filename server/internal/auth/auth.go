package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

// Middleware authenticates requests and ensures the user exists
func Middleware(next http.Handler, logger *slog.Logger, userService *user.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		accessToken := r.Header.Get("x-stack-access-token")
		if accessToken == "" {
			response.ErrorJSON(w, r, logger, http.StatusUnauthorized, "missing access token", nil)
			return
		}

		// Call Stack Auth API to validate the token and get user info
		authUser, err := authenticateWithStack(accessToken, logger)
		if err != nil {
			response.ErrorJSON(w, r, logger, http.StatusUnauthorized, "invalid access token", err)
			return
		}

		// Ensure the user exists in our database
		dbUser, err := userService.EnsureUser(r.Context(), authUser.ID)
		if err != nil {
			response.ErrorJSON(w, r, logger, http.StatusInternalServerError, "failed to ensure user", err)
			return
		}

		// Add the user to the context
		ctx := context.WithValue(r.Context(), "user", dbUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type stackAuthResponse struct {
	ID string `json:"id"`
}

func authenticateWithStack(accessToken string, logger *slog.Logger) (*stackAuthResponse, error) {
	req, err := http.NewRequest("GET", "https://api.stack-auth.com/api/v1/users/me", nil)
	if err != nil {
		logger.Error("failed to create auth request", "error", err)
		return nil, err
	}

	projectID := os.Getenv("PROJECT_ID")
	secretKey := os.Getenv("SECRET_SERVER_KEY")

	if projectID == "" || secretKey == "" {
		logger.Error("missing required environment variables", "has_project_id", projectID != "", "has_secret_key", secretKey != "")
		return nil, fmt.Errorf("missing required environment variables")
	}

	req.Header.Set("x-stack-access-type", "server")
	req.Header.Set("x-stack-project-id", projectID)
	req.Header.Set("x-stack-secret-server-key", secretKey)
	req.Header.Set("x-stack-access-token", accessToken)

	logger.Debug("sending auth request to stack-auth", "url", req.URL, "headers", req.Header)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to send auth request", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("auth service returned non-OK status",
			"status", resp.StatusCode,
			"status_text", http.StatusText(resp.StatusCode),
			"response_body", string(body))
		return nil, fmt.Errorf("auth service returned status: %d", resp.StatusCode)
	}

	var user stackAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		logger.Error("failed to decode auth response", "error", err)
		return nil, err
	}

	if user.ID == "" {
		logger.Warn("received empty user ID in auth response", "response_data", user)
		return nil, fmt.Errorf("empty user ID in auth response")
	}

	logger.Info("successfully authenticated user", "user_id", user.ID)
	return &user, nil
}
