package aichat

import (
	"errors"
	"time"
)

const (
	featureKeyAIChatbot = "ai_chatbot"
	defaultModelName    = "googleai/gemini-2.5-flash"

	roleUser      = "user"
	roleAssistant = "assistant"

	statusStreaming = "streaming"
	statusCompleted = "completed"
	statusFailed    = "failed"
)

var (
	ErrFeatureDisabled    = errors.New("ai chat feature is disabled")
	ErrRuntimeUnavailable = errors.New("ai chat runtime is unavailable")
	ErrConversationBusy   = errors.New("ai chat conversation already has an active run")
	ErrGenerationTimeout  = errors.New("ai chat generation timed out")
)

type ValidateRequest struct {
	Prompt string `json:"prompt"`
}

type SendMessageRequest struct {
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

type Conversation struct {
	ID            int32      `json:"id"`
	UserID        string     `json:"-"`
	Title         *string    `json:"title,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
}

type ChatMessage struct {
	ID             int32      `json:"id"`
	ConversationID int32      `json:"conversation_id"`
	UserID         string     `json:"-"`
	Role           string     `json:"role"`
	Content        string     `json:"content"`
	Status         string     `json:"status"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type ChatRun struct {
	ID                 int32      `json:"id"`
	ConversationID     int32      `json:"conversation_id"`
	UserID             string     `json:"-"`
	UserMessageID      int32      `json:"user_message_id"`
	AssistantMessageID int32      `json:"assistant_message_id"`
	Model              string     `json:"model"`
	Status             string     `json:"status"`
	RequestID          *string    `json:"request_id,omitempty"`
	ErrorMessage       *string    `json:"error_message,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

type ConversationDetail struct {
	Conversation *Conversation `json:"conversation"`
	Messages     []ChatMessage `json:"messages"`
}

type PreparedMessageStream struct {
	Conversation     *Conversation
	History          []ChatMessage
	UserMessage      *ChatMessage
	AssistantMessage *ChatMessage
	Run              *ChatRun
	Prompt           string
}

type RuntimeChatMessage struct {
	Role string
	Text string
}

type StreamDone struct {
	ConversationID int32  `json:"conversation_id,omitempty"`
	RunID          int32  `json:"run_id,omitempty"`
	MessageID      int32  `json:"message_id,omitempty"`
	Model          string `json:"model"`
	Text           string `json:"text"`
}

type StreamEvent struct {
	Type           string `json:"type"`
	RequestID      string `json:"request_id,omitempty"`
	ConversationID int32  `json:"conversation_id,omitempty"`
	RunID          int32  `json:"run_id,omitempty"`
	MessageID      int32  `json:"message_id,omitempty"`
	Model          string `json:"model,omitempty"`
	Delta          string `json:"delta,omitempty"`
	Text           string `json:"text,omitempty"`
	Message        string `json:"message,omitempty"`
}
