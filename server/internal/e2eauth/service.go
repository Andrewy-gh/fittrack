package e2eauth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DevAuthHeaderName   = "x-fittrack-dev-e2e-user"
	FeatureKeyAIChatbot = "ai_chatbot"
	// Local E2E grants are intentionally separate from Stripe grants so local
	// bootstrap does not rewrite subscription-sourced access history.
	featureAccessSource      = "local_e2e_auth"
	featureAccessReference   = "local-e2e-bootstrap"
	featureAccessGrantedBy   = "fittrack-local-e2e"
	featureAccessGrantNote   = "Local development E2E bootstrap access"
	defaultSeedAssistantText = "Structured workout draft ready to edit in the workout form."
)

const ensureFeatureAccessGrantQuery = `
INSERT INTO user_feature_access (
	user_id,
	feature_key,
	source,
	source_reference,
	granted_by,
	note
)
SELECT $1, $2, $3, $4, $5, $6
WHERE NOT EXISTS (
	SELECT 1
	FROM user_feature_access
	WHERE user_id = $1
	  AND feature_key = $2
	  AND revoked_at IS NULL
	  AND (expires_at IS NULL OR expires_at > NOW())
)`

const expireLocalAIChatAccessQuery = `
UPDATE user_feature_access
SET expires_at = NOW() - INTERVAL '1 second'
WHERE user_id = $1
  AND feature_key = $2
  AND source = $3
  AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > NOW())`

type userService interface {
	EnsureUser(ctx context.Context, userID string) (db.Users, error)
}

type Service struct {
	logger      *slog.Logger
	queries     *db.Queries
	pool        *pgxpool.Pool
	userService userService
	userID      string
	email       string
	displayName string
}

type BootstrapResponse struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	FeatureKeys []string `json:"feature_keys"`
}

type SeedConversationRequest struct {
	Title                       string                       `json:"title"`
	UserPrompt                  string                       `json:"user_prompt"`
	AssistantText               string                       `json:"assistant_text"`
	LatestWorkoutDraft          workout.CreateWorkoutRequest `json:"latest_workout_draft"`
	ExpireAIChatAccessAfterSeed bool                         `json:"expire_ai_chat_access_after_seed,omitempty"`
}

type SeedConversationResponse struct {
	ConversationID int32 `json:"conversation_id"`
}

func NewService(
	logger *slog.Logger,
	queries *db.Queries,
	pool *pgxpool.Pool,
	userService userService,
	userID string,
	email string,
	displayName string,
) *Service {
	return &Service{
		logger:      logger,
		queries:     queries,
		pool:        pool,
		userService: userService,
		userID:      strings.TrimSpace(userID),
		email:       strings.TrimSpace(email),
		displayName: strings.TrimSpace(displayName),
	}
}

func (s *Service) Bootstrap(ctx context.Context) (*BootstrapResponse, error) {
	if err := s.ensureLocalUserAccess(ctx); err != nil {
		return nil, err
	}

	return &BootstrapResponse{
		UserID:      s.userID,
		Email:       s.email,
		DisplayName: s.displayName,
		FeatureKeys: []string{FeatureKeyAIChatbot},
	}, nil
}

func (s *Service) SeedConversation(
	ctx context.Context,
	request SeedConversationRequest,
) (*SeedConversationResponse, error) {
	if err := s.ensureLocalUserAccess(ctx); err != nil {
		return nil, err
	}

	draftJSON, err := json.Marshal(request.LatestWorkoutDraft)
	if err != nil {
		return nil, fmt.Errorf("marshal seeded workout draft: %w", err)
	}

	now := time.Now().UTC()
	trimmedPrompt := strings.TrimSpace(request.UserPrompt)
	trimmedAssistantText := strings.TrimSpace(request.AssistantText)
	if trimmedAssistantText == "" {
		trimmedAssistantText = defaultSeedAssistantText
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin seeded conversation transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)
	conversationRow, err := qtx.CreateAIChatConversation(ctx, db.CreateAIChatConversationParams{
		UserID: s.userID,
		Title:  textToPg(request.Title),
	})
	if err != nil {
		return nil, fmt.Errorf("create seeded ai chat conversation: %w", err)
	}

	completedTS := pgtype.Timestamptz{Time: now, Valid: true}
	if trimmedPrompt != "" {
		if _, err := qtx.CreateAIChatMessage(ctx, db.CreateAIChatMessageParams{
			ConversationID: conversationRow.ID,
			UserID:         s.userID,
			Role:           "user",
			Content:        trimmedPrompt,
			Status:         "completed",
			ErrorMessage:   pgtype.Text{},
			CompletedAt:    completedTS,
		}); err != nil {
			return nil, fmt.Errorf("create seeded ai chat user message: %w", err)
		}
	}

	if _, err := qtx.CreateAIChatMessage(ctx, db.CreateAIChatMessageParams{
		ConversationID: conversationRow.ID,
		UserID:         s.userID,
		Role:           "assistant",
		Content:        trimmedAssistantText,
		Status:         "completed",
		ErrorMessage:   pgtype.Text{},
		CompletedAt:    completedTS,
	}); err != nil {
		return nil, fmt.Errorf("create seeded ai chat assistant message: %w", err)
	}

	if err := qtx.SetAIChatConversationLatestWorkoutDraft(ctx, db.SetAIChatConversationLatestWorkoutDraftParams{
		ID:      conversationRow.ID,
		UserID:  s.userID,
		Column3: string(draftJSON),
	}); err != nil {
		return nil, fmt.Errorf("set seeded latest workout draft: %w", err)
	}

	if err := qtx.TouchAIChatConversation(ctx, db.TouchAIChatConversationParams{
		ID:            conversationRow.ID,
		UserID:        s.userID,
		LastMessageAt: completedTS,
	}); err != nil {
		return nil, fmt.Errorf("touch seeded ai chat conversation: %w", err)
	}

	if request.ExpireAIChatAccessAfterSeed {
		if _, err := tx.Exec(ctx, expireLocalAIChatAccessQuery, s.userID, FeatureKeyAIChatbot, featureAccessSource); err != nil {
			return nil, fmt.Errorf("expire local e2e ai chat access: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit seeded ai chat conversation: %w", err)
	}

	return &SeedConversationResponse{
		ConversationID: conversationRow.ID,
	}, nil
}

func (s *Service) UserID() string {
	return s.userID
}

func (s *Service) ensureLocalUserAccess(ctx context.Context) error {
	if _, err := s.userService.EnsureUser(ctx, s.userID); err != nil {
		return fmt.Errorf("ensure local e2e user: %w", err)
	}

	if _, err := s.pool.Exec(
		ctx,
		ensureFeatureAccessGrantQuery,
		s.userID,
		FeatureKeyAIChatbot,
		featureAccessSource,
		featureAccessReference,
		featureAccessGrantedBy,
		featureAccessGrantNote,
	); err != nil {
		return fmt.Errorf("grant local e2e feature access: %w", err)
	}

	return nil
}

func textToPg(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
