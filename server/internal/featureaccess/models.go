package featureaccess

import "time"

type FeatureAccessGrant struct {
	FeatureKey      string     `json:"feature_key"`
	Source          string     `json:"source"`
	SourceReference *string    `json:"source_reference,omitempty"`
	GrantedBy       *string    `json:"granted_by,omitempty"`
	Note            *string    `json:"note,omitempty"`
	StartsAt        time.Time  `json:"starts_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}
