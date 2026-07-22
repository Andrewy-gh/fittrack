package aichat

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const deleteAllConversationsLockTimeout = 5 * time.Second

func (r *repository) DeleteAllConversations(ctx context.Context, userID string, stoppedAt time.Time) (*DeleteAllConversationsResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin ai chat history delete transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, fmt.Sprintf("SET LOCAL lock_timeout = '%dms'", deleteAllConversationsLockTimeout.Milliseconds())); err != nil {
		return nil, fmt.Errorf("bound ai chat history delete lock wait: %w", err)
	}

	qtx := r.queries.WithTx(tx)
	conversationIDs, err := qtx.LockAIChatConversationsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lock ai chat conversations for history delete: %w", err)
	}
	runs, err := qtx.LockAIChatRunsByUser(ctx, userID)
	if err != nil {
		return nil, mapDeleteAllLockError("lock ai chat runs for history delete", err)
	}

	stoppedRunIDs := make([]int32, 0)
	completedAt := pgtype.Timestamptz{Time: stoppedAt.UTC(), Valid: true}
	for _, run := range runs {
		if run.Status != statusStreaming {
			continue
		}
		if _, err := qtx.UpdateAIChatMessageStopped(ctx, db.UpdateAIChatMessageStoppedParams{
			ID: run.AssistantMessageID, UserID: userID, CompletedAt: completedAt,
		}); err != nil {
			return nil, fmt.Errorf("stop ai chat message for history delete: %w", err)
		}
		if _, err := qtx.UpdateAIChatRunStopped(ctx, db.UpdateAIChatRunStoppedParams{
			ID: run.ID, UserID: userID, CompletedAt: completedAt,
		}); err != nil {
			return nil, fmt.Errorf("stop ai chat run for history delete: %w", err)
		}
		stoppedRunIDs = append(stoppedRunIDs, run.ID)
	}

	if err := qtx.ClearUserTrainingProfileSourcesByUser(ctx, userID); err != nil {
		return nil, fmt.Errorf("clear training profile provenance for history delete: %w", err)
	}
	deleted, err := qtx.DeleteAIChatConversationsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("delete ai chat history: %w", err)
	}
	if deleted != int64(len(conversationIDs)) {
		return nil, fmt.Errorf("delete ai chat history: locked %d conversations but deleted %d", len(conversationIDs), deleted)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ai chat history delete transaction: %w", err)
	}
	return &DeleteAllConversationsResult{
		ConversationsDeleted: deleted,
		StoppedRunIDs:        stoppedRunIDs,
	}, nil
}

func mapDeleteAllLockError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "55P03" {
		return ErrConversationBusy
	}
	return fmt.Errorf("%s: %w", operation, err)
}
