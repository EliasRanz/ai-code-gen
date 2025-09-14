//go:build integration
// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/database"
	"github.com/stretchr/testify/require"
)

func TestDatabaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Load configuration from environment
	testConfig, err := LoadIntegrationConfig()
	require.NoError(t, err, "Failed to load integration test configuration")

	// Create database connection
	db, err := database.NewConnection(testConfig.Database)
	require.NoError(t, err, "Failed to create database connection")
	defer database.Close(db)

	ctx := context.Background()

	// Basic connection test
	err = db.PingContext(ctx)
	require.NoError(t, err, "Database ping should succeed")
}
