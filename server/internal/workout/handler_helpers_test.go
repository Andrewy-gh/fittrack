package workout

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkoutHandler_decodeWorkoutIDRejectsInvalidBoundaryValues(t *testing.T) {
	handler := NewHandler(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		validator.New(),
		nil,
	)

	tests := []struct {
		name          string
		pathID        string
		setPathValue  bool
		expectedError string
	}{
		{
			name:          "missing",
			expectedError: "Missing workout ID",
		},
		{
			name:          "negative",
			pathID:        "-1",
			setPathValue:  true,
			expectedError: "Invalid workout ID",
		},
		{
			name:          "zero",
			pathID:        "0",
			setPathValue:  true,
			expectedError: "Invalid workout ID",
		},
		{
			name:          "overflow",
			pathID:        "2147483648",
			setPathValue:  true,
			expectedError: "Invalid workout ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/workouts/"+tt.pathID, nil)
			if tt.setPathValue {
				req.SetPathValue("id", tt.pathID)
			}
			w := httptest.NewRecorder()

			id, ok := handler.decodeWorkoutID(w, req)

			assert.False(t, ok)
			assert.Zero(t, id)
			assert.Equal(t, http.StatusBadRequest, w.Code)

			var resp errorResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Contains(t, resp.Message, tt.expectedError)
		})
	}
}

func TestDecodeStrictJSONRejectsTrailingJSON(t *testing.T) {
	var req CreateWorkoutRequest
	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(
		http.MethodPost,
		"/api/workouts",
		strings.NewReader(`{"date":"2023-01-15T10:00:00Z","exercises":[{"name":"Bench Press","sets":[{"reps":10,"setType":"working"}]}]}{"unexpected":true}`),
	)

	err := decodeStrictJSON(w, httpReq, &req)

	require.Error(t, err)
	assert.Equal(t, "2023-01-15T10:00:00Z", req.Date)
}

func TestDecodeStrictJSONRejectsUnknownWorkoutFields(t *testing.T) {
	var req CreateWorkoutRequest
	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(
		http.MethodPost,
		"/api/workouts",
		strings.NewReader(`{"date":"2023-01-15T10:00:00Z","exercises":[{"name":"Bench Press","sets":[{"reps":10,"setType":"working"}]}],"unexpected":true}`),
	)

	err := decodeStrictJSON(w, httpReq, &req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown field "unexpected"`)
}

func TestDecodeStrictJSONRejectsOversizedWorkoutJSON(t *testing.T) {
	var req CreateWorkoutRequest
	w := httptest.NewRecorder()
	httpReq := httptest.NewRequest(
		http.MethodPost,
		"/api/workouts",
		strings.NewReader(`{"date":"2023-01-15T10:00:00Z","notes":"`+strings.Repeat("x", maxWorkoutJSONBodyBytes)+`","exercises":[{"name":"Bench Press","sets":[{"reps":10,"setType":"working"}]}]}`),
	)

	err := decodeStrictJSON(w, httpReq, &req)

	require.Error(t, err)
	var maxBytesErr *http.MaxBytesError
	assert.True(t, errors.As(err, &maxBytesErr))
	assert.Equal(t, int64(maxWorkoutJSONBodyBytes), maxBytesErr.Limit)
}
