package utilities_entity_interfaces_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for testing interfaces

// MockDomainEntity implements DomainEntity for testing
type MockDomainEntity struct {
	id              string
	entityType      utilities.EntityType
	version         int64
	createdAt       time.Time
	updatedAt       time.Time
	valid           bool
	dirtyFields     []string
	data            map[string]interface{}
	validationRules []utilities.ValidationRule
}

func (m *MockDomainEntity) GetID() string {
	return m.id
}

func (m *MockDomainEntity) GetType() utilities.EntityType {
	return m.entityType
}

func (m *MockDomainEntity) GetVersion() int64 {
	return m.version
}

func (m *MockDomainEntity) GetCreatedAt() time.Time {
	return m.createdAt
}

func (m *MockDomainEntity) GetUpdatedAt() time.Time {
	return m.updatedAt
}

func (m *MockDomainEntity) Validate() error {
	if !m.valid {
		return errors.New("validation failed")
	}
	return nil
}

func (m *MockDomainEntity) IsValid() bool {
	return m.valid
}

func (m *MockDomainEntity) GetValidationRules() []utilities.ValidationRule {
	return m.validationRules
}

func (m *MockDomainEntity) ToJSON() ([]byte, error) {
	if m.data == nil {
		return nil, errors.New("no data to serialize")
	}
	return json.Marshal(m.data)
}

func (m *MockDomainEntity) FromJSON(data []byte) error {
	m.data = make(map[string]interface{})
	return json.Unmarshal(data, &m.data)
}

func (m *MockDomainEntity) ToMap() map[string]interface{} {
	return m.data
}

func (m *MockDomainEntity) MarkDirty(field string) {
	m.dirtyFields = append(m.dirtyFields, field)
}

func (m *MockDomainEntity) GetDirtyFields() []string {
	return m.dirtyFields
}

func (m *MockDomainEntity) ClearDirtyFields() {
	m.dirtyFields = []string{}
}

func (m *MockDomainEntity) BeforeSave() error {
	return nil
}

func (m *MockDomainEntity) AfterSave() error {
	return nil
}

func (m *MockDomainEntity) BeforeDelete() error {
	return nil
}

// MockEntityFactory implements EntityFactory for testing
type MockEntityFactory struct {
	entities    map[utilities.EntityType]*MockDomainEntity
	shouldError bool
}

func (f *MockEntityFactory) CreateEntity(entityType utilities.EntityType, data map[string]interface{}) (utilities.DomainEntity, error) {
	if f.shouldError {
		return nil, errors.New("factory error")
	}

	entity := &MockDomainEntity{
		id:          "test-id",
		entityType:  entityType,
		version:     1,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
		valid:       true,
		data:        data,
		dirtyFields: []string{},
	}

	return entity, nil
}

