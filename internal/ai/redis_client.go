package ai

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
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

// NewRedisClient creates a new Redis client
func NewRedisClient(config *RedisConfig) RedisClient {
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

// RedisClientFactory creates Redis clients following factory pattern
type RedisClientFactory struct{}

// NewRedisClientFactory creates a new Redis client factory
func NewRedisClientFactory() *RedisClientFactory {
	return &RedisClientFactory{}
}

// CreateClient creates a Redis client based on configuration
func (f *RedisClientFactory) CreateClient(config *RedisConfig) (RedisClient, error) {
	if config == nil {
		return &stubRedisClient{}, nil
	}

	if err := f.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Redis configuration: %w", err)
	}

	return NewRedisClient(config), nil
}

// validateConfig validates Redis configuration
func (f *RedisClientFactory) validateConfig(config *RedisConfig) error {
	if config.Host == "" {
		return fmt.Errorf("Redis host is required")
	}
	if config.Port <= 0 {
		return fmt.Errorf("Redis port must be positive")
	}
	return nil
}

// HealthCheckRedis performs a health check on Redis client
func HealthCheckRedis(ctx context.Context, client RedisClient) error {
	if client == nil {
		return fmt.Errorf("Redis client is nil")
	}

	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	return nil
}
