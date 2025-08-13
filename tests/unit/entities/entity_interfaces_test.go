// Tests for entity interface pattern and factory implementations
package entities_test

import (
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityInterfacePattern(t *testing.T) {
	t.Run("BaseEntity implements DomainEntity interface", func(t *testing.T) {
		base := utilities.NewBaseEntity("test-id", utilities.EntityTypeUser)

		assert.Equal(t, "test-id", base.GetID())
		assert.Equal(t, utilities.EntityTypeUser, base.GetType())
		assert.Equal(t, int64(1), base.GetVersion())
		assert.True(t, base.IsValid())
		assert.NoError(t, base.Validate())
		assert.Empty(t, base.GetDirtyFields())
		assert.NoError(t, base.BeforeSave())
		assert.NoError(t, base.AfterSave())
		assert.NoError(t, base.BeforeDelete())
	})

	t.Run("BaseEntity change tracking", func(t *testing.T) {
		base := utilities.NewBaseEntity("test-id", utilities.EntityTypeUser)

		base.MarkDirty("field1")
		base.MarkDirty("field2")
		base.MarkDirty("field1") // Duplicate should be ignored

		dirtyFields := base.GetDirtyFields()
		assert.Len(t, dirtyFields, 2)
		assert.Contains(t, dirtyFields, "field1")
		assert.Contains(t, dirtyFields, "field2")

		base.ClearDirtyFields()
		assert.Empty(t, base.GetDirtyFields())
	})

	t.Run("BaseEntity serialization", func(t *testing.T) {
		base := utilities.NewBaseEntity("test-id", utilities.EntityTypeProject)

		// Test ToMap
		data := base.ToMap()
		assert.Equal(t, "test-id", data["id"])
		assert.Equal(t, "project", data["type"])
		assert.Equal(t, int64(1), data["version"])

		// Test ToJSON
		jsonData, err := base.ToJSON()
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "test-id")
		assert.Contains(t, string(jsonData), "project")

		// Test FromJSON
		err = base.FromJSON(jsonData)
		assert.NoError(t, err) // Base implementation should not error
	})
}

