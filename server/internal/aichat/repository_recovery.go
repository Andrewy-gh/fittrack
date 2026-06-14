package aichat

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
)

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

	lastSequence, err := r.GetLatestStreamSequence(ctx, run.ID, userID)
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
		LastSequence:     lastSequence,
	}, nil
}

func (r *repository) LoadPreparedRunForResume(ctx context.Context, conversationID int32, runID int32, userID string, afterSequence int32) (*PreparedResumeStream, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	conversation, err := r.GetConversation(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}

	runRow, err := r.queries.GetAIChatRun(ctx, db.GetAIChatRunParams{
		ID:     runID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("get ai chat run for resume: %w", err)
	}

	run, err := mapRun(runRow)
	if err != nil {
		return nil, err
	}
	if run.ConversationID != conversationID {
		return nil, pgx.ErrNoRows
	}

	assistantMessageRow, err := r.queries.GetAIChatMessage(ctx, db.GetAIChatMessageParams{
		ID:     run.AssistantMessageID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get ai chat assistant message for resume: %w", err)
	}

	assistantMessage, err := mapMessage(assistantMessageRow)
	if err != nil {
		return nil, err
	}

	lastSequence, err := r.GetLatestStreamSequence(ctx, run.ID, userID)
	if err != nil {
		return nil, err
	}

	return &PreparedResumeStream{
		Conversation:     conversation,
		AssistantMessage: assistantMessage,
		Run:              run,
		AfterSequence:    afterSequence,
		LastSequence:     lastSequence,
	}, nil
}

func (r *repository) ListStreamChunksAfter(ctx context.Context, runID int32, userID string, afterSequence int32) ([]StreamChunk, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.queries.ListAIChatStreamChunksAfter(ctx, db.ListAIChatStreamChunksAfterParams{
		RunID:    runID,
		UserID:   userID,
		Sequence: afterSequence,
	})
	if err != nil {
		return nil, fmt.Errorf("list ai chat stream chunks after sequence: %w", err)
	}

	chunks := make([]StreamChunk, 0, len(rows))
	for _, row := range rows {
		chunks = append(chunks, StreamChunk{
			Delta:    row.DeltaText,
			Sequence: row.Sequence,
		})
	}

	return chunks, nil
}
