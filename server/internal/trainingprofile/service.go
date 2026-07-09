package trainingprofile

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

type Service struct {
	logger *slog.Logger
	repo   Repository
}

func NewService(logger *slog.Logger, repo Repository) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

func (s *Service) Get(ctx context.Context) (*ProfileResponse, error) {
	userID, ok := user.Current(ctx)
	if !ok || userID == "" {
		return nil, &apperrors.Unauthorized{Resource: "training profile", UserID: ""}
	}

	profile, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get training profile: %w", err)
	}
	return profile, nil
}

func (s *Service) Upsert(ctx context.Context, req UpdateProfileRequest) (*ProfileResponse, error) {
	userID, ok := user.Current(ctx)
	if !ok || userID == "" {
		return nil, &apperrors.Unauthorized{Resource: "training profile", UserID: ""}
	}

	normalized, err := validateProfileRequest(req)
	if err != nil {
		return nil, err
	}

	profile, err := s.repo.Upsert(ctx, userID, *normalized)
	if err != nil {
		return nil, fmt.Errorf("failed to update training profile: %w", err)
	}
	return profile, nil
}
