package aichat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateConversation(ctx context.Context, userID string) (*Conversation, error)
	GetConversation(ctx context.Context, conversationID int32, userID string) (*Conversation, error)
	ListMessages(ctx context.Context, conversationID int32, userID string) ([]ChatMessage, error)
	GetActiveRunForConversation(ctx context.Context, conversationID int32, userID string) (*ChatRun, error)
	LoadPreparedRunForRecovery(ctx context.Context, runID int32, userID string) (*PreparedMessageStream, error)
	PrepareMessageStream(ctx context.Context, conversationID int32, userID string, prompt string, model string, requestID string) (*PreparedMessageStream, error)
	UpdateStreamingRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, updatedAt time.Time) error
	MarkRunAwaitingRecovery(ctx context.Context, prepared *PreparedMessageStream, partialText string, updatedAt time.Time) error
	ClaimRunRecovery(ctx context.Context, runID int32, userID string) error
	CompleteRun(ctx context.Context, prepared *PreparedMessageStream, assistantText string, completedAt time.Time) (*ChatMessage, *ChatRun, error)
	FailRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, failure error, completedAt time.Time) error
}

type repository struct {
	logger  *slog.Logger
	queries *db.Queries
	pool    *pgxpool.Pool
}

const streamingRunStaleAfter = chatStreamTimeout + 15*time.Second

func NewRepository(logger *slog.Logger, queries *db.Queries, pool *pgxpool.Pool) Repository {
	return &repository{
		logger:  logger,
		queries: queries,
		pool:    pool,
	}
}

