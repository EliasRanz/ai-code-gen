// Package ai contains tests for AI application use cases  
package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/domain/ai"
	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
)

// Mock implementations for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveGeneration(ctx context.Context, generation ai.GenerationHistory) error {
	args := m.Called(ctx, generation)
	return args.Error(0)
}

func (m *MockRepository) GetHistory(ctx context.Context, userID common.UserID, limit int) ([]ai.GenerationHistory, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]ai.GenerationHistory), args.Error(1)
}

func (m *MockRepository) GetQuotaUsage(ctx context.Context, userID common.UserID) (ai.QuotaStatus, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(ai.QuotaStatus), args.Error(1)
}

func (m *MockRepository) UpdateQuotaUsage(ctx context.Context, userID common.UserID, tokens int) error {
	args := m.Called(ctx, userID, tokens)
	return args.Error(0)
}

type MockLLMService struct {
	mock.Mock
}

func (m *MockLLMService) Generate(ctx context.Context, req ai.GenerationRequest) (ai.GenerationResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(ai.GenerationResult), args.Error(1)
}

func (m *MockLLMService) GenerateStream(ctx context.Context, req ai.GenerationRequest, ch chan<- ai.StreamChunk) error {
	args := m.Called(ctx, req, ch)
	return args.Error(0)
}

func (m *MockLLMService) Stream(ctx context.Context, req ai.GenerationRequest, ch chan<- string) error {
	args := m.Called(ctx, req, ch)
	return args.Error(0)
}

func (m *MockLLMService) Validate(ctx context.Context, code string) (ai.ValidationResult, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(ai.ValidationResult), args.Error(1)
}

type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) Allow(userID common.UserID) bool {
	args := m.Called(userID)
	return args.Bool(0)
}

