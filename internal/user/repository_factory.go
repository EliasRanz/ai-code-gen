package user

import (
	"fmt"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"gorm.io/gorm"
)

// PostgreSQLRepositoryFactory implements RepositoryFactory for PostgreSQL databases
type PostgreSQLRepositoryFactory struct{}

// NewPostgreSQLRepositoryFactory creates a new PostgreSQL repository factory
func NewPostgreSQLRepositoryFactory() RepositoryFactory {
	return &PostgreSQLRepositoryFactory{}
}

// CreateProjectRepository creates a new PostgreSQL project repository
func (f *PostgreSQLRepositoryFactory) CreateProjectRepository(db interface{}) (ProjectRepository, error) {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("invalid database type, expected *gorm.DB")
	}

	return NewPostgreSQLProjectRepository(gormDB)
}

// RepositoryManager manages repository instances following factory pattern
type RepositoryManager struct {
	factory RepositoryFactory
	db      *gorm.DB
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(db *gorm.DB) *RepositoryManager {
	return &RepositoryManager{
		factory: NewPostgreSQLRepositoryFactory(),
		db:      db,
	}
}

// GetProjectRepository returns a project repository instance
func (m *RepositoryManager) GetProjectRepository() (ProjectRepository, error) {
	return m.factory.CreateProjectRepository(m.db)
}

// HealthCheck checks the health of all managed repositories
func (m *RepositoryManager) HealthCheck() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close closes all managed repository connections
func (m *RepositoryManager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}

// ConfigurableRepositoryFactory allows configuration-based repository creation
type ConfigurableRepositoryFactory struct {
	config utilities.DatabaseConfig
}

// NewConfigurableRepositoryFactory creates a new configurable repository factory
func NewConfigurableRepositoryFactory(config utilities.DatabaseConfig) *ConfigurableRepositoryFactory {
	return &ConfigurableRepositoryFactory{
		config: config,
	}
}

// CreateProjectRepository creates a project repository based on configuration
func (f *ConfigurableRepositoryFactory) CreateProjectRepository(db interface{}) (ProjectRepository, error) {
	// In the future, this could switch between different database types based on config
	// For now, only PostgreSQL is supported
	return NewPostgreSQLRepositoryFactory().CreateProjectRepository(db)
}

// ValidateConfig validates the database configuration
func (f *ConfigurableRepositoryFactory) ValidateConfig() error {
	if f.config.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if f.config.Port <= 0 {
		return fmt.Errorf("database port must be positive")
	}
	if f.config.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if f.config.Username == "" {
		return fmt.Errorf("database username is required")
	}
	return nil
}
