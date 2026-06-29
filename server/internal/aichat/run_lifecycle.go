package aichat

import "time"

type chatRunLifecycleState string

const (
	chatRunLifecycleInvalid             chatRunLifecycleState = "invalid"
	chatRunLifecycleQueued              chatRunLifecycleState = "queued"
	chatRunLifecycleOwnedGeneration     chatRunLifecycleState = "owned_generation"
	chatRunLifecycleRecoverableStaleRun chatRunLifecycleState = "recoverable_stale_run"
	chatRunLifecycleTerminal            chatRunLifecycleState = "terminal"
)

type chatRunLifecycle struct {
	run   *ChatRun
	now   time.Time
	state chatRunLifecycleState
}

func newChatRunLifecycle(run *ChatRun, now time.Time) chatRunLifecycle {
	lifecycle := chatRunLifecycle{
		run:   run,
		now:   now.UTC(),
		state: chatRunLifecycleInvalid,
	}
	lifecycle.state = lifecycle.resolveState()
	return lifecycle
}

func (l chatRunLifecycle) State() chatRunLifecycleState {
	return l.state
}

func (l chatRunLifecycle) IsQueued() bool {
	return l.state == chatRunLifecycleQueued
}

func (l chatRunLifecycle) IsOwnedGeneration() bool {
	return l.state == chatRunLifecycleOwnedGeneration
}

func (l chatRunLifecycle) IsRecoverableStaleRun() bool {
	return l.state == chatRunLifecycleRecoverableStaleRun
}

func (l chatRunLifecycle) IsTerminal() bool {
	return l.state == chatRunLifecycleTerminal
}

func (l chatRunLifecycle) ShouldRecover() bool {
	return l.IsRecoverableStaleRun()
}

func (l chatRunLifecycle) CanClaimGeneration() bool {
	if l.GenerationAttemptsExhausted() {
		return false
	}
	return l.IsQueued() || l.IsRecoverableStaleRun()
}

func (l chatRunLifecycle) GenerationAttemptsExhausted() bool {
	return l.run != nil && l.run.GenerationAttempt >= maxGenerationAttempts
}

func (l chatRunLifecycle) resolveState() chatRunLifecycleState {
	if l.run == nil {
		return chatRunLifecycleInvalid
	}
	if l.run.Status != statusStreaming || isTerminalGenerationStatus(l.run.GenerationStatus) {
		return chatRunLifecycleTerminal
	}

	switch l.run.GenerationStatus {
	case generationStatusQueued:
		if isStreamingRunStale(l.run.UpdatedAt, l.now) {
			return chatRunLifecycleRecoverableStaleRun
		}
		return chatRunLifecycleQueued
	case generationStatusGenerating:
		if l.run.LeaseExpiresAt != nil && l.now.After(l.run.LeaseExpiresAt.UTC()) {
			return chatRunLifecycleRecoverableStaleRun
		}
		if l.run.GenerationOwner != nil && l.run.LeaseExpiresAt != nil {
			return chatRunLifecycleOwnedGeneration
		}
		return chatRunLifecycleInvalid
	default:
		return chatRunLifecycleInvalid
	}
}

func isTerminalGenerationStatus(status string) bool {
	switch status {
	case generationStatusCompleted, generationStatusFailed, generationStatusInterrupted:
		return true
	default:
		return false
	}
}
