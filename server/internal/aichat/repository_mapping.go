package aichat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5/pgtype"
)

func buildConversationTitle(prompt string) string {
	title := strings.Join(strings.Fields(prompt), " ")
	title = strings.TrimSpace(title)
	return truncateWithEllipsis(title, 80)
}

func truncateForStorage(value string, maxLen int) string {
	return truncateWithEllipsis(value, maxLen)
}

func truncateWithEllipsis(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if value == "" || maxLen <= 0 {
		return ""
	}

	if utf8.RuneCountInString(value) <= maxLen {
		return value
	}

	if maxLen <= 3 {
		return string([]rune(value)[:maxLen])
	}

	runes := []rune(value)
	return strings.TrimSpace(string(runes[:maxLen-3])) + "..."
}

func isStreamingRunStale(updatedAt time.Time, now time.Time) bool {
	return !updatedAt.IsZero() && now.UTC().Sub(updatedAt.UTC()) > streamingRunStaleAfter
}

func textToPg(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func textPtrToPg(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return textToPg(*value)
}

func timePtrToPg(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func textValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func mapConversation(row db.AiChatConversation) (*Conversation, error) {
	createdAt, err := timeFromPg(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := timeFromPg(row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	lastMessageAt, err := timePtrFromPg(row.LastMessageAt)
	if err != nil {
		return nil, err
	}
	latestWorkoutDraft, err := parseStoredWorkoutDraft(row.LatestWorkoutDraft)
	if err != nil {
		return nil, err
	}
	latestWorkoutDraftStatus, err := draftStatusFromConversationRow(row)
	if err != nil {
		return nil, err
	}

	return &Conversation{
		ID:                       row.ID,
		UserID:                   row.UserID,
		Title:                    textPtr(row.Title),
		LatestWorkoutDraft:       latestWorkoutDraft,
		LatestWorkoutDraftStatus: latestWorkoutDraftStatus,
		CreatedAt:                createdAt,
		UpdatedAt:                updatedAt,
		LastMessageAt:            lastMessageAt,
	}, nil
}

func mapConversationSummary(row db.ListAIChatConversationsByUserRow) (*ConversationSummary, error) {
	createdAt, err := timeFromPg(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := timeFromPg(row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	lastMessageAt, err := timePtrFromPg(row.LastMessageAt)
	if err != nil {
		return nil, err
	}

	return &ConversationSummary{
		ID:            row.ID,
		Title:         textPtr(row.Title),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		LastMessageAt: lastMessageAt,
	}, nil
}

func mapMessage(row db.AiChatMessage) (*ChatMessage, error) {
	createdAt, err := timeFromPg(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := timeFromPg(row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	completedAt, err := timePtrFromPg(row.CompletedAt)
	if err != nil {
		return nil, err
	}

	return &ChatMessage{
		ID:             row.ID,
		ConversationID: row.ConversationID,
		UserID:         row.UserID,
		Role:           row.Role,
		Content:        row.Content,
		Status:         row.Status,
		ErrorMessage:   textPtr(row.ErrorMessage),
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		CompletedAt:    completedAt,
	}, nil
}

func mapRun(row db.AiChatRun) (*ChatRun, error) {
	createdAt, err := timeFromPg(row.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := timeFromPg(row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	startedAt, err := timeFromPg(row.StartedAt)
	if err != nil {
		return nil, err
	}
	completedAt, err := timePtrFromPg(row.CompletedAt)
	if err != nil {
		return nil, err
	}
	leaseExpiresAt, err := timePtrFromPg(row.GenerationLeaseExpiresAt)
	if err != nil {
		return nil, err
	}
	heartbeatAt, err := timePtrFromPg(row.GenerationHeartbeatAt)
	if err != nil {
		return nil, err
	}
	interruptedAt, err := timePtrFromPg(row.InterruptedAt)
	if err != nil {
		return nil, err
	}
	workoutDraft, err := parseStoredWorkoutDraft(row.WorkoutDraft)
	if err != nil {
		return nil, err
	}

	return &ChatRun{
		ID:                 row.ID,
		ConversationID:     row.ConversationID,
		UserID:             row.UserID,
		UserMessageID:      row.UserMessageID,
		AssistantMessageID: row.AssistantMessageID,
		Model:              row.Model,
		Status:             row.Status,
		RequestID:          textPtr(row.RequestID),
		ErrorMessage:       textPtr(row.ErrorMessage),
		WorkoutDraft:       workoutDraft,
		GenerationStatus:   row.GenerationStatus,
		GenerationOwner:    textPtr(row.GenerationOwner),
		LeaseExpiresAt:     leaseExpiresAt,
		HeartbeatAt:        heartbeatAt,
		GenerationAttempt:  row.GenerationAttempt,
		InterruptedAt:      interruptedAt,
		InterruptionReason: textPtr(row.InterruptionReason),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		StartedAt:          startedAt,
		CompletedAt:        completedAt,
	}, nil
}

func marshalWorkoutDraft(draft *workout.CreateWorkoutRequest) ([]byte, error) {
	if draft == nil {
		return nil, nil
	}
	normalizeWorkoutDraft(draft)
	return json.Marshal(draft)
}

func parseStoredWorkoutDraft(payload []byte) (*workout.CreateWorkoutRequest, error) {
	if len(payload) == 0 {
		return nil, nil
	}

	var draft workout.CreateWorkoutRequest
	if err := json.Unmarshal(payload, &draft); err != nil {
		return nil, fmt.Errorf("decode stored workout draft: %w", err)
	}

	normalizeWorkoutDraft(&draft)
	if err := validateWorkoutDraft(&draft); err != nil {
		return nil, fmt.Errorf("validate stored workout draft: %w", err)
	}

	return &draft, nil
}

func timeFromPg(ts pgtype.Timestamptz) (time.Time, error) {
	if !ts.Valid {
		return time.Time{}, fmt.Errorf("invalid timestamptz value")
	}
	return ts.Time.UTC(), nil
}

func timePtrFromPg(ts pgtype.Timestamptz) (*time.Time, error) {
	if !ts.Valid {
		return nil, nil
	}
	utc := ts.Time.UTC()
	return &utc, nil
}

func textPtr(txt pgtype.Text) *string {
	if !txt.Valid {
		return nil
	}
	value := txt.String
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}

func int4PtrToPg(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func intPtrFromPg(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return int32Ptr(value.Int32)
}

func draftStatusFromConversationRow(row db.AiChatConversation) (*LatestWorkoutDraftStatus, error) {
	sourceRunID := intPtrFromPg(row.LatestWorkoutDraftSourceRunID)
	savedWorkoutID := intPtrFromPg(row.LatestWorkoutDraftSavedWorkoutID)
	savedAt, err := timePtrFromPg(row.LatestWorkoutDraftSavedAt)
	if err != nil {
		return nil, err
	}
	if sourceRunID == nil && savedWorkoutID == nil && savedAt == nil {
		return nil, nil
	}
	return &LatestWorkoutDraftStatus{
		SourceRunID:    sourceRunID,
		IsSaved:        savedWorkoutID != nil,
		SavedWorkoutID: savedWorkoutID,
		SavedAt:        savedAt,
	}, nil
}
