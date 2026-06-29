package aichat

import (
	"testing"
	"time"
)

func TestChatRunLifecycleStates(t *testing.T) {
	now := time.Date(2026, 6, 29, 15, 0, 0, 0, time.UTC)
	staleUpdatedAt := now.Add(-streamingRunStaleAfter - time.Second)
	freshUpdatedAt := now.Add(-streamingRunStaleAfter + time.Second)
	expiredLease := now.Add(-time.Second)
	validLease := now.Add(time.Second)
	owner := "api:worker-1"

	tests := []struct {
		name              string
		run               *ChatRun
		wantState         chatRunLifecycleState
		wantQueued        bool
		wantOwned         bool
		wantRecoverable   bool
		wantTerminal      bool
		wantCanClaim      bool
		wantAttemptsSpent bool
	}{
		{
			name:      "nil run",
			wantState: chatRunLifecycleInvalid,
		},
		{
			name: "completed run is terminal",
			run: &ChatRun{
				Status:           statusCompleted,
				GenerationStatus: generationStatusCompleted,
			},
			wantState:    chatRunLifecycleTerminal,
			wantTerminal: true,
		},
		{
			name: "fresh queued streaming run can be claimed but does not need recovery",
			run: &ChatRun{
				Status:           statusStreaming,
				GenerationStatus: generationStatusQueued,
				UpdatedAt:        freshUpdatedAt,
			},
			wantState:    chatRunLifecycleQueued,
			wantQueued:   true,
			wantCanClaim: true,
		},
		{
			name: "stale queued streaming run is recoverable",
			run: &ChatRun{
				Status:           statusStreaming,
				GenerationStatus: generationStatusQueued,
				UpdatedAt:        staleUpdatedAt,
			},
			wantState:       chatRunLifecycleRecoverableStaleRun,
			wantRecoverable: true,
			wantCanClaim:    true,
		},
		{
			name: "fresh owned generation is not claimable",
			run: &ChatRun{
				Status:           statusStreaming,
				GenerationStatus: generationStatusGenerating,
				GenerationOwner:  &owner,
				LeaseExpiresAt:   &validLease,
			},
			wantState: chatRunLifecycleOwnedGeneration,
			wantOwned: true,
		},
		{
			name: "expired owned generation is recoverable and claimable",
			run: &ChatRun{
				Status:           statusStreaming,
				GenerationStatus: generationStatusGenerating,
				GenerationOwner:  &owner,
				LeaseExpiresAt:   &expiredLease,
			},
			wantState:       chatRunLifecycleRecoverableStaleRun,
			wantRecoverable: true,
			wantCanClaim:    true,
		},
		{
			name: "generating run without lease is invalid",
			run: &ChatRun{
				Status:           statusStreaming,
				GenerationStatus: generationStatusGenerating,
				GenerationOwner:  &owner,
			},
			wantState: chatRunLifecycleInvalid,
		},
		{
			name: "interrupted generation is terminal even when run status is failed",
			run: &ChatRun{
				Status:           statusFailed,
				GenerationStatus: generationStatusInterrupted,
			},
			wantState:    chatRunLifecycleTerminal,
			wantTerminal: true,
		},
		{
			name: "exhausted stale generation still needs recovery but cannot be claimed",
			run: &ChatRun{
				Status:            statusStreaming,
				GenerationStatus:  generationStatusGenerating,
				GenerationOwner:   &owner,
				LeaseExpiresAt:    &expiredLease,
				GenerationAttempt: maxGenerationAttempts,
			},
			wantState:         chatRunLifecycleRecoverableStaleRun,
			wantRecoverable:   true,
			wantAttemptsSpent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := newChatRunLifecycle(tt.run, now)

			if lifecycle.State() != tt.wantState {
				t.Fatalf("state = %q, want %q", lifecycle.State(), tt.wantState)
			}
			if lifecycle.IsQueued() != tt.wantQueued {
				t.Fatalf("IsQueued = %t, want %t", lifecycle.IsQueued(), tt.wantQueued)
			}
			if lifecycle.IsOwnedGeneration() != tt.wantOwned {
				t.Fatalf("IsOwnedGeneration = %t, want %t", lifecycle.IsOwnedGeneration(), tt.wantOwned)
			}
			if lifecycle.IsRecoverableStaleRun() != tt.wantRecoverable {
				t.Fatalf("IsRecoverableStaleRun = %t, want %t", lifecycle.IsRecoverableStaleRun(), tt.wantRecoverable)
			}
			if lifecycle.IsTerminal() != tt.wantTerminal {
				t.Fatalf("IsTerminal = %t, want %t", lifecycle.IsTerminal(), tt.wantTerminal)
			}
			if lifecycle.ShouldRecover() != tt.wantRecoverable {
				t.Fatalf("ShouldRecover = %t, want %t", lifecycle.ShouldRecover(), tt.wantRecoverable)
			}
			if lifecycle.CanClaimGeneration() != tt.wantCanClaim {
				t.Fatalf("CanClaimGeneration = %t, want %t", lifecycle.CanClaimGeneration(), tt.wantCanClaim)
			}
			if lifecycle.GenerationAttemptsExhausted() != tt.wantAttemptsSpent {
				t.Fatalf("GenerationAttemptsExhausted = %t, want %t", lifecycle.GenerationAttemptsExhausted(), tt.wantAttemptsSpent)
			}
		})
	}
}

func TestShouldRecoverRunUsesLifecycle(t *testing.T) {
	now := time.Date(2026, 6, 29, 15, 0, 0, 0, time.UTC)
	expiredLease := now.Add(-time.Second)
	validLease := now.Add(time.Second)

	if !shouldRecoverRun(&ChatRun{
		Status:           statusStreaming,
		GenerationStatus: generationStatusGenerating,
		LeaseExpiresAt:   &expiredLease,
	}, now) {
		t.Fatal("expired generating run should be recoverable")
	}
	if shouldRecoverRun(&ChatRun{
		Status:           statusStreaming,
		GenerationStatus: generationStatusGenerating,
		LeaseExpiresAt:   &validLease,
	}, now) {
		t.Fatal("fresh generating run should not be recoverable")
	}
}
