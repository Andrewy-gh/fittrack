package trainingprofile

import (
	"encoding/json"
	"fmt"
	"strings"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

type ProfileResponse struct {
	PrimaryGoal                     *string   `json:"primary_goal"`
	ExperienceLevel                 *string   `json:"experience_level"`
	PreferredSessionDurationMinutes *int32    `json:"preferred_session_duration_minutes"`
	UsualTrainingLocation           *string   `json:"usual_training_location"`
	AvailableEquipment              []string  `json:"available_equipment"`
	AvoidedExercises                []string  `json:"avoided_exercises"`
	MovementLimitations             *[]string `json:"movement_limitations"`
}

type UpdateProfileRequest struct {
	PrimaryGoal                     *string   `json:"primary_goal"`
	ExperienceLevel                 *string   `json:"experience_level"`
	PreferredSessionDurationMinutes *int32    `json:"preferred_session_duration_minutes"`
	UsualTrainingLocation           *string   `json:"usual_training_location"`
	AvailableEquipment              []string  `json:"available_equipment"`
	AvoidedExercises                []string  `json:"avoided_exercises"`
	MovementLimitations             *[]string `json:"movement_limitations"`
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
		AvailableEquipment:  []string{},
		AvoidedExercises:    []string{},
		MovementLimitations: nil,
	}
}

func profileResponseFromRow(row db.UserTrainingProfile) (*ProfileResponse, error) {
	profile := emptyProfileResponse()
	profile.PrimaryGoal = pgTextPtr(row.PrimaryGoal)
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
