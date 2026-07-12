package aichat

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/billing"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockFeatureAccessService struct {
	mock.Mock
}

func (m *mockFeatureAccessService) HasCurrentUserFeatureAccess(ctx context.Context, featureKey string) (bool, error) {
	args := m.Called(ctx, featureKey)
	return args.Bool(0), args.Error(1)
}

type mockRuntime struct {
	mock.Mock
}

func (m *mockRuntime) ModelName() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockRuntime) Available() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockRuntime) GenerateValidation(ctx context.Context, prompt string) (*ValidationOutput, error) {
	args := m.Called(ctx, prompt)
	output, _ := args.Get(0).(*ValidationOutput)
	return output, args.Error(1)
}

func (m *mockRuntime) StreamValidation(ctx context.Context, prompt string, onChunk func(string) error) (*StreamDone, error) {
	args := m.Called(ctx, prompt, onChunk)
	done, _ := args.Get(0).(*StreamDone)
	return done, args.Error(1)
}

func (m *mockRuntime) StreamChat(ctx context.Context, prompt string, history []RuntimeChatMessage, onChunk func(string) error) (*StreamDone, error) {
	args := m.Called(ctx, prompt, history, onChunk)
	done, _ := args.Get(0).(*StreamDone)
	return done, args.Error(1)
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) ListWorkoutsWithSets(ctx context.Context, userID string, filter WorkoutHistoryFilter) ([]ChatWorkoutView, error) {
	args := m.Called(ctx, userID, filter)
	workouts, _ := args.Get(0).([]ChatWorkoutView)
	return workouts, args.Error(1)
}

func (m *mockRepository) ResolveExerciseNames(ctx context.Context, userID string, query string) ([]string, error) {
	args := m.Called(ctx, userID, query)
	names, _ := args.Get(0).([]string)
	return names, args.Error(1)
}

func (m *mockRepository) TrainingSnapshot(ctx context.Context, userID string) (*TrainingSnapshot, error) {
	args := m.Called(ctx, userID)
	snapshot, _ := args.Get(0).(*TrainingSnapshot)
	return snapshot, args.Error(1)
}

func (m *mockRepository) TrainingProfile(ctx context.Context, userID string) (*TrainingProfile, error) {
	args := m.Called(ctx, userID)
	profile, _ := args.Get(0).(*TrainingProfile)
	return profile, args.Error(1)
}

func (m *mockRepository) UpdateTrainingProfile(ctx context.Context, userID string, update TrainingProfileUpdate) (*TrainingProfile, error) {
	args := m.Called(ctx, userID, update)
	profile, _ := args.Get(0).(*TrainingProfile)
	return profile, args.Error(1)
}

func (m *mockRepository) ExerciseStats(ctx context.Context, userID string, exerciseName string, window string) (*ExerciseStatsView, error) {
	args := m.Called(ctx, userID, exerciseName, window)
	stats, _ := args.Get(0).(*ExerciseStatsView)
	return stats, args.Error(1)
}

func (m *mockRepository) CreateConversation(ctx context.Context, userID string) (*Conversation, error) {
	args := m.Called(ctx, userID)
	conversation, _ := args.Get(0).(*Conversation)
	return conversation, args.Error(1)
}

func (m *mockRepository) GetConversation(ctx context.Context, conversationID int32, userID string) (*Conversation, error) {
	args := m.Called(ctx, conversationID, userID)
	conversation, _ := args.Get(0).(*Conversation)
	return conversation, args.Error(1)
}

func (m *mockRepository) DeleteConversation(ctx context.Context, conversationID int32, userID string) error {
	args := m.Called(ctx, conversationID, userID)
	return args.Error(0)
}

func (m *mockRepository) ListConversations(ctx context.Context, userID string, limit int32) ([]ConversationSummary, error) {
	args := m.Called(ctx, userID, limit)
	conversations, _ := args.Get(0).([]ConversationSummary)
	return conversations, args.Error(1)
}

func (m *mockRepository) SaveLatestWorkoutDraft(ctx context.Context, request SaveLatestWorkoutDraftRequest) (*SavedLatestWorkoutDraft, error) {
	args := m.Called(ctx, request)
	resp, _ := args.Get(0).(*SavedLatestWorkoutDraft)
	return resp, args.Error(1)
}

func (m *mockRepository) ListMessages(ctx context.Context, conversationID int32, userID string) ([]ChatMessage, error) {
	args := m.Called(ctx, conversationID, userID)
	messages, _ := args.Get(0).([]ChatMessage)
	return messages, args.Error(1)
}

func (m *mockRepository) GetActiveRunForConversation(ctx context.Context, conversationID int32, userID string) (*ChatRun, error) {
	args := m.Called(ctx, conversationID, userID)
	run, _ := args.Get(0).(*ChatRun)
	return run, args.Error(1)
}

func (m *mockRepository) GetLatestStreamSequence(ctx context.Context, runID int32, userID string) (int32, error) {
	args := m.Called(ctx, runID, userID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *mockRepository) LoadPreparedRunForRecovery(ctx context.Context, runID int32, userID string) (*PreparedMessageStream, error) {
	args := m.Called(ctx, runID, userID)
	prepared, _ := args.Get(0).(*PreparedMessageStream)
	return prepared, args.Error(1)
}

func (m *mockRepository) LoadPreparedRunForResume(ctx context.Context, conversationID int32, runID int32, userID string, afterSequence int32) (*PreparedResumeStream, error) {
	args := m.Called(ctx, conversationID, runID, userID, afterSequence)
	prepared, _ := args.Get(0).(*PreparedResumeStream)
	return prepared, args.Error(1)
}

func (m *mockRepository) ListStreamChunksAfter(ctx context.Context, runID int32, userID string, afterSequence int32) ([]StreamChunk, error) {
	args := m.Called(ctx, runID, userID, afterSequence)
	chunks, _ := args.Get(0).([]StreamChunk)
	return chunks, args.Error(1)
}

func (m *mockRepository) PrepareMessageStream(ctx context.Context, conversationID int32, userID string, prompt string, model string, requestID string) (*PreparedMessageStream, error) {
	args := m.Called(ctx, conversationID, userID, prompt, model, requestID)
	prepared, _ := args.Get(0).(*PreparedMessageStream)
	return prepared, args.Error(1)
}

func (m *mockRepository) ClaimRunGeneration(ctx context.Context, run *ChatRun, owner runOwner, now time.Time) error {
	args := m.Called(ctx, run, owner, now)
	return args.Error(0)
}

func (m *mockRepository) HeartbeatRunGeneration(ctx context.Context, run *ChatRun, owner runOwner, now time.Time) (bool, error) {
	args := m.Called(ctx, run, owner, now)
	return args.Bool(0), args.Error(1)
}

func (m *mockRepository) OwnsRunGeneration(ctx context.Context, run *ChatRun, owner runOwner) (bool, error) {
	args := m.Called(ctx, run, owner)
	return args.Bool(0), args.Error(1)
}

func (m *mockRepository) AppendStreamChunk(ctx context.Context, prepared *PreparedMessageStream, delta string, partialText string, updatedAt time.Time) (int32, error) {
	args := m.Called(ctx, prepared, delta, partialText, updatedAt)
	return args.Get(0).(int32), args.Error(1)
}

func (m *mockRepository) InterruptRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, reason string, completedAt time.Time) error {
	args := m.Called(ctx, prepared, partialText, reason, completedAt)
	return args.Error(0)
}