func (r *repository) CreateConversation(ctx context.Context, userID string) (*Conversation, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	row, err := r.queries.CreateAIChatConversation(ctx, db.CreateAIChatConversationParams{
		UserID: userID,
		Title:  pgtype.Text{},
	})
	if err != nil {
		return nil, fmt.Errorf("create ai chat conversation: %w", err)
	}

	conversation, err := mapConversation(row)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (r *repository) GetConversation(ctx context.Context, conversationID int32, userID string) (*Conversation, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	row, err := r.queries.GetAIChatConversation(ctx, db.GetAIChatConversationParams{
		ID:     conversationID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("get ai chat conversation: %w", err)
	}

	conversation, err := mapConversation(row)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (r *repository) ListMessages(ctx context.Context, conversationID int32, userID string) ([]ChatMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := r.queries.ListAIChatMessagesByConversation(ctx, db.ListAIChatMessagesByConversationParams{
		ConversationID: conversationID,
		UserID:         userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list ai chat messages: %w", err)
	}

	messages := make([]ChatMessage, 0, len(rows))
	for _, row := range rows {
		message, err := mapMessage(row)
		if err != nil {
			return nil, err
		}
		messages = append(messages, *message)
	}

	return messages, nil
}

func (r *repository) GetActiveRunForConversation(ctx context.Context, conversationID int32, userID string) (*ChatRun, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	row, err := r.queries.GetActiveAIChatRunForConversation(ctx, db.GetActiveAIChatRunForConversationParams{
		ConversationID: conversationID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("get active ai chat run: %w", err)
	}

	run, err := mapRun(row)
	if err != nil {
		return nil, err
	}

	return run, nil
}

func (r *repository) LoadPreparedRunForRecovery(ctx context.Context, runID int32, userID string) (*PreparedMessageStream, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	runRow, err := r.queries.GetAIChatRun(ctx, db.GetAIChatRunParams{
		ID:     runID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("get ai chat run for recovery: %w", err)
	}

	run, err := mapRun(runRow)
	if err != nil {
		return nil, err
	}

	conversation, err := r.GetConversation(ctx, run.ConversationID, userID)
	if err != nil {
		return nil, err
	}

	userMessageRow, err := r.queries.GetAIChatMessage(ctx, db.GetAIChatMessageParams{
		ID:     run.UserMessageID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get ai chat user message for recovery: %w", err)
	}

	assistantMessageRow, err := r.queries.GetAIChatMessage(ctx, db.GetAIChatMessageParams{
		ID:     run.AssistantMessageID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get ai chat assistant message for recovery: %w", err)
	}

	historyRows, err := r.queries.ListAIChatMessagesByConversation(ctx, db.ListAIChatMessagesByConversationParams{
		ConversationID: run.ConversationID,
		UserID:         userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list ai chat history for recovery: %w", err)
	}

	history := make([]ChatMessage, 0, len(historyRows))
	for _, row := range historyRows {
		if row.ID >= run.UserMessageID {
			break
		}
		message, err := mapMessage(row)
		if err != nil {
			return nil, err
		}
		history = append(history, *message)
	}

	userMessage, err := mapMessage(userMessageRow)
	if err != nil {
		return nil, err
	}

	assistantMessage, err := mapMessage(assistantMessageRow)
	if err != nil {
		return nil, err
	}

	return &PreparedMessageStream{
		Conversation:     conversation,
		History:          history,
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		Run:              run,
		Prompt:           userMessage.Content,
	}, nil
}

func (r *repository) PrepareMessageStream(ctx context.Context, conversationID int32, userID string, prompt string, model string, requestID string) (*PreparedMessageStream, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin ai chat stream transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	conversationRow, err := qtx.GetAIChatConversation(ctx, db.GetAIChatConversationParams{
		ID:     conversationID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("get ai chat conversation for stream: %w", err)
	}

	conversation, err := mapConversation(conversationRow)
	if err != nil {
		return nil, err
	}

	activeRunRow, err := qtx.GetActiveAIChatRunForConversation(ctx, db.GetActiveAIChatRunForConversationParams{
		ConversationID: conversationID,
		UserID:         userID,
	})
	switch {
	case err == nil:
		activeRun, err := mapRun(activeRunRow)
		if err != nil {
			return nil, err
		}
		if !isStreamingRunStale(activeRun.UpdatedAt, time.Now().UTC()) {
			return nil, ErrConversationBusy
		}
		if err := r.failStaleRun(ctx, qtx, conversation, activeRun); err != nil {
			return nil, err
		}
	case errors.Is(err, pgx.ErrNoRows):
	default:
		return nil, fmt.Errorf("get active ai chat run: %w", err)
	}

	historyRows, err := qtx.ListAIChatMessagesByConversation(ctx, db.ListAIChatMessagesByConversationParams{
		ConversationID: conversationID,
		UserID:         userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list ai chat history: %w", err)
	}

	completedAt := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	userMessageRow, err := qtx.CreateAIChatMessage(ctx, db.CreateAIChatMessageParams{
		ConversationID: conversationID,
		UserID:         userID,
		Role:           roleUser,
		Content:        prompt,
		Status:         statusCompleted,
		ErrorMessage:   pgtype.Text{},
		CompletedAt:    completedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create ai chat user message: %w", err)
	}

	assistantMessageRow, err := qtx.CreateAIChatMessage(ctx, db.CreateAIChatMessageParams{
		ConversationID: conversationID,
		UserID:         userID,
		Role:           roleAssistant,
		Content:        "",
		Status:         statusStreaming,
		ErrorMessage:   pgtype.Text{},
		CompletedAt:    pgtype.Timestamptz{},
	})
	if err != nil {
		return nil, fmt.Errorf("create ai chat assistant message: %w", err)
	}

	runRow, err := qtx.CreateAIChatRun(ctx, db.CreateAIChatRunParams{
		ConversationID:     conversationID,
		UserID:             userID,
		UserMessageID:      userMessageRow.ID,
		AssistantMessageID: assistantMessageRow.ID,
		Model:              model,
		Status:             statusStreaming,
		RequestID:          textToPg(requestID),
		ErrorMessage:       pgtype.Text{},
		CompletedAt:        pgtype.Timestamptz{},
	})
	if err != nil {
		if db.IsUniqueConstraintError(err) {
			return nil, ErrConversationBusy
		}
		return nil, fmt.Errorf("create ai chat run: %w", err)
	}

	now := time.Now().UTC()
	if err := qtx.TouchAIChatConversation(ctx, db.TouchAIChatConversationParams{
		ID:            conversationID,
		UserID:        userID,
		LastMessageAt: pgtype.Timestamptz{Time: now, Valid: true},
	}); err != nil {
		return nil, fmt.Errorf("touch ai chat conversation: %w", err)
	}

	title := buildConversationTitle(prompt)
	if title != "" {
		if _, err := qtx.SetAIChatConversationTitleIfEmpty(ctx, db.SetAIChatConversationTitleIfEmptyParams{
			ID:     conversationID,
			UserID: userID,
			Title:  textToPg(title),
		}); err != nil {
			return nil, fmt.Errorf("set ai chat conversation title: %w", err)
		}
		if conversation.Title == nil {
			conversation.Title = &title
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ai chat stream transaction: %w", err)
	}

	history := make([]ChatMessage, 0, len(historyRows))
	for _, row := range historyRows {
		message, err := mapMessage(row)
		if err != nil {
			return nil, err
		}
		history = append(history, *message)
	}

	userMessage, err := mapMessage(userMessageRow)
	if err != nil {
		return nil, err
	}

	assistantMessage, err := mapMessage(assistantMessageRow)
	if err != nil {
		return nil, err
	}

	run, err := mapRun(runRow)
	if err != nil {
		return nil, err
	}

	return &PreparedMessageStream{
		Conversation:     conversation,
		History:          history,
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		Run:              run,
		Prompt:           prompt,
	}, nil
}

func (r *repository) UpdateStreamingRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, updatedAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin ai chat streaming update transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	if _, err := qtx.UpdateAIChatMessageStreaming(ctx, db.UpdateAIChatMessageStreamingParams{
		ID:      prepared.AssistantMessage.ID,
		UserID:  prepared.AssistantMessage.UserID,
		Content: partialText,
	}); err != nil {
		return fmt.Errorf("update ai chat assistant streaming message: %w", err)
	}

	if err := qtx.TouchAIChatRun(ctx, db.TouchAIChatRunParams{
		ID:     prepared.Run.ID,
		UserID: prepared.Run.UserID,
	}); err != nil {
		return fmt.Errorf("touch ai chat streaming run: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit ai chat streaming update transaction: %w", err)
	}

	prepared.AssistantMessage.Content = partialText
	prepared.AssistantMessage.UpdatedAt = updatedAt.UTC()
	prepared.Run.UpdatedAt = updatedAt.UTC()

	return nil
}

func (r *repository) MarkRunAwaitingRecovery(ctx context.Context, prepared *PreparedMessageStream, partialText string, updatedAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin ai chat recovery handoff transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	partialText = strings.TrimSpace(partialText)
	if partialText != "" {
		if _, err := qtx.UpdateAIChatMessageStreaming(ctx, db.UpdateAIChatMessageStreamingParams{
			ID:      prepared.AssistantMessage.ID,
			UserID:  prepared.AssistantMessage.UserID,
			Content: partialText,
		}); err != nil {
			return fmt.Errorf("update ai chat assistant message for recovery handoff: %w", err)
		}
	}

	runRow, err := qtx.MarkAIChatRunAwaitingRecovery(ctx, db.MarkAIChatRunAwaitingRecoveryParams{
		ID:           prepared.Run.ID,
		UserID:       prepared.Run.UserID,
		ErrorMessage: textToPg(runAwaitingRecoveryMarker),
	})
	if err != nil {
		return fmt.Errorf("mark ai chat run awaiting recovery: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit ai chat recovery handoff transaction: %w", err)
	}

	if partialText != "" {
		prepared.AssistantMessage.Content = partialText
		prepared.AssistantMessage.UpdatedAt = updatedAt.UTC()
	}

	run, err := mapRun(runRow)
	if err != nil {
		return err
	}
	prepared.Run = run

	return nil
}

func (r *repository) ClaimRunRecovery(ctx context.Context, runID int32, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	runRow, err := r.queries.ClaimAIChatRunRecovery(ctx, db.ClaimAIChatRunRecoveryParams{
		ID:           runID,
		UserID:       userID,
		ErrorMessage: textToPg(runAwaitingRecoveryMarker),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		return fmt.Errorf("claim ai chat recovery run: %w", err)
	}

	_, err = mapRun(runRow)
	return err
}

func (r *repository) CompleteRun(ctx context.Context, prepared *PreparedMessageStream, assistantText string, completedAt time.Time) (*ChatMessage, *ChatRun, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin ai chat completion transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	completedTS := pgtype.Timestamptz{Time: completedAt.UTC(), Valid: true}

	messageRow, err := qtx.UpdateAIChatMessageCompleted(ctx, db.UpdateAIChatMessageCompletedParams{
		ID:          prepared.AssistantMessage.ID,
		UserID:      prepared.AssistantMessage.UserID,
		Content:     assistantText,
		CompletedAt: completedTS,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("complete ai chat assistant message: %w", err)
	}

	runRow, err := qtx.UpdateAIChatRunCompleted(ctx, db.UpdateAIChatRunCompletedParams{
		ID:          prepared.Run.ID,
		UserID:      prepared.Run.UserID,
		CompletedAt: completedTS,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("complete ai chat run: %w", err)
	}

	if err := qtx.TouchAIChatConversation(ctx, db.TouchAIChatConversationParams{
		ID:            prepared.Conversation.ID,
		UserID:        prepared.Conversation.UserID,
		LastMessageAt: completedTS,
	}); err != nil {
		return nil, nil, fmt.Errorf("touch ai chat conversation after completion: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit ai chat completion transaction: %w", err)
	}

	message, err := mapMessage(messageRow)
	if err != nil {
		return nil, nil, err
	}

	run, err := mapRun(runRow)
	if err != nil {
		return nil, nil, err
	}

	return message, run, nil
}

func (r *repository) FailRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, failure error, completedAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin ai chat failure transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	completedTS := pgtype.Timestamptz{Time: completedAt.UTC(), Valid: true}
	errorText := truncateForStorage(strings.TrimSpace(failure.Error()), 512)

	if _, err := qtx.UpdateAIChatMessageFailed(ctx, db.UpdateAIChatMessageFailedParams{
		ID:           prepared.AssistantMessage.ID,
		UserID:       prepared.AssistantMessage.UserID,
		Content:      partialText,
		ErrorMessage: textToPg(errorText),
		CompletedAt:  completedTS,
	}); err != nil {
		return fmt.Errorf("fail ai chat assistant message: %w", err)
	}

	if _, err := qtx.UpdateAIChatRunFailed(ctx, db.UpdateAIChatRunFailedParams{
		ID:           prepared.Run.ID,
		UserID:       prepared.Run.UserID,
		ErrorMessage: textToPg(errorText),
		CompletedAt:  completedTS,
	}); err != nil {
		return fmt.Errorf("fail ai chat run: %w", err)
	}

	if err := qtx.TouchAIChatConversation(ctx, db.TouchAIChatConversationParams{
		ID:            prepared.Conversation.ID,
		UserID:        prepared.Conversation.UserID,
		LastMessageAt: completedTS,
	}); err != nil {
		return fmt.Errorf("touch ai chat conversation after failure: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit ai chat failure transaction: %w", err)
	}

	return nil
}

func (r *repository) failStaleRun(ctx context.Context, qtx *db.Queries, conversation *Conversation, activeRun *ChatRun) error {
	assistantRow, err := qtx.GetAIChatMessage(ctx, db.GetAIChatMessageParams{
		ID:     activeRun.AssistantMessageID,
		UserID: activeRun.UserID,
	})
	if err != nil {
		return fmt.Errorf("get stale ai chat assistant message: %w", err)
	}

	completedTS := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	errorText := truncateForStorage(streamInterruptedFailureMessage, 512)
	if _, err := qtx.UpdateAIChatMessageFailed(ctx, db.UpdateAIChatMessageFailedParams{
		ID:           assistantRow.ID,
		UserID:       assistantRow.UserID,
		Content:      assistantRow.Content,
		ErrorMessage: textToPg(errorText),
		CompletedAt:  completedTS,
	}); err != nil {
		return fmt.Errorf("fail stale ai chat assistant message: %w", err)
	}

	if _, err := qtx.UpdateAIChatRunFailed(ctx, db.UpdateAIChatRunFailedParams{
		ID:           activeRun.ID,
		UserID:       activeRun.UserID,
		ErrorMessage: textToPg(errorText),
		CompletedAt:  completedTS,
	}); err != nil {
		return fmt.Errorf("fail stale ai chat run: %w", err)
	}

	if err := qtx.TouchAIChatConversation(ctx, db.TouchAIChatConversationParams{
		ID:            conversation.ID,
		UserID:        conversation.UserID,
		LastMessageAt: completedTS,
	}); err != nil {
		return fmt.Errorf("touch ai chat conversation after stale run recovery: %w", err)
	}

	r.logger.Warn("recovered stale ai chat run before starting a new stream",
		"conversation_id", conversation.ID,
		"run_id", activeRun.ID,
		"updated_at", activeRun.UpdatedAt,
	)

	return nil
}

func buildConversationTitle(prompt string) string {
	title := strings.Join(strings.Fields(prompt), " ")
	title = strings.TrimSpace(title)
	return truncateWithEllipsis(title, 80)
}

func truncateForStorage(value string, maxLen int) string {
	return truncateWithEllipsis(value, maxLen)
}

func truncateWithEllipsis(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if value == "" || maxLen <= 0 {
		return ""
	}

	if utf8.RuneCountInString(value) <= maxLen {
		return value
	}

	if maxLen <= 3 {
		return string([]rune(value)[:maxLen])
	}

	runes := []rune(value)
	return strings.TrimSpace(string(runes[:maxLen-3])) + "..."
}

func isStreamingRunStale(updatedAt time.Time, now time.Time) bool {
	return !updatedAt.IsZero() && now.UTC().Sub(updatedAt.UTC()) > streamingRunStaleAfter
}

func textToPg(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func mapConversation(row db.AiChatConversation) (*Conversation, error) {
	createdAt, err := timeFromPg(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := timeFromPg(row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	lastMessageAt, err := timePtrFromPg(row.LastMessageAt)
	if err != nil {
		return nil, err
	}

	return &Conversation{
		ID:            row.ID,
		UserID:        row.UserID,
		Title:         textPtr(row.Title),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		LastMessageAt: lastMessageAt,
	}, nil
}

func mapMessage(row db.AiChatMessage) (*ChatMessage, error) {
	createdAt, err := timeFromPg(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := timeFromPg(row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	completedAt, err := timePtrFromPg(row.CompletedAt)
	if err != nil {
		return nil, err
	}

	return &ChatMessage{
		ID:             row.ID,
		ConversationID: row.ConversationID,
		UserID:         row.UserID,
		Role:           row.Role,
		Content:        row.Content,
		Status:         row.Status,
		ErrorMessage:   textPtr(row.ErrorMessage),
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		CompletedAt:    completedAt,
	}, nil
}

func mapRun(row db.AiChatRun) (*ChatRun, error) {
	createdAt, err := timeFromPg(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := timeFromPg(row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	startedAt, err := timeFromPg(row.StartedAt)
	if err != nil {
		return nil, err
	}
	completedAt, err := timePtrFromPg(row.CompletedAt)
	if err != nil {
		return nil, err
	}

	return &ChatRun{
		ID:                 row.ID,
		ConversationID:     row.ConversationID,
		UserID:             row.UserID,
		UserMessageID:      row.UserMessageID,
		AssistantMessageID: row.AssistantMessageID,
		Model:              row.Model,
		Status:             row.Status,
		RequestID:          textPtr(row.RequestID),
		ErrorMessage:       textPtr(row.ErrorMessage),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		StartedAt:          startedAt,
		CompletedAt:        completedAt,
	}, nil
}

func timeFromPg(ts pgtype.Timestamptz) (time.Time, error) {
	if !ts.Valid {
		return time.Time{}, fmt.Errorf("invalid timestamptz value")
	}
	return ts.Time.UTC(), nil
}

func timePtrFromPg(ts pgtype.Timestamptz) (*time.Time, error) {
	if !ts.Valid {
		return nil, nil
	}
	utc := ts.Time.UTC()
	return &utc, nil
}

func textPtr(txt pgtype.Text) *string {
	if !txt.Valid {
		return nil
	}
	value := txt.String
	return &value
}

var _ Repository = (*repository)(nil)
