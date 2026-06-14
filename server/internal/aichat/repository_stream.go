package aichat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/billing"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

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

func (r *repository) GetLatestStreamSequence(ctx context.Context, runID int32, userID string) (int32, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sequence, err := r.queries.GetLatestAIChatStreamChunkSequence(ctx, db.GetLatestAIChatStreamChunkSequenceParams{
		RunID:  runID,
		UserID: userID,
	})
	if err != nil {
		return 0, fmt.Errorf("get latest ai chat stream sequence: %w", err)
	}

	return sequence, nil
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

	allowed, err := qtx.ConsumeAIChatTrialPromptForCurrentSubscription(ctx, db.ConsumeAIChatTrialPromptForCurrentSubscriptionParams{
		UserID:    userID,
		PromptCap: r.trialPromptCap,
	})
	if err != nil {
		return nil, fmt.Errorf("consume ai chat trial prompt: %w", err)
	}
	if !allowed.Valid || !allowed.Bool {
		return nil, billing.ErrTrialPromptLimitExceeded
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
		LastSequence:     0,
	}, nil
}

func (r *repository) ClaimRunGeneration(ctx context.Context, run *ChatRun, owner runOwner, now time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	runRow, err := r.queries.ClaimAIChatRunGeneration(ctx, db.ClaimAIChatRunGenerationParams{
		ID:     run.ID,
		UserID: run.UserID,
		GenerationOwner: pgtype.Text{
			String: owner.Value(),
			Valid:  true,
		},
		GenerationLeaseExpiresAt: pgtype.Timestamptz{
			Time:  owner.LeaseExpiresAt(now),
			Valid: true,
		},
		GenerationHeartbeatAt: pgtype.Timestamptz{
			Time:  now.UTC(),
			Valid: true,
		},
		GenerationAttempt: maxGenerationAttempts,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		return fmt.Errorf("claim ai chat run generation: %w", err)
	}

	claimedRun, err := mapRun(runRow)
	if err != nil {
		return err
	}
	*run = *claimedRun

	return nil
}

func (r *repository) HeartbeatRunGeneration(ctx context.Context, run *ChatRun, owner runOwner, now time.Time) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.queries.HeartbeatAIChatRunGeneration(ctx, db.HeartbeatAIChatRunGenerationParams{
		ID:     run.ID,
		UserID: run.UserID,
		GenerationLeaseExpiresAt: pgtype.Timestamptz{
			Time:  owner.LeaseExpiresAt(now),
			Valid: true,
		},
		GenerationHeartbeatAt: pgtype.Timestamptz{
			Time:  now.UTC(),
			Valid: true,
		},
		GenerationOwner: pgtype.Text{
			String: owner.Value(),
			Valid:  true,
		},
	})
	if err != nil {
		return false, fmt.Errorf("heartbeat ai chat run generation: %w", err)
	}

	return rows > 0, nil
}

func (r *repository) AppendStreamChunk(ctx context.Context, prepared *PreparedMessageStream, delta string, partialText string, updatedAt time.Time) (int32, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin ai chat streaming update transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	nextSequence := prepared.LastSequence + 1
	if _, err := qtx.CreateAIChatStreamChunk(ctx, db.CreateAIChatStreamChunkParams{
		RunID:           prepared.Run.ID,
		UserID:          prepared.Run.UserID,
		Sequence:        nextSequence,
		DeltaText:       delta,
		GenerationOwner: textValue(prepared.Run.GenerationOwner),
	}); err != nil {
		return 0, fmt.Errorf("create ai chat stream chunk: %w", err)
	}

	if _, err := qtx.UpdateAIChatMessageStreaming(ctx, db.UpdateAIChatMessageStreamingParams{
		ID:      prepared.AssistantMessage.ID,
		UserID:  prepared.AssistantMessage.UserID,
		Content: partialText,
	}); err != nil {
		return 0, fmt.Errorf("update ai chat assistant streaming message: %w", err)
	}

	if err := qtx.TouchAIChatRun(ctx, db.TouchAIChatRunParams{
		ID:     prepared.Run.ID,
		UserID: prepared.Run.UserID,
	}); err != nil {
		return 0, fmt.Errorf("touch ai chat streaming run: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit ai chat streaming update transaction: %w", err)
	}

	prepared.AssistantMessage.Content = partialText
	prepared.AssistantMessage.UpdatedAt = updatedAt.UTC()
	prepared.Run.UpdatedAt = updatedAt.UTC()
	prepared.LastSequence = nextSequence

	return nextSequence, nil
}