func (m *mockRepository) CompleteRun(ctx context.Context, prepared *PreparedMessageStream, assistantText string, workoutDraft *workout.CreateWorkoutRequest, completedAt time.Time) (*ChatMessage, *ChatRun, error) {
	args := m.Called(ctx, prepared, assistantText, workoutDraft, completedAt)
	message, _ := args.Get(0).(*ChatMessage)
	run, _ := args.Get(1).(*ChatRun)
	return message, run, args.Error(2)
}

func (m *mockRepository) FailRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, failure error, completedAt time.Time) error {
	args := m.Called(ctx, prepared, partialText, failure, completedAt)
	return args.Error(0)
}

func (m *mockRepository) StopRun(ctx context.Context, conversationID, runID int32, userID string, stoppedAt time.Time) (*StopRunResponse, error) {
	args := m.Called(ctx, conversationID, runID, userID, stoppedAt)
	result, _ := args.Get(0).(*StopRunResponse)
	return result, args.Error(1)
}

type mockRecoveryDispatcher struct {
	mock.Mock
}

func (m *mockRecoveryDispatcher) EnqueueRunRecovery(ctx context.Context, request RunRecoveryRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

type mockPremiumAccessService struct {
	mock.Mock
}

func (m *mockPremiumAccessService) EnsureAIChatPromptAllowed(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockWorkoutDraftSaver struct {
	mock.Mock
}

func (m *mockWorkoutDraftSaver) SaveWorkoutTx(ctx context.Context, qtx *db.Queries, requestBody workout.CreateWorkoutRequest, userID string) (int32, error) {
	args := m.Called(ctx, qtx, requestBody, userID)
	return args.Get(0).(int32), args.Error(1)
}

func TestServiceValidate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("returns feature-disabled error when user lacks ai access", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		runtime := new(mockRuntime)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, runtime, repo, nil)

		runtime.On("Available").Return(true).Once()
		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(false, nil).Once()

		resp, err := service.Validate(context.Background(), "test prompt")

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrFeatureDisabled)
		runtime.AssertExpectations(t)
		featureAccess.AssertExpectations(t)
	})

	t.Run("returns runtime-unavailable before checking feature access", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		runtime := new(mockRuntime)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, runtime, repo, nil)

		runtime.On("Available").Return(false).Once()

		resp, err := service.Validate(context.Background(), "test prompt")

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrRuntimeUnavailable)
		runtime.AssertExpectations(t)
		featureAccess.AssertExpectations(t)
	})

	t.Run("returns structured validation response", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		runtime := new(mockRuntime)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, runtime, repo, nil)

		expected := &ValidationOutput{
			Summary:  "Viable.",
			NextStep: "Build persistence.",
		}

		runtime.On("Available").Return(true).Once()
		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
		runtime.On("GenerateValidation", mock.Anything, "test prompt").Return(expected, nil).Once()
		runtime.On("ModelName").Return(defaultModelName).Once()

		resp, err := service.Validate(context.Background(), "test prompt")

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, defaultModelName, resp.Model)
		assert.Equal(t, expected, resp.Output)
		runtime.AssertExpectations(t)
		featureAccess.AssertExpectations(t)
	})

	t.Run("propagates auth errors from feature access", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		runtime := new(mockRuntime)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, runtime, repo, nil)
		expectedErr := apperrors.NewUnauthorized("feature access", "")

		runtime.On("Available").Return(true).Once()
		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(false, expectedErr).Once()

		resp, err := service.Validate(context.Background(), "test prompt")

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
		runtime.AssertExpectations(t)
		featureAccess.AssertExpectations(t)
	})

	t.Run("propagates runtime errors", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		runtime := new(mockRuntime)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, runtime, repo, nil)
		expectedErr := errors.New("boom")

		runtime.On("Available").Return(true).Once()
		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
		runtime.On("GenerateValidation", mock.Anything, "test prompt").Return(nil, expectedErr).Once()

		resp, err := service.Validate(context.Background(), "test prompt")

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
		runtime.AssertExpectations(t)
		featureAccess.AssertExpectations(t)
	})
}

func TestServiceCreateConversation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")
	now := time.Date(2026, 3, 26, 17, 20, 0, 0, time.UTC)

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("CreateConversation", mock.Anything, "user-123").Return(&Conversation{
		ID:        41,
		UserID:    "user-123",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil).Once()

	conversation, err := service.CreateConversation(ctx)

	require.NoError(t, err)
	require.NotNil(t, conversation)
	assert.Equal(t, int32(41), conversation.ID)
	runtime.AssertNotCalled(t, "Available")
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceListConversations_AllowsReadWithoutRuntime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")
	now := time.Date(2026, 3, 26, 17, 20, 0, 0, time.UTC)

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("ListConversations", mock.Anything, "user-123", conversationListLimit).Return([]ConversationSummary{
		{
			ID:        41,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil).Once()

	conversations, err := service.ListConversations(ctx)

	require.NoError(t, err)
	require.Len(t, conversations, 1)
	assert.Equal(t, int32(41), conversations[0].ID)
	runtime.AssertNotCalled(t, "Available")
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceDeleteConversation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("deletes an owned conversation with feature access", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, new(mockRuntime), repo, nil)
		ctx := user.WithContext(context.Background(), "user-123")

		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
		repo.On("DeleteConversation", mock.Anything, int32(41), "user-123").Return(nil).Once()

		require.NoError(t, service.DeleteConversation(ctx, 41))
		featureAccess.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("requires feature access", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, new(mockRuntime), repo, nil)
		ctx := user.WithContext(context.Background(), "user-123")

		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(false, nil).Once()

		assert.ErrorIs(t, service.DeleteConversation(ctx, 41), ErrFeatureDisabled)
		repo.AssertNotCalled(t, "DeleteConversation", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("maps a missing or foreign conversation to not found", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, new(mockRuntime), repo, nil)
		ctx := user.WithContext(context.Background(), "user-123")

		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
		repo.On("DeleteConversation", mock.Anything, int32(41), "user-123").Return(pgx.ErrNoRows).Once()

		var notFound *apperrors.NotFound
		require.ErrorAs(t, service.DeleteConversation(ctx, 41), &notFound)
	})

	t.Run("preserves active-run conflict", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, new(mockRuntime), repo, nil)
		ctx := user.WithContext(context.Background(), "user-123")

		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
		repo.On("DeleteConversation", mock.Anything, int32(41), "user-123").Return(ErrConversationBusy).Once()

		assert.ErrorIs(t, service.DeleteConversation(ctx, 41), ErrConversationBusy)
	})
}

