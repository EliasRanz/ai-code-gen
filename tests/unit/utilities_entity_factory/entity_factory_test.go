package utilities_entity_factory_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// Test DefaultEntityFactory creation
func TestNewEntityFactory(t *testing.T) {
	factory := utilities.NewEntityFactory()

	assert.NotNil(t, factory)
	assert.IsType(t, &utilities.DefaultEntityFactory{}, factory)
}

// Test entity type listing
func TestDefaultEntityFactory_ListEntityTypes(t *testing.T) {
	factory := utilities.NewEntityFactory()

	types := factory.ListEntityTypes()

	assert.Len(t, types, 3)
	assert.Contains(t, types, utilities.EntityTypeUser)
	assert.Contains(t, types, utilities.EntityTypeProject)
	assert.Contains(t, types, utilities.EntityTypeGeneration)
}

// Test User Entity Creation
func TestDefaultEntityFactory_CreateEntity_User_Success(t *testing.T) {
	factory := utilities.NewEntityFactory()

	data := map[string]interface{}{
		"id":       "user123",
		"email":    "test@example.com",
		"username": "testuser",
		"name":     "Test User",
	}

	entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)

	require.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, utilities.EntityTypeUser, entity.GetType())
	assert.Equal(t, "user123", entity.GetID())

	// Verify it's a user entity
	userEntity, ok := entity.(*utilities.BasicUserEntity)
	require.True(t, ok)
	assert.Equal(t, "test@example.com", userEntity.GetEmail())
	assert.Equal(t, "testuser", userEntity.GetUsername())
	assert.Equal(t, "Test User", userEntity.Name)
}

