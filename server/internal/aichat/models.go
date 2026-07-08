package aichat

import (
	"errors"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
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
	ErrFeatureDisabled               = errors.New("ai chat feature is disabled")
	ErrRuntimeUnavailable            = errors.New("ai chat runtime is unavailable")
	ErrRecoveryUnavailable           = errors.New("ai chat recovery is unavailable")
	ErrConversationBusy              = errors.New("ai chat conversation already has an active run")
	ErrGenerationTimeout             = errors.New("ai chat generation timed out")
	ErrStreamDisconnected            = errors.New("ai chat stream disconnected before completion")
	ErrStreamNotStarted              = errors.New("ai chat stream failed before delivery started")
	ErrStreamAwaitingRecovery        = errors.New("ai chat stream interrupted and awaiting recovery")
	ErrLatestWorkoutDraftUnavailable = errors.New("latest ai workout draft is unavailable")
	ErrLatestWorkoutDraftSuperseded  = errors.New("latest ai workout draft changed before it could be marked saved")
)

const (
	streamInterruptedFailureMessage     = "ai chat stream was interrupted before completion"
	interruptionReasonStalePartial      = "stale_partial"
	interruptionReasonAttemptsExhausted = "generation_attempts_exhausted"
)

type ValidateRequest struct {
	Prompt string `json:"prompt"`
}

type SendMessageRequest struct {
	Prompt string `json:"prompt"`
}

type RecoverMessageResponse struct {
	ConversationID int32  `json:"conversation_id"`
	RunID          int32  `json:"run_id,omitempty"`
	Status         string `json:"status"`
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
	ID                       int32                         `json:"id"`
	UserID                   string                        `json:"-"`
	Title                    *string                       `json:"title,omitempty"`
	LatestWorkoutDraft       *workout.CreateWorkoutRequest `json:"latest_workout_draft,omitempty"`
	LatestWorkoutDraftStatus *LatestWorkoutDraftStatus     `json:"latest_workout_draft_status,omitempty"`
	CreatedAt                time.Time                     `json:"created_at"`
	UpdatedAt                time.Time                     `json:"updated_at"`
	LastMessageAt            *time.Time                    `json:"last_message_at,omitempty"`
}

