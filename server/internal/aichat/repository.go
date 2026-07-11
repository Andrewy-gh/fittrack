package aichat

import (
	"context"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SaveWorkoutDraftTx func(ctx context.Context, qtx *db.Queries, draft workout.CreateWorkoutRequest, userID string) (int32, error)

type SaveLatestWorkoutDraftRequest struct {
	ConversationID int32
	UserID         string
	SavedAt        time.Time
	SaveWorkout    SaveWorkoutDraftTx
}

type SavedLatestWorkoutDraft struct {
	Conversation *Conversation
	WorkoutID    int32
}

type Repository interface {
	ChatDataReader
	CreateConversation(ctx context.Context, userID string) (*Conversation, error)
	ListConversations(ctx context.Context, userID string, limit int32) ([]ConversationSummary, error)
	GetConversation(ctx context.Context, conversationID int32, userID string) (*Conversation, error)
	DeleteConversation(ctx context.Context, conversationID int32, userID string) error
	SaveLatestWorkoutDraft(ctx context.Context, request SaveLatestWorkoutDraftRequest) (*SavedLatestWorkoutDraft, error)
	ListMessages(ctx context.Context, conversationID int32, userID string) ([]ChatMessage, error)
	GetActiveRunForConversation(ctx context.Context, conversationID int32, userID string) (*ChatRun, error)
	GetLatestStreamSequence(ctx context.Context, runID int32, userID string) (int32, error)
	LoadPreparedRunForRecovery(ctx context.Context, runID int32, userID string) (*PreparedMessageStream, error)
	LoadPreparedRunForResume(ctx context.Context, conversationID int32, runID int32, userID string, afterSequence int32) (*PreparedResumeStream, error)
	ListStreamChunksAfter(ctx context.Context, runID int32, userID string, afterSequence int32) ([]StreamChunk, error)
	PrepareMessageStream(ctx context.Context, conversationID int32, userID string, prompt string, model string, requestID string) (*PreparedMessageStream, error)
	ClaimRunGeneration(ctx context.Context, run *ChatRun, owner runOwner, now time.Time) error
	HeartbeatRunGeneration(ctx context.Context, run *ChatRun, owner runOwner, now time.Time) (bool, error)
	AppendStreamChunk(ctx context.Context, prepared *PreparedMessageStream, delta string, partialText string, updatedAt time.Time) (int32, error)
	InterruptRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, reason string, completedAt time.Time) error
	CompleteRun(ctx context.Context, prepared *PreparedMessageStream, assistantText string, workoutDraft *workout.CreateWorkoutRequest, completedAt time.Time) (*ChatMessage, *ChatRun, error)
	FailRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, failure error, completedAt time.Time) error
}

type repository struct {
	logger         *slog.Logger
	queries        *db.Queries
	pool           *pgxpool.Pool
	trialPromptCap int32
}

const streamingRunStaleAfter = chatStreamTimeout + 15*time.Second

func NewRepository(logger *slog.Logger, queries *db.Queries, pool *pgxpool.Pool, trialPromptCap ...int) Repository {
	var cap int32
	if len(trialPromptCap) > 0 {
		cap = int32(trialPromptCap[0])
	}
	return &repository{
		logger:         logger,
		queries:        queries,
		pool:           pool,
		trialPromptCap: cap,
	}
}

var _ Repository = (*repository)(nil)
