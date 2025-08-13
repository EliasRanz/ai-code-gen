// Package utilities contains entity factory and validation implementations
package utilities

import (
	"encoding/json"
	"fmt"
	"time"
)

// BaseEntity provides common entity implementation
type BaseEntity struct {
	id          string
	entityType  EntityType
	version     int64
	createdAt   time.Time
	updatedAt   time.Time
	dirtyFields []string
}

// NewBaseEntity creates a new base entity
func NewBaseEntity(id string, entityType EntityType) BaseEntity {
	now := time.Now()
	return BaseEntity{
		id:          id,
		entityType:  entityType,
		version:     1,
		createdAt:   now,
		updatedAt:   now,
		dirtyFields: []string{},
	}
}

// GetID returns the entity ID
func (b *BaseEntity) GetID() string {
	return b.id
}

// GetType returns the entity type
func (b *BaseEntity) GetType() EntityType {
	return b.entityType
}

// GetVersion returns the entity version
func (b *BaseEntity) GetVersion() int64 {
	return b.version
}

// GetCreatedAt returns the creation time
func (b *BaseEntity) GetCreatedAt() time.Time {
	return b.createdAt
}

// GetUpdatedAt returns the last update time
func (b *BaseEntity) GetUpdatedAt() time.Time {
	return b.updatedAt
}

// MarkDirty marks a field as dirty
func (b *BaseEntity) MarkDirty(field string) {
	for _, f := range b.dirtyFields {
		if f == field {
			return
		}
	}
	b.dirtyFields = append(b.dirtyFields, field)
}

// GetDirtyFields returns dirty fields
func (b *BaseEntity) GetDirtyFields() []string {
	return b.dirtyFields
}

// ClearDirtyFields clears dirty fields
func (b *BaseEntity) ClearDirtyFields() {
	b.dirtyFields = []string{}
}

// IsValid returns true if entity passes validation
func (b *BaseEntity) IsValid() bool {
	// Base entity is always valid, specific entities override this
	return true
}

// Validate validates the base entity
func (b *BaseEntity) Validate() error {
	// Base validation - specific entities override this
	return nil
}

// GetValidationRules returns validation rules for the entity
func (b *BaseEntity) GetValidationRules() []ValidationRule {
	// Base entity has no rules, specific entities override this
	return []ValidationRule{}
}

// ToJSON serializes entity to JSON
func (b *BaseEntity) ToJSON() ([]byte, error) {
	data := b.ToMap()
	return json.Marshal(data)
}

// FromJSON deserializes entity from JSON
func (b *BaseEntity) FromJSON(data []byte) error {
	var mapData map[string]interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		return err
	}
	// Base implementation - specific entities should override
	return nil
}

// ToMap converts entity to map
func (b *BaseEntity) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         b.id,
		"type":       string(b.entityType),
		"version":    b.version,
		"created_at": b.createdAt,
		"updated_at": b.updatedAt,
	}
}

// BeforeSave is called before saving
func (b *BaseEntity) BeforeSave() error {
	return nil
}

// AfterSave is called after saving
func (b *BaseEntity) AfterSave() error {
	b.ClearDirtyFields()
	return nil
}

// BeforeDelete is called before deleting
func (b *BaseEntity) BeforeDelete() error {
	return nil
}

// DefaultEntityFactory implements EntityFactory interface
type DefaultEntityFactory struct{}

// NewEntityFactory creates a new entity factory
func NewEntityFactory() EntityFactory {
	return &DefaultEntityFactory{}
}

// CreateEntity creates an entity from map data
func (f *DefaultEntityFactory) CreateEntity(entityType EntityType, data map[string]interface{}) (DomainEntity, error) {
	switch entityType {
	case EntityTypeUser:
		return f.createUserEntity(data)
	case EntityTypeProject:
		return f.createProjectEntity(data)
	case EntityTypeGeneration:
		return f.createGenerationEntity(data)
	default:
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}
}

// CreateFromJSON creates an entity from JSON data
func (f *DefaultEntityFactory) CreateFromJSON(entityType EntityType, jsonData []byte) (DomainEntity, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return f.CreateEntity(entityType, data)
}

// ListEntityTypes returns supported entity types
func (f *DefaultEntityFactory) ListEntityTypes() []EntityType {
	return []EntityType{
		EntityTypeUser,
		EntityTypeProject,
		EntityTypeGeneration,
	}
}

