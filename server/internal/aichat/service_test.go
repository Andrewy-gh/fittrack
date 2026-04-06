package aichat

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
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

func (m *mockRepository) LoadPreparedRunForRecovery(ctx context.Context, runID int32, userID string) (*PreparedMessageStream, error) {
	args := m.Called(ctx, runID, userID)
	prepared, _ := args.Get(0).(*PreparedMessageStream)
	return prepared, args.Error(1)
}

func (m *mockRepository) PrepareMessageStream(ctx context.Context, conversationID int32, userID string, prompt string, model string, requestID string) (*PreparedMessageStream, error) {
	args := m.Called(ctx, conversationID, userID, prompt, model, requestID)
	prepared, _ := args.Get(0).(*PreparedMessageStream)
	return prepared, args.Error(1)
}

func (m *mockRepository) UpdateStreamingRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, updatedAt time.Time) error {
	args := m.Called(ctx, prepared, partialText, updatedAt)
	return args.Error(0)
}

func (m *mockRepository) MarkRunAwaitingRecovery(ctx context.Context, prepared *PreparedMessageStream, partialText string, updatedAt time.Time) error {
	args := m.Called(ctx, prepared, partialText, updatedAt)
	return args.Error(0)
}

func (m *mockRepository) ClaimRunRecovery(ctx context.Context, runID int32, userID string) error {
	args := m.Called(ctx, runID, userID)
	return args.Error(0)
}

func (m *mockRepository) CompleteRun(ctx context.Context, prepared *PreparedMessageStream, assistantText string, completedAt time.Time) (*ChatMessage, *ChatRun, error) {
	args := m.Called(ctx, prepared, assistantText, completedAt)
	message, _ := args.Get(0).(*ChatMessage)
	run, _ := args.Get(1).(*ChatRun)
	return message, run, args.Error(2)
}

func (m *mockRepository) FailRun(ctx context.Context, prepared *PreparedMessageStream, partialText string, failure error, completedAt time.Time) error {
	args := m.Called(ctx, prepared, partialText, failure, completedAt)
	return args.Error(0)
}

type mockRecoveryDispatcher struct {
	mock.Mock
}