func (f *MockEntityFactory) CreateFromJSON(entityType utilities.EntityType, jsonData []byte) (utilities.DomainEntity, error) {
	if f.shouldError {
		return nil, errors.New("factory json error")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	return f.CreateEntity(entityType, data)
}

func (f *MockEntityFactory) ListEntityTypes() []utilities.EntityType {
	return []utilities.EntityType{
		utilities.EntityTypeUser,
		utilities.EntityTypeProject,
		utilities.EntityTypeGeneration,
	}
}

// MockEntityValidator implements EntityValidator for testing
type MockEntityValidator struct {
	shouldError bool
	rules       map[utilities.EntityType]map[string][]utilities.ValidationRule
}

func (v *MockEntityValidator) ValidateEntity(entity utilities.DomainEntity) error {
	if v.shouldError {
		return errors.New("validation error")
	}
	return entity.Validate()
}

func (v *MockEntityValidator) ValidateField(entity utilities.DomainEntity, field string, value interface{}) error {
	if v.shouldError {
		return errors.New("field validation error")
	}
	return nil
}

func (v *MockEntityValidator) GetFieldRules(entityType utilities.EntityType, field string) []utilities.ValidationRule {
	if v.rules != nil {
		if typeRules, exists := v.rules[entityType]; exists {
			if fieldRules, exists := typeRules[field]; exists {
				return fieldRules
			}
		}
	}
	return []utilities.ValidationRule{}
}

// MockUser implements User interface for testing
type MockUser struct {
	*MockDomainEntity
	email       string
	username    string
	password    string
	roles       []string
	permissions []string
}

func (u *MockUser) GetEmail() string {
	return u.email
}

func (u *MockUser) GetUsername() string {
	return u.username
}

func (u *MockUser) SetPassword(password string) error {
	u.password = password
	u.MarkDirty("password")
	return nil
}

func (u *MockUser) ValidatePassword(password string) bool {
	return u.password == password
}

func (u *MockUser) GetRoles() []string {
	return u.roles
}

func (u *MockUser) HasPermission(permission string) bool {
	for _, p := range u.permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// MockProject implements Project interface for testing
type MockProject struct {
	*MockDomainEntity
	ownerID     string
	name        string
	status      utilities.ProjectStatus
	generations []utilities.Generation
}

func (p *MockProject) GetOwnerID() string {
	return p.ownerID
}

func (p *MockProject) GetName() string {
	return p.name
}

func (p *MockProject) GetStatus() utilities.ProjectStatus {
	return p.status
}

func (p *MockProject) SetStatus(status utilities.ProjectStatus) error {
	p.status = status
	p.MarkDirty("status")
	return nil
}

func (p *MockProject) GetGenerations() []utilities.Generation {
	return p.generations
}

func (p *MockProject) AddGeneration(generation utilities.Generation) error {
	p.generations = append(p.generations, generation)
	p.MarkDirty("generations")
	return nil
}

// MockGeneration implements Generation interface for testing
type MockGeneration struct {
	*MockDomainEntity
	projectID  string
	prompt     string
	content    string
	provider   string
	tokensUsed int
}

func (g *MockGeneration) GetProjectID() string {
	return g.projectID
}

func (g *MockGeneration) GetPrompt() string {
	return g.prompt
}

func (g *MockGeneration) GetContent() string {
	return g.content
}

func (g *MockGeneration) GetProvider() string {
	return g.provider
}

func (g *MockGeneration) GetTokensUsed() int {
	return g.tokensUsed
}

func (g *MockGeneration) SetContent(content string) error {
	g.content = content
	g.MarkDirty("content")
	return nil
}

// Tests for EntityType constants
func TestEntityType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		value    utilities.EntityType
		expected string
	}{
		{"EntityTypeUser", utilities.EntityTypeUser, "user"},
		{"EntityTypeProject", utilities.EntityTypeProject, "project"},
		{"EntityTypeGeneration", utilities.EntityTypeGeneration, "generation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.value))
		})
	}
}

// Tests for ValidationRule struct
func TestValidationRule(t *testing.T) {
	t.Run("creation and access", func(t *testing.T) {
		rule := utilities.ValidationRule{
			Field:   "email",
			Rule:    "required|email",
			Message: "Email is required and must be valid",
		}

		assert.Equal(t, "email", rule.Field)
		assert.Equal(t, "required|email", rule.Rule)
		assert.Equal(t, "Email is required and must be valid", rule.Message)
	})
}

