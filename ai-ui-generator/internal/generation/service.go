package generation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/ai-code-gen/ai-ui-generator/internal/auth"
	"github.com/ai-code-gen/ai-ui-generator/internal/llm"
)

/*
TODO: Transition from Stubbed to Production LLM Integration

CURRENT STATE: 
The VLLM client has both real HTTP implementation and fallback stubs.
The generation service is production-ready but needs real LLM providers.

MIGRATION STEPS:

PHASE 1: Add OpenAI as Primary Provider (RECOMMENDED FIRST)
1. Implement OpenAI client (see openai_client_todo.go)
2. Add OpenAI configuration to config files
3. Update generation service to use OpenAI for specific models
4. Add environment variable for API key management
5. Test with actual OpenAI API

PHASE 2: Add Local LLM Support (Ollama/VLLM)
1. Set up local Ollama instance for development
2. Configure VLLM server for production deployment
3. Implement proper authentication for VLLM endpoints
4. Add model management and loading automation

PHASE 3: Multi-Provider Architecture
1. Implement provider factory (see factory_todo.go)
2. Add intelligent routing based on model requirements
3. Implement cost optimization and failover logic
4. Add provider health monitoring

CONFIGURATION MIGRATION:
Current: Only VLLM config with fallback stubs
Target: Multi-provider config with intelligent routing

DEPLOYMENT CONSIDERATIONS:
1. API Key Management:
   - Use environment variables or secret management
   - Implement key rotation capabilities
   - Add key validation on startup

2. Cost Control:
   - Implement usage quotas per user/API key
   - Add cost tracking and alerts
   - Monitor token usage and billing

3. Performance Optimization:
   - Connection pooling for HTTP clients
   - Response caching for frequent requests
   - Load balancing across provider instances

4. Monitoring and Observability:
   - Add metrics for provider performance
   - Implement distributed tracing
   - Log provider selection decisions

TESTING STRATEGY:
1. Unit Tests: Mock all provider APIs
2. Integration Tests: Optional real API tests
3. Load Tests: Test with multiple providers
4. Failover Tests: Verify provider switching

SECURITY CONSIDERATIONS:
1. API key encryption at rest
2. Rate limiting per provider
3. Input sanitization for all providers
4. Audit logging for API usage
*/

/*
TODO: CRITICAL - Replace Stubbed LLM Implementation

IMMEDIATE ACTIONS REQUIRED:
1. Replace single VLLM client with multi-provider factory
2. Remove all fallback stub logic from VLLM client
3. Implement real OpenAI, Claude, and Ollama clients
4. Add cost tracking and usage analytics
5. Implement provider failover and load balancing

See docs/STUB_TO_PRODUCTION_MIGRATION.md for complete migration guide.
See TODO_MASTER_LIST.md for detailed implementation steps.

FILES TO IMPLEMENT:
- internal/llm/openai_client.go (from openai_client_todo.go)
- internal/llm/claude_client.go (from claude_client_todo.go)  
- internal/llm/ollama_client.go (from ollama_client_todo.go)
- internal/llm/factory.go (from factory_todo.go)

STUBBED CODE TO REMOVE:
- Line ~572-640 in internal/llm/vllm_client.go (fallback stubs)
- All mock response generation logic
- Hardcoded responses for testing

CONFIGURATION TO UPDATE:
- Add multi-provider configuration
- Add API key management
- Add routing rules for provider selection

TIMELINE: 1-2 weeks for complete migration
*/

// RedisClient interface for Redis operations
type RedisClient interface {
	Ping(ctx context.Context) error
	Close() error
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
}

// redisClientImpl implements RedisClient using go-redis
type redisClientImpl struct {
	client *redis.Client
}

func (r *redisClientImpl) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *redisClientImpl) Close() error {
	return r.client.Close()
}

func (r *redisClientImpl) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return r.client.Publish(ctx, channel, message)
}

