package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// JsonProvider loads configuration from JSON files
type JsonProvider struct {
	filePath  string
	data      ConfigData
	watching  bool
	callbacks []func(ConfigData)
	lastMod   time.Time
}

// NewJsonProvider creates a new JSON provider
func NewJsonProvider(filePath string) (*JsonProvider, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", filePath)
	}

	return &JsonProvider{
		filePath:  filePath,
		data:      make(ConfigData),
		callbacks: make([]func(ConfigData), 0),
	}, nil
}

// Load loads configuration from JSON file
func (p *JsonProvider) Load(ctx context.Context) (ConfigData, error) {
	// Read file content
	content, err := ioutil.ReadFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to ConfigData
	data := make(ConfigData)
	for key, value := range jsonData {
		data[key] = value
	}

	// Update file modification time
	if stat, err := os.Stat(p.filePath); err == nil {
		p.lastMod = stat.ModTime()
	}

	p.data = data
	return data, nil
}

// Watch watches for file changes
func (p *JsonProvider) Watch(ctx context.Context, callback func(ConfigData)) error {
	p.callbacks = append(p.callbacks, callback)
	p.watching = true

	// Start file watcher in goroutine
	go p.watchFile(ctx)

	return nil
}

// watchFile monitors file for changes
func (p *JsonProvider) watchFile(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5) // Check every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !p.watching {
				return
			}

			// Check if file was modified
			if stat, err := os.Stat(p.filePath); err == nil {
				if stat.ModTime().After(p.lastMod) {
					// File was modified, reload configuration
					if data, err := p.Load(ctx); err == nil {
						// Notify all callbacks
						for _, callback := range p.callbacks {
							callback(data)
						}
					}
				}
			}
		}
	}
}

// Get retrieves a specific configuration value
func (p *JsonProvider) Get(ctx context.Context, key string) (interface{}, error) {
	if val, exists := p.data[key]; exists {
		return val, nil
	}

	return nil, fmt.Errorf("key '%s' not found", key)
}

// Validate validates the configuration data
func (p *JsonProvider) Validate(ctx context.Context, data ConfigData) error {
	// Basic validation - ensure JSON structure is valid
	for key, value := range data {
		if value == nil {
			return fmt.Errorf("configuration key '%s' has nil value", key)
		}
	}

	return nil
}

// HealthCheck performs a health check on the provider
func (p *JsonProvider) HealthCheck(ctx context.Context) error {
	// Check if file exists and is readable
	if _, err := os.Stat(p.filePath); err != nil {
		return fmt.Errorf("configuration file not accessible: %w", err)
	}

	// Try to read and parse the file
	if _, err := p.Load(ctx); err != nil {
		return fmt.Errorf("configuration file is not valid: %w", err)
	}

	return nil
}

// Close closes the provider and cleans up resources
func (p *JsonProvider) Close() error {
	p.watching = false
	p.callbacks = nil
	p.data = nil

	return nil
}