func TestServiceGetConversation_AllowsReadWithoutRuntime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")
	now := time.Date(2026, 3, 26, 17, 25, 0, 0, time.UTC)

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("GetConversation", mock.Anything, int32(41), "user-123").Return(&Conversation{
		ID:        41,
		UserID:    "user-123",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil).Once()
	repo.On("ListMessages", mock.Anything, int32(41), "user-123").Return([]ChatMessage{
		{
			ID:             1,
			ConversationID: 41,
			UserID:         "user-123",
			Role:           roleUser,
			Content:        "hello",
			Status:         statusCompleted,
			CreatedAt:      now,
			UpdatedAt:      now,
			CompletedAt:    &now,
		},
	}, nil).Once()
	repo.On("GetActiveRunForConversation", mock.Anything, int32(41), "user-123").Return((*ChatRun)(nil), pgx.ErrNoRows).Once()

	detail, err := service.GetConversation(ctx, 41)

	require.NoError(t, err)
	require.NotNil(t, detail)
	require.NotNil(t, detail.Conversation)
	assert.Len(t, detail.Messages, 1)
	runtime.AssertNotCalled(t, "Available")
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceGetConversation_IncludesActiveRunSequence(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")
	now := time.Date(2026, 3, 26, 17, 26, 0, 0, time.UTC)

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("GetConversation", mock.Anything, int32(41), "user-123").Return(&Conversation{
		ID:        41,
		UserID:    "user-123",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil).Once()
	repo.On("ListMessages", mock.Anything, int32(41), "user-123").Return([]ChatMessage{
		{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Role:           roleAssistant,
			Content:        "partial",
			Status:         statusStreaming,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}, nil).Once()
	repo.On("GetActiveRunForConversation", mock.Anything, int32(41), "user-123").Return(&ChatRun{
		ID:                 51,
		ConversationID:     41,
		UserID:             "user-123",
		AssistantMessageID: 61,
		Status:             statusStreaming,
	}, nil).Once()
	repo.On("GetLatestStreamSequence", mock.Anything, int32(51), "user-123").Return(int32(4), nil).Once()

	detail, err := service.GetConversation(ctx, 41)

	require.NoError(t, err)
	require.NotNil(t, detail)
	require.NotNil(t, detail.ActiveRun)
	assert.Equal(t, int32(51), detail.ActiveRun.ID)
	assert.Equal(t, int32(61), detail.ActiveRun.AssistantMessageID)
	assert.Equal(t, int32(4), detail.ActiveRun.LatestSequence)
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceGetConversation_IncludesLatestWorkoutDraft(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")
	now := time.Date(2026, 4, 21, 17, 26, 0, 0, time.UTC)
	workoutFocus := "pull"
	latestWorkoutDraft := &workout.CreateWorkoutRequest{
		Date:         "2026-04-21T12:00:00Z",
		WorkoutFocus: &workoutFocus,
		Exercises: []workout.ExerciseInput{
			{
				Name: "Chest Supported Row",
				Sets: []workout.SetInput{
					{Reps: 10, SetType: "working"},
				},
			},
		},
	}

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("GetConversation", mock.Anything, int32(41), "user-123").Return(&Conversation{
		ID:                 41,
		UserID:             "user-123",
		LatestWorkoutDraft: latestWorkoutDraft,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil).Once()
	repo.On("ListMessages", mock.Anything, int32(41), "user-123").Return([]ChatMessage{}, nil).Once()
	repo.On("GetActiveRunForConversation", mock.Anything, int32(41), "user-123").Return((*ChatRun)(nil), pgx.ErrNoRows).Once()

	detail, err := service.GetConversation(ctx, 41)

	require.NoError(t, err)
	require.NotNil(t, detail)
	require.NotNil(t, detail.Conversation)
	require.NotNil(t, detail.Conversation.LatestWorkoutDraft)
	assert.Equal(t, "2026-04-21T12:00:00Z", detail.Conversation.LatestWorkoutDraft.Date)
	require.NotNil(t, detail.Conversation.LatestWorkoutDraft.WorkoutFocus)
	assert.Equal(t, "pull", *detail.Conversation.LatestWorkoutDraft.WorkoutFocus)
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceSaveLatestWorkoutDraft(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("saves the latest draft through the repository transaction boundary", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		repo := new(mockRepository)
		workoutSaver := new(mockWorkoutDraftSaver)
		service := NewService(logger, featureAccess, nil, repo, workoutSaver)
		ctx := user.WithContext(context.Background(), "user-123")
		savedWorkoutID := int32(88)
		expectedSavedAt := time.Date(2026, 4, 21, 17, 30, 0, 0, time.UTC)
		draft := workout.CreateWorkoutRequest{
			Date: "2026-04-21T12:00:00Z",
			Exercises: []workout.ExerciseInput{
				{
					Name: "Chest Supported Row",
					Sets: []workout.SetInput{
						{Reps: 10, SetType: "working"},
					},
				},
			},
		}

		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
		workoutSaver.On("SaveWorkoutTx", mock.Anything, (*db.Queries)(nil), draft, "user-123").Return(savedWorkoutID, nil).Once()
		repo.On("SaveLatestWorkoutDraft", mock.Anything, mock.MatchedBy(func(request SaveLatestWorkoutDraftRequest) bool {
			return request.ConversationID == 41 &&
				request.UserID == "user-123" &&
				!request.SavedAt.IsZero() &&
				request.SaveWorkout != nil
		})).Run(func(args mock.Arguments) {
			request := args.Get(1).(SaveLatestWorkoutDraftRequest)
			actualSavedAt := request.SavedAt
			assert.WithinDuration(t, time.Now().UTC(), actualSavedAt, 5*time.Second)

			workoutID, err := request.SaveWorkout(context.Background(), nil, draft, "user-123")
			require.NoError(t, err)
			assert.Equal(t, savedWorkoutID, workoutID)
		}).Return(&SavedLatestWorkoutDraft{
			WorkoutID: savedWorkoutID,
			Conversation: &Conversation{
				ID: 41,
				LatestWorkoutDraftStatus: &LatestWorkoutDraftStatus{
					IsSaved:        true,
					SavedWorkoutID: &savedWorkoutID,
					SavedAt:        &expectedSavedAt,
				},
			},
		}, nil).Once()

		resp, err := service.SaveLatestWorkoutDraft(ctx, 41)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, savedWorkoutID, resp.WorkoutID)
		require.NotNil(t, resp.Conversation)
		require.NotNil(t, resp.Conversation.LatestWorkoutDraftStatus)
		assert.True(t, resp.Conversation.LatestWorkoutDraftStatus.IsSaved)
		assert.Equal(t, &savedWorkoutID, resp.Conversation.LatestWorkoutDraftStatus.SavedWorkoutID)
		featureAccess.AssertExpectations(t)
		workoutSaver.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("returns the existing saved workout when the latest draft is already saved", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, nil, repo, nil)
		ctx := user.WithContext(context.Background(), "user-123")
		savedWorkoutID := int32(88)

		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
		repo.On("SaveLatestWorkoutDraft", mock.Anything, mock.MatchedBy(func(request SaveLatestWorkoutDraftRequest) bool {
			return request.ConversationID == 41 && request.UserID == "user-123" && request.SaveWorkout != nil
		})).Return(&SavedLatestWorkoutDraft{
			WorkoutID: savedWorkoutID,
			Conversation: &Conversation{
				ID: 41,
				LatestWorkoutDraftStatus: &LatestWorkoutDraftStatus{
					IsSaved:        true,
					SavedWorkoutID: &savedWorkoutID,
				},
			},
		}, nil).Once()

		resp, err := service.SaveLatestWorkoutDraft(ctx, 41)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, savedWorkoutID, resp.WorkoutID)
		featureAccess.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("returns an explicit error when there is no latest draft to save", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, nil, repo, nil)
		ctx := user.WithContext(context.Background(), "user-123")

		featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
		repo.On("SaveLatestWorkoutDraft", mock.Anything, mock.MatchedBy(func(request SaveLatestWorkoutDraftRequest) bool {
			return request.ConversationID == 41 && request.UserID == "user-123" && request.SaveWorkout != nil
		})).Return((*SavedLatestWorkoutDraft)(nil), ErrLatestWorkoutDraftUnavailable).Once()

		resp, err := service.SaveLatestWorkoutDraft(ctx, 41)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrLatestWorkoutDraftUnavailable)
		featureAccess.AssertExpectations(t)
		repo.AssertExpectations(t)
	})
}

func TestServicePrepareMessageStream_RequiresRuntime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")

	runtime.On("Available").Return(false).Once()

	prepared, err := service.PrepareMessageStream(ctx, 41, "hello", "req-123")

	require.Error(t, err)
	assert.Nil(t, prepared)
	assert.ErrorIs(t, err, ErrRuntimeUnavailable)
	repo.AssertNotCalled(t, "PrepareMessageStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	featureAccess.AssertNotCalled(t, "HasCurrentUserFeatureAccess", mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
}

func TestServicePrepareMessageStream_StopsWhenTrialPromptCapReached(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")

	runtime.On("Available").Return(true).Once()
	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	runtime.On("ModelName").Return(defaultModelName).Once()
	repo.On("PrepareMessageStream", mock.Anything, int32(41), "user-123", "hello", defaultModelName, "req-123").
		Return((*PreparedMessageStream)(nil), billing.ErrTrialPromptLimitExceeded).
		Once()

	prepared, err := service.PrepareMessageStream(ctx, 41, "hello", "req-123")

	require.ErrorIs(t, err, billing.ErrTrialPromptLimitExceeded)
	assert.Nil(t, prepared)
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
	runtime.AssertExpectations(t)
}

func TestServiceRequestMessageRecovery_QueuesActiveRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	recovery := new(mockRecoveryDispatcher)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	service.SetRecoveryDispatcher(recovery)
	ctx := user.WithContext(context.Background(), "user-123")
	expiredLease := time.Now().UTC().Add(-time.Second)

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("GetConversation", mock.Anything, int32(41), "user-123").Return(&Conversation{ID: 41, UserID: "user-123"}, nil).Once()
	repo.On("GetActiveRunForConversation", mock.Anything, int32(41), "user-123").Return(&ChatRun{
		ID:               51,
		ConversationID:   41,
		UserID:           "user-123",
		Status:           statusStreaming,
		GenerationStatus: generationStatusGenerating,
		LeaseExpiresAt:   &expiredLease,
	}, nil).Once()
	recovery.On("EnqueueRunRecovery", mock.Anything, RunRecoveryRequest{
		ConversationID: 41,
		RunID:          51,
		UserID:         "user-123",
		Reason:         recoverReasonStreamReconnect,
	}).Return(nil).Once()

	resp, err := service.RequestMessageRecovery(ctx, 41, "")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, recoverStatusQueued, resp.Status)
	assert.Equal(t, int32(51), resp.RunID)
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
	recovery.AssertExpectations(t)
}

