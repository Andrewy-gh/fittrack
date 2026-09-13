package recommendation

import (
	"context"
	"errors"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

var ErrUnauthorized = errors.New("authentication required")
var ErrNotFound = errors.New("exercise not found")
var ErrInvalid = errors.New("readiness must be normal, great, or sluggish")

type Repository interface {
	Load(context.Context, string, int32, time.Time) (TrainingContext, error)
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

func (s *Service) Get(ctx context.Context, exerciseID int32, readiness string) (Result, error) {
	owner, ok := user.Current(ctx)
	if !ok || owner == "" {
		return Result{}, ErrUnauthorized
	}
	if !validReadiness(readiness) {
		return Result{}, ErrInvalid
	}
	now := s.now()
	trainingContext, err := s.repo.Load(ctx, owner, exerciseID, now)
	if err != nil {
		return Result{}, err
	}
	result := RecommendTraining(now, trainingContext, readiness)
	result.ExerciseID = exerciseID
	return result, nil
}
