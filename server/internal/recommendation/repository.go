package recommendation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type repository struct{ queries *db.Queries }

func NewRepository(queries *db.Queries) Repository { return &repository{queries: queries} }

func (r *repository) Load(ctx context.Context, owner string, id int32, now time.Time) (*Range, *Session, error) {
	if _, err := r.queries.GetExercise(ctx, db.GetExerciseParams{ID: id, UserID: owner}); err != nil {
		return nil, nil, lookupError(err)
	}
	var baseline *Range
	row, err := r.queries.GetExercisePrescription(ctx, db.GetExercisePrescriptionParams{ExerciseID: id, UserID: owner})
	if err == nil {
		baseline = &Range{Min: int(row.MinSets), Max: int(row.MaxSets)}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, err
	}
	history, err := r.queries.GetRecommendationHistory(ctx, db.GetRecommendationHistoryParams{ExerciseID: id, UserID: owner, AsOf: pgtype.Timestamptz{Time: now, Valid: true}})
	if err != nil {
		return nil, nil, err
	}
	var previous *Session
	if len(history) > 0 {
		previous = &Session{WorkoutID: history[0].WorkoutID, Date: history[0].Date.Time.UTC(), WorkingSets: int(history[0].WorkingSets)}
	}
	return baseline, previous, nil
}

func (r *repository) SavePrescription(ctx context.Context, owner string, id int32, baseline *Range) error {
	if _, err := r.queries.GetExercise(ctx, db.GetExerciseParams{ID: id, UserID: owner}); err != nil {
		return lookupError(err)
	}
	if baseline == nil {
		return r.queries.DeleteExercisePrescription(ctx, db.DeleteExercisePrescriptionParams{ExerciseID: id, UserID: owner})
	}
	return r.queries.UpsertExercisePrescription(ctx, db.UpsertExercisePrescriptionParams{ExerciseID: id, UserID: owner, MinSets: int32(baseline.Min), MaxSets: int32(baseline.Max)})
}

func (r *repository) Saved(ctx context.Context, owner string, workoutID int32) ([]Snapshot, error) {
	raw, err := r.queries.GetWorkoutRecommendationContext(ctx, db.GetWorkoutRecommendationContextParams{ID: workoutID, UserID: owner})
	if err != nil {
		return nil, lookupError(err)
	}
	snapshots := []Snapshot{}
	if err := json.Unmarshal(raw, &snapshots); err != nil {
		return nil, fmt.Errorf("read recommendation context: %w", err)
	}
	return snapshots, nil
}

func lookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// SaveContext records the displayed recommendation atomically with actual sets, without recalculating it.
func SaveContext(ctx context.Context, qtx *db.Queries, owner string, workoutID int32, exerciseIDs map[string]int32, snapshots []Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	for _, snapshot := range snapshots {
		if err := ValidateSnapshot(snapshot); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidContext, err)
		}
		if seen[snapshot.ExerciseName] || exerciseIDs[snapshot.ExerciseName] != snapshot.Recommendation.ExerciseID {
			return fmt.Errorf("%w: recommendation does not match logged exercise", ErrInvalidContext)
		}
		seen[snapshot.ExerciseName] = true
		if previous := snapshot.Recommendation.Previous; previous != nil {
			if _, err := qtx.GetWorkout(ctx, db.GetWorkoutParams{ID: previous.WorkoutID, UserID: owner}); err != nil {
				return fmt.Errorf("%w: invalid history reference", ErrInvalidContext)
			}
		}
	}
	raw, err := json.Marshal(snapshots)
	if err != nil {
		return err
	}
	return qtx.SaveWorkoutRecommendationContext(ctx, db.SaveWorkoutRecommendationContextParams{WorkoutID: workoutID, UserID: owner, RecommendationContext: string(raw)})
}