func TestServiceRequestMessageRecovery_QueuesStaleClaimedRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	recovery := new(mockRecoveryDispatcher)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	service.SetRecoveryDispatcher(recovery)
	ctx := user.WithContext(context.Background(), "user-123")
	expiredLease := time.Now().UTC().Add(-time.Second)

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("GetConversation", mock.Anything, int32(41), "user-123").Return(&Conversation{ID: 41, UserID: "user-123"}, nil).Once()
	repo.On("GetActiveRunForConversation", mock.Anything, int32(41), "user-123").Return(&ChatRun{
		ID:               51,
		ConversationID:   41,
		UserID:           "user-123",
		Status:           statusStreaming,
		GenerationStatus: generationStatusGenerating,
		LeaseExpiresAt:   &expiredLease,
	}, nil).Once()
	recovery.On("EnqueueRunRecovery", mock.Anything, RunRecoveryRequest{
		ConversationID: 41,
		RunID:          51,
		UserID:         "user-123",
		Reason:         recoverReasonStreamReconnect,
	}).Return(nil).Once()

	resp, err := service.RequestMessageRecovery(ctx, 41, "")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, recoverStatusQueued, resp.Status)
	assert.Equal(t, int32(51), resp.RunID)
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
	recovery.AssertExpectations(t)
}

func TestServiceRequestMessageRecovery_ReturnsUnavailableWithoutDispatcher(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()

	resp, err := service.RequestMessageRecovery(ctx, 41, "")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrRecoveryUnavailable)
	repo.AssertNotCalled(t, "GetConversation", mock.Anything, mock.Anything, mock.Anything)
	featureAccess.AssertExpectations(t)
}