// Helper methods for entity creation
func (f *DefaultEntityFactory) createUserEntity(data map[string]interface{}) (DomainEntity, error) {
	id, ok := data["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required for user entity")
	}
	email, ok := data["email"].(string)
	if !ok {
		return nil, fmt.Errorf("email is required for user entity")
	}
	username, _ := data["username"].(string)
	name, _ := data["name"].(string)

	// Create a basic user entity representation
	now := time.Now()
	return &BasicUserEntity{
		BaseEntity: NewBaseEntity(id, EntityTypeUser),
		ID:         UserID(id),
		Email:      email,
		Username:   username,
		Name:       name,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (f *DefaultEntityFactory) createProjectEntity(data map[string]interface{}) (DomainEntity, error) {
	id, ok := data["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required for project entity")
	}
	name, ok := data["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required for project entity")
	}
	userID, ok := data["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id is required for project entity")
	}

	// Create a basic project entity representation
	now := time.Now()
	return &BasicProjectEntity{
		BaseEntity: NewBaseEntity(id, EntityTypeProject),
		ID:         ProjectID(id),
		Name:       name,
		UserID:     UserID(userID),
		Status:     ProjectStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (f *DefaultEntityFactory) createGenerationEntity(data map[string]interface{}) (DomainEntity, error) {
	id, ok := data["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required for generation entity")
	}
	content, ok := data["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required for generation entity")
	}
	projectID, ok := data["project_id"].(string)
	if !ok {
		return nil, fmt.Errorf("project_id is required for generation entity")
	}

	// Create a basic generation entity representation
	now := time.Now()
	return &BasicGenerationEntity{
		BaseEntity: NewBaseEntity(id, EntityTypeGeneration),
		ID:         id,
		Content:    content,
		ProjectID:  ProjectID(projectID),
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// Basic entity implementations for factory pattern

// BasicUserEntity implements DomainEntity and User interfaces
type BasicUserEntity struct {
	BaseEntity
	ID        UserID
	Email     string
	Username  string
	Name      string
	Roles     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u BasicUserEntity) Validate() error {
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}

func (u BasicUserEntity) GetEmail() string                      { return u.Email }
func (u BasicUserEntity) GetUsername() string                   { return u.Username }
func (u BasicUserEntity) SetPassword(password string) error     { return fmt.Errorf("not implemented") }
func (u BasicUserEntity) ValidatePassword(password string) bool { return false }
func (u BasicUserEntity) GetRoles() []string                    { return u.Roles }
func (u BasicUserEntity) HasPermission(permission string) bool  { return false }

// BasicProjectEntity implements DomainEntity and Project interfaces
type BasicProjectEntity struct {
	BaseEntity
	ID        ProjectID
	Name      string
	UserID    UserID
	Status    ProjectStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p BasicProjectEntity) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func (p BasicProjectEntity) GetOwnerID() string                         { return string(p.UserID) }
func (p BasicProjectEntity) GetName() string                            { return p.Name }
func (p BasicProjectEntity) GetStatus() ProjectStatus                   { return p.Status }
func (p *BasicProjectEntity) SetStatus(status ProjectStatus) error      { p.Status = status; return nil }
func (p BasicProjectEntity) GetGenerations() []Generation               { return []Generation{} }
func (p *BasicProjectEntity) AddGeneration(generation Generation) error { return nil }

// BasicGenerationEntity implements DomainEntity and Generation interfaces
type BasicGenerationEntity struct {
	BaseEntity
	ID        string
	Content   string
	ProjectID ProjectID
	Provider  string
	Tokens    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (g BasicGenerationEntity) Validate() error {
	if g.Content == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}

func (g BasicGenerationEntity) GetProjectID() string             { return string(g.ProjectID) }
func (g BasicGenerationEntity) GetPrompt() string                { return "" }
func (g BasicGenerationEntity) GetContent() string               { return g.Content }
func (g BasicGenerationEntity) GetProvider() string              { return g.Provider }
func (g BasicGenerationEntity) GetTokensUsed() int               { return g.Tokens }
func (g *BasicGenerationEntity) SetContent(content string) error { g.Content = content; return nil }

// DefaultEntityValidator implements EntityValidator interface
type DefaultEntityValidator struct {
	rules map[EntityType]map[string][]ValidationRule
}

// NewEntityValidator creates a new entity validator
func NewEntityValidator() EntityValidator {
	return &DefaultEntityValidator{
		rules: make(map[EntityType]map[string][]ValidationRule),
	}
}

// ValidateEntity validates an entire entity
func (v *DefaultEntityValidator) ValidateEntity(entity DomainEntity) error {
	return entity.Validate()
}

// ValidateField validates a specific field
func (v *DefaultEntityValidator) ValidateField(entity DomainEntity, field string, value interface{}) error {
	rules := v.GetFieldRules(entity.GetType(), field)
	for _, rule := range rules {
		if err := v.applyRule(rule, value); err != nil {
			return fmt.Errorf("validation failed for field %s: %w", field, err)
		}
	}
	return nil
}

// GetFieldRules returns validation rules for a field
func (v *DefaultEntityValidator) GetFieldRules(entityType EntityType, field string) []ValidationRule {
	if typeRules, exists := v.rules[entityType]; exists {
		if fieldRules, exists := typeRules[field]; exists {
			return fieldRules
		}
	}
	return []ValidationRule{}
}

// applyRule applies a validation rule
func (v *DefaultEntityValidator) applyRule(rule ValidationRule, value interface{}) error {
	switch rule.Rule {
	case "required":
		if value == nil || value == "" {
			return fmt.Errorf(rule.Message)
		}
	case "min_length":
		if str, ok := value.(string); ok && len(str) < 3 {
			return fmt.Errorf(rule.Message)
		}
	}
	return nil
}
