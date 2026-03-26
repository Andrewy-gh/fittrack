package aichat

import (
	"context"
	"errors"
)

type featureAccessService interface {
	HasCurrentUserFeatureAccess(ctx context.Context, featureKey string) (bool, error)
}

type runtime interface {
	ModelName() string
	Available() bool
	GenerateValidation(ctx context.Context, prompt string) (*ValidationOutput, error)
	StreamValidation(ctx context.Context, prompt string, onChunk func(string) error) (*StreamDone, error)
}

type Service struct {
	featureAccess featureAccessService
	runtime       runtime
}

func NewService(featureAccess featureAccessService, runtime runtime) *Service {
	return &Service{
		featureAccess: featureAccess,
		runtime:       runtime,
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

func (s *Service) ensureAllowed(ctx context.Context) error {
	if s.runtime == nil || !s.runtime.Available() {
		return ErrRuntimeUnavailable
	}

	hasAccess, err := s.featureAccess.HasCurrentUserFeatureAccess(ctx, featureKeyAIChatbot)
	if err != nil {
		return err
	}
	if !hasAccess {
		return ErrFeatureDisabled
	}

	return nil
}

func isClientSafeError(err error) bool {
	return errors.Is(err, ErrFeatureDisabled) || errors.Is(err, ErrRuntimeUnavailable)
}
