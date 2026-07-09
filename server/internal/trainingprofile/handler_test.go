package trainingprofile

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubProfileService struct {
	getProfile    *ProfileResponse
	getErr        error
	upsertProfile *ProfileResponse
	upsertErr     error
	upsertReq     UpdateProfileRequest
}

func (s *stubProfileService) Get(context.Context) (*ProfileResponse, error) {
	return s.getProfile, s.getErr
}

func (s *stubProfileService) Upsert(_ context.Context, req UpdateProfileRequest) (*ProfileResponse, error) {
	s.upsertReq = req
	return s.upsertProfile, s.upsertErr
}

func TestHandlerGet(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("writes empty first-time profile shape", func(t *testing.T) {
		handler := NewHandler(logger, &stubProfileService{
			getProfile: emptyProfileResponse(),
		})
		req := httptest.NewRequest(http.MethodGet, "/api/training-profile", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, `{
			"primary_goal": null,
			"experience_level": null,
			"preferred_session_duration_minutes": null,
			"usual_training_location": null,
			"available_equipment": [],
			"avoided_exercises": [],
			"movement_limitations": null
		}`, rr.Body.String())
	})

	t.Run("preserves movement limitations tri-state", func(t *testing.T) {
		limitations := []string{}
		handler := NewHandler(logger, &stubProfileService{
			getProfile: &ProfileResponse{
				AvailableEquipment:  []string{},
				AvoidedExercises:    []string{},
				MovementLimitations: &limitations,
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/api/training-profile", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"movement_limitations":[]`)
	})

	t.Run("maps unauthorized", func(t *testing.T) {
		handler := NewHandler(logger, &stubProfileService{
			getErr: &apperrors.Unauthorized{Resource: "training profile"},
		})
		req := httptest.NewRequest(http.MethodGet, "/api/training-profile", nil)
		rr := httptest.NewRecorder()

		handler.Get(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "not authorized")
	})
}

func TestHandlerUpsert(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("decodes full document and writes response", func(t *testing.T) {
		goal := "strength"
		duration := int32(45)
		limitations := []string{"knee pain"}
		handler := NewHandler(logger, &stubProfileService{
			upsertProfile: &ProfileResponse{
				PrimaryGoal:                     &goal,
				PreferredSessionDurationMinutes: &duration,
				AvailableEquipment:              []string{"dumbbells"},
				AvoidedExercises:                []string{"burpees"},
				MovementLimitations:             &limitations,
			},
		})
		req := httptest.NewRequest(http.MethodPut, "/api/training-profile", strings.NewReader(`{
			"primary_goal": "strength",
			"experience_level": null,
			"preferred_session_duration_minutes": 45,
			"usual_training_location": null,
			"available_equipment": ["dumbbells"],
			"avoided_exercises": ["burpees"],
			"movement_limitations": ["knee pain"]
		}`))
		rr := httptest.NewRecorder()

		handler.Upsert(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"primary_goal":"strength"`)
		assert.Contains(t, rr.Body.String(), `"movement_limitations":["knee pain"]`)
	})

	t.Run("maps validation errors to bad request", func(t *testing.T) {
		handler := NewHandler(logger, &stubProfileService{
			upsertErr: &ValidationError{Field: "primary_goal", Message: "must be supported"},
		})
		req := httptest.NewRequest(http.MethodPut, "/api/training-profile", strings.NewReader(`{
			"primary_goal": "powerbuilding",
			"experience_level": null,
			"preferred_session_duration_minutes": null,
			"usual_training_location": null,
			"available_equipment": [],
			"avoided_exercises": [],
			"movement_limitations": null
		}`))
		rr := httptest.NewRecorder()

		handler.Upsert(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "primary_goal")
	})

	t.Run("rejects unknown JSON fields", func(t *testing.T) {
		handler := NewHandler(logger, &stubProfileService{})
		req := httptest.NewRequest(http.MethodPut, "/api/training-profile", strings.NewReader(`{"unexpected": true}`))
		rr := httptest.NewRecorder()

		handler.Upsert(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("maps service failures", func(t *testing.T) {
		handler := NewHandler(logger, &stubProfileService{
			upsertErr: errors.New("database unavailable"),
		})
		req := httptest.NewRequest(http.MethodPut, "/api/training-profile", strings.NewReader(`{
			"primary_goal": null,
			"experience_level": null,
			"preferred_session_duration_minutes": null,
			"usual_training_location": null,
			"available_equipment": [],
			"avoided_exercises": [],
			"movement_limitations": null
		}`))
		rr := httptest.NewRecorder()

		handler.Upsert(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "internal error")
	})
}

func TestProfileResponseJSONTriState(t *testing.T) {
	base := emptyProfileResponse()
	raw, err := json.Marshal(base)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"movement_limitations":null`)

	none := []string{}
	base.MovementLimitations = &none
	raw, err = json.Marshal(base)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"movement_limitations":[]`)

	values := []string{"no deep squats"}
	base.MovementLimitations = &values
	raw, err = json.Marshal(base)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"movement_limitations":["no deep squats"]`)
}
