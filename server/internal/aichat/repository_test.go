package aichat

import (
	"strings"
	"testing"
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
