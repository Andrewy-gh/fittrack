package errors

import (
	"errors"
	"fmt"
	"testing"
)

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

func TestNewUnauthorized(t *testing.T) {
	resource := "exercise"
	userID := "user123"
	err := NewUnauthorized(resource, userID)

	if err.Resource != resource {
		t.Errorf("NewUnauthorized() Resource = %v, want %v", err.Resource, resource)
	}
	if err.UserID != userID {
		t.Errorf("NewUnauthorized() UserID = %v, want %v", err.UserID, userID)
	}

	expected := "user user123 is not authorized to access exercise"
	if got := err.Error(); got != expected {
		t.Errorf("NewUnauthorized().Error() = %v, want %v", got, expected)
	}
}

func TestNewNotFound(t *testing.T) {
	resource := "workout"
	id := "456"
	err := NewNotFound(resource, id)

	if err.Resource != resource {
		t.Errorf("NewNotFound() Resource = %v, want %v", err.Resource, resource)
	}
	if err.ID != id {
		t.Errorf("NewNotFound() ID = %v, want %v", err.ID, id)
	}

	expected := "workout with id 456 not found"
	if got := err.Error(); got != expected {
		t.Errorf("NewNotFound().Error() = %v, want %v", got, expected)
	}
}

func TestUnauthorized_ErrorsAs(t *testing.T) {
	originalErr := &Unauthorized{
		Resource: "exercise",
		UserID:   "user123",
	}

	// Test with direct error
	var unauthorizedErr *Unauthorized
	if !errors.As(originalErr, &unauthorizedErr) {
		t.Error("errors.As() failed to match Unauthorized error")
	}
	if unauthorizedErr.Resource != "exercise" {
		t.Errorf("errors.As() Resource = %v, want %v", unauthorizedErr.Resource, "exercise")
	}

	// Test with wrapped error
	wrappedErr := fmt.Errorf("failed to get exercise: %w", originalErr)
	var wrappedUnauthorizedErr *Unauthorized
	if !errors.As(wrappedErr, &wrappedUnauthorizedErr) {
		t.Error("errors.As() failed to match wrapped Unauthorized error")
	}
	if wrappedUnauthorizedErr.Resource != "exercise" {
		t.Errorf("errors.As() on wrapped error, Resource = %v, want %v", wrappedUnauthorizedErr.Resource, "exercise")
	}
}

func TestNotFound_ErrorsAs(t *testing.T) {
	originalErr := &NotFound{
		Resource: "workout",
		ID:       "789",
	}

	// Test with direct error
	var notFoundErr *NotFound
	if !errors.As(originalErr, &notFoundErr) {
		t.Error("errors.As() failed to match NotFound error")
	}
	if notFoundErr.ID != "789" {
		t.Errorf("errors.As() ID = %v, want %v", notFoundErr.ID, "789")
	}

	// Test with wrapped error
	wrappedErr := fmt.Errorf("failed to get workout: %w", originalErr)
	var wrappedNotFoundErr *NotFound
	if !errors.As(wrappedErr, &wrappedNotFoundErr) {
		t.Error("errors.As() failed to match wrapped NotFound error")
	}
	if wrappedNotFoundErr.ID != "789" {
		t.Errorf("errors.As() on wrapped error, ID = %v, want %v", wrappedNotFoundErr.ID, "789")
	}
}