type ConversationSummary struct {
	ID            int32      `json:"id"`
	Title         *string    `json:"title,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
}

type LatestWorkoutDraftStatus struct {
	SourceRunID    *int32     `json:"source_run_id,omitempty"`
	IsSaved        bool       `json:"is_saved"`
	SavedWorkoutID *int32     `json:"saved_workout_id,omitempty"`
	SavedAt        *time.Time `json:"saved_at,omitempty"`
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
	ID                 int32                         `json:"id"`
	ConversationID     int32                         `json:"conversation_id"`
	UserID             string                        `json:"-"`
	UserMessageID      int32                         `json:"user_message_id"`
	AssistantMessageID int32                         `json:"assistant_message_id"`
	Model              string                        `json:"model"`
	Status             string                        `json:"status"`
	RequestID          *string                       `json:"request_id,omitempty"`
	ErrorMessage       *string                       `json:"error_message,omitempty"`
	WorkoutDraft       *workout.CreateWorkoutRequest `json:"workout_draft,omitempty"`
	GenerationStatus   string                        `json:"generation_status"`
	GenerationOwner    *string                       `json:"generation_owner,omitempty"`
	LeaseExpiresAt     *time.Time                    `json:"generation_lease_expires_at,omitempty"`
	HeartbeatAt        *time.Time                    `json:"generation_heartbeat_at,omitempty"`
	GenerationAttempt  int32                         `json:"generation_attempt"`
	InterruptedAt      *time.Time                    `json:"interrupted_at,omitempty"`
	InterruptionReason *string                       `json:"interruption_reason,omitempty"`
	CreatedAt          time.Time                     `json:"created_at"`
	UpdatedAt          time.Time                     `json:"updated_at"`
	StartedAt          time.Time                     `json:"started_at"`
	CompletedAt        *time.Time                    `json:"completed_at,omitempty"`
}

type ConversationDetail struct {
	Conversation *Conversation        `json:"conversation"`
	Messages     []ChatMessage        `json:"messages"`
	ActiveRun    *ConversationRunView `json:"active_run,omitempty"`
}

type SaveLatestWorkoutDraftResponse struct {
	Conversation *Conversation `json:"conversation"`
	WorkoutID    int32         `json:"workout_id"`
}

type PreparedMessageStream struct {
	Conversation     *Conversation
	History          []ChatMessage
	UserMessage      *ChatMessage
	AssistantMessage *ChatMessage
	Run              *ChatRun
	Prompt           string
	LastSequence     int32
}

type PreparedResumeStream struct {
	Conversation     *Conversation
	AssistantMessage *ChatMessage
	Run              *ChatRun
	AfterSequence    int32
	LastSequence     int32
}

type RunRecoveryRequest struct {
	ConversationID int32
	RunID          int32
	UserID         string
	Reason         string
}

type RuntimeChatMessage struct {
	Role string
	Text string
}

type WorkoutHistoryFilter struct {
	LastN        int
	StartDate    *time.Time
	EndDate      *time.Time
	ExerciseName string
	WorkoutFocus string
}

type ChatWorkoutView struct {
	Date      string             `json:"date"`
	Focus     string             `json:"focus,omitempty"`
	Notes     string             `json:"notes,omitempty"`
	Exercises []ChatExerciseView `json:"exercises,omitempty"`
}

type ChatExerciseView struct {
	Name string   `json:"name"`
	Sets []string `json:"sets,omitempty"`
}

type ExerciseStatsView struct {
	ExerciseName    string                    `json:"exercise_name"`
	Window          string                    `json:"window"`
	BestE1RM        *ExerciseBestE1RMView     `json:"best_e1rm,omitempty"`
	Trend           []ExerciseStatsTrendPoint `json:"trend,omitempty"`
	LastSessionSets []string                  `json:"last_session_sets,omitempty"`
	LastSessionDate string                    `json:"last_session_date,omitempty"`
	SessionCount    int                       `json:"session_count"`
	Message         string                    `json:"message,omitempty"`
}

type ExerciseBestE1RMView struct {
	Weight float64 `json:"weight"`
	Date   string  `json:"date,omitempty"`
}

type ExerciseStatsTrendPoint struct {
	Date      string  `json:"date"`
	BestE1RM  float64 `json:"best_e1rm,omitempty"`
	AvgE1RM   float64 `json:"avg_e1rm,omitempty"`
	Volume    float64 `json:"volume,omitempty"`
	WorkoutID int32   `json:"workout_id,omitempty"`
}

type TrainingSnapshot struct {
	LastWorkoutDate string   `json:"last_workout_date,omitempty"`
	WorkoutsLast30D int64    `json:"workouts_last_30d"`
	TopExercises    []string `json:"top_exercises,omitempty"`
}

type StreamDone struct {
	ConversationID int32                         `json:"conversation_id,omitempty"`
	RunID          int32                         `json:"run_id,omitempty"`
	MessageID      int32                         `json:"message_id,omitempty"`
	Model          string                        `json:"model"`
	Text           string                        `json:"text"`
	Sequence       int32                         `json:"sequence,omitempty"`
	WorkoutDraft   *workout.CreateWorkoutRequest `json:"workout_draft,omitempty"`
	ToolCalls      []string                      `json:"tool_calls,omitempty"`
}

type StreamChunk struct {
	Delta    string
	Sequence int32
}

type ConversationRunView struct {
	ID                 int32  `json:"id"`
	AssistantMessageID int32  `json:"assistant_message_id"`
	Status             string `json:"status"`
	LatestSequence     int32  `json:"latest_sequence"`
}

type StreamEvent struct {
	Type           string                        `json:"type"`
	RequestID      string                        `json:"request_id,omitempty"`
	ConversationID int32                         `json:"conversation_id,omitempty"`
	RunID          int32                         `json:"run_id,omitempty"`
	MessageID      int32                         `json:"message_id,omitempty"`
	Model          string                        `json:"model,omitempty"`
	Delta          string                        `json:"delta,omitempty"`
	Text           string                        `json:"text,omitempty"`
	Message        string                        `json:"message,omitempty"`
	Sequence       int32                         `json:"sequence,omitempty"`
	WorkoutDraft   *workout.CreateWorkoutRequest `json:"workout_draft,omitempty"`
}
