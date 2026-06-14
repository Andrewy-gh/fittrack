package aichat

import (
	"context"
	"log/slog"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
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
