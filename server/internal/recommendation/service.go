package recommendation

import (
	"context"
	"errors"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

var ErrUnauthorized = errors.New("authentication required")
var ErrNotFound = errors.New("exercise or workout not found")
var ErrInvalid = errors.New("working-set baseline must be an ordered range from 1 to 20")

type Repository interface {
	Load(context.Context, string, int32, time.Time) (*Range, *Session, error)
	SavePrescription(context.Context, string, int32, *Range) error
	Saved(context.Context, string, int32) ([]Snapshot, error)
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
	now := s.now()
	baseline, previous, err := s.repo.Load(ctx, owner, exerciseID, now)
	if err != nil {
		return Result{}, err
	}
	result := Recommend(now, baseline, previous, readiness)
	result.ExerciseID = exerciseID
	return result, nil
}

func (s *Service) Prescribe(ctx context.Context, exerciseID int32, baseline *Range) error {
	owner, ok := user.Current(ctx)
	if !ok || owner == "" {
		return ErrUnauthorized
	}
	if baseline != nil && !baseline.Valid() {
		return ErrInvalid
	}
	return s.repo.SavePrescription(ctx, owner, exerciseID, baseline)
}

func (s *Service) Saved(ctx context.Context, workoutID int32) ([]Snapshot, error) {
	owner, ok := user.Current(ctx)
	if !ok || owner == "" {
		return nil, ErrUnauthorized
	}
	return s.repo.Saved(ctx, owner, workoutID)
}
