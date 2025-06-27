package generation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/EliasRanz/ai-code-gen/internal/llm"
)

// RedisClient defines the interface for Redis operations
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
		log.Info().Msg("Redis config not provided, using stub client")
		return &stubRedisClient{}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	return &redisClientImpl{client: rdb}
}

// stubRedisClient provides a no-op implementation for testing
type stubRedisClient struct{}

func (s *stubRedisClient) Ping(ctx context.Context) error {
	log.Debug().Msg("Stub Redis ping")
	return nil
}

func (s *stubRedisClient) Close() error {
	log.Debug().Msg("Stub Redis close")
	return nil
}

func (s *stubRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	log.Debug().Str("channel", channel).Msg("Stub Redis publish")
	// Return a stub IntCmd that always reports success
	cmd := redis.NewIntCmd(ctx, "publish", channel, message)
	cmd.SetVal(1) // Simulate one subscriber received the message
	return cmd
}

func (s *stubRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	log.Debug().Strs("channels", channels).Msg("Stub Redis subscribe")
	// Return a stub PubSub - in real implementation this would be a working PubSub
	return &redis.PubSub{}
}

// publishToRedis publishes generation response to Redis channels
func (s *Service) publishToRedis(resp *llm.GenerationResponse, userID, projectID string) {
	ctx := context.Background()

	message := gin.H{
		"response":   resp,
		"user_id":    userID,
		"project_id": projectID,
		"timestamp":  time.Now().UTC(),
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal message for Redis")
		return
	}

	// Publish to user-specific channel
	if userID != "" {
		channel := fmt.Sprintf("user:%s:generations", userID)
		if err := s.redisClient.Publish(ctx, channel, jsonMessage).Err(); err != nil {
			log.Error().Err(err).Str("channel", channel).Msg("Failed to publish to user channel")
		}
	}

	// Publish to project-specific channel
	if projectID != "" {
		channel := fmt.Sprintf("project:%s:generations", projectID)
		if err := s.redisClient.Publish(ctx, channel, jsonMessage).Err(); err != nil {
			log.Error().Err(err).Str("channel", channel).Msg("Failed to publish to project channel")
		}
	}

	// Publish to global channel
	if err := s.redisClient.Publish(ctx, "global:generations", jsonMessage).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to publish to global channel")
	}
}

// SubscribeToUserChannel subscribes to user-specific generation events
func (s *Service) SubscribeToUserChannel(ctx context.Context, userID string) (*redis.PubSub, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	channel := fmt.Sprintf("user:%s:generations", userID)
	pubsub := s.redisClient.Subscribe(ctx, channel)

	return pubsub, nil
}

// SubscribeToProjectChannel subscribes to project-specific generation events
func (s *Service) SubscribeToProjectChannel(ctx context.Context, projectID string) (*redis.PubSub, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	channel := fmt.Sprintf("project:%s:generations", projectID)
	pubsub := s.redisClient.Subscribe(ctx, channel)

	return pubsub, nil
}

// SubscribeToGlobalChannel subscribes to global generation events
func (s *Service) SubscribeToGlobalChannel(ctx context.Context) (*redis.PubSub, error) {
	pubsub := s.redisClient.Subscribe(ctx, "global:generations")
	return pubsub, nil
}