func (m *mockRecoveryDispatcher) EnqueueRunRecovery(ctx context.Context, request RunRecoveryRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func TestServiceValidate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("returns feature-disabled error when user lacks ai access", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		runtime := new(mockRuntime)
		repo := new(mockRepository)
		service := NewService(logger, featureAccess, runtime, repo)

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
		service := NewService(logger, featureAccess, runtime, repo)

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
		service := NewService(logger, featureAccess, runtime, repo)

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
		service := NewService(logger, featureAccess, runtime, repo)
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
		service := NewService(logger, featureAccess, runtime, repo)
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
	service := NewService(logger, featureAccess, runtime, repo)
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

func TestServiceGetConversation_AllowsReadWithoutRuntime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo)
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

	detail, err := service.GetConversation(ctx, 41)

	require.NoError(t, err)
	require.NotNil(t, detail)
	require.NotNil(t, detail.Conversation)
	assert.Len(t, detail.Messages, 1)
	runtime.AssertNotCalled(t, "Available")
	featureAccess.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServicePrepareMessageStream_RequiresRuntime(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo)
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

func TestServiceRequestMessageRecovery_QueuesActiveRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	recovery := new(mockRecoveryDispatcher)
	service := NewService(logger, featureAccess, runtime, repo)
	service.SetRecoveryDispatcher(recovery)
	ctx := user.WithContext(context.Background(), "user-123")

	featureAccess.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil).Once()
	repo.On("GetConversation", mock.Anything, int32(41), "user-123").Return(&Conversation{ID: 41, UserID: "user-123"}, nil).Once()
	repo.On("GetActiveRunForConversation", mock.Anything, int32(41), "user-123").Return(&ChatRun{
		ID:             51,
		ConversationID: 41,
		UserID:         "user-123",
		Status:         statusStreaming,
		ErrorMessage:   stringPtr(runAwaitingRecoveryMarker),
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
	service := NewService(logger, featureAccess, runtime, repo)
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
	service := NewService(logger, featureAccess, runtime, repo)
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
	service := NewService(logger, featureAccess, runtime, repo)
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
	service := NewService(logger, featureAccess, runtime, repo)
	now := time.Date(2026, 3, 26, 17, 30, 0, 0, time.UTC)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			ErrorMessage:       stringPtr(runAwaitingRecoveryMarker),
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

	runtime.On("StreamChat", mock.Anything, "new prompt", []RuntimeChatMessage{
		{Role: roleUser, Text: "previous user"},
		{Role: roleAssistant, Text: "previous assistant"},
	}, mock.Anything).Run(func(args mock.Arguments) {
		onChunk := args.Get(3).(func(string) error)
		_ = onChunk("hello ")
		_ = onChunk("world")
	}).Return(&StreamDone{
		Model: defaultModelName,
		Text:  "hello world",
	}, nil).Once()
	repo.On("UpdateStreamingRun", mock.Anything, prepared, "hello", mock.AnythingOfType("time.Time")).Return(nil).Once()
	repo.On("CompleteRun", mock.Anything, prepared, "hello world", mock.AnythingOfType("time.Time")).Return(&ChatMessage{
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

	done, err := service.StreamMessage(context.Background(), prepared, func(string) error { return nil })

	require.NoError(t, err)
	require.NotNil(t, done)
	assert.Equal(t, "hello world", done.Text)
	assert.Equal(t, int32(51), done.RunID)
	repo.AssertNotCalled(t, "FailRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceStreamMessage_LeavesRunStreamingOnDisconnect(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			ErrorMessage:       stringPtr(runAwaitingRecoveryMarker),
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
	repo.On("UpdateStreamingRun", mock.Anything, prepared, "partial", mock.AnythingOfType("time.Time")).Return(nil).Once()
	repo.On("MarkRunAwaitingRecovery", mock.Anything, prepared, "partial", mock.AnythingOfType("time.Time")).Return(nil).Once()

	done, err := service.StreamMessage(context.Background(), prepared, func(string) error {
		return errors.New("broken pipe")
	})

	require.Error(t, err)
	assert.Nil(t, done)
	assert.ErrorIs(t, err, ErrStreamAwaitingRecovery)
	repo.AssertNotCalled(t, "FailRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "CompleteRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceStreamMessage_FailsRunOnRuntimeError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			ErrorMessage:       stringPtr(runAwaitingRecoveryMarker),
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
	repo.On("UpdateStreamingRun", mock.Anything, prepared, "partial", mock.AnythingOfType("time.Time")).Return(nil).Once()
	repo.On("FailRun", mock.Anything, prepared, "partial", expectedErr, mock.AnythingOfType("time.Time")).Return(nil).Once()

	done, err := service.StreamMessage(context.Background(), prepared, func(string) error { return nil })

	require.Error(t, err)
	assert.Nil(t, done)
	assert.ErrorIs(t, err, expectedErr)
	repo.AssertNotCalled(t, "CompleteRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceRecoverStreamingRun_CompletesActiveRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo)
	now := time.Date(2026, 3, 26, 17, 30, 0, 0, time.UTC)
	prepared := &PreparedMessageStream{
		Conversation: &Conversation{ID: 41, UserID: "user-123"},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			ErrorMessage:       stringPtr(runAwaitingRecoveryMarker),
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
	repo.On("ClaimRunRecovery", mock.Anything, int32(51), "user-123").Return(nil).Once()
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
	repo.On("UpdateStreamingRun", mock.Anything, prepared, "hello", mock.AnythingOfType("time.Time")).Return(nil).Once()
	repo.On("CompleteRun", mock.Anything, prepared, "hello world", mock.AnythingOfType("time.Time")).Return(&ChatMessage{
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

func TestServiceRecoverStreamingRun_IgnoresRunsWithoutRecoveryMarker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo)
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
	repo.AssertNotCalled(t, "ClaimRunRecovery", mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestServiceAbortPreparedMessageStream_PersistsFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	featureAccess := new(mockFeatureAccessService)
	runtime := new(mockRuntime)
	repo := new(mockRepository)
	service := NewService(logger, featureAccess, runtime, repo)
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
	service := NewService(logger, featureAccess, runtime, repo)
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

	done, err := service.StreamMessage(ctx, prepared, func(string) error { return nil })

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
	service := NewService(logger, featureAccess, runtime, repo)
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
	repo.On("UpdateStreamingRun", mock.Anything, prepared, "partial", mock.AnythingOfType("time.Time")).Return(nil).Once()
	repo.On(
		"FailRun",
		mock.MatchedBy(func(ctx context.Context) bool { return ctx.Err() == nil }),
		prepared,
		"partial",
		expectedErr,
		mock.AnythingOfType("time.Time"),
	).Return(nil).Once()

	done, err := service.StreamMessage(ctx, prepared, func(string) error { return nil })

	require.Error(t, err)
	assert.Nil(t, done)
	assert.ErrorIs(t, err, expectedErr)
	repo.AssertNotCalled(t, "CompleteRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	runtime.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestStreamProgressSink_ThrottlesPersistence(t *testing.T) {
	repo := new(mockRepository)
	prepared := &PreparedMessageStream{
		Conversation:     &Conversation{ID: 41, UserID: "user-123"},
		Run:              &ChatRun{ID: 51, UserID: "user-123"},
		AssistantMessage: &ChatMessage{ID: 61, UserID: "user-123"},
	}
	base := time.Date(2026, 3, 26, 18, 0, 0, 0, time.UTC)
	sink := newStreamProgressSink(context.Background(), repo, prepared)
	times := []time.Time{
		base,
		base.Add(500 * time.Millisecond),
		base.Add(1500 * time.Millisecond),
	}
	sink.now = func() time.Time {
		next := times[0]
		times = times[1:]
		return next
	}

	repo.On("UpdateStreamingRun", mock.Anything, prepared, "hello", base).Return(nil).Once()
	repo.On("UpdateStreamingRun", mock.Anything, prepared, "hello world", base.Add(1500*time.Millisecond)).Return(nil).Once()

	require.NoError(t, sink.maybePersist("hello"))
	require.NoError(t, sink.maybePersist("hello again"))
	require.NoError(t, sink.maybePersist("hello world"))
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
