package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/observability"
	authtest "github.com/EliasRanz/ai-code-gen/tests/unit/auth"
)

func TestCheckRoleIntegration(t *testing.T) {
	t.Run("end-to-end role check success", func(t *testing.T) {
		// Setup
		logger := observability.NewLogger("debug", "console")
		userRepo := new(authtest.MockUserRepository)

		// Create use case with real dependencies
		checkRoleUC := auth.NewCheckRole(userRepo)

		// Setup user data
		userData := auth.User{
			ID:       auth.UserID("test-user-id"),
			Email:    "test@example.com",
			Username: "testuser",
			Name:     "Test User",
			Roles:    []string{"admin", "editor"},
			Active:   true,
		}

		userRepo.On("GetByID", mock.Anything, auth.UserID("test-user-id")).Return(userData, nil)

		// Execute the use case directly (simulating end-to-end flow)
		req := auth.CheckRoleRequest{
			UserID:       auth.UserID("test-user-id"),
			RequiredRole: "editor",
		}

		resp, err := checkRoleUC.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Authorized)
		assert.Equal(t, []string{"admin", "editor"}, resp.UserRoles)
		assert.Empty(t, resp.Reason)

		logger.Info("Integration test completed", map[string]interface{}{
			"test":       "check_role_success",
			"authorized": resp.Authorized,
		})

		userRepo.AssertExpectations(t)
	})

	t.Run("end-to-end role check failure", func(t *testing.T) {
		// Setup
		logger := observability.NewLogger("debug", "console")
		userRepo := new(authtest.MockUserRepository)

		// Create use case with real dependencies
		checkRoleUC := auth.NewCheckRole(userRepo)

		// Setup user data with insufficient role
		userData := auth.User{
			ID:       auth.UserID("test-user-id"),
			Email:    "test@example.com",
			Username: "testuser",
			Name:     "Test User",
			Roles:    []string{"viewer"},
			Active:   true,
		}

		userRepo.On("GetByID", mock.Anything, auth.UserID("test-user-id")).Return(userData, nil)

		// Execute the use case directly (simulating end-to-end flow)
		req := auth.CheckRoleRequest{
			UserID:       auth.UserID("test-user-id"),
			RequiredRole: "admin",
		}

		resp, err := checkRoleUC.Execute(context.Background(), req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Authorized)
		assert.Equal(t, []string{"viewer"}, resp.UserRoles)
		assert.Equal(t, "user does not have required role: admin", resp.Reason)

		logger.Info("Integration test completed", map[string]interface{}{
			"test":       "check_role_failure",
			"authorized": resp.Authorized,
		})

		userRepo.AssertExpectations(t)
	})
}