func (m *MockRateLimiter) Reset(userID common.UserID) {
	m.Called(userID)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishGenerationEvent(ctx context.Context, userID common.UserID, tokens int) error {
	args := m.Called(ctx, userID, tokens)
	return args.Error(0)
}

func TestStreamCodeUseCase_Execute_ModelNameCapture(t *testing.T) {
	ctx := context.Background()
	userID := common.UserID("test-user")

	mockRepo := new(MockRepository)
	mockLLM := new(MockLLMService)
	mockRateLimiter := new(MockRateLimiter)
	mockPublisher := new(MockEventPublisher)

	useCase := NewStreamCodeUseCase(mockRepo, mockLLM, mockRateLimiter, mockPublisher)

	request := StreamCodeRequest{
		Prompt:     "Generate a React component",
		Language:   "javascript",
		Framework:  "react",
		Complexity: "simple",
		UserID:     userID,
	}

	t.Run("should capture model name from stream chunks", func(t *testing.T) {
		responseChan := make(chan StreamCodeResponse, 10)

		// Setup quota check
		quota := ai.QuotaStatus{
			UserID:     userID,
			DailyLimit: 1000,
			UsedToday:  100,
			Remaining:  900,
			ResetTime:  "2024-01-01T00:00:00Z",
		}
		mockRepo.On("GetQuotaUsage", ctx, userID).Return(quota, nil)

		// Setup rate limiter
		mockRateLimiter.On("Allow", userID).Return(true)

		// Setup streaming response with model name
		mockLLM.On("GenerateStream", ctx, mock.AnythingOfType("ai.GenerationRequest"), mock.AnythingOfType("chan<- ai.StreamChunk")).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- ai.StreamChunk)

			// Send chunks with model information
			chunks := []ai.StreamChunk{
				{
					Content:    "const ",
					TokenCount: 1,
					Model:      "gpt-4-turbo", // Model name in first chunk
					IsComplete: false,
				},
				{
					Content:    "MyComponent = () => {\n",
					TokenCount: 3,
					Model:      "gpt-4-turbo", // Model name repeated
					IsComplete: false,
				},
				{
					Content:    "  return <div>Hello World</div>;\n};",
					TokenCount: 8,
					Model:      "gpt-4-turbo",
					IsComplete: true,
				},
			}

			for _, chunk := range chunks {
				ch <- chunk
			}
		}).Return(nil)

		// Setup repository calls
		var capturedHistory ai.GenerationHistory
		mockRepo.On("SaveGeneration", ctx, mock.AnythingOfType("ai.GenerationHistory")).Run(func(args mock.Arguments) {
			capturedHistory = args.Get(1).(ai.GenerationHistory)
		}).Return(nil)

		mockRepo.On("UpdateQuotaUsage", ctx, userID, 12).Return(nil)
		mockPublisher.On("PublishGenerationEvent", ctx, userID, 12).Return(nil)

		// Execute
		go func() {
			defer close(responseChan)
			err := useCase.Execute(ctx, request, responseChan)
			require.NoError(t, err)
		}()

		// Collect responses
		var responses []StreamCodeResponse
		for response := range responseChan {
			responses = append(responses, response)
		}

		// Verify streaming responses
		assert.Len(t, responses, 4) // 3 chunks + 1 complete
		assert.Equal(t, "chunk", responses[0].Type)
		assert.Equal(t, "const ", responses[0].Content)
		assert.Equal(t, "complete", responses[3].Type)
		assert.True(t, responses[3].IsComplete)

		// Verify model name was captured correctly
		assert.Equal(t, "gpt-4-turbo", capturedHistory.Model)
		assert.Equal(t, userID, capturedHistory.UserID)
		assert.Equal(t, request.Prompt, capturedHistory.Prompt)
		assert.Equal(t, "const MyComponent = () => {\n  return <div>Hello World</div>;\n};", capturedHistory.Code)
		assert.Equal(t, 12, capturedHistory.Tokens)

		// Verify all expectations
		mockRepo.AssertExpectations(t)
		mockLLM.AssertExpectations(t)
		mockRateLimiter.AssertExpectations(t)
		mockPublisher.AssertExpectations(t)
	})

	t.Run("should use fallback model name when not provided in chunks", func(t *testing.T) {
		responseChan := make(chan StreamCodeResponse, 10)

		// Reset mocks
		mockRepo = new(MockRepository)
		mockLLM = new(MockLLMService)
		mockRateLimiter = new(MockRateLimiter)
		mockPublisher = new(MockEventPublisher)

		useCase = NewStreamCodeUseCase(mockRepo, mockLLM, mockRateLimiter, mockPublisher)

		// Setup quota check
		quota := ai.QuotaStatus{
			UserID:     userID,
			DailyLimit: 1000,
			UsedToday:  100,
			Remaining:  900,
			ResetTime:  "2024-01-01T00:00:00Z",
		}
		mockRepo.On("GetQuotaUsage", ctx, userID).Return(quota, nil)

		// Setup rate limiter
		mockRateLimiter.On("Allow", userID).Return(true)

		// Setup streaming response WITHOUT model name
		mockLLM.On("GenerateStream", ctx, mock.AnythingOfType("ai.GenerationRequest"), mock.AnythingOfType("chan<- ai.StreamChunk")).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- ai.StreamChunk)

			// Send chunks without model information
			chunks := []ai.StreamChunk{
				{
					Content:    "console.log('hello');",
					TokenCount: 5,
					// Model field is empty
					IsComplete: true,
				},
			}

			for _, chunk := range chunks {
				ch <- chunk
			}
		}).Return(nil)

		// Setup repository calls
		var capturedHistory ai.GenerationHistory
		mockRepo.On("SaveGeneration", ctx, mock.AnythingOfType("ai.GenerationHistory")).Run(func(args mock.Arguments) {
			capturedHistory = args.Get(1).(ai.GenerationHistory)
		}).Return(nil)

		mockRepo.On("UpdateQuotaUsage", ctx, userID, 5).Return(nil)
		mockPublisher.On("PublishGenerationEvent", ctx, userID, 5).Return(nil)

		// Execute
		go func() {
			defer close(responseChan)
			err := useCase.Execute(ctx, request, responseChan)
			require.NoError(t, err)
		}()

		// Collect responses
		var responses []StreamCodeResponse
		for response := range responseChan {
			responses = append(responses, response)
		}

		// Verify fallback model name was used
		assert.Equal(t, "unknown-model", capturedHistory.Model)
		assert.Equal(t, "console.log('hello');", capturedHistory.Code)

		// Verify all expectations
		mockRepo.AssertExpectations(t)
		mockLLM.AssertExpectations(t)
		mockRateLimiter.AssertExpectations(t)
		mockPublisher.AssertExpectations(t)
	})

	t.Run("should capture model from first chunk with model info", func(t *testing.T) {
		responseChan := make(chan StreamCodeResponse, 10)

		// Reset mocks
		mockRepo = new(MockRepository)
		mockLLM = new(MockLLMService)
		mockRateLimiter = new(MockRateLimiter)
		mockPublisher = new(MockEventPublisher)

		useCase = NewStreamCodeUseCase(mockRepo, mockLLM, mockRateLimiter, mockPublisher)

		// Setup quota check
		quota := ai.QuotaStatus{
			UserID:     userID,
			DailyLimit: 1000,
			UsedToday:  100,
			Remaining:  900,
			ResetTime:  "2024-01-01T00:00:00Z",
		}
		mockRepo.On("GetQuotaUsage", ctx, userID).Return(quota, nil)

		// Setup rate limiter
		mockRateLimiter.On("Allow", userID).Return(true)

		// Setup streaming response with mixed model availability
		mockLLM.On("GenerateStream", ctx, mock.AnythingOfType("ai.GenerationRequest"), mock.AnythingOfType("chan<- ai.StreamChunk")).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- ai.StreamChunk)

			// First chunk has no model, second has model
			chunks := []ai.StreamChunk{
				{
					Content:    "function ",
					TokenCount: 1,
					// No model
					IsComplete: false,
				},
				{
					Content:    "test() { return true; }",
					TokenCount: 6,
					Model:      "claude-3-sonnet", // Model appears here
					IsComplete: true,
				},
			}

			for _, chunk := range chunks {
				ch <- chunk
			}
		}).Return(nil)

		// Setup repository calls
		var capturedHistory ai.GenerationHistory
		mockRepo.On("SaveGeneration", ctx, mock.AnythingOfType("ai.GenerationHistory")).Run(func(args mock.Arguments) {
			capturedHistory = args.Get(1).(ai.GenerationHistory)
		}).Return(nil)

		mockRepo.On("UpdateQuotaUsage", ctx, userID, 7).Return(nil)
		mockPublisher.On("PublishGenerationEvent", ctx, userID, 7).Return(nil)

		// Execute
		go func() {
			defer close(responseChan)
			err := useCase.Execute(ctx, request, responseChan)
			require.NoError(t, err)
		}()

		// Collect responses
		var responses []StreamCodeResponse
		for response := range responseChan {
			responses = append(responses, response)
		}

		// Verify model name from second chunk was captured
		assert.Equal(t, "claude-3-sonnet", capturedHistory.Model)
		assert.Equal(t, "function test() { return true; }", capturedHistory.Code)

		// Verify all expectations
		mockRepo.AssertExpectations(t)
		mockLLM.AssertExpectations(t)
		mockRateLimiter.AssertExpectations(t)
		mockPublisher.AssertExpectations(t)
	})
}

