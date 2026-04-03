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

type recoveryDispatcher interface {
	EnqueueRunRecovery(ctx context.Context, request RunRecoveryRequest) error
}

type Service struct {
	logger        *slog.Logger
	featureAccess featureAccessService
	runtime       runtime
	repo          Repository
	recovery      recoveryDispatcher
}

const (
	streamProgressPersistInterval = 1 * time.Second
	streamProgressPersistTimeout  = 5 * time.Second
	recoverStatusQueued           = "queued"
	recoverStatusNotNeeded        = "not_needed"
	recoverReasonStreamReconnect  = "stream_reconnect"
	runAwaitingRecoveryMarker     = "ai chat stream interrupted and awaiting recovery handoff"
)

func NewService(logger *slog.Logger, featureAccess featureAccessService, runtime runtime, repo Repository) *Service {
	return &Service{
		logger:        logger,
		featureAccess: featureAccess,
		runtime:       runtime,
		repo:          repo,
	}
}

func (s *Service) SetRecoveryDispatcher(dispatcher recoveryDispatcher) {
	s.recovery = dispatcher
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
	if !isRunAwaitingRecovery(run) {
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
	return s.executePreparedRun(ctx, prepared, onChunk, true)
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
	if !isRunAwaitingRecovery(prepared.Run) {
		return nil
	}
	if err := s.repo.ClaimRunRecovery(ctx, request.RunID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	prepared.Run.ErrorMessage = nil

	_, err = s.executePreparedRun(ctx, prepared, nil, false)
	return err
}

func (s *Service) executePreparedRun(ctx context.Context, prepared *PreparedMessageStream, onChunk func(string) error, allowRecovery bool) (*StreamDone, error) {
	history := toRuntimeHistory(prepared.History)
	var partial strings.Builder
	persistCtx := context.WithoutCancel(ctx)
	progressSink := newStreamProgressSink(persistCtx, s.repo, prepared)
	ctx = withForegroundStreamDebug(ctx, onChunk != nil)

	done, err := s.runtime.StreamChat(ctx, prepared.Prompt, history, func(delta string) error {
		partial.WriteString(delta)
		partialText := strings.TrimSpace(partial.String())
		if err := progressSink.maybePersist(partialText); err != nil {
			s.logger.Warn("failed to persist ai chat streaming progress", "error", err, "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID)
		}
		if onChunk == nil {
			return nil
		}
		if err := onChunk(delta); err != nil {
			if persistErr := progressSink.forcePersist(partialText); persistErr != nil {
				s.logger.Warn("failed to persist ai chat streaming progress after stream write error", "error", persistErr, "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID)
			}
			return normalizeStreamChunkError(ctx, err)
		}
		return nil
	})
	if err != nil {
		if allowRecovery && errors.Is(err, ErrStreamDisconnected) {
			if markErr := s.repo.MarkRunAwaitingRecovery(persistCtx, prepared, strings.TrimSpace(partial.String()), time.Now().UTC()); markErr != nil {
				return nil, markErr
			}
			return nil, ErrStreamAwaitingRecovery
		}
		if failErr := s.failPreparedRun(persistCtx, prepared, partial.String(), err); failErr != nil {
			s.logger.Error("failed to persist ai chat run failure", "error", failErr, "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID)
		}
		return nil, err
	}

	text := strings.TrimSpace(done.Text)
	if text == "" {
		text = strings.TrimSpace(partial.String())
	}

	message, run, err := s.repo.CompleteRun(persistCtx, prepared, text, time.Now().UTC())
	if err != nil {
		persistErr := err
		if failErr := s.failPreparedRun(persistCtx, prepared, text, err); failErr != nil {
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

func (s *Service) AbortPreparedMessageStream(ctx context.Context, prepared *PreparedMessageStream, failure error) error {
	if prepared == nil {
		return nil
	}
	return s.failPreparedRun(context.WithoutCancel(ctx), prepared, prepared.AssistantMessage.Content, failure)
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
	return s.repo.FailRun(ctx, prepared, strings.TrimSpace(partialText), normalizeStreamFailure(failure), time.Now().UTC())
}

type streamProgressSink struct {
	ctx               context.Context
	repo              Repository
	prepared          *PreparedMessageStream
	now               func() time.Time
	lastPersistedAt   time.Time
	lastPersistedText string
}

func newStreamProgressSink(ctx context.Context, repo Repository, prepared *PreparedMessageStream) *streamProgressSink {
	return &streamProgressSink{
		ctx:      ctx,
		repo:     repo,
		prepared: prepared,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *streamProgressSink) maybePersist(partialText string) error {
	return s.persist(partialText, false)
}

func (s *streamProgressSink) forcePersist(partialText string) error {
	return s.persist(partialText, true)
}

func (s *streamProgressSink) persist(partialText string, force bool) error {
	partialText = strings.TrimSpace(partialText)
	if partialText == "" || partialText == s.lastPersistedText {
		return nil
	}

	now := s.now().UTC()
	if !force && !s.lastPersistedAt.IsZero() && now.Sub(s.lastPersistedAt) < streamProgressPersistInterval {
		return nil
	}

	persistCtx, cancel := context.WithTimeout(s.ctx, streamProgressPersistTimeout)
	defer cancel()

	if err := s.repo.UpdateStreamingRun(persistCtx, s.prepared, partialText, now); err != nil {
		return err
	}

	s.lastPersistedAt = now
	s.lastPersistedText = partialText

	return nil
}

func normalizeStreamChunkError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return ErrStreamDisconnected
	}

	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "broken pipe"),
		strings.Contains(lower, "connection reset"),
		strings.Contains(lower, "context canceled"),
		strings.Contains(lower, "client disconnected"),
		strings.Contains(lower, "transport is closing"),
		strings.Contains(lower, "eof"):
		return ErrStreamDisconnected
	default:
		return err
	}
}

func normalizeStreamFailure(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrStreamDisconnected) || errors.Is(err, ErrStreamNotStarted) {
		return err
	}
	return err
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
		errors.Is(err, ErrRecoveryUnavailable) ||
		errors.Is(err, ErrConversationBusy) ||
		errors.Is(err, ErrGenerationTimeout)
}

func isRunAwaitingRecovery(run *ChatRun) bool {
	if run == nil || run.ErrorMessage == nil {
		return false
	}
	return strings.TrimSpace(*run.ErrorMessage) == runAwaitingRecoveryMarker
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