func TestServiceRequestMessageRecovery_NoopsWithoutRecoveryMarker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	recovery := new(mockRecoveryDispatcher)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	service.SetRecoveryDispatcher(recovery)
	ctx := user.WithContext(context.Background(), "user-123")

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("GetConversation", mock.Anything, int32(41), "user-123").Return(&Conversation{ID: 41, UserID: "user-123"}, nil).Once()
	repo.On("GetActiveRunForConversation", mock.Anything, int32(41), "user-123").Return(&ChatRun{
		ID:             51,
		ConversationID: 41,
		UserID:         "user-123",
		Status:         statusStreaming,
	}, nil).Once()

	resp, err := service.RequestMessageRecovery(ctx, 41, "")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, recoverStatusNotNeeded, resp.Status)
	assert.Zero(t, resp.RunID)
	recovery.AssertNotCalled(t, "EnqueueRunRecovery", mock.Anything, mock.Anything)
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceRequestMessageRecovery_NoopsWithoutActiveRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	recovery := new(mockRecoveryDispatcher)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	service.SetRecoveryDispatcher(recovery)
	ctx := user.WithContext(context.Background(), "user-123")

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("GetConversation", mock.Anything, int32(41), "user-123").Return(&Conversation{ID: 41, UserID: "user-123"}, nil).Once()
	repo.On("GetActiveRunForConversation", mock.Anything, int32(41), "user-123").Return((*ChatRun)(nil), pgx.ErrNoRows).Once()

	resp, err := service.RequestMessageRecovery(ctx, 41, "")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, recoverStatusNotNeeded, resp.Status)
	assert.Zero(t, resp.RunID)
	recovery.AssertNotCalled(t, "EnqueueRunRecovery", mock.Anything, mock.Anything)
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceStreamMessage_CompletesRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	now := time.Date(2026, 3, 26, 17, 30, 0, 0, time.UTC)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			UserMessageID:      60,
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		History: []ChatMessage{
			{Role: roleUser, Content: "previous user", Status: statusCompleted},
			{Role: roleAssistant, Content: "previous assistant", Status: statusCompleted},
		},
		Prompt: "new prompt",
	}
	workoutDraft := &workout.CreateWorkoutRequest{
		Date: "2026-03-26T17:30:00Z",
		Exercises: []workout.ExerciseInput{
			{
				Name: "Goblet Squat",
				Sets: []workout.SetInput{
					{Reps: 10, SetType: "working"},
				},
			},
		},
	}

	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{
		{Role: roleUser, Text: "previous user"},
		{Role: roleAssistant, Text: "previous assistant"},
	}, mock.Anything).Run(func(args mock.Arguments) {
		source, ok := trainingProfileSourceFromContext(args.Get(0).(context.Context))
		require.True(t, ok)
		assert.Equal(t, int32(41), source.ConversationID)
		assert.Equal(t, int32(60), source.MessageID)
		onChunk := args.Get(3).(func(string) error)
		_ = onChunk("hello ")
		_ = onChunk("world")
	}).Return(&StreamDone{
		Model:        defaultModelName,
		Text:         "hello world",
		WorkoutDraft: workoutDraft,
		ToolCalls:    []string{workoutDraftToolName, updateTrainingProfileToolName},
	}, nil).Once()
	repo.On("AppendStreamChunk", mock.Anything, prepared, "hello ", "hello", mock.AnythingOfType("time.Time")).Return(int32(1), nil).Once()
	repo.On("AppendStreamChunk", mock.Anything, prepared, "world", "hello world", mock.AnythingOfType("time.Time")).Return(int32(2), nil).Once()
	repo.On("CompleteRun", mock.Anything, prepared, "hello world", workoutDraft, mock.AnythingOfType("time.Time")).Return(&ChatMessage{
		ID:             61,
		ConversationID: 41,
		UserID:         "user-123",
		Status:         statusCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	}, &ChatRun{
		ID:                 51,
		ConversationID:     41,
		UserID:             "user-123",
		AssistantMessageID: 61,
		Model:              defaultModelName,
		Status:             statusCompleted,
		WorkoutDraft:       workoutDraft,
		CreatedAt:          now,
		UpdatedAt:          now,
		StartedAt:          now,
		CompletedAt:        &now,
	}, nil).Once()

	done, err := service.StreamMessage(context.Background(), prepared, func(StreamChunk) error { return nil })

	require.NoError(t, err)
	require.NotNil(t, done)
	assert.Equal(t, "hello world", done.Text)
	assert.Equal(t, int32(51), done.RunID)
	assert.Equal(t, workoutDraft, done.WorkoutDraft)
	assert.Equal(t, []string{workoutDraftToolName, updateTrainingProfileToolName}, done.ToolCalls)
	repo.AssertNotCalled(t, "FailRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceStreamMessage_FailsRunOnDisconnect(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		Prompt: "new prompt",
	}

	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{}, mock.Anything).Run(func(args mock.Arguments) {
		onChunk := args.Get(3).(func(string) error)
		err := onChunk("partial ")
		require.ErrorIs(t, err, ErrStreamDisconnected)
	}).Return((*StreamDone)(nil), ErrStreamDisconnected).Once()
	repo.On("AppendStreamChunk", mock.Anything, prepared, "partial ", "partial", mock.AnythingOfType("time.Time")).Return(int32(1), nil).Once()
	repo.On("FailRun", mock.Anything, prepared, "partial", ErrStreamDisconnected, mock.AnythingOfType("time.Time")).Return(nil).Once()

	done, err := service.StreamMessage(context.Background(), prepared, func(StreamChunk) error {
		return errors.New("broken pipe")
	})

	require.Error(t, err)
	assert.Nil(t, done)
	assert.ErrorIs(t, err, ErrStreamDisconnected)
	repo.AssertNotCalled(t, "CompleteRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceStreamMessage_FailsRunOnRuntimeError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		Prompt: "new prompt",
	}
	expectedErr := errors.New("provider failed")

	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{}, mock.Anything).Run(func(args mock.Arguments) {
		onChunk := args.Get(3).(func(string) error)
		_ = onChunk("partial ")
	}).Return((*StreamDone)(nil), expectedErr).Once()
	repo.On("AppendStreamChunk", mock.Anything, prepared, "partial ", "partial", mock.AnythingOfType("time.Time")).Return(int32(1), nil).Once()
	repo.On("FailRun", mock.Anything, prepared, "partial", expectedErr, mock.AnythingOfType("time.Time")).Return(nil).Once()

	done, err := service.StreamMessage(context.Background(), prepared, func(StreamChunk) error { return nil })

	require.Error(t, err)
	assert.Nil(t, done)
	assert.ErrorIs(t, err, expectedErr)
	repo.AssertNotCalled(t, "CompleteRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceStartMessageGeneration_ContinuesAfterCallerContextCanceled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	now := time.Date(2026, 3, 26, 17, 30, 0, 0, time.UTC)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:               51,
			ConversationID:   41,
			UserID:           "user-123",
			Model:            defaultModelName,
			Status:           statusStreaming,
			GenerationStatus: generationStatusQueued,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		Prompt: "new prompt",
	}
	completed := make(chan struct{})

	repo.On("ClaimRunGeneration", mock.Anything, prepared.Run, mock.Anything, mock.AnythingOfType("time.Time")).Run(func(args mock.Arguments) {
		run := args.Get(1).(*ChatRun)
		run.GenerationStatus = generationStatusGenerating
		run.GenerationOwner = stringPtr("api:test")
	}).Return(nil).Once()
	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{}, mock.Anything).Run(func(args mock.Arguments) {
		streamCtx := args.Get(0).(context.Context)
		require.NoError(t, streamCtx.Err())
	}).Return(&StreamDone{
		Model: defaultModelName,
		Text:  "hello",
	}, nil).Once()
	repo.On("CompleteRun", mock.Anything, prepared, "hello", (*workout.CreateWorkoutRequest)(nil), mock.AnythingOfType("time.Time")).Run(func(mock.Arguments) {
		close(completed)
	}).Return(&ChatMessage{
		ID:             61,
		ConversationID: 41,
		UserID:         "user-123",
		Status:         statusCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	}, &ChatRun{
		ID:             51,
		ConversationID: 41,
		UserID:         "user-123",
		Model:          defaultModelName,
		Status:         statusCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
		StartedAt:      now,
		CompletedAt:    &now,
	}, nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, service.StartMessageGeneration(ctx, prepared))
	cancel()

	select {
	case <-completed:
	case <-time.After(time.Second):
		t.Fatal("generation did not complete after caller context cancellation")
	}

	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceStartMessageGeneration_StopDuringClaimCancelsBeforeRuntimeStarts(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:               51,
			ConversationID:   41,
			UserID:           "user-123",
			Status:           statusStreaming,
			GenerationStatus: generationStatusQueued,
		},
		AssistantMessage: &ChatMessage{ID: 61, ConversationID: 41, UserID: "user-123", Status: statusStreaming},
		Prompt:           "new prompt",
	}
	claimStarted := make(chan struct{})
	releaseClaim := make(chan struct{})
	startResult := make(chan error, 1)

	repo.On("ClaimRunGeneration", mock.Anything, prepared.Run, mock.Anything, mock.AnythingOfType("time.Time")).Run(func(args mock.Arguments) {
		close(claimStarted)
		<-releaseClaim
	}).Return(context.Canceled).Once()
	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("StopRun", mock.Anything, int32(41), int32(51), "user-123", mock.AnythingOfType("time.Time")).Return(&StopRunResponse{
		ConversationID: 41,
		RunID:          51,
		MessageID:      61,
		Status:         statusStopped,
	}, nil).Once()

	go func() {
		startResult <- service.StartMessageGeneration(context.Background(), prepared)
	}()
	<-claimStarted

	_, err := service.StopRun(user.WithContext(context.Background(), "user-123"), 41, 51)
	require.NoError(t, err)
	close(releaseClaim)
	require.ErrorIs(t, <-startResult, context.Canceled)
	runtime.AssertNotCalled(t, "StreamChat", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestServiceStartMessageGeneration_RemoteStopCancelsOnOwnershipPoll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:               51,
			ConversationID:   41,
			UserID:           "user-123",
			Status:           statusStreaming,
			GenerationStatus: generationStatusQueued,
		},
		AssistantMessage: &ChatMessage{ID: 61, ConversationID: 41, UserID: "user-123", Status: statusStreaming},
		Prompt:           "new prompt",
	}
	canceled := make(chan struct{})

	repo.On("ClaimRunGeneration", mock.Anything, prepared.Run, mock.Anything, mock.AnythingOfType("time.Time")).Run(func(args mock.Arguments) {
		prepared.Run.GenerationStatus = generationStatusGenerating
		prepared.Run.GenerationOwner = stringPtr(args.Get(2).(runOwner).Value())
	}).Return(nil).Once()
	repo.On("OwnsRunGeneration", mock.Anything, prepared.Run, mock.Anything).Return(false, nil).Once()
	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{}, mock.Anything).Run(func(args mock.Arguments) {
		<-args.Get(0).(context.Context).Done()
		close(canceled)
	}).Return((*StreamDone)(nil), context.Canceled).Once()

	require.NoError(t, service.StartMessageGeneration(context.Background(), prepared))
	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("generation was not canceled after durable ownership was cleared")
	}
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceRecoverStreamingRun_CompletesActiveRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	now := time.Date(2026, 3, 26, 17, 30, 0, 0, time.UTC)
	expiredLease := now.Add(-time.Second)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			GenerationStatus:   generationStatusGenerating,
			LeaseExpiresAt:     &expiredLease,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		History: []ChatMessage{
			{Role: roleUser, Content: "previous user", Status: statusCompleted},
		},
		Prompt: "new prompt",
	}

	runtime.On("Available").Return(true).Once()
	repo.On("LoadPreparedRunForRecovery", mock.Anything, int32(51), "user-123").Return(prepared, nil).Once()
	repo.On("ClaimRunGeneration", mock.Anything, prepared.Run, mock.Anything, mock.AnythingOfType("time.Time")).Run(func(args mock.Arguments) {
		run := args.Get(1).(*ChatRun)
		run.GenerationStatus = generationStatusGenerating
		run.GenerationOwner = stringPtr("inngest:run-51")
		run.UpdatedAt = now
	}).Return(nil).Once()
	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{
		{Role: roleUser, Text: "previous user"},
	}, mock.Anything).Run(func(args mock.Arguments) {
		onChunk := args.Get(3).(func(string) error)
		require.NoError(t, onChunk("hello "))
		require.NoError(t, onChunk("world"))
	}).Return(&StreamDone{
		Model: defaultModelName,
		Text:  "hello world",
	}, nil).Once()
	repo.On("AppendStreamChunk", mock.Anything, prepared, "hello ", "hello", mock.AnythingOfType("time.Time")).Return(int32(1), nil).Once()
	repo.On("AppendStreamChunk", mock.Anything, prepared, "world", "hello world", mock.AnythingOfType("time.Time")).Return(int32(2), nil).Once()
	repo.On("CompleteRun", mock.Anything, prepared, "hello world", (*workout.CreateWorkoutRequest)(nil), mock.AnythingOfType("time.Time")).Return(&ChatMessage{
		ID:             61,
		ConversationID: 41,
		UserID:         "user-123",
		Status:         statusCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	}, &ChatRun{
		ID:                 51,
		ConversationID:     41,
		UserID:             "user-123",
		AssistantMessageID: 61,
		Model:              defaultModelName,
		Status:             statusCompleted,
		CreatedAt:          now,
		UpdatedAt:          now,
		StartedAt:          now,
		CompletedAt:        &now,
	}, nil).Once()

	err := service.RecoverStreamingRun(context.Background(), RunRecoveryRequest{
		ConversationID: 41,
		RunID:          51,
		UserID:         "user-123",
		Reason:         recoverReasonStreamReconnect,
	})

	require.NoError(t, err)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceRecoverStreamingRun_InterruptsStalePartialRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	now := time.Date(2026, 3, 26, 17, 30, 0, 0, time.UTC)
	expiredLease := now.Add(-time.Second)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			GenerationStatus:   generationStatusGenerating,
			LeaseExpiresAt:     &expiredLease,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Content:        "partial answer",
			Status:         statusStreaming,
		},
		Prompt:       "new prompt",
		LastSequence: 1,
	}

	runtime.On("Available").Return(true).Once()
	repo.On("LoadPreparedRunForRecovery", mock.Anything, int32(51), "user-123").Return(prepared, nil).Once()
	repo.On("InterruptRun", mock.Anything, prepared, "partial answer", interruptionReasonStalePartial, mock.AnythingOfType("time.Time")).Return(nil).Once()

	err := service.RecoverStreamingRun(context.Background(), RunRecoveryRequest{
		ConversationID: 41,
		RunID:          51,
		UserID:         "user-123",
		Reason:         recoverReasonStreamReconnect,
	})

	require.NoError(t, err)
	repo.AssertNotCalled(t, "ClaimRunGeneration", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertNotCalled(t, "StreamChat", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceRecoverStreamingRun_InterruptsRunWhenGenerationAttemptsExhausted(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	now := time.Date(2026, 3, 26, 17, 30, 0, 0, time.UTC)
	expiredLease := now.Add(-time.Second)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			GenerationStatus:   generationStatusGenerating,
			LeaseExpiresAt:     &expiredLease,
			GenerationAttempt:  maxGenerationAttempts,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		Prompt: "new prompt",
	}

	runtime.On("Available").Return(true).Once()
	repo.On("LoadPreparedRunForRecovery", mock.Anything, int32(51), "user-123").Return(prepared, nil).Once()
	repo.On("InterruptRun", mock.Anything, prepared, "", interruptionReasonAttemptsExhausted, mock.AnythingOfType("time.Time")).Return(nil).Once()

	err := service.RecoverStreamingRun(context.Background(), RunRecoveryRequest{
		ConversationID: 41,
		RunID:          51,
		UserID:         "user-123",
		Reason:         recoverReasonStreamReconnect,
	})

	require.NoError(t, err)
	repo.AssertNotCalled(t, "ClaimRunGeneration", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertNotCalled(t, "StreamChat", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceRecoverStreamingRun_ReclaimsStaleClaimedRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	now := time.Date(2026, 3, 26, 17, 30, 0, 0, time.UTC)
	expiredLease := now.Add(-time.Second)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			GenerationStatus:   generationStatusGenerating,
			LeaseExpiresAt:     &expiredLease,
			UpdatedAt:          now,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		History: []ChatMessage{
			{Role: roleUser, Content: "previous user", Status: statusCompleted},
		},
		Prompt: "new prompt",
	}

	runtime.On("Available").Return(true).Once()
	repo.On("LoadPreparedRunForRecovery", mock.Anything, int32(51), "user-123").Return(prepared, nil).Once()
	repo.On("ClaimRunGeneration", mock.Anything, prepared.Run, mock.Anything, mock.AnythingOfType("time.Time")).Run(func(args mock.Arguments) {
		run := args.Get(1).(*ChatRun)
		run.GenerationStatus = generationStatusGenerating
		run.GenerationOwner = stringPtr("inngest:run-51")
		run.UpdatedAt = now
	}).Return(nil).Once()
	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{
		{Role: roleUser, Text: "previous user"},
	}, mock.Anything).Run(func(args mock.Arguments) {
		onChunk := args.Get(3).(func(string) error)
		require.NoError(t, onChunk("hello "))
		require.NoError(t, onChunk("world"))
	}).Return(&StreamDone{
		Model: defaultModelName,
		Text:  "hello world",
	}, nil).Once()
	repo.On("AppendStreamChunk", mock.Anything, prepared, "hello ", "hello", mock.AnythingOfType("time.Time")).Return(int32(1), nil).Once()
	repo.On("AppendStreamChunk", mock.Anything, prepared, "world", "hello world", mock.AnythingOfType("time.Time")).Return(int32(2), nil).Once()
	repo.On("CompleteRun", mock.Anything, prepared, "hello world", (*workout.CreateWorkoutRequest)(nil), mock.AnythingOfType("time.Time")).Return(&ChatMessage{
		ID:             61,
		ConversationID: 41,
		UserID:         "user-123",
		Status:         statusCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	}, &ChatRun{
		ID:                 51,
		ConversationID:     41,
		UserID:             "user-123",
		AssistantMessageID: 61,
		Model:              defaultModelName,
		Status:             statusCompleted,
		CreatedAt:          now,
		UpdatedAt:          now,
		StartedAt:          now,
		CompletedAt:        &now,
	}, nil).Once()

	err := service.RecoverStreamingRun(context.Background(), RunRecoveryRequest{
		ConversationID: 41,
		RunID:          51,
		UserID:         "user-123",
		Reason:         recoverReasonStreamReconnect,
	})

	require.NoError(t, err)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceRecoverStreamingRun_SkipsFreshClaimedRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	validLease := time.Now().UTC().Add(generationLeaseDuration)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			GenerationStatus:   generationStatusGenerating,
			LeaseExpiresAt:     &validLease,
			UpdatedAt:          time.Now().UTC(),
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		Prompt: "new prompt",
	}

	runtime.On("Available").Return(true).Once()
	repo.On("LoadPreparedRunForRecovery", mock.Anything, int32(51), "user-123").Return(prepared, nil).Once()

	err := service.RecoverStreamingRun(context.Background(), RunRecoveryRequest{
		ConversationID: 41,
		RunID:          51,
		UserID:         "user-123",
		Reason:         recoverReasonStreamReconnect,
	})

	require.NoError(t, err)
	repo.AssertNotCalled(t, "ClaimRunGeneration", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
	runtime.AssertExpectations(t)
}

func TestServiceResumeMessageStream_ReplaysWhileRunActiveAndCompletes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	now := time.Date(2026, 3, 26, 17, 31, 0, 0, time.UTC)
	prepared := &PreparedResumeStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Content:        "hello",
			Status:         statusStreaming,
		},
		AfterSequence: 1,
		LastSequence:  1,
	}
	completed := &PreparedResumeStream{
		Conversation: prepared.Conversation,
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusCompleted,
			WorkoutDraft: &workout.CreateWorkoutRequest{
				Date: "2026-03-26T17:31:00Z",
				Exercises: []workout.ExerciseInput{
					{
						Name: "Bench Press",
						Sets: []workout.SetInput{
							{Reps: 8, SetType: "working"},
						},
					},
				},
			},
			CompletedAt: &now,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Content:        "hello world",
			Status:         statusCompleted,
			CompletedAt:    &now,
		},
		AfterSequence: 4,
		LastSequence:  4,
	}

	repo.On("ListStreamChunksAfter", mock.Anything, int32(51), "user-123", int32(1)).Return([]StreamChunk{
		{Delta: " ", Sequence: 2},
		{Delta: "world", Sequence: 4},
	}, nil).Once()
	repo.On("LoadPreparedRunForResume", mock.Anything, int32(41), int32(51), "user-123", int32(4)).Return(completed, nil).Once()

	var seen []StreamChunk
	done, err := service.ResumeMessageStream(context.Background(), prepared, func(chunk StreamChunk) error {
		seen = append(seen, chunk)
		return nil
	})

	require.NoError(t, err)
	require.NotNil(t, done)
	assert.Equal(t, []StreamChunk{
		{Delta: " ", Sequence: 2},
		{Delta: "world", Sequence: 4},
	}, seen)
	assert.Equal(t, "hello world", done.Text)
	assert.Equal(t, int32(4), done.Sequence)
	require.NotNil(t, done.WorkoutDraft)
	assert.Equal(t, "2026-03-26T17:31:00Z", done.WorkoutDraft.Date)
	repo.AssertExpectations(t)
}

