package exercise

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	_ "time/tzdata"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
)

type goalCheckQueries interface {
	GetExerciseGoalCheck(context.Context, db.GetExerciseGoalCheckParams) ([]db.GetExerciseGoalCheckRow, error)
}
type GoalCheckHandler struct {
	queries goalCheckQueries
	logger  *slog.Logger
	now     func() time.Time
}

func NewGoalCheckHandler(queries goalCheckQueries, logger *slog.Logger, now func() time.Time) *GoalCheckHandler {
	return &GoalCheckHandler{queries: queries, logger: logger, now: now}
}

// Get godoc
// @Summary Compare logged exercise training with general strength references
// @Tags exercises
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Param timezone query string false "IANA timezone (default UTC)"
// @Success 200 {object} exercise.StrengthGoalCheck
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /exercises/{id}/goal-check [get]
func (h *GoalCheckHandler) Get(w http.ResponseWriter, r *http.Request) {
	owner, ok := user.Current(r.Context())
	if !ok {
		response.ErrorJSON(w, r, h.logger, 401, "Unauthorized", nil)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		response.ErrorJSON(w, r, h.logger, 400, "Invalid exercise ID", nil)
		return
	}
	timezone := r.URL.Query().Get("timezone")
	if timezone == "" {
		timezone = "UTC"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil || timezone == "Local" {
		response.ErrorJSON(w, r, h.logger, 400, "Invalid IANA timezone", nil)
		return
	}
	window := goalCheckWindow(h.now(), location)
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	rows, err := h.queries.GetExerciseGoalCheck(ctx, db.GetExerciseGoalCheckParams{
		ExerciseID: int32(id), UserID: owner,
		WindowStart: pgtype.Timestamptz{Time: window.Start, Valid: true}, AsOf: pgtype.Timestamptz{Time: window.End, Valid: true},
	})
	if err != nil {
		response.ErrorJSON(w, r, h.logger, 500, "Could not load Goal Check", err)
		return
	}
	if len(rows) == 0 {
		response.ErrorJSON(w, r, h.logger, 404, "Exercise not found", nil)
		return
	}
	first := rows[0]
	goals := first.Goals
	if len(goals) == 0 && first.PrimaryGoal.Valid {
		goals = []string{first.PrimaryGoal.String}
	}
	var reference *GoalCheckReference
	value, err := floatPtrFromNumeric(first.Historical1rm)
	if err != nil {
		response.ErrorJSON(w, r, h.logger, 500, "Could not read saved reference", err)
		return
	}
	if value != nil && positiveFinite(*value) {
		reference = &GoalCheckReference{Value: *value, Origin: "manual_unverified"}
		if first.Historical1rmUpdatedAt.Valid {
			reference.UpdatedAt = &first.Historical1rmUpdatedAt.Time
		}
		if first.Historical1rmSourceWorkoutID.Valid {
			reference.Origin = "workout_estimate"
			reference.SourceWorkoutID = &first.Historical1rmSourceWorkoutID.Int32
		}
	}
	sets := make([]goalCheckSet, 0, len(rows))
	for _, row := range rows {
		if !row.SetID.Valid || !row.WorkoutID.Valid || !row.Date.Valid || !row.Reps.Valid {
			continue
		}
		weight, err := floatPtrFromNumeric(row.Weight)
		if err != nil {
			response.ErrorJSON(w, r, h.logger, 500, "Could not read logged load", err)
			return
		}
		sets = append(sets, goalCheckSet{ID: row.SetID.Int32, WorkoutID: row.WorkoutID.Int32, Date: row.Date.Time, Weight: weight, Reps: int(row.Reps.Int32)})
	}
	result := calculateGoalCheck(int32(id), goals, window, location, sets, reference)
	if err := response.JSON(w, 200, result); err != nil {
		h.logger.Error("write goal check", "error", err)
	}
}