func (r *redisClientImpl) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// newRedisClient creates a new Redis client
func newRedisClient(config *RedisConfig) RedisClient {
	if config == nil {
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	return &redisClientImpl{client: rdb}
}

// stubRedisClient is a stub implementation of RedisClient for fallback
type stubRedisClient struct{}

func (s *stubRedisClient) Ping(ctx context.Context) error {
	log.Debug().Msg("Redis ping (stubbed)")
	return nil
}

func (s *stubRedisClient) Close() error {
	log.Debug().Msg("Redis close (stubbed)")
	return nil
}

func (s *stubRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	log.Debug().Str("channel", channel).Msg("Redis publish (stubbed)")
	return nil
}

func (s *stubRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	log.Debug().Strs("channels", channels).Msg("Redis subscribe (stubbed)")
	return nil
}

// Service provides AI generation functionality
type Service struct {
	llmClient   llm.LLMClient
	redis       RedisClient
	authService *auth.Service
}

// Config holds configuration for the generation service
type Config struct {
	LLMConfig   *llm.VLLMConfig `json:"llm"`
	RedisConfig *RedisConfig    `json:"redis"`
}

// RedisConfig holds Redis configuration for pub/sub
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// GenerationRequest represents the incoming generation request
type GenerationRequest struct {
	Model       string                 `json:"model" binding:"required"`
	Prompt      string                 `json:"prompt" binding:"required"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	ProjectID   string                 `json:"project_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// StreamResponse represents a server-sent event response
type StreamResponse struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
	ID    string      `json:"id,omitempty"`
}

// NewService creates a new generation service
func NewService(config *Config, authService *auth.Service) (*Service, error) {
	// Initialize LLM client
	llmClient := llm.NewVLLMClient(config.LLMConfig)

	// Initialize Redis client
	var redisClient RedisClient
	if config.RedisConfig != nil {
		redisClient = newRedisClient(config.RedisConfig)

		// Test Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx); err != nil {
			log.Warn().Err(err).Msg("Redis connection failed, falling back to stub")
			redisClient = &stubRedisClient{}
		} else {
			log.Info().Msg("Redis connection established for generation service")
		}
	} else {
		log.Info().Msg("Redis not configured, using stub implementation")
		redisClient = &stubRedisClient{}
	}

	return &Service{
		llmClient:   llmClient,
		redis:       redisClient,
		authService: authService,
	}, nil
}

// StreamGenerationHandler handles the /generate/stream SSE endpoint
func (s *Service) StreamGenerationHandler(c *gin.Context) {
	// Extract user information from context (set by auth middleware)
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, exists := auth.GetUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "User account is inactive"})
		return
	}

	var req GenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set up SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Convert to LLM request
	llmReq := &llm.GenerationRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
		UserID:      userID,
		ProjectID:   req.ProjectID,
		Metadata:    req.Metadata,
	}

	// Start streaming generation
	ctx := c.Request.Context()
	respChan, err := s.llmClient.GenerateStream(ctx, llmReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start generation stream")
		s.writeSSEError(c, "generation_failed", "Failed to start generation")
		return
	}

	// Stream the response
	s.streamResponse(c, respChan, userID, req.ProjectID)
}

// NonStreamGenerationHandler handles non-streaming generation requests
func (s *Service) NonStreamGenerationHandler(c *gin.Context) {
	// Extract user information from context (set by auth middleware)
	userID, exists := auth.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, exists := auth.GetUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "User account is inactive"})
		return
	}

	var req GenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to LLM request
	llmReq := &llm.GenerationRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
		UserID:      userID,
		ProjectID:   req.ProjectID,
		Metadata:    req.Metadata,
	}

	// Generate response
	ctx := c.Request.Context()
	resp, err := s.llmClient.Generate(ctx, llmReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Generation failed"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetModelsHandler returns available models
func (s *Service) GetModelsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	models, err := s.llmClient.GetModels(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get models")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get models"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

// HealthHandler checks the health of the generation service
func (s *Service) HealthHandler(c *gin.Context) {
	ctx := c.Request.Context()

	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"services": gin.H{
			"llm":   "unknown",
			"redis": "unknown",
		},
	}

	// Check LLM health
	if err := s.llmClient.Health(ctx); err != nil {
		health["services"].(gin.H)["llm"] = "unhealthy"
		health["status"] = "degraded"
		log.Warn().Err(err).Msg("LLM service health check failed")
	} else {
		health["services"].(gin.H)["llm"] = "healthy"
	}

	// Check Redis health (if available)
	if s.redis != nil {
		if err := s.redis.Ping(ctx); err != nil {
			health["services"].(gin.H)["redis"] = "unhealthy"
			health["status"] = "degraded"
			log.Warn().Err(err).Msg("Redis health check failed")
		} else {
			health["services"].(gin.H)["redis"] = "healthy"
		}
	} else {
		health["services"].(gin.H)["redis"] = "disabled"
	}

	status := http.StatusOK
	if health["status"] == "degraded" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, health)
}

// streamResponse handles the actual streaming of responses
func (s *Service) streamResponse(c *gin.Context, respChan <-chan *llm.GenerationResponse, userID, projectID string) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		log.Error().Msg("Streaming unsupported")
		s.writeSSEError(c, "streaming_unsupported", "Streaming not supported")
		return
	}

	// Send initial event
	s.writeSSEEvent(c, "generation_started", gin.H{
		"message":    "Generation started",
		"user_id":    userID,
		"project_id": projectID,
	}, "")
	flusher.Flush()

	// Stream responses
	for resp := range respChan {
		// Publish to Redis for horizontal scaling (stubbed)
		if s.redis != nil {
			s.publishToRedis(resp, userID, projectID)
		}

		// Send the response chunk
		s.writeSSEEvent(c, "generation_chunk", resp, resp.ID)
		flusher.Flush()

		// Check if generation is complete
		if len(resp.Choices) > 0 && resp.Choices[0].FinishReason != nil {
			s.writeSSEEvent(c, "generation_complete", gin.H{
				"message":       "Generation completed",
				"finish_reason": *resp.Choices[0].FinishReason,
				"usage":         resp.Usage,
			}, "")
			flusher.Flush()
			break
		}
	}

	// Send final event
	s.writeSSEEvent(c, "stream_end", gin.H{"message": "Stream ended"}, "")
	flusher.Flush()
}

