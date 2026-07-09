package trainingprofile

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	maxProfileListItems = 20
	maxProfileListChars = 120
)

var allowedGoals = map[string]struct{}{
	"strength":        {},
	"hypertrophy":     {},
	"endurance":       {},
	"general_fitness": {},
	"weight_loss":     {},
	"mobility":        {},
}

var allowedExperienceLevels = map[string]struct{}{
	"beginner":     {},
	"intermediate": {},
	"advanced":     {},
}

var allowedLocations = map[string]struct{}{
	"gym":     {},
	"home":    {},
	"outdoor": {},
	"travel":  {},
}

func validateProfileRequest(req UpdateProfileRequest) (*UpdateProfileRequest, error) {
	normalized := UpdateProfileRequest{
		PrimaryGoal:                     normalizeOptionalText(req.PrimaryGoal),
		ExperienceLevel:                 normalizeOptionalText(req.ExperienceLevel),
		PreferredSessionDurationMinutes: req.PreferredSessionDurationMinutes,
		UsualTrainingLocation:           normalizeOptionalText(req.UsualTrainingLocation),
		AvailableEquipment:              cleanStringList(req.AvailableEquipment),
		AvoidedExercises:                cleanStringList(req.AvoidedExercises),
		MovementLimitations:             nil,
	}

	if normalized.PrimaryGoal != nil {
		if _, ok := allowedGoals[*normalized.PrimaryGoal]; !ok {
			return nil, &ValidationError{Field: "primary_goal", Message: "must be strength, hypertrophy, endurance, general_fitness, weight_loss, mobility, or null"}
		}
	}
	if normalized.ExperienceLevel != nil {
		if _, ok := allowedExperienceLevels[*normalized.ExperienceLevel]; !ok {
			return nil, &ValidationError{Field: "experience_level", Message: "must be beginner, intermediate, advanced, or null"}
		}
	}
	if normalized.PreferredSessionDurationMinutes != nil {
		duration := *normalized.PreferredSessionDurationMinutes
		if duration < 10 || duration > 240 {
			return nil, &ValidationError{Field: "preferred_session_duration_minutes", Message: "must be between 10 and 240 minutes, or null"}
		}
	}
	if normalized.UsualTrainingLocation != nil {
		if _, ok := allowedLocations[*normalized.UsualTrainingLocation]; !ok {
			return nil, &ValidationError{Field: "usual_training_location", Message: "must be gym, home, outdoor, travel, or null"}
		}
	}
	if req.MovementLimitations != nil {
		values := cleanStringList(*req.MovementLimitations)
		normalized.MovementLimitations = &values
	}

	return &normalized, nil
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func pgTextPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	trimmed := strings.TrimSpace(value.String)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func cleanStringList(values []string) []string {
	cleaned := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if len(trimmed) > maxProfileListChars {
			trimmed = trimmed[:maxProfileListChars]
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		cleaned = append(cleaned, trimmed)
		if len(cleaned) == maxProfileListItems {
			break
		}
	}
	return cleaned
}