func (r *repository) InterruptRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, reason string, completedAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin ai chat interruption transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	completedTS := pgtype.Timestamptz{Time: completedAt.UTC(), Valid: true}
	errorText := truncateForStorage(streamInterruptedFailureMessage, 512)

	if _, err := qtx.UpdateAIChatMessageFailed(ctx, db.UpdateAIChatMessageFailedParams{
		ID:           prepared.AssistantMessage.ID,
		UserID:       prepared.AssistantMessage.UserID,
		Content:      strings.TrimSpace(partialText),
		ErrorMessage: textToPg(errorText),
		CompletedAt:  completedTS,
	}); err != nil {
		return fmt.Errorf("interrupt ai chat assistant message: %w", err)
	}

	if _, err := qtx.UpdateAIChatRunInterrupted(ctx, db.UpdateAIChatRunInterruptedParams{
		ID:                               prepared.Run.ID,
		UserID:                           prepared.Run.UserID,
		ErrorMessage:                     textToPg(errorText),
		CompletedAt:                      completedTS,
		InterruptionReason:               textToPg(reason),
		ExpectedGenerationStatus:         prepared.Run.GenerationStatus,
		ExpectedGenerationOwner:          textPtrToPg(prepared.Run.GenerationOwner),
		ExpectedGenerationLeaseExpiresAt: timePtrToPg(prepared.Run.LeaseExpiresAt),
		ExpectedGenerationAttempt:        prepared.Run.GenerationAttempt,
	}); err != nil {
		return fmt.Errorf("interrupt ai chat run: %w", err)
	}

	if err := qtx.TouchAIChatConversation(ctx, db.TouchAIChatConversationParams{
		ID:            prepared.Conversation.ID,
		UserID:        prepared.Conversation.UserID,
		LastMessageAt: completedTS,
	}); err != nil {
		return fmt.Errorf("touch ai chat conversation after interruption: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit ai chat interruption transaction: %w", err)
	}

	return nil
}

func (r *repository) CompleteRun(ctx context.Context, prepared *PreparedMessageStream, assistantText string, workoutDraft *workout.CreateWorkoutRequest, completedAt time.Time) (*ChatMessage, *ChatRun, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin ai chat completion transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	completedTS := pgtype.Timestamptz{Time: completedAt.UTC(), Valid: true}
	workoutDraftJSON, err := marshalWorkoutDraft(workoutDraft)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal ai chat workout draft: %w", err)
	}

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
		Column4:     string(workoutDraftJSON),
		Column5:     textValue(prepared.Run.GenerationOwner),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("complete ai chat run: %w", err)
	}

	if len(workoutDraftJSON) > 0 {
		if err := qtx.SetAIChatConversationLatestWorkoutDraft(ctx, db.SetAIChatConversationLatestWorkoutDraftParams{
			ID:      prepared.Conversation.ID,
			UserID:  prepared.Conversation.UserID,
			Column3: string(workoutDraftJSON),
			LatestWorkoutDraftSourceRunID: pgtype.Int4{
				Int32: prepared.Run.ID,
				Valid: true,
			},
		}); err != nil {
			return nil, nil, fmt.Errorf("persist latest ai chat workout draft: %w", err)
		}
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
		Column5:      textValue(prepared.Run.GenerationOwner),
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