// Tests for DomainEntity interface
func TestDomainEntity_Interface(t *testing.T) {
	now := time.Now()
	entity := &MockDomainEntity{
		id:         "test-123",
		entityType: utilities.EntityTypeUser,
		version:    5,
		createdAt:  now,
		updatedAt:  now.Add(time.Hour),
		valid:      true,
		data:       map[string]interface{}{"name": "John"},
		validationRules: []utilities.ValidationRule{
			{Field: "name", Rule: "required", Message: "Name is required"},
		},
	}

	t.Run("entity identification", func(t *testing.T) {
		assert.Equal(t, "test-123", entity.GetID())
		assert.Equal(t, utilities.EntityTypeUser, entity.GetType())
		assert.Equal(t, int64(5), entity.GetVersion())
		assert.Equal(t, now, entity.GetCreatedAt())
		assert.Equal(t, now.Add(time.Hour), entity.GetUpdatedAt())
	})

	t.Run("validation", func(t *testing.T) {
		assert.NoError(t, entity.Validate())
		assert.True(t, entity.IsValid())

		rules := entity.GetValidationRules()
		assert.Len(t, rules, 1)
		assert.Equal(t, "name", rules[0].Field)
	})

	t.Run("serialization", func(t *testing.T) {
		jsonData, err := entity.ToJSON()
		assert.NoError(t, err)
		assert.Contains(t, string(jsonData), "John")

		dataMap := entity.ToMap()
		assert.Equal(t, "John", dataMap["name"])
	})

	t.Run("deserialization", func(t *testing.T) {
		jsonData := `{"age": 30, "city": "New York"}`
		err := entity.FromJSON([]byte(jsonData))
		assert.NoError(t, err)

		dataMap := entity.ToMap()
		assert.Equal(t, float64(30), dataMap["age"])
		assert.Equal(t, "New York", dataMap["city"])
	})

	t.Run("change tracking", func(t *testing.T) {
		entity.ClearDirtyFields()
		assert.Empty(t, entity.GetDirtyFields())

		entity.MarkDirty("email")
		entity.MarkDirty("name")

		dirtyFields := entity.GetDirtyFields()
		assert.Len(t, dirtyFields, 2)
		assert.Contains(t, dirtyFields, "email")
		assert.Contains(t, dirtyFields, "name")

		entity.ClearDirtyFields()
		assert.Empty(t, entity.GetDirtyFields())
	})

	t.Run("lifecycle events", func(t *testing.T) {
		assert.NoError(t, entity.BeforeSave())
		assert.NoError(t, entity.AfterSave())
		assert.NoError(t, entity.BeforeDelete())
	})
}

// Tests for EntityFactory interface
func TestEntityFactory_Interface(t *testing.T) {
	factory := &MockEntityFactory{}

	t.Run("create entity from data", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
		assert.NoError(t, err)
		assert.NotNil(t, entity)
		assert.Equal(t, utilities.EntityTypeUser, entity.GetType())
		assert.Equal(t, "test-id", entity.GetID())
	})

	t.Run("create entity from JSON", func(t *testing.T) {
		jsonData := `{"name":"Jane Doe","email":"jane@example.com"}`

		entity, err := factory.CreateFromJSON(utilities.EntityTypeUser, []byte(jsonData))
		assert.NoError(t, err)
		assert.NotNil(t, entity)
		assert.Equal(t, utilities.EntityTypeUser, entity.GetType())
	})

	t.Run("list entity types", func(t *testing.T) {
		types := factory.ListEntityTypes()
		assert.Len(t, types, 3)
		assert.Contains(t, types, utilities.EntityTypeUser)
		assert.Contains(t, types, utilities.EntityTypeProject)
		assert.Contains(t, types, utilities.EntityTypeGeneration)
	})

	t.Run("error handling", func(t *testing.T) {
		factory.shouldError = true

		_, err := factory.CreateEntity(utilities.EntityTypeUser, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "factory error")

		_, err = factory.CreateFromJSON(utilities.EntityTypeUser, []byte(`{}`))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "factory json error")
	})
}

