package aichat

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
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

func TestIsStreamingRunStale(t *testing.T) {
	now := time.Date(2026, 3, 26, 18, 30, 0, 0, time.UTC)

	if isStreamingRunStale(now.Add(-streamingRunStaleAfter+time.Second), now) {
		t.Fatal("run newer than stale threshold should not be stale")
	}
	if !isStreamingRunStale(now.Add(-streamingRunStaleAfter-time.Second), now) {
		t.Fatal("run older than stale threshold should be stale")
	}
}
