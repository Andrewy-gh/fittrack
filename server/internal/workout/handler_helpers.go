package workout

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

const maxWorkoutJSONBodyBytes = 256 << 10

func (h *WorkoutHandler) decodeWorkoutID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	raw := strings.TrimSpace(r.PathValue("id"))
	if raw == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing workout ID", nil)
		return 0, false
	}

	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || parsed <= 0 {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid workout ID", err)
		return 0, false
	}

	return int32(parsed), true
}

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return request.DecodeStrictJSON(w, r, dst, maxWorkoutJSONBodyBytes)
}