// Tests for EntityValidator interface
func TestEntityValidator_Interface(t *testing.T) {
	validator := &MockEntityValidator{
		rules: map[utilities.EntityType]map[string][]utilities.ValidationRule{
			utilities.EntityTypeUser: {
				"email": {
					{Field: "email", Rule: "required", Message: "Email is required"},
					{Field: "email", Rule: "email", Message: "Email must be valid"},
				},
			},
		},
	}

	entity := &MockDomainEntity{valid: true}

	t.Run("validate entity", func(t *testing.T) {
		err := validator.ValidateEntity(entity)
		assert.NoError(t, err)
	})

	t.Run("validate field", func(t *testing.T) {
		err := validator.ValidateField(entity, "email", "test@example.com")
		assert.NoError(t, err)
	})

	t.Run("get field rules", func(t *testing.T) {
		rules := validator.GetFieldRules(utilities.EntityTypeUser, "email")
		assert.Len(t, rules, 2)
		assert.Equal(t, "email", rules[0].Field)
		assert.Equal(t, "required", rules[0].Rule)
	})

	t.Run("error handling", func(t *testing.T) {
		validator.shouldError = true

		err := validator.ValidateEntity(entity)
		assert.Error(t, err)

		err = validator.ValidateField(entity, "email", "invalid")
		assert.Error(t, err)
	})
}

// Tests for User interface
func TestUser_Interface(t *testing.T) {
	user := &MockUser{
		MockDomainEntity: &MockDomainEntity{
			id:         "user-123",
			entityType: utilities.EntityTypeUser,
			valid:      true,
		},
		email:       "user@example.com",
		username:    "testuser",
		password:    "hashedpassword",
		roles:       []string{"user", "admin"},
		permissions: []string{"read", "write", "delete"},
	}

	t.Run("user-specific methods", func(t *testing.T) {
		assert.Equal(t, "user@example.com", user.GetEmail())
		assert.Equal(t, "testuser", user.GetUsername())

		err := user.SetPassword("newpassword")
		assert.NoError(t, err)
		assert.True(t, user.ValidatePassword("newpassword"))
		assert.False(t, user.ValidatePassword("oldpassword"))

		roles := user.GetRoles()
		assert.Len(t, roles, 2)
		assert.Contains(t, roles, "user")
		assert.Contains(t, roles, "admin")

		assert.True(t, user.HasPermission("read"))
		assert.True(t, user.HasPermission("write"))
		assert.False(t, user.HasPermission("admin"))
	})

	t.Run("implements DomainEntity", func(t *testing.T) {
		var entity utilities.DomainEntity = user
		assert.Equal(t, "user-123", entity.GetID())
		assert.Equal(t, utilities.EntityTypeUser, entity.GetType())
	})
}

// Tests for Project interface
func TestProject_Interface(t *testing.T) {
	project := &MockProject{
		MockDomainEntity: &MockDomainEntity{
			id:         "project-456",
			entityType: utilities.EntityTypeProject,
			valid:      true,
		},
		ownerID:     "user-123",
		name:        "Test Project",
		status:      utilities.ProjectStatusActive,
		generations: []utilities.Generation{},
	}

	generation := &MockGeneration{
		MockDomainEntity: &MockDomainEntity{
			id:         "gen-789",
			entityType: utilities.EntityTypeGeneration,
		},
		projectID: "project-456",
		prompt:    "Generate code",
		content:   "// Generated code here",
	}

	t.Run("project-specific methods", func(t *testing.T) {
		assert.Equal(t, "user-123", project.GetOwnerID())
		assert.Equal(t, "Test Project", project.GetName())
		assert.Equal(t, utilities.ProjectStatusActive, project.GetStatus())

		err := project.SetStatus(utilities.ProjectStatusInactive)
		assert.NoError(t, err)
		assert.Equal(t, utilities.ProjectStatusInactive, project.GetStatus())

		assert.Empty(t, project.GetGenerations())

		err = project.AddGeneration(generation)
		assert.NoError(t, err)

		generations := project.GetGenerations()
		assert.Len(t, generations, 1)
		assert.Equal(t, "gen-789", generations[0].GetID())
	})

	t.Run("implements DomainEntity", func(t *testing.T) {
		var entity utilities.DomainEntity = project
		assert.Equal(t, "project-456", entity.GetID())
		assert.Equal(t, utilities.EntityTypeProject, entity.GetType())
	})
}

