package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Error struct {
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, data any) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	return err
}

func ErrorJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, message string, err error) {
	logger.Error(message, "error", err, "path", r.URL.Path)

	resp := Error{
		Message: message,
	}

	if jsonErr := JSON(w, status, resp); jsonErr != nil {
		logger.Error("failed to write error response", "error", jsonErr, "path", r.URL.Path)
	}
}
