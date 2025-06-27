package generation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/generation"
	"github.com/EliasRanz/ai-code-gen/internal/llm"
	"github.com/EliasRanz/ai-code-gen/internal/user"
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

// mockAuthService creates a minimal auth service for testing
func mockAuthService() *auth.Service {
	// Return nil since the generation service doesn't actually use the auth service directly
	// The auth is handled by middleware, not by the service itself
	return nil
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
			service := generation.NewService(tt.llmClient, tt.redisClient, mockAuthService())

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
				// Set valid user context
				testUser := &user.User{
					ID:       "test-user",
					Email:    "test@example.com",
					IsActive: true,
				}
				c.Set("user_id", testUser.ID)
				c.Set("user", testUser)
			},
			requestBody:    `{"model": "test-model", "prompt": "test prompt"}`,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name: "invalid request body",
			setupContext: func(c *gin.Context) {
				testUser := &user.User{
					ID:       "test-user",
					Email:    "test@example.com",
					IsActive: true,
				}
				c.Set("user_id", testUser.ID)
				c.Set("user", testUser)
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
		{
			name: "forbidden - inactive user",
			setupContext: func(c *gin.Context) {
				testUser := &user.User{
					ID:       "test-user",
					Email:    "test@example.com",
					IsActive: false, // User is inactive
				}
				c.Set("user_id", testUser.ID)
				c.Set("user", testUser)
			},
			requestBody:    `{"model": "test-model", "prompt": "test prompt"}`,
			expectedStatus: http.StatusForbidden,
			expectedError:  "User account is inactive",
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
			mockLLM.On("GenerateStream", mock.Anything, mock.Anything).Return((<-chan *llm.GenerationResponse)(respChan), nil)

			service := generation.NewService(mockLLM, mockRedis, mockAuthService())

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

func TestNonStreamGenerationHandler(t *testing.T) {
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
				testUser := &user.User{
					ID:       "test-user",
					Email:    "test@example.com",
					IsActive: true,
				}
				c.Set("user_id", testUser.ID)
				c.Set("user", testUser)
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
				testUser := &user.User{
					ID:       "test-user",
					Email:    "test@example.com",
					IsActive: true,
				}
				c.Set("user_id", testUser.ID)
				c.Set("user", testUser)
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
		{
			name: "forbidden - inactive user",
			setupContext: func(c *gin.Context) {
				testUser := &user.User{
					ID:       "test-user",
					Email:    "test@example.com",
					IsActive: false, // User is inactive
				}
				c.Set("user_id", testUser.ID)
				c.Set("user", testUser)
			},
			requestBody:    `{"model": "test-model", "prompt": "test prompt"}`,
			expectedStatus: http.StatusForbidden,
			expectedError:  "User account is inactive",
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

			service := generation.NewService(mockLLM, mockRedis, mockAuthService())

			// Create test request
			r := gin.New()
			r.POST("/generate", func(c *gin.Context) {
				tt.setupContext(c)
				service.NonStreamGenerationHandler(c)
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

func TestGetModelsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMocks     func(*MockLLMClient)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful models retrieval",
			setupMocks: func(mockLLM *MockLLMClient) {
				models := []llm.Model{
					{
						ID:          "model1",
						Name:        "Test Model 1",
						Description: "Test model 1 description",
						Provider:    "test",
						MaxTokens:   4096,
					},
					{
						ID:          "model2",
						Name:        "Test Model 2",
						Description: "Test model 2 description",
						Provider:    "test",
						MaxTokens:   2048,
					},
				}
				mockLLM.On("GetModels", mock.Anything).Return(models, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "models retrieval failure",
			setupMocks: func(mockLLM *MockLLMClient) {
				mockLLM.On("GetModels", mock.Anything).Return([]llm.Model(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to get models",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMClient)
			mockRedis := new(MockRedisClient)

			service := generation.NewService(mockLLM, mockRedis, mockAuthService())

			tt.setupMocks(mockLLM)

			r := gin.New()
			r.GET("/models", service.GetModelsHandler)

			req := httptest.NewRequest("GET", "/models", nil)
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

func TestHealthHandler(t *testing.T) {
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
			expectedHealth: "degraded",
		},
		{
			name: "redis service unhealthy",
			setupMocks: func(mockLLM *MockLLMClient, mockRedis *MockRedisClient) {
				mockLLM.On("Health", mock.Anything).Return(nil)
				mockRedis.On("Ping", mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "degraded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := new(MockLLMClient)
			mockRedis := new(MockRedisClient)

			tt.setupMocks(mockLLM, mockRedis)

			service := generation.NewService(mockLLM, mockRedis, mockAuthService())

			r := gin.New()
			r.GET("/health", service.HealthHandler)

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
	mockLLM := new(MockLLMClient)
	mockRedis := new(MockRedisClient)
	service := generation.NewService(mockLLM, mockRedis, mockAuthService())

	ctx := context.Background()

	tests := []struct {
		name    string
		method  func() (*redis.PubSub, error)
		channel string
	}{
		{
			name: "subscribe to user channel",
			method: func() (*redis.PubSub, error) {
				return service.SubscribeToUserChannel(ctx, "user-123")
			},
			channel: "generation:user:user-123",
		},
		{
			name: "subscribe to project channel",
			method: func() (*redis.PubSub, error) {
				return service.SubscribeToProjectChannel(ctx, "project-456")
			},
			channel: "generation:project:project-456",
		},
		{
			name: "subscribe to global channel",
			method: func() (*redis.PubSub, error) {
				return service.SubscribeToGlobalChannel(ctx)
			},
			channel: "generation:global",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedPubSub := &redis.PubSub{}
			mockRedis.On("Subscribe", ctx, []string{tt.channel}).Return(expectedPubSub)

			pubsub, err := tt.method()

			assert.NoError(t, err)
			assert.Equal(t, expectedPubSub, pubsub)
			mockRedis.AssertExpectations(t)
		})
	}
}

func TestRedisSubscriptionsWithoutRedis(t *testing.T) {
	mockLLM := new(MockLLMClient)
	service := generation.NewService(mockLLM, nil, mockAuthService())
	ctx := context.Background()

	tests := []struct {
		name   string
		method func() (*redis.PubSub, error)
	}{
		{
			name: "user channel without redis",
			method: func() (*redis.PubSub, error) {
				return service.SubscribeToUserChannel(ctx, "user-123")
			},
		},
		{
			name: "project channel without redis",
			method: func() (*redis.PubSub, error) {
				return service.SubscribeToProjectChannel(ctx, "project-456")
			},
		},
		{
			name: "global channel without redis",
			method: func() (*redis.PubSub, error) {
				return service.SubscribeToGlobalChannel(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubsub, err := tt.method()

			assert.Error(t, err)
			assert.Nil(t, pubsub)
			assert.Contains(t, err.Error(), "redis not available")
		})
	}
}
