package aichat

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Service) StopRun(ctx context.Context, conversationID, runID int32) (*StopRunResponse, error) {
	if err := s.ensureFeatureAccess(ctx); err != nil {
		return nil, err
	}
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := s.repo.StopRun(ctx, conversationID, runID, userID, time.Now().UTC())
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, newConversationNotFound(conversationID)
	}
	if err != nil {
		return nil, err
	}
	if result.Status == statusStopped {
		s.cancelMu.Lock()
		cancel := s.runCancels[runID]
		s.cancelMu.Unlock()
		if cancel != nil {
			cancel()
		}
		s.logger.Info("ai chat run stopped", "conversation_id", conversationID, "run_id", runID, "sequence", result.Sequence)
	}
	return result, nil
}
