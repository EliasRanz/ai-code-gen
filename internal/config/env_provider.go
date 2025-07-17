package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// EnvironmentProvider loads configuration from environment variables
type EnvironmentProvider struct {
	prefix    string
	data      ConfigData
	watching  bool
	callbacks []func(ConfigData)
}

// NewEnvironmentProvider creates a new environment variable provider
func NewEnvironmentProvider(prefix string) *EnvironmentProvider {
	return &EnvironmentProvider{
		prefix:    prefix,
		data:      make(ConfigData),
		callbacks: make([]func(ConfigData), 0),
	}
}

// Load loads configuration from environment variables
func (p *EnvironmentProvider) Load(ctx context.Context) (ConfigData, error) {
	data := make(ConfigData)

	// Get all environment variables
	environ := os.Environ()

	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Filter by prefix if specified
		if p.prefix != "" && !strings.HasPrefix(key, p.prefix) {
			continue
		}

		// Remove prefix from key
		if p.prefix != "" {
			key = strings.TrimPrefix(key, p.prefix)
			key = strings.TrimPrefix(key, "_") // Remove leading underscore
		}

		// Convert key to lowercase and replace underscores with dots
		key = strings.ToLower(key)
		key = strings.ReplaceAll(key, "_", ".")

		// Try to parse value as different types
		data[key] = parseValue(value)
	}

	p.data = data
	return data, nil
}

// Watch watches for environment variable changes
func (p *EnvironmentProvider) Watch(ctx context.Context, callback func(ConfigData)) error {
	p.callbacks = append(p.callbacks, callback)
	p.watching = true

	return nil
}

// Get retrieves a specific configuration value
func (p *EnvironmentProvider) Get(ctx context.Context, key string) (interface{}, error) {
	if val, exists := p.data[key]; exists {
		return val, nil
	}

	// Try to get directly from environment with prefix
	envKey := strings.ToUpper(key)
	envKey = strings.ReplaceAll(envKey, ".", "_")

	if p.prefix != "" {
		envKey = fmt.Sprintf("%s_%s", p.prefix, envKey)
	}

	if value := os.Getenv(envKey); value != "" {
		parsed := parseValue(value)
		p.data[key] = parsed
		return parsed, nil
	}

	return nil, fmt.Errorf("key '%s' not found", key)
}

// Validate validates the configuration data
func (p *EnvironmentProvider) Validate(ctx context.Context, data ConfigData) error {
	for key, value := range data {
		if value == nil {
			return fmt.Errorf("configuration key '%s' has nil value", key)
		}
	}

	return nil
}

// HealthCheck performs a health check on the provider
func (p *EnvironmentProvider) HealthCheck(ctx context.Context) error {
	environ := os.Environ()
	if environ == nil {
		return fmt.Errorf("unable to access environment variables")
	}

	return nil
}

// Close closes the provider and cleans up resources
func (p *EnvironmentProvider) Close() error {
	p.watching = false
	p.callbacks = nil
	p.data = nil

	return nil
}

// parseValue attempts to parse a string value into appropriate Go types
func parseValue(value string) interface{} {
	// Try boolean
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Try integer
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		if intVal >= int64(^uint(0)>>1)*-1 && intVal <= int64(^uint(0)>>1) {
			return int(intVal)
		}
		return intVal
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Try duration
	if durVal, err := time.ParseDuration(value); err == nil {
		return durVal.String()
	}

	// Check for comma-separated values (slice)
	if strings.Contains(value, ",") {
		parts := strings.Split(value, ",")
		var trimmed []string
		for _, part := range parts {
			trimmed = append(trimmed, strings.TrimSpace(part))
		}
		return trimmed
	}

	// Return as string
	return value
}
