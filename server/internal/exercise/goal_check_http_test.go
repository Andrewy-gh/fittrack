package exercise

import (
	"context"
	"encoding/json"
	"errors"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"
)

type goalCheckQueryFunc func(context.Context, db.GetExerciseGoalCheckParams) ([]db.GetExerciseGoalCheckRow, error)

func (f goalCheckQueryFunc) GetExerciseGoalCheck(c context.Context, p db.GetExerciseGoalCheckParams) ([]db.GetExerciseGoalCheckRow, error) {
	return f(c, p)
}
func TestGoalCheckHTTP(t *testing.T) {
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name, id, timezone, owner string
		rows                      []db.GetExerciseGoalCheckRow
		err                       error
		status                    int
	}{
		{name: "unauthenticated", id: "1", status: 401},
		{name: "invalid id", id: "0", owner: "a", status: 400},
		{name: "overflow id", id: "99999999999999", owner: "a", status: 400},
		{name: "invalid timezone", id: "1", owner: "a", timezone: "Invalid/Zone", status: 400},
		{name: "local timezone", id: "1", owner: "a", timezone: "Local", status: 400},
		{name: "not owned", id: "1", owner: "a", status: 404},
		{name: "database failure", id: "1", owner: "a", status: 500, err: errors.New("database failed")},
		{name: "empty legacy strength", id: "1", owner: "a", status: 200, rows: []db.GetExerciseGoalCheckRow{{PrimaryGoal: pgtype.Text{String: "strength", Valid: true}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := NewGoalCheckHandler(goalCheckQueryFunc(func(_ context.Context, p db.GetExerciseGoalCheckParams) ([]db.GetExerciseGoalCheckRow, error) {
				require.Equal(t, tc.owner, p.UserID)
				require.Equal(t, now, p.AsOf.Time)
				return tc.rows, tc.err
			}), slog.New(slog.NewTextHandler(io.Discard, nil)), func() time.Time { return now })
			r := httptest.NewRequest("GET", "/api/exercises/"+tc.id+"/goal-check?timezone="+tc.timezone, nil)
			r.SetPathValue("id", tc.id)
			if tc.owner != "" {
				r = r.WithContext(user.WithContext(r.Context(), tc.owner))
			}
			w := httptest.NewRecorder()
			h.Get(w, r)
			require.Equal(t, tc.status, w.Code)
			if tc.status == 200 {
				var result StrengthGoalCheck
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
				require.True(t, result.Applicable)
				require.NotNil(t, result.Sessions)
				require.Nil(t, result.WorkingSets.Average)
			}
		})
	}
}
