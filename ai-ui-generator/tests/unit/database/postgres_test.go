package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	config := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "test",
		DBName:   "test",
		SSLMode:  "disable",
	}

	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "test", config.User)
	assert.Equal(t, "test", config.Password)
	assert.Equal(t, "test", config.DBName)
	assert.Equal(t, "disable", config.SSLMode)
}

func TestConnection_Close(t *testing.T) {
	conn := &Connection{}
	err := conn.Close()
	assert.NoError(t, err)
}

func TestConnection_Health_NilDB(t *testing.T) {
	conn := &Connection{}
	err := conn.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

func TestConnection_Migrate_NilDB(t *testing.T) {
	conn := &Connection{}
	err := conn.Migrate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

func TestConnection_GetMigrationVersion_NilDB(t *testing.T) {
	conn := &Connection{}
	version, dirty, err := conn.GetMigrationVersion()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
	assert.Equal(t, uint(0), version)
	assert.False(t, dirty)
}