func TestStreamCodeUseCase_Execute_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	userID := common.UserID("test-user")

	t.Run("should handle streaming error with proper model fallback", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockLLM := new(MockLLMService)
		mockRateLimiter := new(MockRateLimiter)
		mockPublisher := new(MockEventPublisher)

		useCase := NewStreamCodeUseCase(mockRepo, mockLLM, mockRateLimiter, mockPublisher)

		request := StreamCodeRequest{
			Prompt:     "Generate code",
			Language:   "javascript",
			Complexity: "simple",
			UserID:     userID,
		}

		responseChan := make(chan StreamCodeResponse, 10)

		// Setup quota check
		quota := ai.QuotaStatus{
			UserID:     userID,
			DailyLimit: 1000,
			UsedToday:  100,
			Remaining:  900,
			ResetTime:  "2024-01-01T00:00:00Z",
		}
		mockRepo.On("GetQuotaUsage", ctx, userID).Return(quota, nil)

		// Setup rate limiter
		mockRateLimiter.On("Allow", userID).Return(true)

		// Setup streaming response with error
		mockLLM.On("GenerateStream", ctx, mock.AnythingOfType("ai.GenerationRequest"), mock.AnythingOfType("chan<- ai.StreamChunk")).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- ai.StreamChunk)

			// Send one successful chunk, then error
			ch <- ai.StreamChunk{
				Content:    "const x = ",
				TokenCount: 3,
				Model:      "test-model",
				IsComplete: false,
			}

			ch <- ai.StreamChunk{
				Error: assert.AnError,
			}
		}).Return(nil)

		// Execute
		go func() {
			defer close(responseChan)
			err := useCase.Execute(ctx, request, responseChan)
			assert.Error(t, err)
		}()

		// Collect responses
		var responses []StreamCodeResponse
		for response := range responseChan {
			responses = append(responses, response)
		}

		// Should get one successful chunk and one error
		assert.Len(t, responses, 2)
		assert.Equal(t, "chunk", responses[0].Type)
		assert.Equal(t, "const x = ", responses[0].Content)
		assert.Equal(t, "error", responses[1].Type)
		assert.NotEmpty(t, responses[1].Error)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLLM.AssertExpectations(t)
		mockRateLimiter.AssertExpectations(t)
	})
}
