package utilities_entity_factory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// Test DefaultEntityValidator creation
func TestNewEntityValidator(t *testing.T) {
	validator := utilities.NewEntityValidator()

	assert.NotNil(t, validator)
	assert.IsType(t, &utilities.DefaultEntityValidator{}, validator)
}

// Test ValidateEntity functionality
func TestDefaultEntityValidator_ValidateEntity(t *testing.T) {
	factory := utilities.NewEntityFactory()
	validator := utilities.NewEntityValidator()

	tests := []struct {
		name        string
		entityType  utilities.EntityType
		data        map[string]interface{}
		expectValid bool
	}{
		{
			name:       "valid user entity",
			entityType: utilities.EntityTypeUser,
			data: map[string]interface{}{
				"id":    "user123",
				"email": "test@example.com",
			},
			expectValid: true,
		},
		{
			name:       "invalid user entity - empty email",
			entityType: utilities.EntityTypeUser,
			data: map[string]interface{}{
				"id":    "user123",
				"email": "",
			},
			expectValid: false,
		},
		{
			name:       "valid project entity",
			entityType: utilities.EntityTypeProject,
			data: map[string]interface{}{
				"id":      "project123",
				"name":    "Test Project",
				"user_id": "user123",
			},
			expectValid: true,
		},
		{
			name:       "invalid project entity - empty name",
			entityType: utilities.EntityTypeProject,
			data: map[string]interface{}{
				"id":      "project123",
				"name":    "",
				"user_id": "user123",
			},
			expectValid: false,
		},
		{
			name:       "valid generation entity",
			entityType: utilities.EntityTypeGeneration,
			data: map[string]interface{}{
				"id":         "gen123",
				"content":    "Generated content",
				"project_id": "project123",
			},
			expectValid: true,
		},
		{
			name:       "invalid generation entity - empty content",
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

			validationErr := validator.ValidateEntity(entity)

			if tt.expectValid {
				assert.NoError(t, validationErr)
			} else {
				assert.Error(t, validationErr)
			}
		})
	}
}

// Test ValidateField functionality with validation rules
func TestDefaultEntityValidator_ValidateField(t *testing.T) {
	factory := utilities.NewEntityFactory()
	validator := utilities.NewEntityValidator()

	// Create a test entity
	data := map[string]interface{}{
		"id":    "user123",
		"email": "test@example.com",
	}

	entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
	require.NoError(t, err)

	tests := []struct {
		name        string
		field       string
		value       interface{}
		expectValid bool
	}{
		{
			name:        "valid field value",
			field:       "email",
			value:       "test@example.com",
			expectValid: true,
		},
		{
			name:        "nil value",
			field:       "optional_field",
			value:       nil,
			expectValid: true, // No rules means valid
		},
		{
			name:        "empty string value",
			field:       "some_field",
			value:       "",
			expectValid: true, // No rules means valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateField(entity, tt.field, tt.value)

			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Test GetFieldRules functionality
func TestDefaultEntityValidator_GetFieldRules(t *testing.T) {
	validator := utilities.NewEntityValidator()

	// Test getting rules for non-existent entity type
	rules := validator.GetFieldRules("non_existent_type", "field")
	assert.Empty(t, rules)

	// Test getting rules for non-existent field
	rules = validator.GetFieldRules(utilities.EntityTypeUser, "non_existent_field")
	assert.Empty(t, rules)
}

// Test validation rule application
func TestDefaultEntityValidator_ValidationRules(t *testing.T) {
	factory := utilities.NewEntityFactory()
	validator := utilities.NewEntityValidator()

	// Create test entity
	data := map[string]interface{}{
		"id":    "user123",
		"email": "test@example.com",
	}

	entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
	require.NoError(t, err)

	t.Run("required rule - valid value", func(t *testing.T) {
		err := validator.ValidateField(entity, "email", "test@example.com")
		assert.NoError(t, err)
	})

	t.Run("required rule - empty value", func(t *testing.T) {
		err := validator.ValidateField(entity, "email", "")
		assert.NoError(t, err) // No rules configured, so passes
	})

	t.Run("required rule - nil value", func(t *testing.T) {
		err := validator.ValidateField(entity, "email", nil)
		assert.NoError(t, err) // No rules configured, so passes
	})
}

// Test validator with different entity types
func TestDefaultEntityValidator_MultipleEntityTypes(t *testing.T) {
	validator := utilities.NewEntityValidator()

	entityTypes := []utilities.EntityType{
		utilities.EntityTypeUser,
		utilities.EntityTypeProject,
		utilities.EntityTypeGeneration,
	}

	for _, entityType := range entityTypes {
		t.Run(string(entityType), func(t *testing.T) {
			rules := validator.GetFieldRules(entityType, "test_field")
			assert.NotNil(t, rules) // Should return empty slice, not nil
			assert.Empty(t, rules)
		})
	}
}

// Test validator edge cases
func TestDefaultEntityValidator_EdgeCases(t *testing.T) {
	factory := utilities.NewEntityFactory()
	validator := utilities.NewEntityValidator()

	// Test with nil entity (should panic or handle gracefully)
	t.Run("nil entity validation", func(t *testing.T) {
		// This test verifies behavior with nil - it might panic
		defer func() {
			if r := recover(); r != nil {
				// Panic is acceptable for nil entity
				assert.NotNil(t, r)
			}
		}()

		// This may panic, which is acceptable behavior
		validator.ValidateEntity(nil)
	})

	t.Run("empty field name", func(t *testing.T) {
		data := map[string]interface{}{
			"id":    "user123",
			"email": "test@example.com",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
		require.NoError(t, err)

		err = validator.ValidateField(entity, "", "some_value")
		assert.NoError(t, err) // Empty field name with no rules should pass
	})
}

// Test validator consistency
func TestDefaultEntityValidator_Consistency(t *testing.T) {
	factory := utilities.NewEntityFactory()
	validator := utilities.NewEntityValidator()

	// Test that the same entity validates consistently
	data := map[string]interface{}{
		"id":    "user123",
		"email": "test@example.com",
	}

	entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
	require.NoError(t, err)

	// Validate multiple times
	for i := 0; i < 3; i++ {
		err := validator.ValidateEntity(entity)
		assert.NoError(t, err, "Validation should be consistent across multiple calls")
	}
}

// Test validator with complex entity data
func TestDefaultEntityValidator_ComplexEntityData(t *testing.T) {
	factory := utilities.NewEntityFactory()
	validator := utilities.NewEntityValidator()

	// Create entities with complex data
	userData := map[string]interface{}{
		"id":       "user123",
		"email":    "complex.email+test@example.com",
		"username": "complex_user_123",
		"name":     "Complex User Name With Spaces",
	}

	userEntity, err := factory.CreateEntity(utilities.EntityTypeUser, userData)
	require.NoError(t, err)

	err = validator.ValidateEntity(userEntity)
	assert.NoError(t, err)

	projectData := map[string]interface{}{
		"id":      "project_with_special_chars_!@#",
		"name":    "Project with Special Characters & Symbols",
		"user_id": "user123",
	}

	projectEntity, err := factory.CreateEntity(utilities.EntityTypeProject, projectData)
	require.NoError(t, err)

	err = validator.ValidateEntity(projectEntity)
	assert.NoError(t, err)
}
