package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/joho/godotenv"
)

// IntegrationTestConfig holds configuration for integration tests
type IntegrationTestConfig struct {
	Redis    cache.CacheConfig
	Database *config.DatabaseConfig
	Test     TestConfig
}

// TestConfig holds test-specific configuration
type TestConfig struct {
	Timeout     time.Duration
	Concurrency int
	Operations  int
}

// LoadIntegrationConfig loads configuration for integration tests based on environment
func LoadIntegrationConfig() (*IntegrationTestConfig, error) {
	// Determine environment (default to local)
	env := os.Getenv("INTEGRATION_ENV")
	if env == "" {
		env = "local"
	}

	// Get the directory of this source file for proper path resolution
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	// Build absolute path to the .env file
	sourceDir := filepath.Dir(filename)
	envFile := filepath.Join(sourceDir, fmt.Sprintf(".env.%s", env))

	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", envFile, err)
	}

	testConfig := &IntegrationTestConfig{}

	// Load Redis configuration
	redisPort, err := strconv.Atoi(getEnvWithDefault("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %w", err)
	}

	redisDB, err := strconv.Atoi(getEnvWithDefault("REDIS_DB", "1"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	maxConnections, err := strconv.Atoi(getEnvWithDefault("REDIS_MAX_CONNECTIONS", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_MAX_CONNECTIONS: %w", err)
	}

	maxIdleConnections, err := strconv.Atoi(getEnvWithDefault("REDIS_MAX_IDLE_CONNECTIONS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_MAX_IDLE_CONNECTIONS: %w", err)
	}

	connectionTimeout, err := time.ParseDuration(getEnvWithDefault("REDIS_CONNECTION_TIMEOUT", "5s"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_CONNECTION_TIMEOUT: %w", err)
	}

	idleTimeout, err := time.ParseDuration(getEnvWithDefault("REDIS_IDLE_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_IDLE_TIMEOUT: %w", err)
	}

	failureThreshold, err := strconv.Atoi(getEnvWithDefault("REDIS_FAILURE_THRESHOLD", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_FAILURE_THRESHOLD: %w", err)
	}

	requestVolumeThreshold, err := strconv.Atoi(getEnvWithDefault("REDIS_REQUEST_VOLUME_THRESHOLD", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_REQUEST_VOLUME_THRESHOLD: %w", err)
	}

	recoveryTimeout, err := time.ParseDuration(getEnvWithDefault("REDIS_RECOVERY_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_RECOVERY_TIMEOUT: %w", err)
	}

	testConfig.Redis = cache.CacheConfig{
		Host:                   getEnvWithDefault("REDIS_HOST", "localhost"),
		Port:                   redisPort,
		Password:               os.Getenv("REDIS_PASSWORD"),
		DB:                     redisDB,
		MaxConnections:         maxConnections,
		MaxIdleConnections:     maxIdleConnections,
		ConnectionTimeout:      connectionTimeout,
		IdleTimeout:            idleTimeout,
		FailureThreshold:       failureThreshold,
		RequestVolumeThreshold: requestVolumeThreshold,
		RecoveryTimeout:        recoveryTimeout,
		DefaultTTL:             5 * time.Minute,
	}

	// Load Database configuration
	dbPort, err := strconv.Atoi(getEnvWithDefault("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	maxOpenConns, err := strconv.Atoi(getEnvWithDefault("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdleConns, err := strconv.Atoi(getEnvWithDefault("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}

	testConfig.Database = &config.DatabaseConfig{
		Host:            getEnvWithDefault("DB_HOST", "localhost"),
		Port:            dbPort,
		User:            getEnvWithDefault("DB_USER", "postgres"),
		Password:        getEnvWithDefault("DB_PASSWORD", "postgres"),
		DBName:          getEnvWithDefault("DB_NAME", "ai_code_gen_test"),
		SSLMode:         getEnvWithDefault("DB_SSLMODE", "disable"),
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: getEnvWithDefault("DB_CONN_MAX_LIFETIME", "30m"),
		ConnMaxIdleTime: getEnvWithDefault("DB_CONN_MAX_IDLE_TIME", "5m"),
	}

	// Load Test configuration
	testTimeout, err := time.ParseDuration(getEnvWithDefault("TEST_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("invalid TEST_TIMEOUT: %w", err)
	}

	testConcurrency, err := strconv.Atoi(getEnvWithDefault("TEST_CONCURRENCY", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid TEST_CONCURRENCY: %w", err)
	}

	testOperations, err := strconv.Atoi(getEnvWithDefault("TEST_OPERATIONS", "50"))
	if err != nil {
		return nil, fmt.Errorf("invalid TEST_OPERATIONS: %w", err)
	}

	testConfig.Test = TestConfig{
		Timeout:     testTimeout,
		Concurrency: testConcurrency,
		Operations:  testOperations,
	}

	return testConfig, nil
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ValidateIntegrationEnvironment validates that required services are available
func ValidateIntegrationEnvironment(config *IntegrationTestConfig) error {
	// TODO: Add health checks for Redis and Database connections
	// This could include ping operations to ensure services are available
	return nil
}
