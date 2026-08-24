package errors

import "testing"

func TestUnauthorized_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Unauthorized
		expected string
	}{
		{
			name: "with resource and user ID",
			err: &Unauthorized{
				Resource: "exercise",
				UserID:   "user123",
			},
			expected: "user user123 is not authorized to access exercise",
		},
		{
			name: "with resource only",
			err: &Unauthorized{
				Resource: "workout",
			},
			expected: "not authorized to access workout",
		},
		{
			name:     "empty fields",
			err:      &Unauthorized{},
			expected: "unauthorized access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Unauthorized.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNotFound_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *NotFound
		expected string
	}{
		{
			name: "with resource and ID",
			err: &NotFound{
				Resource: "exercise",
				ID:       "123",
			},
			expected: "exercise with id 123 not found",
		},
		{
			name: "with resource only",
			err: &NotFound{
				Resource: "workout",
			},
			expected: "workout not found",
		},
		{
			name:     "empty fields",
			err:      &NotFound{},
			expected: "resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("NotFound.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}
