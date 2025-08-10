package response

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message" example:"Error message"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool `json:"success" example:"true"`
}
