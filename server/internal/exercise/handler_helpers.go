package exercise

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

const maxExerciseJSONBodyBytes = 32 << 10

func (h *ExerciseHandler) decodeExerciseID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	raw := strings.TrimSpace(r.PathValue("id"))
	if raw == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing exercise ID", nil)
		return 0, false
	}

	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || parsed <= 0 {
		// ParseInt's *strconv.NumError includes the raw path segment. Keep
		// only its syntax/range cause so the public message is input-independent.
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID", errors.Unwrap(err))
		return 0, false
	}

	return int32(parsed), true
}

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return request.DecodeStrictJSON(w, r, dst, maxExerciseJSONBodyBytes)
}
