package workout

import (
	"context"
	"fmt"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

// MARK: ListWorkoutFocusValues
func (wr *workoutRepository) ListWorkoutFocusValues(ctx context.Context, userID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Execute the query to get distinct workout focus values
	rows, err := wr.queries.ListWorkoutFocusValues(ctx, userID)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			wr.logger.Error("list workout focus values query failed - RLS policy violation",
				"error", err,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			wr.logger.Error("list workout focus values query failed", "error", err, "user_id", userID)
		}
		return nil, fmt.Errorf("failed to list workout focus values: %w", err)
	}

	// Convert []pgtype.Text to []string
	var focusValues []string
	for _, row := range rows {
		if row.Valid {
			focusValues = append(focusValues, row.String)
		}
	}

	// Ensure we always return an empty slice, not nil
	if focusValues == nil {
		focusValues = []string{}
	}

	// Log empty results that might indicate RLS filtering
	if len(focusValues) == 0 {
		wr.logger.Debug("list workout focus values returned empty results",
			"user_id", userID,
			"potential_rls_filtering", true)
	}

	return focusValues, nil
}

// MARK: GetContributionData
func (wr *workoutRepository) GetContributionData(ctx context.Context, userID string) ([]db.GetContributionDataRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := wr.queries.GetContributionData(ctx, userID)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			wr.logger.Error("get contribution data query failed - RLS policy violation",
				"error", err,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			wr.logger.Error("get contribution data query failed", "error", err, "user_id", userID)
		}
		return nil, fmt.Errorf("failed to get contribution data: %w", err)
	}

	if rows == nil {
		rows = []db.GetContributionDataRow{}
	}

	// Log empty results that might indicate RLS filtering
	if len(rows) == 0 {
		wr.logger.Debug("get contribution data returned empty results",
			"user_id", userID,
			"potential_rls_filtering", true)
	}

	return rows, nil
}
