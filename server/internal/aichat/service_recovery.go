package aichat

import (
	"context"
	"errors"
	"fmt"
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

	now := time.Now().UTC()
	switch decision := decideRunRecovery(prepared, now).(type) {
	case keepRunRecovery:
		return nil
	case interruptRunRecovery:
		return s.repo.InterruptRun(
			context.WithoutCancel(ctx),
			prepared,
			decision.text,
			decision.reason,
			now,
		)
	case claimRunRecovery:
		// A lost claim is a rejected proposal; it must not call the provider.
	default:
		return fmt.Errorf("unexpected ai chat recovery decision %T", decision)
	}

	owner := newInngestRunOwner(prepared.Run.ID)
	claimCtx, claimCancel := s.registerRunCancellation(ctx, prepared.Run.ID, owner)
	if err := s.repo.ClaimRunGeneration(claimCtx, prepared.Run, owner, now); err != nil {
		s.unregisterRunCancellation(prepared.Run.ID, owner)
		claimCancel()
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	generationCtx, timeoutCancel := context.WithTimeout(claimCtx, chatGenerationTimeout)
	cancel := func() {
		timeoutCancel()
		claimCancel()
	}

	_, err = s.executeOwnedRun(generationCtx, prepared, owner, cancel)
	return err
}

func shouldRecoverRun(run *ChatRun, now time.Time) bool {
	return newChatRunLifecycle(run, now).ShouldRecover()
}
