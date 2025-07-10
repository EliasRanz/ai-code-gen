package generation_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/generation"
	"github.com/EliasRanz/ai-code-gen/internal/llm"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient for testing
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

// MockLLMClient for testing
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) Generate(ctx context.Context, req *llm.GenerationRequest) (*llm.GenerationResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*llm.GenerationResponse), args.Error(1)
}

func (m *MockLLMClient) GenerateStream(ctx context.Context, req *llm.GenerationRequest) (<-chan *llm.GenerationResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(<-chan *llm.GenerationResponse), args.Error(1)
}

func (m *MockLLMClient) GetModels(ctx context.Context) ([]llm.Model, error) {
	args := m.Called(ctx)
	return args.Get(0).([]llm.Model), args.Error(1)
}

func (m *MockLLMClient) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockLLMClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		llmClient   llm.LLMClient
		redisClient generation.RedisClient
		expectError bool
	}{
		{
			name:        "service with valid clients",
			llmClient:   &MockLLMClient{},
			redisClient: &MockRedisClient{},
			expectError: false,
		},
		{
			name:        "service with nil redis client",
			llmClient:   &MockLLMClient{},
			redisClient: nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := generation.NewService(tt.llmClient, tt.redisClient)

			assert.NotNil(t, service)
			// Note: Fields are private, so we can't test them directly
			// Instead we should test the service behavior
		})
	}
}

func TestStreamGenerationHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful stream request",
			setupContext: func(c *gin.Context) {
				c.Set("authenticated", true)
				c.Set("user_id", "test-user")
			},
			requestBody:    `{"model": "test-model", "prompt": "test prompt"}`,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name: "invalid request body",
			setupContext: func(c *gin.Context) {
				c.Set("authenticated", true)
				c.Set("user_id", "test-user")
			},
			requestBody:    `{"invalid": "json"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Error:Field validation for 'Model' failed on the 'required' tag",
		},
		{
			name: "unauthorized - no user context",
			setupContext: func(c *gin.Context) {
				// Don't set any user context
			},
			requestBody:    `{"model": "test-model", "prompt": "test prompt"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service with mocks
			mockLLM := new(MockLLMClient)
			mockRedis := new(MockRedisClient)

			// Set up mock expectations for streaming (even for error cases)
			respChan := make(chan *llm.GenerationResponse, 1)
			close(respChan) // Close immediately to simulate no responses
			mockLLM.On("GenerateStream", mock.Anything, mock.AnythingOfType("*llm.GenerationRequest")).Return((<-chan *llm.GenerationResponse)(respChan), nil).Maybe()

			service := generation.NewService(mockLLM, mockRedis)

			// Create test request
			r := gin.New()
			r.POST("/generate/stream", func(c *gin.Context) {
				tt.setupContext(c)
				service.StreamGenerationHandler(c)
			})

			req := httptest.NewRequest("POST", "/generate/stream", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestRequestResponseHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		setupMocks     func(*MockLLMClient)
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful generation",
			setupContext: func(c *gin.Context) {
				c.Set("authenticated", true)
				c.Set("user_id", "test-user")
			},
			setupMocks: func(mockLLM *MockLLMClient) {
				resp := &llm.GenerationResponse{
					ID: "test-response",
					Choices: []llm.Choice{
						{
							Text:         "Generated text",
							FinishReason: nil,
						},
					},
				}
				mockLLM.On("Generate", mock.Anything, mock.Anything).Return(resp, nil)
			},
			requestBody:    `{"model": "test-model", "prompt": "test prompt"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "generation failure",
			setupContext: func(c *gin.Context) {
				c.Set("authenticated", true)
				c.Set("user_id", "test-user")
			},
			setupMocks: func(mockLLM *MockLLMClient) {
				mockLLM.On("Generate", mock.Anything, mock.Anything).Return((*llm.GenerationResponse)(nil), assert.AnError)
			},
			requestBody:    `{"model": "test-model", "prompt": "test prompt"}`,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Generation failed",
		},
		{
			name: "unauthorized - no user context",
			setupContext: func(c *gin.Context) {
				// Don't set any user context
			},
			requestBody:    `{"model": "test-model", "prompt": "test prompt"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service with mocks
			mockLLM := new(MockLLMClient)
			mockRedis := new(MockRedisClient)

			if tt.setupMocks != nil {
				tt.setupMocks(mockLLM)
			}

			service := generation.NewService(mockLLM, mockRedis)

			// Create test request
			r := gin.New()
			r.POST("/generate", func(c *gin.Context) {
				tt.setupContext(c)
				service.RequestResponseHandler(c)
			})

			req := httptest.NewRequest("POST", "/generate", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockLLM.AssertExpectations(t)
		})
	}
}

func TestHealthCheckHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMocks     func(*MockLLMClient, *MockRedisClient)
		expectedStatus int
		expectedHealth string
	}{
		{
			name: "all services healthy",
			setupMocks: func(mockLLM *MockLLMClient, mockRedis *MockRedisClient) {
				mockLLM.On("Health", mock.Anything).Return(nil)
				mockRedis.On("Ping", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedHealth: "healthy",
		},
		{
			name: "llm service unhealthy",
			setupMocks: func(mockLLM *MockLLMClient, mockRedis *MockRedisClient) {
				mockLLM.On("Health", mock.Anything).Return(assert.AnError)
				mockRedis.On("Ping", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "error",
		},
		{
			name: "redis service unhealthy",
			setupMocks: func(mockLLM *MockLLMClient, mockRedis *MockRedisClient) {
				mockLLM.On("Health", mock.Anything).Return(nil)
				mockRedis.On("Ping", mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMClient)
			mockRedis := new(MockRedisClient)

			tt.setupMocks(mockLLM, mockRedis)

			service := generation.NewService(mockLLM, mockRedis)

			r := gin.New()
			r.GET("/health", service.HealthCheckHandler)

			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedHealth)

			mockLLM.AssertExpectations(t)
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestRedisSubscriptions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(mockRedis *MockRedisClient)
		execute func(service *generation.Service) (*redis.PubSub, error)
	}{
		{
			name: "subscribe to user channel",
			setup: func(mockRedis *MockRedisClient) {
				mockRedis.On("Subscribe", ctx, []string{"user:user-123:generations"}).Return(&redis.PubSub{}).Once()
			},
			execute: func(service *generation.Service) (*redis.PubSub, error) {
				return service.SubscribeToUserChannel(ctx, "user-123")
			},
		},
		{
			name: "subscribe to project channel",
			setup: func(mockRedis *MockRedisClient) {
				mockRedis.On("Subscribe", ctx, []string{"project:project-456:generations"}).Return(&redis.PubSub{}).Once()
			},
			execute: func(service *generation.Service) (*redis.PubSub, error) {
				return service.SubscribeToProjectChannel(ctx, "project-456")
			},
		},
		{
			name: "subscribe to global channel",
			setup: func(mockRedis *MockRedisClient) {
				mockRedis.On("Subscribe", ctx, []string{"global:generations"}).Return(&redis.PubSub{}).Once()
			},
			execute: func(service *generation.Service) (*redis.PubSub, error) {
				return service.SubscribeToGlobalChannel(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMClient)
			mockRedis := new(MockRedisClient)
			service := generation.NewService(mockLLM, mockRedis)

			tt.setup(mockRedis)

			pubsub, err := tt.execute(service)

			assert.NoError(t, err)
			assert.NotNil(t, pubsub)
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestRedisSubscriptionsWithoutRedis(t *testing.T) {
	mockLLM := new(MockLLMClient)
	service := generation.NewService(mockLLM, nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		method        func() (*redis.PubSub, error)
		expectedError string
	}{
		{
			name: "user_channel_without_redis",
			method: func() (*redis.PubSub, error) {
				return service.SubscribeToUserChannel(ctx, "user-123")
			},
			expectedError: "redis client is not initialized",
		},
		{
			name: "project_channel_without_redis",
			method: func() (*redis.PubSub, error) {
				return service.SubscribeToProjectChannel(ctx, "project-456")
			},
			expectedError: "redis client is not initialized",
		},
		{
			name: "global_channel_without_redis",
			method: func() (*redis.PubSub, error) {
				return service.SubscribeToGlobalChannel(ctx)
			},
			expectedError: "redis client is not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubsub, err := tt.method()

			assert.Error(t, err)
			assert.Nil(t, pubsub)
			assert.Equal(t, tt.expectedError, err.Error())
		})
	}
}