func TestServiceResumeMessageStream_EndsWhenRunAwaitsRecovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	expiredLease := time.Now().UTC().Add(-time.Second)
	prepared := &PreparedResumeStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			GenerationStatus:   generationStatusGenerating,
			LeaseExpiresAt:     &expiredLease,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Content:        "partial",
			Status:         statusStreaming,
		},
		AfterSequence: 1,
		LastSequence:  1,
	}
	awaitingRecovery := &PreparedResumeStream{
		Conversation: prepared.Conversation,
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			GenerationStatus:   generationStatusGenerating,
			LeaseExpiresAt:     &expiredLease,
		},
		AssistantMessage: prepared.AssistantMessage,
		AfterSequence:    1,
		LastSequence:     1,
	}

	repo.On("ListStreamChunksAfter", mock.Anything, int32(51), "user-123", int32(1)).Return([]StreamChunk{
		{Delta: " more", Sequence: 2},
	}, nil).Once()
	repo.On("LoadPreparedRunForResume", mock.Anything, int32(41), int32(51), "user-123", int32(2)).Return(awaitingRecovery, nil).Once()

	var seen []StreamChunk
	done, err := service.ResumeMessageStream(context.Background(), prepared, func(chunk StreamChunk) error {
		seen = append(seen, chunk)
		return nil
	})

	require.ErrorIs(t, err, ErrStreamAwaitingRecovery)
	assert.Nil(t, done)
	assert.Equal(t, []StreamChunk{{Delta: " more", Sequence: 2}}, seen)
	repo.AssertExpectations(t)
}

