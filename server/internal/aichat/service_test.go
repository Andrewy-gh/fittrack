package aichat

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
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

func TestServiceValidate(t *testing.T) {
	t.Run("returns feature-disabled error when user lacks ai access", func(t *testing.T) {
		featureAccess := new(mockFeatureAccessService)
		runtime := new(mockRuntime)
		service := NewService(featureAccess, runtime)

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
		service := NewService(featureAccess, runtime)

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
		service := NewService(featureAccess, runtime)

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
		service := NewService(featureAccess, runtime)
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
		service := NewService(featureAccess, runtime)
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
