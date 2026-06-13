package aichat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
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
	logger            *slog.Logger
	featureAccess     featureAccessService
	runtime           runtime
	repo              Repository
	recovery          recoveryDispatcher
	workoutDraftSaver workout.TxSaver
}

const (
	streamResumePollInterval     = 250 * time.Millisecond
	recoverStatusQueued          = "queued"
	recoverStatusNotNeeded       = "not_needed"
	recoverReasonStreamReconnect = "stream_reconnect"
)

func NewService(logger *slog.Logger, featureAccess featureAccessService, runtime runtime, repo Repository, workoutDraftSaver workout.TxSaver) *Service {
	return &Service{
		logger:            logger,
		featureAccess:     featureAccess,
		runtime:           runtime,
		repo:              repo,
		workoutDraftSaver: workoutDraftSaver,
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

func (s *Service) StreamMessage(ctx context.Context, prepared *PreparedMessageStream, onChunk func(StreamChunk) error) (*StreamDone, error) {
	return s.executePreparedRun(ctx, prepared, onChunk)
}

func (s *Service) StartMessageGeneration(ctx context.Context, prepared *PreparedMessageStream) error {
	owner := newAPIRunOwner()
	now := time.Now().UTC()
	if err := s.repo.ClaimRunGeneration(context.WithoutCancel(ctx), prepared.Run, owner, now); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConversationBusy
		}
		return err
	}

	go s.runOwnedGeneration(context.WithoutCancel(ctx), prepared, owner)
	return nil
}

func (s *Service) PrepareResumeMessageStream(ctx context.Context, conversationID int32, runID int32, afterSequence int32) (*PreparedResumeStream, error) {
	if err := s.ensureFeatureAccess(ctx); err != nil {
		return nil, err
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	prepared, err := s.repo.LoadPreparedRunForResume(ctx, conversationID, runID, userID, afterSequence)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newConversationNotFound(conversationID)
		}
		return nil, err
	}

	return prepared, nil
}

func (s *Service) ResumeMessageStream(ctx context.Context, prepared *PreparedResumeStream, onChunk func(StreamChunk) error) (*StreamDone, error) {
	currentSequence := prepared.AfterSequence

	for {
		chunks, err := s.repo.ListStreamChunksAfter(ctx, prepared.Run.ID, prepared.Run.UserID, currentSequence)
		if err != nil {
			return nil, err
		}
		for _, chunk := range chunks {
			currentSequence = chunk.Sequence
			if onChunk == nil {
				continue
			}
			if err := onChunk(chunk); err != nil {
				return nil, normalizeStreamChunkError(ctx, err)
			}
		}

		refreshed, err := s.repo.LoadPreparedRunForResume(ctx, prepared.Conversation.ID, prepared.Run.ID, prepared.Run.UserID, currentSequence)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, newConversationNotFound(prepared.Conversation.ID)
			}
			return nil, err
		}
		prepared = refreshed

		if prepared.Run.Status != statusStreaming {
			if prepared.Run.Status == statusCompleted {
				return &StreamDone{
					ConversationID: prepared.Conversation.ID,
					RunID:          prepared.Run.ID,
					MessageID:      prepared.AssistantMessage.ID,
					Model:          prepared.Run.Model,
					Text:           prepared.AssistantMessage.Content,
					Sequence:       prepared.LastSequence,
					WorkoutDraft:   prepared.Run.WorkoutDraft,
				}, nil
			}

			failureMessage := "failed to resume ai chat response"
			if prepared.AssistantMessage.ErrorMessage != nil {
				failureMessage = *prepared.AssistantMessage.ErrorMessage
			} else if prepared.Run.ErrorMessage != nil {
				failureMessage = *prepared.Run.ErrorMessage
			}
			return nil, errors.New(failureMessage)
		}

		if shouldStopResumeAndRecover(prepared.Run, time.Now().UTC()) {
			return nil, ErrStreamAwaitingRecovery
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(streamResumePollInterval):
		}
	}
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

func (s *Service) runOwnedGeneration(ctx context.Context, prepared *PreparedMessageStream, owner runOwner) {
	if _, err := s.executeOwnedRun(ctx, prepared, owner); err != nil {
		s.logger.Error("failed to complete owned ai chat generation",
			"error", err,
			"conversation_id", prepared.Conversation.ID,
			"run_id", prepared.Run.ID,
			"owner", owner.Value(),
		)
	}
}

func (s *Service) executeOwnedRun(ctx context.Context, prepared *PreparedMessageStream, owner runOwner) (*StreamDone, error) {
	generationCtx, cancel := context.WithTimeout(ctx, chatGenerationTimeout)
	defer cancel()

	heartbeatDone := make(chan struct{})
	go s.heartbeatRunGeneration(generationCtx, cancel, prepared, owner, heartbeatDone)

	done, err := s.executePreparedRun(generationCtx, prepared, nil)
	cancel()
	<-heartbeatDone
	return done, err
}

