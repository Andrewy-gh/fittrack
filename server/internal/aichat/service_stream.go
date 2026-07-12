package aichat

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

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
	claimCtx, claimCancel := s.registerRunCancellation(context.WithoutCancel(ctx), prepared.Run.ID, owner)
	if err := s.repo.ClaimRunGeneration(claimCtx, prepared.Run, owner, now); err != nil {
		s.unregisterRunCancellation(prepared.Run.ID, owner)
		claimCancel()
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConversationBusy
		}
		return err
	}
	generationCtx, timeoutCancel := context.WithTimeout(claimCtx, chatGenerationTimeout)
	cancel := func() {
		timeoutCancel()
		claimCancel()
	}

	go s.runOwnedGeneration(generationCtx, prepared, owner, cancel)
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
			if prepared.Run.Status == statusCompleted || prepared.Run.Status == statusStopped {
				return &StreamDone{
					ConversationID: prepared.Conversation.ID,
					RunID:          prepared.Run.ID,
					MessageID:      prepared.AssistantMessage.ID,
					Model:          prepared.Run.Model,
					Text:           prepared.AssistantMessage.Content,
					Sequence:       prepared.LastSequence,
					WorkoutDraft:   prepared.Run.WorkoutDraft,
					Status:         prepared.Run.Status,
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

func (s *Service) runOwnedGeneration(ctx context.Context, prepared *PreparedMessageStream, owner runOwner, cancel context.CancelFunc) {
	if _, err := s.executeOwnedRun(ctx, prepared, owner, cancel); err != nil {
		if errors.Is(err, context.Canceled) {
			s.logger.Info("ai chat generation canceled", "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID, "owner", owner.Value())
			return
		}
		s.logger.Error("failed to complete owned ai chat generation",
			"error", err,
			"conversation_id", prepared.Conversation.ID,
			"run_id", prepared.Run.ID,
			"owner", owner.Value(),
		)
	}
}

func (s *Service) executeOwnedRun(generationCtx context.Context, prepared *PreparedMessageStream, owner runOwner, cancel context.CancelFunc) (*StreamDone, error) {
	defer cancel()
	defer s.unregisterRunCancellation(prepared.Run.ID, owner)

	heartbeatDone := make(chan struct{})
	go s.heartbeatRunGeneration(generationCtx, cancel, prepared, owner, heartbeatDone)

	done, err := s.executePreparedRun(generationCtx, prepared, nil)
	cancel()
	<-heartbeatDone
	return done, err
}

func (s *Service) registerRunCancellation(ctx context.Context, runID int32, owner runOwner) (context.Context, context.CancelFunc) {
	generationCtx, cancel := context.WithCancel(ctx)
	s.cancelMu.Lock()
	s.runCancels[runID] = runCancellation{owner: owner.Value(), cancel: cancel}
	s.cancelMu.Unlock()
	return generationCtx, cancel
}

func (s *Service) unregisterRunCancellation(runID int32, owner runOwner) {
	s.cancelMu.Lock()
	defer s.cancelMu.Unlock()
	if registration, ok := s.runCancels[runID]; ok && registration.owner == owner.Value() {
		delete(s.runCancels, runID)
	}
}

func (s *Service) heartbeatRunGeneration(ctx context.Context, cancel context.CancelFunc, prepared *PreparedMessageStream, owner runOwner, done chan<- struct{}) {
	defer close(done)

	stopTicker := time.NewTicker(generationStopPollInterval)
	defer stopTicker.Stop()
	heartbeatTicker := time.NewTicker(generationHeartbeatInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-stopTicker.C:
			owned, err := s.repo.OwnsRunGeneration(context.WithoutCancel(ctx), prepared.Run, owner)
			if err != nil {
				s.logger.Error("failed to check ai chat generation ownership",
					"error", err,
					"conversation_id", prepared.Conversation.ID,
					"run_id", prepared.Run.ID,
					"owner", owner.Value(),
				)
				continue
			}
			if !owned {
				s.logger.Warn("ai chat generation lost ownership",
					"conversation_id", prepared.Conversation.ID,
					"run_id", prepared.Run.ID,
					"owner", owner.Value(),
				)
				cancel()
				return
			}
		case now := <-heartbeatTicker.C:
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
	runtimeCtx := contextWithTrainingProfileSource(ctx, prepared.Conversation.ID, prepared.Run.UserMessageID)
	done, err := s.runtime.StreamChat(runtimeCtx, prepared.Prompt, history, func(delta string) error {
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
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
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
		ToolCalls:      done.ToolCalls,
	}, nil
}

func (s *Service) AbortPreparedMessageStream(ctx context.Context, prepared *PreparedMessageStream, failure error) error {
	if prepared == nil {
		return nil
	}
	return s.failPreparedRun(context.WithoutCancel(ctx), prepared, prepared.AssistantMessage.Content, failure)
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
		isVisibleStoppedAssistant := message.Role == roleAssistant &&
			message.Status == statusStopped
		if message.Status != statusCompleted && !isVisibleStoppedAssistant {
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

func shouldStopResumeAndRecover(run *ChatRun, now time.Time) bool {
	return shouldRecoverRun(run, now)
}
