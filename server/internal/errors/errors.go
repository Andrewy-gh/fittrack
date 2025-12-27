// Package errors provides shared error types for the FitTrack application.
// These error types are used across different service layers to maintain
// consistency in error handling and reporting.
package errors

import "fmt"

// Unauthorized represents an authorization error when a user attempts
// to access or modify a resource they don't own.
type Unauthorized struct {
	Resource string // The type of resource (e.g., "exercise", "workout")
	UserID   string // The ID of the user attempting the action
}

// Error returns the error message for an Unauthorized error.
func (e *Unauthorized) Error() string {
	if e.Resource != "" && e.UserID != "" {
		return fmt.Sprintf("user %s is not authorized to access %s", e.UserID, e.Resource)
	}
	if e.Resource != "" {
		return fmt.Sprintf("not authorized to access %s", e.Resource)
	}
	return "unauthorized access"
}

// NotFound represents an error when a requested resource cannot be found.
type NotFound struct {
	Resource string // The type of resource (e.g., "exercise", "workout")
	ID       string // The ID of the resource that was not found
}

// Error returns the error message for a NotFound error.
func (e *NotFound) Error() string {
	if e.Resource != "" && e.ID != "" {
		return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
	}
	if e.Resource != "" {
		return fmt.Sprintf("%s not found", e.Resource)
	}
	return "resource not found"
}

// NewUnauthorized creates a new Unauthorized error.
func NewUnauthorized(resource, userID string) *Unauthorized {
	return &Unauthorized{
		Resource: resource,
		UserID:   userID,
	}
}

// NewNotFound creates a new NotFound error.
func NewNotFound(resource, id string) *NotFound {
	return &NotFound{
		Resource: resource,
		ID:       id,
	}
}
