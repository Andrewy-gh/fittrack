package workout

import (
	"context"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAssessmentHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, id, query string
		authed, missing bool
		status          int
	}{
		{"success", "1", "?startDate=2026-09-14&timezone=UTC", true, false, 200},
		{"unauthorized", "1", "?startDate=2026-09-14&timezone=UTC", false, false, 401},
		{"invalid id", "0", "?startDate=2026-09-14&timezone=UTC", true, false, 400},
		{"missing timezone", "1", "?startDate=2026-09-14", true, false, 400},
		{"future", "1", "?startDate=2027-01-01&timezone=UTC", true, false, 400},
		{"foreign exercise", "1", "?startDate=2026-09-14&timezone=UTC", true, true, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &assessmentStub{}
			if tc.missing {
				stub.lookupErr = pgx.ErrNoRows
			}
			handler := NewAssessmentHandler(NewAssessmentService(stub, func() time.Time { return time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC) }), slog.New(slog.NewTextHandler(io.Discard, nil)))
			req := httptest.NewRequest(http.MethodGet, "/api/exercises/"+tc.id+"/assessment"+tc.query, nil)
			req.SetPathValue("id", tc.id)
			if tc.authed {
				req = req.WithContext(user.WithContext(context.Background(), "owner"))
			}
			recorder := httptest.NewRecorder()
			handler.Get(recorder, req)
			require.Equal(t, tc.status, recorder.Code)
			if tc.status == 200 {
				require.Contains(t, recorder.Body.String(), `"sessions":[]`)
				require.Contains(t, recorder.Body.String(), TrainingAssessmentPolicyVersion)
			}
		})
	}
}
