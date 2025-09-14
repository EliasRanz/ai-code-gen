package database_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/config"
)

// TestDatabaseConfigValidation tests database configuration validation
func TestDatabaseConfigValidation(t *testing.T) {
	t.Run("should validate required fields", func(t *testing.T) {
		testCases := []struct {
			name          string
			config        config.DatabaseConfig
			expectedError bool
		}{
			{
				name: "valid config",
				config: config.DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "test",
					DBName:   "testdb",
				},
				expectedError: false,
			},
			{
				name: "missing host",
				config: config.DatabaseConfig{
					Host:     "",
					Port:     5432,
					User:     "test",
					Password: "test",
					DBName:   "testdb",
				},
				expectedError: true,
			},
			{
				name: "missing database name",
				config: config.DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "test",
					Password: "test",
					DBName:   "",
				},
				expectedError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := validateDatabaseConfig(tc.config)
				if tc.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

// validateDatabaseConfig validates database configuration without connecting
func validateDatabaseConfig(cfg config.DatabaseConfig) error {
	if cfg.Host == "" {
		return errors.New("host is required")
	}
	if cfg.DBName == "" {
		return errors.New("database name is required")
	}
	if cfg.User == "" {
		return errors.New("user is required")
	}
	if cfg.Port <= 0 {
		return errors.New("port must be positive")
	}
	return nil
}