func (s *Service) heartbeatRunGeneration(ctx context.Context, cancel context.CancelFunc, prepared *PreparedMessageStream, owner runOwner, done chan<- struct{}) {
	defer close(done)

	ticker := time.NewTicker(generationHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			ok, err := s.repo.HeartbeatRunGeneration(context.WithoutCancel(ctx), prepared.Run, owner, now.UTC())
			if err != nil {
				s.logger.Error("failed to heartbeat ai chat generation",
					"error", err,
					"conversation_id", prepared.Conversation.ID,
					"run_id", prepared.Run.ID,
					"owner", owner.Value(),
				)
				continue
			}
			if !ok {
				s.logger.Warn("ai chat generation lost ownership lease",
					"conversation_id", prepared.Conversation.ID,
					"run_id", prepared.Run.ID,
					"owner", owner.Value(),
				)
				cancel()
				return
			}
		}
	}
}

func (s *Service) executePreparedRun(ctx context.Context, prepared *PreparedMessageStream, onChunk func(StreamChunk) error) (*StreamDone, error) {
	traceStartedAt := time.Now()
	history := toRuntimeHistory(prepared.History)
	var partial strings.Builder
	persistCtx := context.WithoutCancel(ctx)

	// Trace marker: measures the whole service layer around model streaming and persistence.
	logAIChatTrace(s.logger, "service_stream_started",
		"conversation_id", prepared.Conversation.ID,
		"run_id", prepared.Run.ID,
		"message_id", prepared.AssistantMessage.ID,
		"history_messages", len(history),
	)
	firstChunkPersisted := false
	done, err := s.runtime.StreamChat(ctx, prepared.Prompt, history, func(delta string) error {
		partial.WriteString(delta)
		partialText := strings.TrimSpace(partial.String())
		sequence, err := s.repo.AppendStreamChunk(persistCtx, prepared, delta, partialText, time.Now().UTC())
		if err != nil {
			return err
		}
		if !firstChunkPersisted {
			firstChunkPersisted = true
			// Trace marker: shows when the first model delta was persisted before SSE delivery.
			logAIChatTrace(s.logger, "first_chunk_persisted",
				"elapsed_ms", time.Since(traceStartedAt).Milliseconds(),
				"conversation_id", prepared.Conversation.ID,
				"run_id", prepared.Run.ID,
				"message_id", prepared.AssistantMessage.ID,
				"sequence", sequence,
			)
		}
		if onChunk == nil {
			return nil
		}
		if err := onChunk(StreamChunk{Delta: delta, Sequence: sequence}); err != nil {
			return normalizeStreamChunkError(ctx, err)
		}
		return nil
	})
	if err != nil {
		if failErr := s.failPreparedRun(persistCtx, prepared, partial.String(), err); failErr != nil {
			s.logger.Error("failed to persist ai chat run failure", "error", failErr, "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID)
		}
		return nil, err
	}
	// Trace marker: captures total runtime latency before final DB completion.
	logAIChatTrace(s.logger, "runtime_stream_finished",
		"elapsed_ms", time.Since(traceStartedAt).Milliseconds(),
		"conversation_id", prepared.Conversation.ID,
		"run_id", prepared.Run.ID,
		"message_id", prepared.AssistantMessage.ID,
		"has_workout_draft", done.WorkoutDraft != nil,
	)

	text := strings.TrimSpace(done.Text)
	if text == "" {
		text = strings.TrimSpace(partial.String())
	}

	completeStartedAt := time.Now()
	message, run, err := s.repo.CompleteRun(persistCtx, prepared, text, done.WorkoutDraft, time.Now().UTC())
	if err != nil {
		persistErr := err
		if failErr := s.failPreparedRun(persistCtx, prepared, text, err); failErr != nil {
			s.logger.Error("failed to mark ai chat run failed after completion error", "error", failErr, "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID)
		}
		return nil, persistErr
	}
	// Trace marker: separates model latency from final persistence latency.
	logAIChatTrace(s.logger, "complete_run_finished",
		"elapsed_ms", time.Since(traceStartedAt).Milliseconds(),
		"duration_ms", time.Since(completeStartedAt).Milliseconds(),
		"conversation_id", prepared.Conversation.ID,
		"run_id", run.ID,
		"message_id", message.ID,
		"has_workout_draft", done.WorkoutDraft != nil,
	)
	prepared.AssistantMessage = message
	prepared.Run = run

	return &StreamDone{
		ConversationID: prepared.Conversation.ID,
		RunID:          run.ID,
		MessageID:      message.ID,
		Model:          run.Model,
		Text:           text,
		Sequence:       prepared.LastSequence,
		WorkoutDraft:   done.WorkoutDraft,
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
	partialText = strings.TrimSpace(partialText)
	failure = normalizeStreamFailure(failure)
	completedAt := time.Now().UTC()
	if err := s.repo.FailRun(ctx, prepared, partialText, failure, completedAt); err != nil {
		return err
	}
	failureMessage := failure.Error()

	prepared.AssistantMessage.Content = partialText
	prepared.AssistantMessage.Status = statusFailed
	prepared.AssistantMessage.ErrorMessage = &failureMessage
	prepared.AssistantMessage.CompletedAt = &completedAt
	prepared.AssistantMessage.UpdatedAt = completedAt
	prepared.Run.Status = statusFailed
	prepared.Run.ErrorMessage = &failureMessage
	prepared.Run.CompletedAt = &completedAt
	prepared.Run.UpdatedAt = completedAt

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

func shouldStopResumeAndRecover(run *ChatRun, now time.Time) bool {
	return shouldRecoverRun(run, now)
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
