package aichat

import (
	"context"
	"errors"
	"strings"
	"time"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
)

func (s *Service) RequestMessageRecovery(ctx context.Context, conversationID int32, reason string) (*RecoverMessageResponse, error) {
	if err := s.ensureFeatureAccess(ctx); err != nil {
		return nil, err
	}
	if s.recovery == nil {
		return nil, ErrRecoveryUnavailable
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.GetConversation(ctx, conversationID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newConversationNotFound(conversationID)
		}
		return nil, err
	}

	run, err := s.repo.GetActiveRunForConversation(ctx, conversationID, userID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return &RecoverMessageResponse{
			ConversationID: conversationID,
			Status:         recoverStatusNotNeeded,
		}, nil
	case err != nil:
		return nil, err
	}
	if !shouldRecoverRun(run, time.Now().UTC()) {
		return &RecoverMessageResponse{
			ConversationID: conversationID,
			Status:         recoverStatusNotNeeded,
		}, nil
	}

	request := RunRecoveryRequest{
		ConversationID: conversationID,
		RunID:          run.ID,
		UserID:         userID,
		Reason:         strings.TrimSpace(reason),
	}
	if request.Reason == "" {
		request.Reason = recoverReasonStreamReconnect
	}

	if err := s.recovery.EnqueueRunRecovery(ctx, request); err != nil {
		return nil, err
	}

	return &RecoverMessageResponse{
		ConversationID: conversationID,
		RunID:          run.ID,
		Status:         recoverStatusQueued,
	}, nil
}

func (s *Service) RecoverStreamingRun(ctx context.Context, request RunRecoveryRequest) error {
	if s.runtime == nil || !s.runtime.Available() {
		return ErrRuntimeUnavailable
	}

	userID := strings.TrimSpace(request.UserID)
	if userID == "" {
		return apperrors.NewUnauthorized("ai chat", "")
	}

	ctx = user.WithContext(ctx, userID)
	prepared, err := s.repo.LoadPreparedRunForRecovery(ctx, request.RunID, userID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil
	case err != nil:
		return err
	}

	if prepared.Run == nil || prepared.Run.Status != statusStreaming {
		return nil
	}
	now := time.Now().UTC()
	if !shouldRecoverRun(prepared.Run, now) {
		return nil
	}
	if prepared.LastSequence > 0 || strings.TrimSpace(prepared.AssistantMessage.Content) != "" {
		return s.repo.InterruptRun(
			context.WithoutCancel(ctx),
			prepared,
			strings.TrimSpace(prepared.AssistantMessage.Content),
			interruptionReasonStalePartial,
			time.Now().UTC(),
		)
	}
	if generationAttemptsExhausted(prepared.Run) {
		return s.repo.InterruptRun(
			context.WithoutCancel(ctx),
			prepared,
			strings.TrimSpace(prepared.AssistantMessage.Content),
			interruptionReasonAttemptsExhausted,
			time.Now().UTC(),
		)
	}

	owner := newInngestRunOwner(prepared.Run.ID)
	if err := s.repo.ClaimRunGeneration(ctx, prepared.Run, owner, now); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	_, err = s.executeOwnedRun(ctx, prepared, owner)
	return err
}

func shouldRecoverRun(run *ChatRun, now time.Time) bool {
	if run == nil || run.Status != statusStreaming {
		return false
	}
	switch run.GenerationStatus {
	case generationStatusGenerating:
		return run.LeaseExpiresAt != nil && now.UTC().After(run.LeaseExpiresAt.UTC())
	case generationStatusQueued:
		return isStreamingRunStale(run.UpdatedAt, now)
	default:
		return false
	}
}

func generationAttemptsExhausted(run *ChatRun) bool {
	return run != nil && run.GenerationAttempt >= maxGenerationAttempts
}
