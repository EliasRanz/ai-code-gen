package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/zerolog/log"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connection holds database connection and metadata
type Connection struct {
	DB     *sql.DB
	Config *Config
}

// Close closes the database connection
func (c *Connection) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// Health checks database health
func (c *Connection) Health() error {
	if c.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection with ping
	if err := c.DB.PingContext(ctx); err != nil {
		log.Error().Err(err).Msg("Database ping failed")
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test with a simple query
	var version string
	err := c.DB.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		log.Error().Err(err).Msg("Database version query failed")
		return fmt.Errorf("database query failed: %w", err)
	}

	log.Debug().Str("version", version).Msg("Database health check passed")
	return nil
}

// Migrate runs database migrations using golang-migrate
func (c *Connection) Migrate() error {
	if c.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	log.Info().Msg("Running database migrations")

	// Create postgres driver instance
	driver, err := postgres.WithInstance(c.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get the absolute path to migrations directory
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Run migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Info().Msg("No new migrations to apply")
	} else {
		log.Info().Msg("Database migrations completed successfully")
	}

	return nil
}

// GetMigrationVersion returns the current migration version
func (c *Connection) GetMigrationVersion() (uint, bool, error) {
	if c.DB == nil {
		return 0, false, fmt.Errorf("database connection is nil")
	}

	// Create postgres driver instance
	driver, err := postgres.WithInstance(c.DB, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get the absolute path to migrations directory
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migrations path: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Get current version
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return 0, false, nil // No migrations applied yet
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
