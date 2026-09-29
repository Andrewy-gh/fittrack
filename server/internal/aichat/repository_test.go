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
