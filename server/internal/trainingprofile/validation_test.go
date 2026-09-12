package trainingprofile

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProfileRequest(t *testing.T) {
	duration := int32(45)
	limitations := []string{" knee pain ", "Knee Pain", strings.Repeat("x", 140)}

	req, err := validateProfileRequest(UpdateProfileRequest{
		PrimaryGoal:                     ptrString("strength"),
		ExperienceLevel:                 ptrString("intermediate"),
		PreferredSessionDurationMinutes: &duration,
		UsualTrainingLocation:           ptrString("home"),
		AvailableEquipment:              []string{" dumbbells ", "Dumbbells", ""},
		AvoidedExercises:                []string{"burpees"},
		MovementLimitations:             &limitations,
	})

	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, "strength", *req.PrimaryGoal)
	assert.Equal(t, []string{"dumbbells"}, req.AvailableEquipment)
	require.NotNil(t, req.MovementLimitations)
	assert.Equal(t, []string{"knee pain", strings.Repeat("x", 120)}, *req.MovementLimitations)
}

func TestValidateProfileRequestRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		req  UpdateProfileRequest
		want string
	}{
		{
			name: "goal",
			req:  UpdateProfileRequest{PrimaryGoal: ptrString("power")},
			want: "primary_goal",
		},
		{
			name: "experience",
			req:  UpdateProfileRequest{ExperienceLevel: ptrString("expert")},
			want: "experience_level",
		},
		{
			name: "duration",
			req:  UpdateProfileRequest{PreferredSessionDurationMinutes: ptrInt32(9)},
			want: "preferred_session_duration_minutes",
		},
		{
			name: "location",
			req:  UpdateProfileRequest{UsualTrainingLocation: ptrString("office")},
			want: "usual_training_location",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateProfileRequest(tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func ptrString(value string) *string {
	return &value
}

func ptrInt32(value int32) *int32 {
	return &value
}

func TestMultipleGoals(t *testing.T) {
	for _, tc := range []struct {
		name    string
		goals   []string
		want    []string
		invalid bool
	}{
		{"multiple", []string{"strength", "mobility"}, []string{"strength", "mobility"}, false},
		{"deduplicate", []string{"strength", "strength"}, []string{"strength"}, false},
		{"clear", []string{}, []string{}, false},
		{"invalid", []string{"strength", "power"}, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateProfileRequest(UpdateProfileRequest{Goals: tc.goals, PrimaryGoal: ptrString("endurance")})
			if tc.invalid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.Goals)
			if len(tc.want) == 0 {
				assert.Nil(t, got.PrimaryGoal)
			} else {
				require.NotNil(t, got.PrimaryGoal)
				assert.Equal(t, tc.want[0], *got.PrimaryGoal)
			}
		})
	}
}
