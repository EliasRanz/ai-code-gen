package utilities

import (
	"context"
	"database/sql"
)

// QueryFilter represents query filtering options
type QueryFilter struct {
	Search string                 `json:"search,omitempty"`
	Status string                 `json:"status,omitempty"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

// Repository defines the core interface that all repository implementations must implement
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, entity interface{}) error
	GetByID(ctx context.Context, id string, entity interface{}) error
	Update(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, id string) error

	// Query operations
	List(ctx context.Context, filter QueryFilter, entities interface{}) error
	Count(ctx context.Context, filter QueryFilter) (int64, error)

	// Transaction management
	BeginTx(ctx context.Context) (Transaction, error)

	// Health and maintenance
	HealthCheck(ctx context.Context) error
	Close() error
}

// Transaction interface for consistent transaction handling
type Transaction interface {
	Commit() error
	Rollback() error
	Repository() Repository
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
}

// RepositoryFactory defines factory pattern for repository instantiation
type RepositoryFactory interface {
	CreateRepository(entityType string, config DatabaseConfig) (Repository, error)
	CreateUserRepository(config DatabaseConfig) (UserRepository, error)
	CreateProjectRepository(config DatabaseConfig) (ProjectRepository, error)
}

// UserRepository defines user-specific repository operations
type UserRepository interface {
	Repository
	GetByEmail(ctx context.Context, email string) (interface{}, error)
	GetByUsername(ctx context.Context, username string) (interface{}, error)
	UpdateLastLogin(ctx context.Context, userID string) error
}

// ProjectRepository defines project-specific repository operations
type ProjectRepository interface {
	Repository
	GetByUserID(ctx context.Context, userID UserID) (interface{}, error)
	GetByStatus(ctx context.Context, status string) (interface{}, error)
}

// TransactionOperation defines operations that can be executed within a transaction
type TransactionOperation interface {
	Execute(ctx context.Context, tx *sql.Tx) (interface{}, error)
}

// OperationType defines types of database operations
type OperationType string

const (
	OperationTypeCreate      OperationType = "create"
	OperationTypeRead        OperationType = "read"
	OperationTypeUpdate      OperationType = "update"
	OperationTypeDelete      OperationType = "delete"
	OperationTypeTransaction OperationType = "transaction"
	OperationTypeBatch       OperationType = "batch"
)
