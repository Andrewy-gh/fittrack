package aichat

import "errors"

// These are application byte limits, not model token or context budgets.
const (
	maxChatBodyBytes   = 64 * 1024
	maxChatPromptBytes = 8 * 1024
)

var (
	errPromptRequired = errors.New("prompt is required")
	errPromptTooLarge = errors.New("prompt exceeds 8 KiB")
)

// prompt must already be trimmed; len counts its UTF-8 bytes.
func validateChatPrompt(prompt string) error {
	if prompt == "" {
		return errPromptRequired
	}
	if len(prompt) > maxChatPromptBytes {
		return errPromptTooLarge
	}
	return nil
}
