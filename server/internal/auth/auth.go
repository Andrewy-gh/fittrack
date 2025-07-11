package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

func Middleware(next http.Handler, logger *slog.Logger) http.Handler {
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

		req, err := http.NewRequest("GET", "https://api.stack-auth.com/api/v1/users/me", nil)
		if err != nil {
			response.ErrorJSON(w, r, logger, http.StatusInternalServerError, "failed to create auth request", err)
			return
		}

		projectId := os.Getenv("PROJECT_ID")
		secretKey := os.Getenv("SECRET_SERVER_KEY")

		req.Header.Set("x-stack-access-type", "server")
		req.Header.Set("x-stack-project-id", projectId)
		req.Header.Set("x-stack-secret-server-key", secretKey)
		req.Header.Set("x-stack-access-token", accessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			response.ErrorJSON(w, r, logger, http.StatusInternalServerError, "failed to perform auth request", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			response.ErrorJSON(w, r, logger, http.StatusUnauthorized, "invalid access token", nil)
			return
		}

		var user struct {
			ID string `json:"id"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			response.ErrorJSON(w, r, logger, http.StatusInternalServerError, "failed to decode user from auth response", err)
			return
		}

		if user.ID == "" {
			response.ErrorJSON(w, r, logger, http.StatusUnauthorized, "unauthorized", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