// writeSSEEvent writes a server-sent event
func (s *Service) writeSSEEvent(c *gin.Context, event string, data interface{}, id string) {
	if id != "" {
		fmt.Fprintf(c.Writer, "id: %s\n", id)
	}
	fmt.Fprintf(c.Writer, "event: %s\n", event)

	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal SSE data")
		dataBytes = []byte(`{"error": "Failed to marshal data"}`)
	}

	// Handle multi-line data
	dataStr := string(dataBytes)
	for _, line := range strings.Split(dataStr, "\n") {
		fmt.Fprintf(c.Writer, "data: %s\n", line)
	}
	fmt.Fprintf(c.Writer, "\n")
}

// writeSSEError writes an error event
func (s *Service) writeSSEError(c *gin.Context, errorCode, message string) {
	s.writeSSEEvent(c, "error", gin.H{
		"error":   errorCode,
		"message": message,
	}, "")
}

// publishToRedis publishes generation events to Redis for horizontal scaling
func (s *Service) publishToRedis(resp *llm.GenerationResponse, userID, projectID string) {
	if s.redis == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare message payload
	message := map[string]interface{}{
		"response":   resp,
		"user_id":    userID,
		"project_id": projectID,
		"timestamp":  time.Now().UTC(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal Redis message")
		return
	}

	// Publish to user-specific channel
	if userID != "" {
		userChannel := fmt.Sprintf("generation:user:%s", userID)
		if err := s.redis.Publish(ctx, userChannel, messageBytes).Err(); err != nil {
			log.Error().Err(err).Str("channel", userChannel).Msg("Failed to publish to user channel")
		} else {
			log.Debug().Str("channel", userChannel).Msg("Published to user channel")
		}
	}

	// Publish to project-specific channel
	if projectID != "" {
		projectChannel := fmt.Sprintf("generation:project:%s", projectID)
		if err := s.redis.Publish(ctx, projectChannel, messageBytes).Err(); err != nil {
			log.Error().Err(err).Str("channel", projectChannel).Msg("Failed to publish to project channel")
		} else {
			log.Debug().Str("channel", projectChannel).Msg("Published to project channel")
		}
	}

	// Publish to global generation channel for monitoring/analytics
	globalChannel := "generation:global"
	if err := s.redis.Publish(ctx, globalChannel, messageBytes).Err(); err != nil {
		log.Error().Err(err).Str("channel", globalChannel).Msg("Failed to publish to global channel")
	} else {
		log.Debug().Str("channel", globalChannel).Msg("Published to global channel")
	}
}

// SubscribeToUserChannel subscribes to user-specific generation events
func (s *Service) SubscribeToUserChannel(ctx context.Context, userID string) (*redis.PubSub, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}

	channel := fmt.Sprintf("generation:user:%s", userID)
	pubsub := s.redis.Subscribe(ctx, channel)
	
	log.Info().Str("channel", channel).Msg("Subscribed to user channel")
	return pubsub, nil
}

// SubscribeToProjectChannel subscribes to project-specific generation events
func (s *Service) SubscribeToProjectChannel(ctx context.Context, projectID string) (*redis.PubSub, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}

	channel := fmt.Sprintf("generation:project:%s", projectID)
	pubsub := s.redis.Subscribe(ctx, channel)
	
	log.Info().Str("channel", channel).Msg("Subscribed to project channel")
	return pubsub, nil
}

// SubscribeToGlobalChannel subscribes to global generation events for monitoring
func (s *Service) SubscribeToGlobalChannel(ctx context.Context) (*redis.PubSub, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}

	channel := "generation:global"
	pubsub := s.redis.Subscribe(ctx, channel)
	
	log.Info().Str("channel", channel).Msg("Subscribed to global channel")
	return pubsub, nil
}

// SetupRoutes sets up the HTTP routes for the generation service
func (s *Service) SetupRoutes(authService *auth.Service) *gin.Engine {
	r := gin.Default()

	// Apply authentication middleware to protected routes
	protected := r.Group("/api/v1")
	protected.Use(auth.JWTMiddleware(authService))

	// Generation endpoints
	protected.POST("/generate/stream", s.StreamGenerationHandler)
	protected.POST("/generate", s.NonStreamGenerationHandler)

	// Model and health endpoints (public)
	r.GET("/api/v1/models", s.GetModelsHandler)
	r.GET("/health", s.HealthHandler)

	return r
}

// Close closes the service and its dependencies
func (s *Service) Close() error {
	var err error

	if s.llmClient != nil {
		if closeErr := s.llmClient.Close(); closeErr != nil {
			err = closeErr
		}
	}

	if s.redis != nil {
		if closeErr := s.redis.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return err
}
