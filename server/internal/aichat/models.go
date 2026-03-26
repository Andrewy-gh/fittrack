package aichat

import "errors"

const (
	featureKeyAIChatbot = "ai_chatbot"
	defaultModelName    = "googleai/gemini-2.5-flash"
)

var (
	ErrFeatureDisabled    = errors.New("ai chat feature is disabled")
	ErrRuntimeUnavailable = errors.New("ai chat runtime is unavailable")
)

type ValidateRequest struct {
	Prompt string `json:"prompt"`
}

type ValidationOutput struct {
	Summary  string `json:"summary" jsonschema:"description=A brief production-readiness summary for this FitTrack AI chat validation run."`
	NextStep string `json:"next_step" jsonschema:"description=One concrete follow-up step for phase 1."`
}

type ValidateResponse struct {
	Model  string            `json:"model"`
	Output *ValidationOutput `json:"output"`
}

type StreamDone struct {
	Model string `json:"model"`
	Text  string `json:"text"`
}

type StreamEvent struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id,omitempty"`
	Model     string `json:"model,omitempty"`
	Delta     string `json:"delta,omitempty"`
	Text      string `json:"text,omitempty"`
	Message   string `json:"message,omitempty"`
}
