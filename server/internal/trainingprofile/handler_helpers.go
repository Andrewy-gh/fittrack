package trainingprofile

import (
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
)

const maxTrainingProfileJSONBodyBytes = 32 << 10

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return request.DecodeStrictJSON(w, r, dst, maxTrainingProfileJSONBodyBytes)
}
