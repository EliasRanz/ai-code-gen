package database

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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

// NewGormConnection creates a new GORM database connection
func NewGormConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// Parse durations from config strings
	if maxLifetime, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
		sqlDB.SetConnMaxLifetime(maxLifetime)
	}
	if maxIdleTime, err := time.ParseDuration(cfg.ConnMaxIdleTime); err == nil {
		sqlDB.SetConnMaxIdleTime(maxIdleTime)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.DBName).
		Msg("GORM Database connection established")

	return db, nil
}