func TestDefaultEntityFactory_CreateEntity_User_RequiredFields(t *testing.T) {
	factory := utilities.NewEntityFactory()

	tests := []struct {
		name        string
		data        map[string]interface{}
		expectedErr string
	}{
		{
			name:        "missing id",
			data:        map[string]interface{}{"email": "test@example.com"},
			expectedErr: "id is required for user entity",
		},
		{
			name:        "missing email",
			data:        map[string]interface{}{"id": "user123"},
			expectedErr: "email is required for user entity",
		},
		{
			name: "invalid id type",
			data: map[string]interface{}{
				"id":    123,
				"email": "test@example.com",
			},
			expectedErr: "id is required for user entity",
		},
		{
			name: "invalid email type",
			data: map[string]interface{}{
				"id":    "user123",
				"email": 123,
			},
			expectedErr: "email is required for user entity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := factory.CreateEntity(utilities.EntityTypeUser, tt.data)

			assert.Error(t, err)
			assert.Nil(t, entity)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestDefaultEntityFactory_CreateEntity_User_OptionalFields(t *testing.T) {
	factory := utilities.NewEntityFactory()

	// Test with minimal data (only required fields)
	data := map[string]interface{}{
		"id":    "user123",
		"email": "test@example.com",
	}

	entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)

	require.NoError(t, err)
	userEntity, ok := entity.(*utilities.BasicUserEntity)
	require.True(t, ok)

	assert.Equal(t, "user123", userEntity.GetID())
	assert.Equal(t, "test@example.com", userEntity.GetEmail())
	assert.Equal(t, "", userEntity.GetUsername()) // Optional field defaults to empty
	assert.Equal(t, "", userEntity.Name)          // Optional field defaults to empty
}

// Test Project Entity Creation
func TestDefaultEntityFactory_CreateEntity_Project_Success(t *testing.T) {
	factory := utilities.NewEntityFactory()

	data := map[string]interface{}{
		"id":      "project123",
		"name":    "Test Project",
		"user_id": "user123",
	}

	entity, err := factory.CreateEntity(utilities.EntityTypeProject, data)

	require.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, utilities.EntityTypeProject, entity.GetType())
	assert.Equal(t, "project123", entity.GetID())

	// Verify it's a project entity
	projectEntity, ok := entity.(*utilities.BasicProjectEntity)
	require.True(t, ok)
	assert.Equal(t, "Test Project", projectEntity.GetName())
	assert.Equal(t, "user123", projectEntity.GetOwnerID())
	assert.Equal(t, utilities.ProjectStatusActive, projectEntity.GetStatus())
}

func TestDefaultEntityFactory_CreateEntity_Project_RequiredFields(t *testing.T) {
	factory := utilities.NewEntityFactory()

	tests := []struct {
		name        string
		data        map[string]interface{}
		expectedErr string
	}{
		{
			name: "missing id",
			data: map[string]interface{}{
				"name":    "Test Project",
				"user_id": "user123",
			},
			expectedErr: "id is required for project entity",
		},
		{
			name: "missing name",
			data: map[string]interface{}{
				"id":      "project123",
				"user_id": "user123",
			},
			expectedErr: "name is required for project entity",
		},
		{
			name: "missing user_id",
			data: map[string]interface{}{
				"id":   "project123",
				"name": "Test Project",
			},
			expectedErr: "user_id is required for project entity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := factory.CreateEntity(utilities.EntityTypeProject, tt.data)

			assert.Error(t, err)
			assert.Nil(t, entity)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// Test Generation Entity Creation
func TestDefaultEntityFactory_CreateEntity_Generation_Success(t *testing.T) {
	factory := utilities.NewEntityFactory()

	data := map[string]interface{}{
		"id":         "gen123",
		"content":    "Generated code content",
		"project_id": "project123",
	}

	entity, err := factory.CreateEntity(utilities.EntityTypeGeneration, data)

	require.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, utilities.EntityTypeGeneration, entity.GetType())
	assert.Equal(t, "gen123", entity.GetID())

	// Verify it's a generation entity
	genEntity, ok := entity.(*utilities.BasicGenerationEntity)
	require.True(t, ok)
	assert.Equal(t, "Generated code content", genEntity.GetContent())
	assert.Equal(t, "project123", genEntity.GetProjectID())
}

func TestDefaultEntityFactory_CreateEntity_Generation_RequiredFields(t *testing.T) {
	factory := utilities.NewEntityFactory()

	tests := []struct {
		name        string
		data        map[string]interface{}
		expectedErr string
	}{
		{
			name: "missing id",
			data: map[string]interface{}{
				"content":    "test content",
				"project_id": "project123",
			},
			expectedErr: "id is required for generation entity",
		},
		{
			name: "missing content",
			data: map[string]interface{}{
				"id":         "gen123",
				"project_id": "project123",
			},
			expectedErr: "content is required for generation entity",
		},
		{
			name: "missing project_id",
			data: map[string]interface{}{
				"id":      "gen123",
				"content": "test content",
			},
			expectedErr: "project_id is required for generation entity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := factory.CreateEntity(utilities.EntityTypeGeneration, tt.data)

			assert.Error(t, err)
			assert.Nil(t, entity)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// Test unknown entity type
func TestDefaultEntityFactory_CreateEntity_UnknownType(t *testing.T) {
	factory := utilities.NewEntityFactory()

	data := map[string]interface{}{"id": "test123"}
	entity, err := factory.CreateEntity("unknown_type", data)

	assert.Error(t, err)
	assert.Nil(t, entity)
	assert.Contains(t, err.Error(), "unknown entity type: unknown_type")
}

// Test CreateFromJSON functionality
func TestDefaultEntityFactory_CreateFromJSON_Success(t *testing.T) {
	factory := utilities.NewEntityFactory()

	tests := []struct {
		name       string
		entityType utilities.EntityType
		jsonData   string
	}{
		{
			name:       "user from JSON",
			entityType: utilities.EntityTypeUser,
			jsonData:   `{"id":"user123","email":"test@example.com","username":"testuser"}`,
		},
		{
			name:       "project from JSON",
			entityType: utilities.EntityTypeProject,
			jsonData:   `{"id":"project123","name":"Test Project","user_id":"user123"}`,
		},
		{
			name:       "generation from JSON",
			entityType: utilities.EntityTypeGeneration,
			jsonData:   `{"id":"gen123","content":"Generated content","project_id":"project123"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := factory.CreateFromJSON(tt.entityType, []byte(tt.jsonData))

			require.NoError(t, err)
			assert.NotNil(t, entity)
			assert.Equal(t, tt.entityType, entity.GetType())
		})
	}
}

func TestDefaultEntityFactory_CreateFromJSON_InvalidJSON(t *testing.T) {
	factory := utilities.NewEntityFactory()

	entity, err := factory.CreateFromJSON(utilities.EntityTypeUser, []byte("invalid json"))

	assert.Error(t, err)
	assert.Nil(t, entity)
	assert.Contains(t, err.Error(), "failed to unmarshal JSON")
}

func TestDefaultEntityFactory_CreateFromJSON_MissingRequiredFields(t *testing.T) {
	factory := utilities.NewEntityFactory()

	// JSON with missing required field
	jsonData := `{"id":"user123"}` // missing email

	entity, err := factory.CreateFromJSON(utilities.EntityTypeUser, []byte(jsonData))

	assert.Error(t, err)
	assert.Nil(t, entity)
	assert.Contains(t, err.Error(), "email is required for user entity")
}

// Test BaseEntity functionality through created entities
func TestEntityFactory_BaseEntityFunctionality(t *testing.T) {
	factory := utilities.NewEntityFactory()

	data := map[string]interface{}{
		"id":    "user123",
		"email": "test@example.com",
	}

	entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
	require.NoError(t, err)

	// Test BaseEntity methods
	assert.Equal(t, "user123", entity.GetID())
	assert.Equal(t, utilities.EntityTypeUser, entity.GetType())
	assert.NotNil(t, entity.GetCreatedAt())
	assert.NotNil(t, entity.GetUpdatedAt())

	// Test ToMap
	entityMap := entity.ToMap()
	assert.Equal(t, "user123", entityMap["id"])
	assert.Equal(t, string(utilities.EntityTypeUser), entityMap["type"])

	// Test ToJSON
	jsonData, err := entity.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), `"id":"user123"`)
	assert.Contains(t, string(jsonData), `"type":"user"`)
}

// Test entity timestamps
func TestEntityFactory_EntityTimestamps(t *testing.T) {
	factory := utilities.NewEntityFactory()

	before := time.Now()

	data := map[string]interface{}{
		"id":    "user123",
		"email": "test@example.com",
	}

	entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
	require.NoError(t, err)

	after := time.Now()

	createdAt := entity.GetCreatedAt()
	updatedAt := entity.GetUpdatedAt()

	assert.True(t, createdAt.After(before) || createdAt.Equal(before))
	assert.True(t, createdAt.Before(after) || createdAt.Equal(after))
	assert.True(t, updatedAt.After(before) || updatedAt.Equal(before))
	assert.True(t, updatedAt.Before(after) || updatedAt.Equal(after))
}

// Test entity validation
func TestEntityFactory_EntityValidation(t *testing.T) {
	factory := utilities.NewEntityFactory()

	tests := []struct {
		name        string
		entityType  utilities.EntityType
		data        map[string]interface{}
		expectValid bool
	}{
		{
			name:       "valid user",
			entityType: utilities.EntityTypeUser,
			data: map[string]interface{}{
				"id":    "user123",
				"email": "test@example.com",
			},
			expectValid: true,
		},
		{
			name:       "user with empty email",
			entityType: utilities.EntityTypeUser,
			data: map[string]interface{}{
				"id":    "user123",
				"email": "",
			},
			expectValid: false,
		},
		{
			name:       "valid project",
			entityType: utilities.EntityTypeProject,
			data: map[string]interface{}{
				"id":      "project123",
				"name":    "Test Project",
				"user_id": "user123",
			},
			expectValid: true,
		},
		{
			name:       "project with empty name",
			entityType: utilities.EntityTypeProject,
			data: map[string]interface{}{
				"id":      "project123",
				"name":    "",
				"user_id": "user123",
			},
			expectValid: false,
		},
		{
			name:       "valid generation",
			entityType: utilities.EntityTypeGeneration,
			data: map[string]interface{}{
				"id":         "gen123",
				"content":    "Generated content",
				"project_id": "project123",
			},
			expectValid: true,
		},
		{
			name:       "generation with empty content",
			entityType: utilities.EntityTypeGeneration,
			data: map[string]interface{}{
				"id":         "gen123",
				"content":    "",
				"project_id": "project123",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := factory.CreateEntity(tt.entityType, tt.data)
			require.NoError(t, err)

			validationErr := entity.Validate()

			if tt.expectValid {
				assert.NoError(t, validationErr)
			} else {
				assert.Error(t, validationErr)
			}
		})
	}
}

// Test entity interface methods
func TestEntityFactory_EntityInterfaceMethods(t *testing.T) {
	factory := utilities.NewEntityFactory()

	t.Run("User entity interface methods", func(t *testing.T) {
		data := map[string]interface{}{
			"id":       "user123",
			"email":    "test@example.com",
			"username": "testuser",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
		require.NoError(t, err)

		userEntity, ok := entity.(*utilities.BasicUserEntity)
		require.True(t, ok)

		assert.Equal(t, "test@example.com", userEntity.GetEmail())
		assert.Equal(t, "testuser", userEntity.GetUsername())
		assert.Empty(t, userEntity.GetRoles())
		assert.False(t, userEntity.HasPermission("any_permission"))
		assert.False(t, userEntity.ValidatePassword("password"))

		err = userEntity.SetPassword("password")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("Project entity interface methods", func(t *testing.T) {
		data := map[string]interface{}{
			"id":      "project123",
			"name":    "Test Project",
			"user_id": "user123",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeProject, data)
		require.NoError(t, err)

		projectEntity, ok := entity.(*utilities.BasicProjectEntity)
		require.True(t, ok)

		assert.Equal(t, "Test Project", projectEntity.GetName())
		assert.Equal(t, "user123", projectEntity.GetOwnerID())
		assert.Equal(t, utilities.ProjectStatusActive, projectEntity.GetStatus())
		assert.Empty(t, projectEntity.GetGenerations())

		err = projectEntity.SetStatus(utilities.ProjectStatusArchived)
		assert.NoError(t, err)
		assert.Equal(t, utilities.ProjectStatusArchived, projectEntity.GetStatus())

		err = projectEntity.AddGeneration(nil)
		assert.NoError(t, err) // Basic implementation allows nil
	})

	t.Run("Generation entity interface methods", func(t *testing.T) {
		data := map[string]interface{}{
			"id":         "gen123",
			"content":    "Generated content",
			"project_id": "project123",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeGeneration, data)
		require.NoError(t, err)

		genEntity, ok := entity.(*utilities.BasicGenerationEntity)
		require.True(t, ok)

		assert.Equal(t, "Generated content", genEntity.GetContent())
		assert.Equal(t, "project123", genEntity.GetProjectID())
		assert.Equal(t, "", genEntity.GetPrompt())
		assert.Equal(t, "", genEntity.GetProvider())
		assert.Equal(t, 0, genEntity.GetTokensUsed())

		err = genEntity.SetContent("New content")
		assert.NoError(t, err)
		assert.Equal(t, "New content", genEntity.GetContent())
	})
}

// Test complex JSON scenarios
func TestEntityFactory_ComplexJSONScenarios(t *testing.T) {
	factory := utilities.NewEntityFactory()

	t.Run("JSON with extra fields", func(t *testing.T) {
		jsonData := `{
			"id": "user123",
			"email": "test@example.com",
			"username": "testuser",
			"name": "Test User",
			"extra_field": "should be ignored",
			"nested": {"data": "ignored"}
		}`

		entity, err := factory.CreateFromJSON(utilities.EntityTypeUser, []byte(jsonData))
		require.NoError(t, err)

		userEntity, ok := entity.(*utilities.BasicUserEntity)
		require.True(t, ok)
		assert.Equal(t, "test@example.com", userEntity.GetEmail())
		assert.Equal(t, "testuser", userEntity.GetUsername())
		assert.Equal(t, "Test User", userEntity.Name)
	})

	t.Run("JSON with null values", func(t *testing.T) {
		jsonData := `{
			"id": "user123",
			"email": "test@example.com",
			"username": null,
			"name": null
		}`

		entity, err := factory.CreateFromJSON(utilities.EntityTypeUser, []byte(jsonData))
		require.NoError(t, err)

		userEntity, ok := entity.(*utilities.BasicUserEntity)
		require.True(t, ok)
		assert.Equal(t, "test@example.com", userEntity.GetEmail())
		assert.Equal(t, "", userEntity.GetUsername()) // null becomes empty string
		assert.Equal(t, "", userEntity.Name)
	})
}
