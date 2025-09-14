// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/EliasRanz/ai-code-gen/tests/mocks"
)

func TestCheckRoleIntegration(t *testing.T) {
	t.Run("end-to-end role check success", func(t *testing.T) {
		// Setup gomock controller
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock user repository
		userRepo := mocks.NewMockAuthUserRepository(ctrl)

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

		// Setup mock expectations
		userRepo.EXPECT().GetByID(gomock.Any(), auth.UserID("test-user-id")).Return(userData, nil)

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

		// Log completion
		logger := observability.NewLogger("debug", "console")
		logger.Info("Integration test completed", map[string]interface{}{
			"test":       "check_role_success",
			"authorized": resp.Authorized,
		})
	})

	t.Run("end-to-end role check failure", func(t *testing.T) {
		// Setup gomock controller
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock user repository
		userRepo := mocks.NewMockAuthUserRepository(ctrl)

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

		// Setup mock expectations
		userRepo.EXPECT().GetByID(gomock.Any(), auth.UserID("test-user-id")).Return(userData, nil)

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

		// Log completion
		logger := observability.NewLogger("debug", "console")
		logger.Info("Integration test completed", map[string]interface{}{
			"test":       "check_role_failure",
			"authorized": resp.Authorized,
		})
	})
}
