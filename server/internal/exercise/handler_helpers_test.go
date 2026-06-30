package exercise

import (
	"encoding/json"
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

func TestExerciseHandler_decodeExerciseIDRejectsInvalidBoundaryValues(t *testing.T) {
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
			expectedError: "Missing exercise ID",
		},
		{
			name:          "negative",
			pathID:        "-1",
			setPathValue:  true,
			expectedError: "Invalid exercise ID",
		},
		{
			name:          "zero",
			pathID:        "0",
			setPathValue:  true,
			expectedError: "Invalid exercise ID",
		},
		{
			name:          "overflow",
			pathID:        "2147483648",
			setPathValue:  true,
			expectedError: "Invalid exercise ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/exercises/"+tt.pathID, nil)
			if tt.setPathValue {
				req.SetPathValue("id", tt.pathID)
			}
			w := httptest.NewRecorder()

			id, ok := handler.decodeExerciseID(w, req)

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
	var req CreateExerciseRequest
	httpReq := httptest.NewRequest(
		http.MethodPost,
		"/api/exercises",
		strings.NewReader(`{"name":"Bench Press"}{"unexpected":true}`),
	)

	err := decodeStrictJSON(httpReq, &req)

	require.Error(t, err)
	assert.Equal(t, CreateExerciseRequest{Name: "Bench Press"}, req)
}
