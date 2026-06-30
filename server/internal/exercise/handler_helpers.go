package exercise

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

func (h *ExerciseHandler) decodeExerciseID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	raw := strings.TrimSpace(r.PathValue("id"))
	if raw == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing exercise ID", nil)
		return 0, false
	}

	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || parsed <= 0 {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID", err)
		return 0, false
	}

	return int32(parsed), true
}

func decodeStrictJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return err
		}
		return errors.New("request body must contain a single JSON value")
	}

	return nil
}
