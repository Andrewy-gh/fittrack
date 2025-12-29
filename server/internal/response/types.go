package response

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message   string `json:"message" example:"Error message"`
	RequestID string `json:"request_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool `json:"success" example:"true"`
}
