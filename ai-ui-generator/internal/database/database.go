package database

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/config"
)

// NewConnection creates a new database connection
func NewConnection(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	// Parse durations from config strings
	if maxLifetime, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
		db.SetConnMaxLifetime(maxLifetime)
	}
	if maxIdleTime, err := time.ParseDuration(cfg.ConnMaxIdleTime); err == nil {
		db.SetConnMaxIdleTime(maxIdleTime)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.DBName).
		Msg("Database connection established")

	return db, nil
}

// Close closes the database connection
func Close(db *sqlx.DB) error {
	if db != nil {
		log.Info().Msg("Closing database connection")
		return db.Close()
	}
	return nil
}
