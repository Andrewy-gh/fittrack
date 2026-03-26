package aichat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
)

type featureAccessService interface {
	HasCurrentUserFeatureAccess(ctx context.Context, featureKey string) (bool, error)
}

type runtime interface {
	ModelName() string
	Available() bool
	GenerateValidation(ctx context.Context, prompt string) (*ValidationOutput, error)
	StreamValidation(ctx context.Context, prompt string, onChunk func(string) error) (*StreamDone, error)
	StreamChat(ctx context.Context, prompt string, history []RuntimeChatMessage, onChunk func(string) error) (*StreamDone, error)
}

type Service struct {
	logger        *slog.Logger
	featureAccess featureAccessService
	runtime       runtime
	repo          Repository
}

func NewService(logger *slog.Logger, featureAccess featureAccessService, runtime runtime, repo Repository) *Service {
	return &Service{
		logger:        logger,
		featureAccess: featureAccess,
		runtime:       runtime,
		repo:          repo,
	}
}

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

	return &ConversationDetail{
		Conversation: conversation,
		Messages:     messages,
	}, nil
}

func (s *Service) PrepareMessageStream(ctx context.Context, conversationID int32, prompt string, requestID string) (*PreparedMessageStream, error) {
	if err := s.ensureAllowed(ctx); err != nil {
		return nil, err
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	prepared, err := s.repo.PrepareMessageStream(ctx, conversationID, userID, prompt, s.runtime.ModelName(), requestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newConversationNotFound(conversationID)
		}
		return nil, err
	}

	return prepared, nil
}

func (s *Service) StreamMessage(ctx context.Context, prepared *PreparedMessageStream, onChunk func(string) error) (*StreamDone, error) {
	history := toRuntimeHistory(prepared.History)
	var partial strings.Builder

	done, err := s.runtime.StreamChat(ctx, prepared.Prompt, history, func(delta string) error {
		partial.WriteString(delta)
		return onChunk(delta)
	})
	if err != nil {
		if failErr := s.failPreparedRun(ctx, prepared, partial.String(), err); failErr != nil {
			s.logger.Error("failed to persist ai chat run failure", "error", failErr, "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID)
		}
		return nil, err
	}

	text := strings.TrimSpace(done.Text)
	if text == "" {
		text = strings.TrimSpace(partial.String())
	}

	message, run, err := s.repo.CompleteRun(ctx, prepared, text, time.Now().UTC())
	if err != nil {
		persistErr := err
		if failErr := s.failPreparedRun(ctx, prepared, text, err); failErr != nil {
			s.logger.Error("failed to mark ai chat run failed after completion error", "error", failErr, "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID)
		}
		return nil, persistErr
	}

	return &StreamDone{
		ConversationID: prepared.Conversation.ID,
		RunID:          run.ID,
		MessageID:      message.ID,
		Model:          run.Model,
		Text:           text,
	}, nil
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

func (s *Service) failPreparedRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, failure error) error {
	return s.repo.FailRun(ctx, prepared, strings.TrimSpace(partialText), failure, time.Now().UTC())
}

func toRuntimeHistory(history []ChatMessage) []RuntimeChatMessage {
	messages := make([]RuntimeChatMessage, 0, len(history))
	for _, message := range history {
		if message.Status != statusCompleted {
			continue
		}
		if message.Role != roleUser && message.Role != roleAssistant {
			continue
		}
		text := strings.TrimSpace(message.Content)
		if text == "" {
			continue
		}
		messages = append(messages, RuntimeChatMessage{
			Role: message.Role,
			Text: text,
		})
	}

	return messages
}

func isClientSafeError(err error) bool {
	return errors.Is(err, ErrFeatureDisabled) ||
		errors.Is(err, ErrRuntimeUnavailable) ||
		errors.Is(err, ErrConversationBusy) ||
		errors.Is(err, ErrGenerationTimeout)
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
