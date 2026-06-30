package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// DecodeStrictJSON decodes a single app-owned JSON request body with a byte cap.
func DecodeStrictJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBytes))
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