func TestServiceResumeMessageStream_EndsWhenClaimedRunTurnsStale(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	now := time.Now().UTC()
	validLease := now.Add(generationLeaseDuration)
	expiredLease := now.Add(-time.Second)
	prepared := &PreparedResumeStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			GenerationStatus:   generationStatusGenerating,
			LeaseExpiresAt:     &validLease,
			UpdatedAt:          now,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Content:        "partial",
			Status:         statusStreaming,
		},
		AfterSequence: 1,
		LastSequence:  1,
	}
	staleClaimed := &PreparedResumeStream{
		Conversation: prepared.Conversation,
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			GenerationStatus:   generationStatusGenerating,
			LeaseExpiresAt:     &expiredLease,
			UpdatedAt:          now,
		},
		AssistantMessage: prepared.AssistantMessage,
		AfterSequence:    1,
		LastSequence:     1,
	}

	repo.On("ListStreamChunksAfter", mock.Anything, int32(51), "user-123", int32(1)).Return([]StreamChunk{}, nil).Once()
	repo.On("LoadPreparedRunForResume", mock.Anything, int32(41), int32(51), "user-123", int32(1)).Return(staleClaimed, nil).Once()

	done, err := service.ResumeMessageStream(context.Background(), prepared, func(StreamChunk) error {
		return nil
	})

	require.ErrorIs(t, err, ErrStreamAwaitingRecovery)
	assert.Nil(t, done)
	repo.AssertExpectations(t)
}

