package trainingprofile

import (
	"encoding/json"
	"fmt"
	"strings"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

type ProfileResponse struct {
	Goals                           []string  `json:"goals" validate:"required"`
	PrimaryGoal                     *string   `json:"primary_goal" extensions:"x-nullable" validate:"required"`
	ExperienceLevel                 *string   `json:"experience_level" extensions:"x-nullable" validate:"required"`
	PreferredSessionDurationMinutes *int32    `json:"preferred_session_duration_minutes" extensions:"x-nullable" validate:"required"`
	UsualTrainingLocation           *string   `json:"usual_training_location" extensions:"x-nullable" validate:"required"`
	AvailableEquipment              []string  `json:"available_equipment" validate:"required"`
	AvoidedExercises                []string  `json:"avoided_exercises" validate:"required"`
	MovementLimitations             *[]string `json:"movement_limitations" extensions:"x-nullable" validate:"required"`
}

type UpdateProfileRequest struct {
	Goals                           []string  `json:"goals"`
	PrimaryGoal                     *string   `json:"primary_goal" extensions:"x-nullable"`
	ExperienceLevel                 *string   `json:"experience_level" extensions:"x-nullable"`
	PreferredSessionDurationMinutes *int32    `json:"preferred_session_duration_minutes" extensions:"x-nullable"`
	UsualTrainingLocation           *string   `json:"usual_training_location" extensions:"x-nullable"`
	AvailableEquipment              []string  `json:"available_equipment"`
	AvoidedExercises                []string  `json:"avoided_exercises"`
	MovementLimitations             *[]string `json:"movement_limitations" extensions:"x-nullable"`
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if strings.TrimSpace(e.Field) == "" {
		return e.Message
	}
	return e.Field + ": " + e.Message
}

func emptyProfileResponse() *ProfileResponse {
	return &ProfileResponse{
		Goals:               []string{},
		AvailableEquipment:  []string{},
		AvoidedExercises:    []string{},
		MovementLimitations: nil,
	}
}

func profileResponseFromRow(row db.UserTrainingProfile) (*ProfileResponse, error) {
	profile := emptyProfileResponse()
	profile.PrimaryGoal = pgTextPtr(row.PrimaryGoal)
	profile.Goals = row.Goals
	if profile.Goals == nil {
		profile.Goals = []string{}
	}
	profile.ExperienceLevel = pgTextPtr(row.ExperienceLevel)
	profile.UsualTrainingLocation = pgTextPtr(row.UsualTrainingLocation)
	if row.PreferredSessionDurationMinutes.Valid {
		value := row.PreferredSessionDurationMinutes.Int32
		profile.PreferredSessionDurationMinutes = &value
	}

	var err error
	profile.AvailableEquipment, err = decodeStringArray(row.AvailableEquipment)
	if err != nil {
		return nil, fmt.Errorf("decode available equipment: %w", err)
	}
	profile.AvoidedExercises, err = decodeStringArray(row.AvoidedExercises)
	if err != nil {
		return nil, fmt.Errorf("decode avoided exercises: %w", err)
	}
	if row.MovementLimitations != nil {
		limitations, err := decodeStringArray(row.MovementLimitations)
		if err != nil {
			return nil, fmt.Errorf("decode movement limitations: %w", err)
		}
		profile.MovementLimitations = &limitations
	}

	return profile, nil
}

func decodeStringArray(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return cleanStringList(values), nil
}
