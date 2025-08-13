package user_test

import (
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserEntity(t *testing.T) {
	t.Run("User creation and basic properties", func(t *testing.T) {
		userID := utilities.UserID("user-123")
		u := user.NewUser(userID, "test@example.com", "testuser", "Test User")

		assert.Equal(t, userID, u.ID)
		assert.Equal(t, "test@example.com", u.Email)
		assert.Equal(t, "testuser", u.Username)
		assert.Equal(t, "Test User", u.Name)
		assert.Equal(t, user.RoleUser, u.Role)
		assert.True(t, u.Active)
		assert.Equal(t, user.StatusActiveUser, u.Status)

		// Test DomainEntity interface implementation
		assert.Equal(t, string(userID), u.GetID())
		assert.Equal(t, utilities.EntityTypeUser, u.GetType())
		assert.Equal(t, int64(1), u.GetVersion())
	})

	t.Run("User validation - valid user", func(t *testing.T) {
		userID := utilities.UserID("user-123")
		u := user.NewUser(userID, "test@example.com", "testuser", "Test User")

		err := u.Validate()
		assert.NoError(t, err)
		assert.True(t, u.IsValid())
	})

	t.Run("User validation - invalid email", func(t *testing.T) {
		userID := utilities.UserID("user-123")
		u := user.NewUser(userID, "", "testuser", "Test User")

		err := u.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
		assert.False(t, u.IsValid())
	})

	t.Run("User validation - invalid username", func(t *testing.T) {
		userID := utilities.UserID("user-123")
		u := user.NewUser(userID, "test@example.com", "", "Test User")

		err := u.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username")
		assert.False(t, u.IsValid())
	})

	t.Run("User ToMap conversion", func(t *testing.T) {
		userID := utilities.UserID("user-123")
		u := user.NewUser(userID, "test@example.com", "testuser", "Test User")

		userMap := u.ToMap()
		assert.Equal(t, string(userID), userMap["id"])
		assert.Equal(t, "test@example.com", userMap["email"])
		assert.Equal(t, "testuser", userMap["username"])
		assert.Equal(t, "Test User", userMap["name"])
		assert.Equal(t, string(user.RoleUser), userMap["role"])
		assert.Equal(t, true, userMap["active"])
		assert.Equal(t, string(user.StatusActiveUser), userMap["status"])
	})

	t.Run("User ToJSON conversion", func(t *testing.T) {
		userID := utilities.UserID("user-123")
		u := user.NewUser(userID, "test@example.com", "testuser", "Test User")

		jsonData, err := u.ToJSON()
		assert.NoError(t, err)
		assert.Contains(t, string(jsonData), "test@example.com")
		assert.Contains(t, string(jsonData), "testuser")
		assert.Contains(t, string(jsonData), "Test User")
	})

	t.Run("User FromJSON conversion", func(t *testing.T) {
		userID := utilities.UserID("user-123")
		u := user.NewUser(userID, "test@example.com", "testuser", "Test User")

		jsonData, err := u.ToJSON()
		require.NoError(t, err)

		var newUser user.User
		err = newUser.FromJSON(jsonData)
		assert.NoError(t, err)
		assert.Equal(t, u.Email, newUser.Email)
		assert.Equal(t, u.Username, newUser.Username)
		assert.Equal(t, u.Name, newUser.Name)
	})
}

func TestUserRolesAndPermissions(t *testing.T) {
	t.Run("Default role assignment", func(t *testing.T) {
		userID := utilities.UserID("user-123")
		u := user.NewUser(userID, "test@example.com", "testuser", "Test User")

		assert.Equal(t, user.RoleUser, u.Role)
	})

	t.Run("Role validation", func(t *testing.T) {
		testCases := []struct {
			role  user.Role
			valid bool
		}{
			{user.RoleUser, true},
			{user.RoleAdmin, true},
			{user.Role("invalid"), false},
		}

		for _, tc := range testCases {
			userID := utilities.UserID("user-123")
			u := user.NewUser(userID, "test@example.com", "testuser", "Test User")
			u.Role = tc.role

			err := u.Validate()
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		}
	})
}

func TestProjectEntity(t *testing.T) {
	t.Run("Project creation and basic properties", func(t *testing.T) {
		projectID := utilities.ProjectID("project-123")
		userID := utilities.UserID("user-123")
		p := user.NewProject(projectID, "Test Project", "A test project", userID)

		assert.Equal(t, projectID, p.ID)
		assert.Equal(t, "Test Project", p.Name)
		assert.Equal(t, "A test project", p.Description)
		assert.Equal(t, userID, p.UserID)
		assert.Equal(t, user.StatusActive, p.Status)

		// Test DomainEntity interface implementation
		assert.Equal(t, string(projectID), p.GetID())
		assert.Equal(t, utilities.EntityTypeProject, p.GetType())
		assert.Equal(t, int64(1), p.GetVersion())
	})

	t.Run("Project validation - valid project", func(t *testing.T) {
		projectID := utilities.ProjectID("project-123")
		userID := utilities.UserID("user-123")
		p := user.NewProject(projectID, "Test Project", "A test project", userID)

		err := p.Validate()
		assert.NoError(t, err)
		assert.True(t, p.IsValid())
	})

	t.Run("Project validation - missing name", func(t *testing.T) {
		projectID := utilities.ProjectID("project-123")
		userID := utilities.UserID("user-123")
		p := user.NewProject(projectID, "", "A test project", userID)

		err := p.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
		assert.False(t, p.IsValid())
	})

	t.Run("Project validation - missing user ID", func(t *testing.T) {
		projectID := utilities.ProjectID("project-123")
		p := user.NewProject(projectID, "Test Project", "A test project", utilities.UserID(""))

		err := p.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user")
		assert.False(t, p.IsValid())
	})

	t.Run("Project ToMap conversion", func(t *testing.T) {
		projectID := utilities.ProjectID("project-123")
		userID := utilities.UserID("user-123")
		p := user.NewProject(projectID, "Test Project", "A test project", userID)

		projectMap := p.ToMap()
		assert.Equal(t, string(projectID), projectMap["id"])
		assert.Equal(t, "Test Project", projectMap["name"])
		assert.Equal(t, "A test project", projectMap["description"])
		assert.Equal(t, string(userID), projectMap["user_id"])
		assert.Equal(t, string(user.StatusActive), projectMap["status"])
	})

	t.Run("Project ToJSON conversion", func(t *testing.T) {
		projectID := utilities.ProjectID("project-123")
		userID := utilities.UserID("user-123")
		p := user.NewProject(projectID, "Test Project", "A test project", userID)

		jsonData, err := p.ToJSON()
		assert.NoError(t, err)
		assert.Contains(t, string(jsonData), "Test Project")
		assert.Contains(t, string(jsonData), "A test project")
		assert.Contains(t, string(jsonData), string(userID))
	})

	t.Run("Project status transitions", func(t *testing.T) {
		projectID := utilities.ProjectID("project-123")
		userID := utilities.UserID("user-123")
		p := user.NewProject(projectID, "Test Project", "A test project", userID)

		// Test status change
		p.Status = user.StatusArchived
		assert.Equal(t, user.StatusArchived, p.Status)

		// Archived project should still be valid
		err := p.Validate()
		assert.NoError(t, err)

		// Test inactive status
		p.Status = user.StatusInactive
		assert.Equal(t, user.StatusInactive, p.Status)
	})
}

func TestProjectInterfaceMethods(t *testing.T) {
	t.Run("Project interface compliance", func(t *testing.T) {
		projectID := utilities.ProjectID("project-123")
		userID := utilities.UserID("user-123")
		p := user.NewProject(projectID, "Test Project", "A test project", userID)

		// Test that project implements the Project interface
		var project utilities.Project = p
		assert.Equal(t, "Test Project", project.GetName())
		assert.Equal(t, utilities.ProjectStatusActive, project.GetStatus())
	})

	t.Run("Project status checking", func(t *testing.T) {
		projectID := utilities.ProjectID("project-123")
		userID := utilities.UserID("user-123")
		p := user.NewProject(projectID, "Test Project", "A test project", userID)

		// Test initial status
		assert.Equal(t, user.StatusActive, p.Status)

		// Test status changes
		p.Status = user.StatusArchived
		assert.Equal(t, user.StatusArchived, p.Status)

		p.Status = user.StatusInactive
		assert.Equal(t, user.StatusInactive, p.Status)
	})
}

func TestUserAndProjectLifecycle(t *testing.T) {
	t.Run("User and project entity lifecycle", func(t *testing.T) {
		// Create user
		userID := utilities.UserID("user-123")
		u := user.NewUser(userID, "test@example.com", "testuser", "Test User")

		// Validate user
		err := u.Validate()
		assert.NoError(t, err)

		// Create project for user
		projectID := utilities.ProjectID("project-123")
		p := user.NewProject(projectID, "Test Project", "A test project", userID)

		// Validate project
		err = p.Validate()
		assert.NoError(t, err)

		// Test entity relationships
		assert.Equal(t, userID, p.UserID)

		// Test entity versions
		assert.Equal(t, int64(1), u.GetVersion())
		assert.Equal(t, int64(1), p.GetVersion())

		// Test entity IDs
		assert.Equal(t, string(userID), u.GetID())
		assert.Equal(t, string(projectID), p.GetID())

		// Test entity types
		assert.Equal(t, utilities.EntityTypeUser, u.GetType())
		assert.Equal(t, utilities.EntityTypeProject, p.GetType())
	})
}
