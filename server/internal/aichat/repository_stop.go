package aichat

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *repository) StopRun(ctx context.Context, conversationID, runID int32, userID string, stoppedAt time.Time) (*StopRunResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin ai chat stop transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := r.queries.WithTx(tx)
	runRow, err := qtx.GetAIChatRunForUpdate(ctx, db.GetAIChatRunForUpdateParams{ID: runID, ConversationID: conversationID, UserID: userID})
	if err != nil {
		return nil, err
	}
	run, err := mapRun(runRow)
	if err != nil {
		return nil, err
	}
	if run.Status == statusStreaming {
		ts := pgtype.Timestamptz{Time: stoppedAt.UTC(), Valid: true}
		if _, err = qtx.UpdateAIChatMessageStopped(ctx, db.UpdateAIChatMessageStoppedParams{ID: run.AssistantMessageID, UserID: userID, CompletedAt: ts}); err != nil {
			return nil, fmt.Errorf("stop ai chat message: %w", err)
		}
		runRow, err = qtx.UpdateAIChatRunStopped(ctx, db.UpdateAIChatRunStoppedParams{ID: runID, UserID: userID, CompletedAt: ts})
		if err != nil {
			return nil, fmt.Errorf("stop ai chat run: %w", err)
		}
		run, err = mapRun(runRow)
		if err != nil {
			return nil, err
		}
	}
	messageRow, err := qtx.GetAIChatMessage(ctx, db.GetAIChatMessageParams{ID: run.AssistantMessageID, UserID: userID})
	if err != nil {
		return nil, err
	}
	message, err := mapMessage(messageRow)
	if err != nil {
		return nil, err
	}
	sequence, err := qtx.GetLatestAIChatStreamChunkSequence(ctx, db.GetLatestAIChatStreamChunkSequenceParams{RunID: runID, UserID: userID})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ai chat stop transaction: %w", err)
	}
	return &StopRunResponse{ConversationID: conversationID, RunID: runID, MessageID: message.ID, Status: run.Status, Text: message.Content, Sequence: sequence}, nil
}
