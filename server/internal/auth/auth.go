package auth

import (
	"encoding/json"
	"net/http"
	"os"
)

func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get("x-stack-access-token")
		if accessToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		req, err := http.NewRequest("GET", "https://api.stack-auth.com/api/v1/users/me", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var user struct {
			ID string `json:"id"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.ID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