func TestServiceRecoverStreamingRun_IgnoresRunsWithoutRecoveryMarker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:             51,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		Prompt: "new prompt",
	}

	runtime.On("Available").Return(true).Once()
	repo.On("LoadPreparedRunForRecovery", mock.Anything, int32(51), "user-123").Return(prepared, nil).Once()

	err := service.RecoverStreamingRun(context.Background(), RunRecoveryRequest{
		ConversationID: 41,
		RunID:          51,
		UserID:         "user-123",
		Reason:         recoverReasonStreamReconnect,
	})

	require.NoError(t, err)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceAbortPreparedMessageStream_PersistsFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:             51,
			ConversationID: 41,
			UserID:         "user-123",
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
	}

	repo.On(
		"FailRun",
		mock.MatchedBy(func(ctx context.Context) bool { return ctx.Err() == nil }),
		prepared,
		"",
		ErrStreamNotStarted,
		mock.AnythingOfType("time.Time"),
	).Return(nil).Once()

	err := service.AbortPreparedMessageStream(context.Background(), prepared, ErrStreamNotStarted)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceStreamMessage_PersistsCompletionWhenRequestContextCanceled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	now := time.Date(2026, 3, 26, 17, 40, 0, 0, time.UTC)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		Prompt: "new prompt",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{}, mock.Anything).
		Return(&StreamDone{
			Model: defaultModelName,
			Text:  "hello world",
		}, nil).Once()
	repo.On(
		"CompleteRun",
		mock.MatchedBy(func(ctx context.Context) bool { return ctx.Err() == nil }),
		prepared,
		"hello world",
		(*workout.CreateWorkoutRequest)(nil),
		mock.AnythingOfType("time.Time"),
	).Return(&ChatMessage{
		ID:             61,
		ConversationID: 41,
		UserID:         "user-123",
		Status:         statusCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    &now,
	}, &ChatRun{
		ID:                 51,
		ConversationID:     41,
		UserID:             "user-123",
		AssistantMessageID: 61,
		Model:              defaultModelName,
		Status:             statusCompleted,
		CreatedAt:          now,
		UpdatedAt:          now,
		StartedAt:          now,
		CompletedAt:        &now,
	}, nil).Once()

	done, err := service.StreamMessage(ctx, prepared, func(StreamChunk) error { return nil })

	require.NoError(t, err)
	require.NotNil(t, done)
	assert.Equal(t, "hello world", done.Text)
	repo.AssertNotCalled(t, "FailRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceStreamMessage_PersistsFailureWhenRequestContextCanceled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo, nil)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Status:         statusStreaming,
		},
		Prompt: "new prompt",
	}
	expectedErr := errors.New("provider failed")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{}, mock.Anything).Run(func(args mock.Arguments) {
		onChunk := args.Get(3).(func(string) error)
		_ = onChunk("partial ")
	}).Return((*StreamDone)(nil), expectedErr).Once()
	repo.On("AppendStreamChunk", mock.Anything, prepared, "partial ", "partial", mock.AnythingOfType("time.Time")).Return(int32(1), nil).Once()
	repo.On(
		"FailRun",
		mock.MatchedBy(func(ctx context.Context) bool { return ctx.Err() == nil }),
		prepared,
		"partial",
		expectedErr,
		mock.AnythingOfType("time.Time"),
	).Return(nil).Once()

	done, err := service.StreamMessage(ctx, prepared, func(StreamChunk) error { return nil })

	require.Error(t, err)
	assert.Nil(t, done)
	assert.ErrorIs(t, err, expectedErr)
	repo.AssertNotCalled(t, "CompleteRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestNormalizeStreamChunkError_MapsDisconnects(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := normalizeStreamChunkError(ctx, errors.New("write tcp: broken pipe"))

	require.ErrorIs(t, err, ErrStreamDisconnected)
}

func stringPtr(value string) *string {
	return &value
}
