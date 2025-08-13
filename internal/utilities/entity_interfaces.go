// Package utilities contains entity interface patterns and shared domain contracts
package utilities

import (
	"time"
)

// EntityType represents the type of entity
type EntityType string

const (
	EntityTypeUser       EntityType = "user"
	EntityTypeProject    EntityType = "project"
	EntityTypeGeneration EntityType = "generation"
)

// ValidationRule represents a validation rule
type ValidationRule struct {
	Field   string
	Rule    string
	Message string
}

// DomainEntity defines the core interface that all domain entities must implement
type DomainEntity interface {
	// Entity identification and metadata
	GetID() string
	GetType() EntityType
	GetVersion() int64
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time

	// Validation and business rules
	Validate() error
	IsValid() bool
	GetValidationRules() []ValidationRule

	// Serialization and persistence
	ToJSON() ([]byte, error)
	FromJSON(data []byte) error
	ToMap() map[string]interface{}

	// Change tracking and auditing
	MarkDirty(field string)
	GetDirtyFields() []string
	ClearDirtyFields()

	// Lifecycle events
	BeforeSave() error
	AfterSave() error
	BeforeDelete() error
}

// EntityFactory defines the interface for entity instantiation
type EntityFactory interface {
	CreateEntity(entityType EntityType, data map[string]interface{}) (DomainEntity, error)
	CreateFromJSON(entityType EntityType, jsonData []byte) (DomainEntity, error)
	ListEntityTypes() []EntityType
}

// EntityValidator defines the interface for consistent entity validation
type EntityValidator interface {
	ValidateEntity(entity DomainEntity) error
	ValidateField(entity DomainEntity, field string, value interface{}) error
	GetFieldRules(entityType EntityType, field string) []ValidationRule
}

// Service-specific entity interfaces

// User interface extends DomainEntity with user-specific methods
type User interface {
	DomainEntity
	GetEmail() string
	GetUsername() string
	SetPassword(password string) error
	ValidatePassword(password string) bool
	GetRoles() []string
	HasPermission(permission string) bool
}

// Project interface extends DomainEntity with project-specific methods
type Project interface {
	DomainEntity
	GetOwnerID() string
	GetName() string
	GetStatus() ProjectStatus
	SetStatus(status ProjectStatus) error
	GetGenerations() []Generation
	AddGeneration(generation Generation) error
}

// Generation interface extends DomainEntity with generation-specific methods
type Generation interface {
	DomainEntity
	GetProjectID() string
	GetPrompt() string
	GetContent() string
	GetProvider() string
	GetTokensUsed() int
	SetContent(content string) error
}

// ProjectStatus represents project status values
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusInactive ProjectStatus = "inactive"
	ProjectStatusArchived ProjectStatus = "archived"
)
