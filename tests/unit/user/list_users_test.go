// Package user contains tests for user application use cases
package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
	userapp "github.com/EliasRanz/ai-code-gen/internal/application/user"
)

// MockUserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id common.UserID) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id common.UserID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, params common.PaginationParams, search string) ([]user.User, error) {
	args := m.Called(ctx, params, search)
	return args.Get(0).([]user.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, search string) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
}

func TestListUsersUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	testUsers := []user.User{
		{
			ID:       common.UserID("user1"),
			Email:    "user1@example.com",
			Username: "user1",
			Name:     "user.User One",
			Active:   true,
			Role:     user.RoleUser,
			Status:   user.StatusActiveUser,
		},
		{
			ID:       common.UserID("user2"),
			Email:    "user2@example.com",
			Username: "user2",
			Name:     "user.User Two",
			Active:   true,
			Role:     user.RoleUser,
			Status:   user.StatusActiveUser,
		},
	}

	t.Run("successful listing with correct total count", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		useCase := userapp.NewListUsersUseCase(mockRepo)

		request := userapp.ListUsersRequest{
			Page:   1,
			Limit:  10,
			Search: "",
		}

		params := common.PaginationParams{
			Page:  1,
			Limit: 10,
		}

		// Setup expectations
		mockRepo.On("List", ctx, params, "").Return(testUsers, nil)
		mockRepo.On("Count", ctx, "").Return(25, nil) // Total 25 users in DB

		// Execute
		response, err := useCase.Execute(ctx, request)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, testUsers, response.Users)
		assert.Equal(t, 25, response.TotalCount) // Should be total count, not len(users)
		assert.Equal(t, int32(1), response.Page)
		assert.Equal(t, int32(10), response.Limit)

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful listing with search and correct count", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		useCase := userapp.NewListUsersUseCase(mockRepo)

		request := userapp.ListUsersRequest{
			Page:   1,
			Limit:  10,
			Search: "john",
		}

		params := common.PaginationParams{
			Page:  1,
			Limit: 10,
		}

		filteredUsers := []user.User{testUsers[0]} // Only one user matches

		// Setup expectations
		mockRepo.On("List", ctx, params, "john").Return(filteredUsers, nil)
		mockRepo.On("Count", ctx, "john").Return(3, nil) // 3 users match search

		// Execute
		response, err := useCase.Execute(ctx, request)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, filteredUsers, response.Users)
		assert.Equal(t, 3, response.TotalCount) // Should be total matching count
		assert.Equal(t, int32(1), response.Page)
		assert.Equal(t, int32(10), response.Limit)

		mockRepo.AssertExpectations(t)
	})

	t.Run("list repository error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		useCase := userapp.NewListUsersUseCase(mockRepo)

		request := userapp.ListUsersRequest{
			Page:  1,
			Limit: 10,
		}

		params := common.PaginationParams{
			Page:  1,
			Limit: 10,
		}

		mockRepo.On("List", ctx, params, "").Return([]user.User{}, assert.AnError)

		response, err := useCase.Execute(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to list users")

		mockRepo.AssertExpectations(t)
	})

	t.Run("count repository error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		useCase := userapp.NewListUsersUseCase(mockRepo)

		request := userapp.ListUsersRequest{
			Page:  1,
			Limit: 10,
		}

		params := common.PaginationParams{
			Page:  1,
			Limit: 10,
		}

		mockRepo.On("List", ctx, params, "").Return(testUsers, nil)
		mockRepo.On("Count", ctx, "").Return(0, assert.AnError)

		response, err := useCase.Execute(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to count users")

		mockRepo.AssertExpectations(t)
	})

	t.Run("default pagination values", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		useCase := userapp.NewListUsersUseCase(mockRepo)

		request := userapp.ListUsersRequest{
			Page:  0, // Should default to 1
			Limit: 0, // Should default to 20
		}

		params := common.PaginationParams{
			Page:  1,
			Limit: 20,
		}

		mockRepo.On("List", ctx, params, "").Return(testUsers, nil)
		mockRepo.On("Count", ctx, "").Return(25, nil)

		response, err := useCase.Execute(ctx, request)

		require.NoError(t, err)
		assert.Equal(t, int32(1), response.Page)
		assert.Equal(t, int32(20), response.Limit)

		mockRepo.AssertExpectations(t)
	})

	t.Run("limit capping", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		useCase := userapp.NewListUsersUseCase(mockRepo)

		request := userapp.ListUsersRequest{
			Page:  1,
			Limit: 150, // Should be capped to 100
		}

		params := common.PaginationParams{
			Page:  1,
			Limit: 100,
		}

		mockRepo.On("List", ctx, params, "").Return(testUsers, nil)
		mockRepo.On("Count", ctx, "").Return(25, nil)

		response, err := useCase.Execute(ctx, request)

		require.NoError(t, err)
		assert.Equal(t, int32(100), response.Limit)

		mockRepo.AssertExpectations(t)
	})
}
