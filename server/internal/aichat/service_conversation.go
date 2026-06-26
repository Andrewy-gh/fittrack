package aichat

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5"
)

const conversationListLimit int32 = 50

func (s *Service) Validate(ctx context.Context, prompt string) (*ValidateResponse, error) {
	if err := s.ensureAllowed(ctx); err != nil {
		return nil, err
	}

	output, err := s.runtime.GenerateValidation(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &ValidateResponse{
		Model:  s.runtime.ModelName(),
		Output: output,
	}, nil
}

func (s *Service) EnsureAllowed(ctx context.Context) error {
	return s.ensureAllowed(ctx)
}

func (s *Service) CurrentUserHasFeatureAccess(ctx context.Context) (bool, error) {
	return s.featureAccess.HasCurrentUserFeatureAccess(ctx, featureKeyAIChatbot)
}

func (s *Service) StreamValidate(ctx context.Context, prompt string, onChunk func(string) error) (*StreamDone, error) {
	if err := s.ensureAllowed(ctx); err != nil {
		return nil, err
	}

	return s.runtime.StreamValidation(ctx, prompt, onChunk)
}

func (s *Service) CreateConversation(ctx context.Context) (*Conversation, error) {
	if err := s.ensureFeatureAccess(ctx); err != nil {
		return nil, err
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.CreateConversation(ctx, userID)
}

func (s *Service) ListConversations(ctx context.Context) ([]ConversationSummary, error) {
	if err := s.ensureFeatureAccess(ctx); err != nil {
		return nil, err
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListConversations(ctx, userID, conversationListLimit)
}

func (s *Service) GetConversation(ctx context.Context, conversationID int32) (*ConversationDetail, error) {
	if err := s.ensureFeatureAccess(ctx); err != nil {
		return nil, err
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	conversation, err := s.repo.GetConversation(ctx, conversationID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newConversationNotFound(conversationID)
		}
		return nil, err
	}

	messages, err := s.repo.ListMessages(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}

	var activeRunView *ConversationRunView
	activeRun, err := s.repo.GetActiveRunForConversation(ctx, conversationID, userID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
	case err != nil:
		return nil, err
	default:
		latestSequence, seqErr := s.repo.GetLatestStreamSequence(ctx, activeRun.ID, userID)
		if seqErr != nil {
			return nil, seqErr
		}
		activeRunView = &ConversationRunView{
			ID:                 activeRun.ID,
			AssistantMessageID: activeRun.AssistantMessageID,
			Status:             activeRun.Status,
			LatestSequence:     latestSequence,
		}
	}

	return &ConversationDetail{
		Conversation: conversation,
		Messages:     messages,
		ActiveRun:    activeRunView,
	}, nil
}

func (s *Service) SaveLatestWorkoutDraft(ctx context.Context, conversationID int32) (*SaveLatestWorkoutDraftResponse, error) {
	if err := s.ensureFeatureAccess(ctx); err != nil {
		return nil, err
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	saved, err := s.repo.SaveLatestWorkoutDraft(ctx, SaveLatestWorkoutDraftRequest{
		ConversationID: conversationID,
		UserID:         userID,
		SavedAt:        time.Now().UTC(),
		SaveWorkout:    s.saveWorkoutDraftTx,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newConversationNotFound(conversationID)
		}
		return nil, err
	}

	return &SaveLatestWorkoutDraftResponse{
		Conversation: saved.Conversation,
		WorkoutID:    saved.WorkoutID,
	}, nil
}

func (s *Service) saveWorkoutDraftTx(ctx context.Context, qtx *db.Queries, draft workout.CreateWorkoutRequest, userID string) (int32, error) {
	if s.workoutDraftSaver == nil {
		return 0, errors.New("ai chat workout draft saver is unavailable")
	}
	return s.workoutDraftSaver.SaveWorkoutTx(ctx, qtx, draft, userID)
}

func (s *Service) ensureAllowed(ctx context.Context) error {
	if s.runtime == nil || !s.runtime.Available() {
		return ErrRuntimeUnavailable
	}

	return s.ensureFeatureAccess(ctx)
}

func (s *Service) ensureFeatureAccess(ctx context.Context) error {
	hasAccess, err := s.featureAccess.HasCurrentUserFeatureAccess(ctx, featureKeyAIChatbot)
	if err != nil {
		return err
	}
	if !hasAccess {
		return ErrFeatureDisabled
	}

	return nil
}

func currentUserID(ctx context.Context) (string, error) {
	userID, ok := user.Current(ctx)
	if !ok || userID == "" {
		return "", apperrors.NewUnauthorized("ai chat", "")
	}
	return userID, nil
}

func newConversationNotFound(conversationID int32) error {
	return apperrors.NewNotFound("ai conversation", fmt.Sprintf("%d", conversationID))
}
