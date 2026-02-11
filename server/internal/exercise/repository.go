package exercise

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type exerciseRepository struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgxpool.Pool
}

// NewRepository creates a new instance of ExerciseRepository
func NewRepository(logger *slog.Logger, queries *db.Queries, conn *pgxpool.Pool) ExerciseRepository {
	return &exerciseRepository{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}
}

func (er *exerciseRepository) ListExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	exerciseRows, err := er.queries.ListExercises(ctx, userID)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("list exercises query failed - RLS policy violation",
				"error", err,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("list exercises query failed", "error", err, "user_id", userID)
		}
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}

	if exerciseRows == nil {
		return []db.Exercise{}, nil
	}

	// Convert ListExercisesRow to Exercise
	exercises := make([]db.Exercise, len(exerciseRows))
	for i, row := range exerciseRows {
		exercises[i] = db.Exercise{
			ID:   row.ID,
			Name: row.Name,
			// UserID is not returned by optimized query but was used for filtering
			UserID: userID,
		}
	}

	// Log empty results that might indicate RLS filtering
	if len(exercises) == 0 {
		er.logger.Debug("list exercises returned empty results",
			"user_id", userID,
			"potential_rls_filtering", true)
	}

	return exercises, nil
}

func (er *exerciseRepository) GetExercise(ctx context.Context, id int32, userID string) (db.Exercise, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := db.GetExerciseParams{
		ID:     id,
		UserID: userID,
	}
	exerciseRow, err := er.queries.GetExercise(ctx, params)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("get exercise query failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("get exercise query failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return db.Exercise{}, fmt.Errorf("failed to get exercise (id: %d): %w", id, err)
	}

	// Convert GetExerciseRow to Exercise
	exercise := db.Exercise{
		ID:   exerciseRow.ID,
		Name: exerciseRow.Name,
		// UserID is not returned by optimized query but was used for filtering
		UserID: userID,
	}

	return exercise, nil
}

func (er *exerciseRepository) GetExerciseDetail(ctx context.Context, id int32, userID string) (db.GetExerciseDetailRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := db.GetExerciseDetailParams{
		ID:     id,
		UserID: userID,
	}
	row, err := er.queries.GetExerciseDetail(ctx, params)
	if err != nil {
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("get exercise detail query failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("get exercise detail query failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return db.GetExerciseDetailRow{}, fmt.Errorf("failed to get exercise detail (id: %d): %w", id, err)
	}

	return row, nil
}

func (er *exerciseRepository) GetOrCreateExercise(ctx context.Context, name string, userID string) (db.Exercise, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := db.GetOrCreateExerciseParams{
		Name:   name,
		UserID: userID,
	}
	exerciseID, err := er.queries.GetOrCreateExercise(ctx, params)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("get or create exercise query failed - RLS policy violation",
				"error", err,
				"exercise_name", name,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("get or create exercise query failed",
				"exercise_name", name,
				"user_id", userID,
				"error", err)
		}
		return db.Exercise{}, fmt.Errorf("failed to get or create exercise: %w", err)
	}

	// Create Exercise from returned ID
	exercise := db.Exercise{
		ID:     exerciseID,
		Name:   name,
		UserID: userID,
	}

	return exercise, nil
}

func (er *exerciseRepository) GetExerciseWithSets(ctx context.Context, id int32, userID string) ([]db.GetExerciseWithSetsRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := db.GetExerciseWithSetsParams{
		ExerciseID: id,
		UserID:     userID,
	}
	sets, err := er.queries.GetExerciseWithSets(ctx, params)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("get exercise with sets query failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("get exercise with sets query failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return nil, fmt.Errorf("failed to get exercise with sets (id: %d): %w", id, err)
	}

	if sets == nil {
		sets = []db.GetExerciseWithSetsRow{}
	}

	// Log empty results that might indicate RLS filtering
	if len(sets) == 0 {
		er.logger.Debug("get exercise with sets returned empty results",
			"exercise_id", id,
			"user_id", userID,
			"potential_rls_filtering", true)
	}

	return sets, nil
}

func (er *exerciseRepository) GetOrCreateExerciseTx(ctx context.Context, qtx *db.Queries, name, userID string) (db.Exercise, error) {
	er.logger.Info("GetOrCreateExerciseTx called", "exercise_name", name, "user_id", userID)

	params := db.GetOrCreateExerciseParams{
		Name:   name,
		UserID: userID,
	}

	er.logger.Info("calling GetOrCreateExercise with params", "params", params)
	exerciseID, err := qtx.GetOrCreateExercise(ctx, params)
	if err != nil {
		er.logger.Error("failed to get or create exercise",
			"error", err,
			"exercise_name", name,
			"user_id", userID)
		return db.Exercise{}, fmt.Errorf("failed to get or create exercise: %w", err)
	}

	// Create Exercise from returned ID
	exercise := db.Exercise{
		ID:     exerciseID,
		Name:   name,
		UserID: userID,
	}

	er.logger.Info("successfully got/created exercise",
		"exercise_id", exercise.ID,
		"exercise_name", exercise.Name,
		"user_id", exercise.UserID)

	return exercise, nil
}

