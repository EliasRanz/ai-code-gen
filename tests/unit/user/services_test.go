package user

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// TestUserEntity tests the User entity methods
func TestUserEntity(t *testing.T) {
	t.Run("IsAdmin should return true for admin role", func(t *testing.T) {
		u := user.User{
			ID:   utilities.UserID("test-id"),
			Role: user.RoleAdmin,
		}

		assert.True(t, u.IsAdmin())
	})

	t.Run("IsAdmin should return false for user role", func(t *testing.T) {
		u := user.User{
			ID:   utilities.UserID("test-id"),
			Role: user.RoleUser,
		}

		assert.False(t, u.IsAdmin())
	})

	t.Run("CanAccessProject should return true for own project", func(t *testing.T) {
		userID := utilities.UserID("test-id")
		u := user.User{
			ID:   userID,
			Role: user.RoleUser,
		}

		assert.True(t, u.CanAccessProject(userID))
	})
}

// TestUserStatuses tests user status constants
func TestUserStatuses(t *testing.T) {
	t.Run("should have correct status values", func(t *testing.T) {
		assert.Equal(t, user.UserStatus("active"), user.StatusActiveUser)
		assert.Equal(t, user.UserStatus("inactive"), user.StatusInactiveUser)
		assert.Equal(t, user.UserStatus("suspended"), user.StatusSuspendedUser)
	})
}

// TestUserRoles tests user role constants
func TestUserRoles(t *testing.T) {
	t.Run("should have correct role values", func(t *testing.T) {
		assert.Equal(t, user.Role("user"), user.RoleUser)
		assert.Equal(t, user.Role("admin"), user.RoleAdmin)
	})
}