func TestEntityFactory(t *testing.T) {
	factory := utilities.NewEntityFactory()

	t.Run("Factory returns supported entity types", func(t *testing.T) {
		types := factory.ListEntityTypes()
		assert.Len(t, types, 3)
		assert.Contains(t, types, utilities.EntityTypeUser)
		assert.Contains(t, types, utilities.EntityTypeProject)
		assert.Contains(t, types, utilities.EntityTypeGeneration)
	})

	t.Run("Create user entity from map", func(t *testing.T) {
		data := map[string]interface{}{
			"id":       "user-123",
			"email":    "test@example.com",
			"username": "testuser",
			"name":     "Test User",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
		require.NoError(t, err)
		assert.NotNil(t, entity)
		assert.Equal(t, utilities.EntityTypeUser, entity.GetType())

		// Test User interface methods
		user, ok := entity.(utilities.User)
		require.True(t, ok)
		assert.Equal(t, "test@example.com", user.GetEmail())
		assert.Equal(t, "testuser", user.GetUsername())
	})

	t.Run("Create project entity from map", func(t *testing.T) {
		data := map[string]interface{}{
			"id":      "project-456",
			"name":    "Test Project",
			"user_id": "user-123",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeProject, data)
		require.NoError(t, err)
		assert.NotNil(t, entity)
		assert.Equal(t, utilities.EntityTypeProject, entity.GetType())

		// Test Project interface methods
		project, ok := entity.(utilities.Project)
		require.True(t, ok)
		assert.Equal(t, "Test Project", project.GetName())
		assert.Equal(t, "user-123", project.GetOwnerID())
		assert.Equal(t, utilities.ProjectStatusActive, project.GetStatus())
	})

	t.Run("Create generation entity from map", func(t *testing.T) {
		data := map[string]interface{}{
			"id":         "gen-789",
			"content":    "generated code here",
			"project_id": "project-456",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeGeneration, data)
		require.NoError(t, err)
		assert.NotNil(t, entity)
		assert.Equal(t, utilities.EntityTypeGeneration, entity.GetType())

		// Test Generation interface methods
		generation, ok := entity.(utilities.Generation)
		require.True(t, ok)
		assert.Equal(t, "generated code here", generation.GetContent())
		assert.Equal(t, "project-456", generation.GetProjectID())
	})

	t.Run("Create entity from JSON", func(t *testing.T) {
		jsonData := `{"id":"user-999","email":"json@example.com","username":"jsonuser"}`

		entity, err := factory.CreateFromJSON(utilities.EntityTypeUser, []byte(jsonData))
		require.NoError(t, err)
		assert.NotNil(t, entity)

		user, ok := entity.(utilities.User)
		require.True(t, ok)
		assert.Equal(t, "json@example.com", user.GetEmail())
		assert.Equal(t, "jsonuser", user.GetUsername())
	})

	t.Run("Factory validation errors", func(t *testing.T) {
		// Test unknown entity type
		_, err := factory.CreateEntity("unknown", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown entity type")

		// Test missing required fields
		_, err = factory.CreateEntity(utilities.EntityTypeUser, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")

		// Test invalid JSON
		_, err = factory.CreateFromJSON(utilities.EntityTypeUser, []byte("invalid json"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal JSON")
	})
}

func TestEntityValidator(t *testing.T) {
	validator := utilities.NewEntityValidator()
	factory := utilities.NewEntityFactory()

	t.Run("Validate user entity", func(t *testing.T) {
		data := map[string]interface{}{
			"id":       "user-123",
			"email":    "test@example.com",
			"username": "testuser",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeUser, data)
		require.NoError(t, err)

		err = validator.ValidateEntity(entity)
		assert.NoError(t, err)
	})

	t.Run("Validate project entity", func(t *testing.T) {
		data := map[string]interface{}{
			"id":      "project-456",
			"name":    "Test Project",
			"user_id": "user-123",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeProject, data)
		require.NoError(t, err)

		err = validator.ValidateEntity(entity)
		assert.NoError(t, err)
	})

	t.Run("Validate generation entity", func(t *testing.T) {
		data := map[string]interface{}{
			"id":         "gen-789",
			"content":    "generated content",
			"project_id": "project-456",
		}

		entity, err := factory.CreateEntity(utilities.EntityTypeGeneration, data)
		require.NoError(t, err)

		err = validator.ValidateEntity(entity)
		assert.NoError(t, err)
	})

	t.Run("Validate invalid entities", func(t *testing.T) {
		// User without email - factory will fail
		data := map[string]interface{}{
			"id":       "user-123",
			"username": "testuser",
		}

		_, err := factory.CreateEntity(utilities.EntityTypeUser, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")

		// Project without name - factory will fail
		data = map[string]interface{}{
			"id":      "project-456",
			"user_id": "user-123",
		}

		_, err = factory.CreateEntity(utilities.EntityTypeProject, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})
}

func TestEntityLifecycleEvents(t *testing.T) {
	t.Run("Entity lifecycle hooks are called", func(t *testing.T) {
		base := utilities.NewBaseEntity("test-id", utilities.EntityTypeUser)

		// Test lifecycle hooks don't error by default
		assert.NoError(t, base.BeforeSave())
		assert.NoError(t, base.AfterSave())
		assert.NoError(t, base.BeforeDelete())

		// Test AfterSave clears dirty fields
		base.MarkDirty("field1")
		assert.Len(t, base.GetDirtyFields(), 1)

		err := base.AfterSave()
		assert.NoError(t, err)
		assert.Empty(t, base.GetDirtyFields())
	})
}

func TestEntityTimestamps(t *testing.T) {
	t.Run("Entity timestamps are set correctly", func(t *testing.T) {
		start := time.Now()
		base := utilities.NewBaseEntity("test-id", utilities.EntityTypeUser)
		end := time.Now()

		createdAt := base.GetCreatedAt()
		updatedAt := base.GetUpdatedAt()

		assert.True(t, createdAt.After(start) || createdAt.Equal(start))
		assert.True(t, createdAt.Before(end) || createdAt.Equal(end))
		assert.True(t, updatedAt.After(start) || updatedAt.Equal(start))
		assert.True(t, updatedAt.Before(end) || updatedAt.Equal(end))
	})
}