func (er *exerciseRepository) GetRecentSetsForExercise(ctx context.Context, id int32, userID string) ([]db.GetRecentSetsForExerciseRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := db.GetRecentSetsForExerciseParams{
		ExerciseID: id,
		UserID:     userID,
	}
	sets, err := er.queries.GetRecentSetsForExercise(ctx, params)
	if err != nil {
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("get recent sets for exercise query failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("get recent sets for exercise query failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return nil, fmt.Errorf("failed to get recent sets for exercise (id: %d): %w", id, err)
	}

	if sets == nil {
		sets = []db.GetRecentSetsForExerciseRow{}
	}

	if len(sets) == 0 {
		er.logger.Debug("get recent sets for exercise returned empty results",
			"exercise_id", id,
			"user_id", userID,
			"potential_rls_filtering", true)
	}

	return sets, nil
}

func (er *exerciseRepository) GetExerciseMetricsHistory(ctx context.Context, req GetExerciseMetricsHistoryRequest, userID string) ([]ExerciseMetricsHistoryPoint, MetricsHistoryBucket, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	switch req.Range {
	case "W":
		{
			rows, err := er.queries.GetExerciseMetricsHistoryRawWeek(ctx, db.GetExerciseMetricsHistoryRawWeekParams{
				ExerciseID: req.ExerciseID,
				UserID:     userID,
			})
			if err != nil {
				return nil, "", fmt.Errorf("get exercise metrics history raw week query failed: %w", err)
			}

			points := make([]ExerciseMetricsHistoryPoint, 0, len(rows))
			for _, row := range rows {
				workoutID := row.WorkoutID
				if !row.WorkoutDay.Valid {
					continue
				}
				points = append(points, ExerciseMetricsHistoryPoint{
					X:                    fmt.Sprintf("%d", workoutID),
					Date:                 row.WorkoutDay.Time,
					WorkoutID:            &workoutID,
					SessionBestE1RM:      row.SessionBestE1rm,
					SessionAvgE1RM:       row.SessionAvgE1rm,
					SessionAvgIntensity:  row.SessionAvgIntensity,
					SessionBestIntensity: row.SessionBestIntensity,
					TotalVolumeWorking:   row.TotalVolumeWorking,
				})
			}
			return points, MetricsHistoryBucketWorkout, nil
		}
	case "M":
		{
			rows, err := er.queries.GetExerciseMetricsHistoryRawMonth(ctx, db.GetExerciseMetricsHistoryRawMonthParams{
				ExerciseID: req.ExerciseID,
				UserID:     userID,
			})
			if err != nil {
				return nil, "", fmt.Errorf("get exercise metrics history raw month query failed: %w", err)
			}

			points := make([]ExerciseMetricsHistoryPoint, 0, len(rows))
			for _, row := range rows {
				workoutID := row.WorkoutID
				if !row.WorkoutDay.Valid {
					continue
				}
				points = append(points, ExerciseMetricsHistoryPoint{
					X:                    fmt.Sprintf("%d", workoutID),
					Date:                 row.WorkoutDay.Time,
					WorkoutID:            &workoutID,
					SessionBestE1RM:      row.SessionBestE1rm,
					SessionAvgE1RM:       row.SessionAvgE1rm,
					SessionAvgIntensity:  row.SessionAvgIntensity,
					SessionBestIntensity: row.SessionBestIntensity,
					TotalVolumeWorking:   row.TotalVolumeWorking,
				})
			}
			return points, MetricsHistoryBucketWorkout, nil
		}
	case "6M":
		{
			rows, err := er.queries.GetExerciseMetricsHistoryWeekly6M(ctx, db.GetExerciseMetricsHistoryWeekly6MParams{
				ExerciseID: req.ExerciseID,
				UserID:     userID,
			})
			if err != nil {
				return nil, "", fmt.Errorf("get exercise metrics history weekly 6M query failed: %w", err)
			}

			points := make([]ExerciseMetricsHistoryPoint, 0, len(rows))
			for _, row := range rows {
				if !row.BucketDay.Valid {
					continue
				}
				day := row.BucketDay.Time
				points = append(points, ExerciseMetricsHistoryPoint{
					X:                    day.Format("2006-01-02"),
					Date:                 day,
					SessionBestE1RM:      row.SessionBestE1rm,
					SessionAvgE1RM:       row.SessionAvgE1rm,
					SessionAvgIntensity:  row.SessionAvgIntensity,
					SessionBestIntensity: row.SessionBestIntensity,
					TotalVolumeWorking:   row.TotalVolumeWorking,
				})
			}
			return points, MetricsHistoryBucketWeek, nil
		}
	case "Y":
		{
			rows, err := er.queries.GetExerciseMetricsHistoryMonthlyYear(ctx, db.GetExerciseMetricsHistoryMonthlyYearParams{
				ExerciseID: req.ExerciseID,
				UserID:     userID,
			})
			if err != nil {
				return nil, "", fmt.Errorf("get exercise metrics history monthly year query failed: %w", err)
			}

			points := make([]ExerciseMetricsHistoryPoint, 0, len(rows))
			for _, row := range rows {
				if !row.BucketDay.Valid {
					continue
				}
				day := row.BucketDay.Time
				points = append(points, ExerciseMetricsHistoryPoint{
					X:                    day.Format("2006-01-02"),
					Date:                 day,
					SessionBestE1RM:      row.SessionBestE1rm,
					SessionAvgE1RM:       row.SessionAvgE1rm,
					SessionAvgIntensity:  row.SessionAvgIntensity,
					SessionBestIntensity: row.SessionBestIntensity,
					TotalVolumeWorking:   row.TotalVolumeWorking,
				})
			}
			return points, MetricsHistoryBucketMonth, nil
		}
	default:
		return nil, "", fmt.Errorf("invalid range: %q", req.Range)
	}
}

func (er *exerciseRepository) UpdateExerciseName(ctx context.Context, id int32, name, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := er.queries.UpdateExerciseName(ctx, db.UpdateExerciseNameParams{
		ID:     id,
		Name:   name,
		UserID: userID,
	}); err != nil {
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("update exercise name failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("update exercise name failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return fmt.Errorf("failed to update exercise name (id: %d): %w", id, err)
	}

	er.logger.Info("exercise name updated successfully",
		"exercise_id", id,
		"user_id", userID)

	return nil
}

func numericFromFloat(val *float64) (pgtype.Numeric, error) {
	if val == nil {
		return pgtype.Numeric{Valid: false}, nil
	}

	var n pgtype.Numeric
	// Keep 2dp to match NUMERIC(8,2) storage.
	if err := n.Scan(fmt.Sprintf("%.2f", *val)); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("failed to convert float to numeric: %w", err)
	}
	return n, nil
}

func (er *exerciseRepository) GetExerciseHistorical1RM(ctx context.Context, id int32, userID string) (db.GetExerciseHistorical1RMRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return er.queries.GetExerciseHistorical1RM(ctx, db.GetExerciseHistorical1RMParams{
		ID:     id,
		UserID: userID,
	})
}

func (er *exerciseRepository) GetExerciseBestE1rmWithWorkout(ctx context.Context, exerciseID int32, userID string) (db.GetExerciseBestE1rmWithWorkoutRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return er.queries.GetExerciseBestE1rmWithWorkout(ctx, db.GetExerciseBestE1rmWithWorkoutParams{
		UserID:     userID,
		ExerciseID: exerciseID,
	})
}

func (er *exerciseRepository) UpdateExerciseHistorical1RMManual(ctx context.Context, id int32, historical1rm *float64, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	n, err := numericFromFloat(historical1rm)
	if err != nil {
		return err
	}

	if err := er.queries.UpdateExerciseHistorical1RMManual(ctx, db.UpdateExerciseHistorical1RMManualParams{
		ID:            id,
		Historical1rm: n,
		UserID:        userID,
	}); err != nil {
		return fmt.Errorf("failed to update exercise historical 1rm manual (id: %d): %w", id, err)
	}

	return nil
}

func (er *exerciseRepository) SetExerciseHistorical1RM(ctx context.Context, id int32, historical1rm *float64, sourceWorkoutID *int32, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	n, err := numericFromFloat(historical1rm)
	if err != nil {
		return err
	}

	src := pgtype.Int4{Valid: false}
	if sourceWorkoutID != nil {
		src = pgtype.Int4{Int32: *sourceWorkoutID, Valid: true}
	}

	if err := er.queries.SetExerciseHistorical1RM(ctx, db.SetExerciseHistorical1RMParams{
		ID:                           id,
		Historical1rm:                n,
		Historical1rmSourceWorkoutID: src,
		UserID:                       userID,
	}); err != nil {
		// Keep the RLS log pattern consistent with other repository methods.
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("set exercise historical 1rm failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("set exercise historical 1rm failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return fmt.Errorf("failed to set exercise historical 1rm (id: %d): %w", id, err)
	}

	return nil
}

func (er *exerciseRepository) DeleteExercise(ctx context.Context, id int32, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := er.queries.DeleteExercise(ctx, db.DeleteExerciseParams{
		ID:     id,
		UserID: userID,
	}); err != nil {
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("delete exercise failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("delete exercise failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return fmt.Errorf("failed to delete exercise (id: %d): %w", id, err)
	}

	er.logger.Info("exercise deleted successfully",
		"exercise_id", id,
		"user_id", userID)

	return nil
}

var _ ExerciseRepository = (*exerciseRepository)(nil)