// Tests for Generation interface
func TestGeneration_Interface(t *testing.T) {
	generation := &MockGeneration{
		MockDomainEntity: &MockDomainEntity{
			id:         "gen-789",
			entityType: utilities.EntityTypeGeneration,
			valid:      true,
		},
		projectID:  "project-456",
		prompt:     "Generate a REST API",
		content:    "// REST API code here",
		provider:   "openai",
		tokensUsed: 1500,
	}

	t.Run("generation-specific methods", func(t *testing.T) {
		assert.Equal(t, "project-456", generation.GetProjectID())
		assert.Equal(t, "Generate a REST API", generation.GetPrompt())
		assert.Equal(t, "// REST API code here", generation.GetContent())
		assert.Equal(t, "openai", generation.GetProvider())
		assert.Equal(t, 1500, generation.GetTokensUsed())

		err := generation.SetContent("// Updated code here")
		assert.NoError(t, err)
		assert.Equal(t, "// Updated code here", generation.GetContent())
	})

	t.Run("implements DomainEntity", func(t *testing.T) {
		var entity utilities.DomainEntity = generation
		assert.Equal(t, "gen-789", entity.GetID())
		assert.Equal(t, utilities.EntityTypeGeneration, entity.GetType())
	})
}

// Tests for ProjectStatus constants
func TestProjectStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		value    utilities.ProjectStatus
		expected string
	}{
		{"ProjectStatusActive", utilities.ProjectStatusActive, "active"},
		{"ProjectStatusInactive", utilities.ProjectStatusInactive, "inactive"},
		{"ProjectStatusArchived", utilities.ProjectStatusArchived, "archived"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.value))
		})
	}
}

// Integration tests
func TestEntityInterfaces_Integration(t *testing.T) {
	t.Run("complete workflow", func(t *testing.T) {
		// Create factory and validator
		factory := &MockEntityFactory{}
		validator := &MockEntityValidator{}

		// Create user entity
		userData := map[string]interface{}{
			"email":    "test@example.com",
			"username": "testuser",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeUser, userData)
		assert.NoError(t, err)
		assert.NotNil(t, entity)

		// Validate entity
		err = validator.ValidateEntity(entity)
		assert.NoError(t, err)

		// Test serialization round-trip
		jsonData, err := entity.ToJSON()
		assert.NoError(t, err)

		newEntity, err := factory.CreateFromJSON(utilities.EntityTypeUser, jsonData)
		assert.NoError(t, err)
		assert.Equal(t, entity.GetType(), newEntity.GetType())

		// Test change tracking
		entity.MarkDirty("email")
		assert.Contains(t, entity.GetDirtyFields(), "email")

		// Test lifecycle
		assert.NoError(t, entity.BeforeSave())
		assert.NoError(t, entity.AfterSave())
	})

	t.Run("interface compatibility", func(t *testing.T) {
		// Verify all mock implementations satisfy interfaces
		var user utilities.User = &MockUser{
			MockDomainEntity: &MockDomainEntity{
				entityType: utilities.EntityTypeUser,
				valid:      true,
			},
		}
		var project utilities.Project = &MockProject{
			MockDomainEntity: &MockDomainEntity{
				entityType: utilities.EntityTypeProject,
				valid:      true,
			},
		}
		var generation utilities.Generation = &MockGeneration{
			MockDomainEntity: &MockDomainEntity{
				entityType: utilities.EntityTypeGeneration,
				valid:      true,
			},
		}
		var factory utilities.EntityFactory = &MockEntityFactory{}
		var validator utilities.EntityValidator = &MockEntityValidator{}

		// Test polymorphism - all can be treated as DomainEntity
		entities := []utilities.DomainEntity{user, project, generation}
		for _, entity := range entities {
			assert.NotEmpty(t, entity.GetType())
		}

		// Test factory and validator work with interfaces
		types := factory.ListEntityTypes()
		assert.Contains(t, types, utilities.EntityTypeUser)

		err := validator.ValidateEntity(user)
		assert.NoError(t, err)
	})
}
