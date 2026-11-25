package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	errorCategoryNone       = "none"
	errorCategoryDatabase   = "database"
	errorCategoryJWT        = "jwt"
	errorCategoryValidation = "validation"
	errorCategoryInternal   = "internal"
)

// JSON writes a JSON response to the HTTP response writer.
//
// Note on error handling: JSON encoding errors are extremely rare in practice and
// typically indicate serious programming bugs (attempting to marshal channels, functions,
// or cyclic data structures). If encoding fails after headers are written, the response
// will be partial. Callers should log encoding errors for debugging. Adding metrics for
// these errors would add complexity without significant value unless errors are observed
// in production logs.
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

// ErrorJSON writes a caller-selected public message, never the error's text.
// Recognized database failures use a generic response. Unclassified failures
// keep the caller's safe message; do not pass an arbitrary err.Error() as message.
func ErrorJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, message string, err error) {
	category := classifyError(err)
	if category == errorCategoryDatabase {
		message = "internal error"
	}
	writeErrorJSON(w, r, logger, status, message, err, category)
}

// InvalidRequestJSON writes a public 400 for a request body that did not decode.
// Decoding provider or stored data is not client validation.
func InvalidRequestJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, message string, err error) {
	writeErrorJSON(w, r, logger, http.StatusBadRequest, message, err, errorCategoryValidation)
}

// InvalidAccessTokenJSON is used only after the access-token verifier rejects
// a token. Its response must not depend on the verifier's diagnostic text.
func InvalidAccessTokenJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	writeErrorJSON(w, r, logger, http.StatusUnauthorized, "unauthorized", err, errorCategoryJWT)
}

func classifyError(err error) string {
	if err == nil {
		return errorCategoryNone
	}
	var pgErr *pgconn.PgError
	var connectErr *pgconn.ConnectError
	var numberErr *strconv.NumError
	var validationErr validator.ValidationErrors
	switch {
	case errors.As(err, &pgErr), errors.As(err, &connectErr):
		return errorCategoryDatabase
	case errors.As(err, &numberErr), errors.Is(err, strconv.ErrSyntax), errors.Is(err, strconv.ErrRange),
		errors.As(err, &validationErr):
		return errorCategoryValidation
	default:
		return errorCategoryInternal
	}
}

func writeErrorJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, message string, err error, category string) {
	requestID := request.GetRequestID(r.Context())
	logAttrs := []any{
		"route", r.Pattern,
		"method", r.Method,
		"status", status,
		"request_id", requestID,
		"response_message", message,
		"error_category", category,
		"error_present", err != nil,
	}
	if err != nil {
		logAttrs = append(logAttrs, "error_type", fmt.Sprintf("%T", err))
	}

	logger.Error("error response", logAttrs...)

	resp := ErrorResponse{
		Message:   message,
		RequestID: requestID,
	}

	if jsonErr := JSON(w, status, resp); jsonErr != nil {
		logger.Error("failed to write error response",
			"route", r.Pattern,
			"method", r.Method,
			"status", status,
			"request_id", requestID,
			"error_type", fmt.Sprintf("%T", jsonErr),
		)
	}
}
