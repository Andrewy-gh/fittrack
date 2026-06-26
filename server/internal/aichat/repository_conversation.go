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

func (r *repository) ListConversations(ctx context.Context, userID string, limit int32) ([]ConversationSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := r.queries.ListAIChatConversationsByUser(ctx, db.ListAIChatConversationsByUserParams{
		UserID: userID,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list ai chat conversations: %w", err)
	}

	conversations := make([]ConversationSummary, 0, len(rows))
	for _, row := range rows {
		conversation, err := mapConversationSummary(row)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, *conversation)
	}

	return conversations, nil
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

func (r *repository) SaveLatestWorkoutDraft(ctx context.Context, request SaveLatestWorkoutDraftRequest) (*SavedLatestWorkoutDraft, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if request.SaveWorkout == nil {
		return nil, errors.New("save latest ai chat workout draft requires a workout saver")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin ai chat latest workout draft save transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	row, err := qtx.GetAIChatConversationForUpdate(ctx, db.GetAIChatConversationForUpdateParams{
		ID:     request.ConversationID,
		UserID: request.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("lock ai chat conversation for latest workout draft save: %w", err)
	}

	conversation, err := mapConversation(row)
	if err != nil {
		return nil, err
	}
	if conversation.LatestWorkoutDraft == nil {
		return nil, ErrLatestWorkoutDraftUnavailable
	}

	if status := conversation.LatestWorkoutDraftStatus; status != nil && status.IsSaved && status.SavedWorkoutID != nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit ai chat latest workout draft save transaction: %w", err)
		}
		return &SavedLatestWorkoutDraft{
			Conversation: conversation,
			WorkoutID:    *status.SavedWorkoutID,
		}, nil
	}

	workoutID, err := request.SaveWorkout(ctx, qtx, *conversation.LatestWorkoutDraft, request.UserID)
	if err != nil {
		return nil, fmt.Errorf("save ai chat latest workout draft workout: %w", err)
	}

	var sourceRunID *int32
	if conversation.LatestWorkoutDraftStatus != nil {
		sourceRunID = conversation.LatestWorkoutDraftStatus.SourceRunID
	}

	updatedRow, err := qtx.MarkAIChatConversationLatestWorkoutDraftSaved(ctx, db.MarkAIChatConversationLatestWorkoutDraftSavedParams{
		ID:                            request.ConversationID,
		UserID:                        request.UserID,
		LatestWorkoutDraftSourceRunID: int4PtrToPg(sourceRunID),
		LatestWorkoutDraftSavedWorkoutID: pgtype.Int4{
			Int32: workoutID,
			Valid: true,
		},
		LatestWorkoutDraftSavedAt: pgtype.Timestamptz{
			Time:  request.SavedAt.UTC(),
			Valid: true,
		},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLatestWorkoutDraftSuperseded
		}
		return nil, fmt.Errorf("mark latest ai chat workout draft saved: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ai chat latest workout draft save transaction: %w", err)
	}

	updatedConversation, err := mapConversation(updatedRow)
	if err != nil {
		return nil, err
	}

	return &SavedLatestWorkoutDraft{
		Conversation: updatedConversation,
		WorkoutID:    workoutID,
	}, nil
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
