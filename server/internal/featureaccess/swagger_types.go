package featureaccess

import "time"

type FeatureAccessResponse struct {
	FeatureKey      string     `json:"feature_key" validate:"required" example:"ai_chatbot"`
	Source          string     `json:"source" validate:"required" example:"manual"`
	SourceReference *string    `json:"source_reference,omitempty" example:"sub_123"`
	GrantedBy       *string    `json:"granted_by,omitempty" example:"andy"`
	Note            *string    `json:"note,omitempty" example:"dev demo access"`
	StartsAt        time.Time  `json:"starts_at" validate:"required" example:"2026-03-25T12:00:00Z"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" example:"2026-04-25T12:00:00Z"`
	CreatedAt       time.Time  `json:"created_at" validate:"required" example:"2026-03-25T12:00:00Z"`
}
