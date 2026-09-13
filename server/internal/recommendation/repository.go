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

func (r *repository) Load(ctx context.Context, owner string, id int32, now time.Time) (TrainingContext, error) {
	var result TrainingContext
	exercise, err := r.queries.GetExercise(ctx, db.GetExerciseParams{ID: id, UserID: owner})
	if err != nil {
		return result, lookupError(err)
	}
	profile, err := r.queries.GetUserTrainingProfile(ctx, owner)
	if err == nil {
		result.Goal, result.Experience = profile.PrimaryGoal.String, profile.ExperienceLevel.String
		if result.Goal == "" && len(profile.Goals) > 0 {
			result.Goal = profile.Goals[0]
		}
		var avoided, limitations []string
		if len(profile.AvoidedExercises) > 0 {
			if err := json.Unmarshal(profile.AvoidedExercises, &avoided); err != nil {
				return result, err
			}
		}
		if len(profile.MovementLimitations) > 0 {
			if err := json.Unmarshal(profile.MovementLimitations, &limitations); err != nil {
				return result, err
			}
		}
		result.Avoided = matchesExercise(avoided, exercise.Name)
		result.HasLimitations = len(limitations) > 0
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}

	history, err := r.queries.GetRecommendationTrainingHistory(ctx, db.GetRecommendationTrainingHistoryParams{
		ExerciseID: id,
		UserID:     owner,
		AsOf:       pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return result, err
	}
	var currentWorkoutID int32
	for _, row := range history {
		if len(result.Sessions) == 0 || row.WorkoutID != currentWorkoutID {
			if row.Date.Time.After(now) || now.Sub(row.Date.Time) > Recency {
				break
			}
			result.Sessions = append(result.Sessions, Session{Date: row.Date.Time.UTC()})
			currentWorkoutID = row.WorkoutID
		}
		if !row.WorkingSetID.Valid {
			continue
		}
		var weight *float64
		value, err := row.Weight.Float64Value()
		if err != nil {
			return result, err
		}
		if value.Valid {
			weight = &value.Float64
		}
		latest := &result.Sessions[len(result.Sessions)-1]
		latest.Sets = append(latest.Sets, WorkingSet{Weight: weight, Reps: int(row.Reps.Int32)})
	}
	return result, nil
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
	}
	raw, err := json.Marshal(snapshots)
	if err != nil {
		return err
	}
	return qtx.SaveWorkoutRecommendationContext(ctx, db.SaveWorkoutRecommendationContextParams{WorkoutID: workoutID, UserID: owner, RecommendationContext: string(raw)})
}
