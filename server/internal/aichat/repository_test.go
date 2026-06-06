package aichat

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

func TestBuildConversationTitle_TruncatesWithoutBreakingUTF8(t *testing.T) {
	prompt := strings.Repeat("🙂", 100)

	title := buildConversationTitle(prompt)

	if !utf8.ValidString(title) {
		t.Fatalf("buildConversationTitle returned invalid UTF-8: %q", title)
	}
	if got := utf8.RuneCountInString(title); got != 80 {
		t.Fatalf("buildConversationTitle rune count = %d, want 80", got)
	}
	if !strings.HasSuffix(title, "...") {
		t.Fatalf("buildConversationTitle should end with ellipsis, got %q", title)
	}
}

func TestTruncateForStorage_TruncatesWithoutBreakingUTF8(t *testing.T) {
	value := strings.Repeat("🙂", 600)

	truncated := truncateForStorage(value, 512)

	if !utf8.ValidString(truncated) {
		t.Fatalf("truncateForStorage returned invalid UTF-8: %q", truncated)
	}
	if got := utf8.RuneCountInString(truncated); got != 512 {
		t.Fatalf("truncateForStorage rune count = %d, want 512", got)
	}
	if !strings.HasSuffix(truncated, "...") {
		t.Fatalf("truncateForStorage should end with ellipsis, got %q", truncated)
	}
}

func TestTextPtrToPg_PreservesNilAsNull(t *testing.T) {
	nullText := textPtrToPg(nil)
	if nullText.Valid {
		t.Fatal("nil pointer should map to SQL NULL")
	}

	owner := "api:worker-1"
	text := textPtrToPg(&owner)
	if !text.Valid || text.String != owner {
		t.Fatalf("owner pointer should map to valid text, got valid=%t value=%q", text.Valid, text.String)
	}
}

func TestNewInngestRunOwner_UniquePerRecoveryClaim(t *testing.T) {
	first := newInngestRunOwner(51).Value()
	second := newInngestRunOwner(51).Value()

	if first == second {
		t.Fatalf("recovery owners should be unique per claim, got %q", first)
	}
	if !strings.HasPrefix(first, "inngest:run-51-") {
		t.Fatalf("recovery owner should include run id for traceability, got %q", first)
	}
	if !strings.HasPrefix(second, "inngest:run-51-") {
		t.Fatalf("recovery owner should include run id for traceability, got %q", second)
	}
}

func TestIsStreamingRunStale(t *testing.T) {
	now := time.Date(2026, 3, 26, 18, 30, 0, 0, time.UTC)

	if isStreamingRunStale(now.Add(-streamingRunStaleAfter+time.Second), now) {
		t.Fatal("run newer than stale threshold should not be stale")
	}
	if !isStreamingRunStale(now.Add(-streamingRunStaleAfter-time.Second), now) {
		t.Fatal("run older than stale threshold should be stale")
	}
}

func TestMarshalWorkoutDraft_ReturnsJSON(t *testing.T) {
	workoutFocus := "full body"
	weight := 50.0
	draft := &workout.CreateWorkoutRequest{
		Date:         "2026-04-22T12:00:00Z",
		WorkoutFocus: &workoutFocus,
		Exercises: []workout.ExerciseInput{
			{
				Name: "Goblet Squat",
				Sets: []workout.SetInput{
					{Weight: &weight, Reps: 10, SetType: "working"},
				},
			},
		},
	}

	payload, err := marshalWorkoutDraft(draft)

	if err != nil {
		t.Fatalf("marshalWorkoutDraft returned error: %v", err)
	}
	if !strings.HasPrefix(string(payload), "{") {
		t.Fatalf("marshalWorkoutDraft should return JSON, got %q", string(payload))
	}

	var decoded workout.CreateWorkoutRequest
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("marshalWorkoutDraft returned invalid JSON: %v", err)
	}
	if decoded.Date != draft.Date {
		t.Fatalf("decoded draft date = %q, want %q", decoded.Date, draft.Date)
	}
}

func TestMarshalWorkoutDraft_NilDraftReturnsNil(t *testing.T) {
	payload, err := marshalWorkoutDraft(nil)

	if err != nil {
		t.Fatalf("marshalWorkoutDraft returned error: %v", err)
	}
	if payload != nil {
		t.Fatal("marshalWorkoutDraft should return nil for nil draft")
	}
}
