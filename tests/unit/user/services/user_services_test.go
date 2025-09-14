package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func rolePtr(r user.Role) *user.Role {
	return &r
}

func statusPtr(s user.UserStatus) *user.UserStatus {
	return &s
}

// Mock interfaces
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id utilities.UserID) (user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, user user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id utilities.UserID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context, params utilities.PaginationParams, search string) ([]user.User, error) {
	args := m.Called(ctx, params, search)
	return args.Get(0).([]user.User), args.Error(1)
}

func (m *MockRepository) Count(ctx context.Context, search string) (int, error) {
	args := m.Called(ctx, search)
	return args.Get(0).(int), args.Error(1)
}

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) ValidateStruct(s interface{}) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockValidator) ValidateUser(user *user.User) error {
	args := m.Called(user)
	return args.Error(0)
}

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) NotifyUserCreated(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyUserUpdated(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyUserDeleted(ctx context.Context, userID utilities.UserID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Test User Creator Service
func TestUserCreator_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     user.CreateUserRequest
		setupMocks  func(*MockRepository, *MockValidator, *MockNotificationService)
		expectError bool
		errorType   string
	}{
		{
			name: "successful user creation",
			request: user.CreateUserRequest{
				Email:     "test@example.com",
				Name:      "Test User",
				AvatarURL: "https://example.com/avatar.jpg",
				Roles:     []string{"user"},
			},
			setupMocks: func(repo *MockRepository, validator *MockValidator, notifier *MockNotificationService) {
				// Validation succeeds
				validator.On("ValidateStruct", mock.AnythingOfType("user.CreateUserRequest")).Return(nil)

				// No existing user
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return(user.User{}, utilities.NewNotFoundError("user not found"))

				// User entity validation succeeds
				validator.On("ValidateUser", mock.AnythingOfType("*user.User")).Return(nil)

				// Repository create succeeds
				repo.On("Create", mock.Anything, mock.AnythingOfType("user.User")).Return(nil)

				// Notification (called async)
				notifier.On("NotifyUserCreated", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "user creation failure - validation error",
			request: user.CreateUserRequest{
				Email: "invalid-email",
				Name:  "",
			},
			setupMocks: func(repo *MockRepository, validator *MockValidator, notifier *MockNotificationService) {
				validator.On("ValidateStruct", mock.AnythingOfType("user.CreateUserRequest")).Return(assert.AnError)
			},
			expectError: true,
			errorType:   "validation",
		},
		{
			name: "user creation failure - user already exists",
			request: user.CreateUserRequest{
				Email: "existing@example.com",
				Name:  "Test User",
			},
			setupMocks: func(repo *MockRepository, validator *MockValidator, notifier *MockNotificationService) {
				// Validation succeeds
				validator.On("ValidateStruct", mock.AnythingOfType("user.CreateUserRequest")).Return(nil)

				// User already exists
				existingUser := user.User{
					ID:    utilities.UserID("existing123"),
					Email: "existing@example.com",
					Name:  "Existing User",
				}
				repo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectError: true,
			errorType:   "conflict",
		},
		{
			name: "user creation failure - entity validation error",
			request: user.CreateUserRequest{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setupMocks: func(repo *MockRepository, validator *MockValidator, notifier *MockNotificationService) {
				// Request validation succeeds
				validator.On("ValidateStruct", mock.AnythingOfType("user.CreateUserRequest")).Return(nil)

				// No existing user
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return(user.User{}, utilities.NewNotFoundError("user not found"))

				// Entity validation fails
				validator.On("ValidateUser", mock.AnythingOfType("*user.User")).Return(assert.AnError)
			},
			expectError: true,
			errorType:   "validation",
		},
		{
			name: "user creation failure - repository error",
			request: user.CreateUserRequest{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setupMocks: func(repo *MockRepository, validator *MockValidator, notifier *MockNotificationService) {
				// Validation succeeds
				validator.On("ValidateStruct", mock.AnythingOfType("user.CreateUserRequest")).Return(nil)

				// No existing user
				repo.On("GetByEmail", mock.Anything, "test@example.com").Return(user.User{}, utilities.NewNotFoundError("user not found"))

				// Entity validation succeeds
				validator.On("ValidateUser", mock.AnythingOfType("*user.User")).Return(nil)

				// Repository create fails
				repo.On("Create", mock.Anything, mock.AnythingOfType("user.User")).Return(assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			repo := &MockRepository{}
			validator := &MockValidator{}
			notifier := &MockNotificationService{}

			tt.setupMocks(repo, validator, notifier)

			// Create service
			userCreator := user.NewUserCreator(repo, validator, notifier)

			// Execute
			response, err := userCreator.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)

				if tt.errorType != "" {
					switch tt.errorType {
					case "validation":
						assert.True(t, utilities.IsValidationError(err))
					case "conflict":
						assert.True(t, utilities.IsConflictError(err))
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, response.User)
				assert.Equal(t, tt.request.Email, response.User.Email)
				assert.Equal(t, tt.request.Name, response.User.Name)
				assert.True(t, response.User.Active)
			}

			// Verify mock expectations
			repo.AssertExpectations(t)
			validator.AssertExpectations(t)
			// Note: notifier is called async, so we can't easily verify it
		})
	}
}

// Test User Retriever Service
func TestUserRetriever_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     user.GetUserRequest
		setupMocks  func(*MockRepository)
		expectError bool
		expectedID  utilities.UserID
	}{
		{
			name: "successful user retrieval",
			request: user.GetUserRequest{
				UserID: utilities.UserID("user123"),
			},
			setupMocks: func(repo *MockRepository) {
				userEntity := user.User{
					ID:     utilities.UserID("user123"),
					Email:  "test@example.com",
					Name:   "Test User",
					Active: true,
					Status: user.StatusActiveUser,
				}

				repo.On("GetByID", mock.Anything, utilities.UserID("user123")).Return(userEntity, nil)
			},
			expectError: false,
			expectedID:  utilities.UserID("user123"),
		},
		{
			name: "user retrieval failure - user not found",
			request: user.GetUserRequest{
				UserID: utilities.UserID("nonexistent"),
			},
			setupMocks: func(repo *MockRepository) {
				repo.On("GetByID", mock.Anything, utilities.UserID("nonexistent")).Return(user.User{}, utilities.NewNotFoundError("user not found"))
			},
			expectError: true,
		},
		{
			name: "user retrieval failure - empty user ID",
			request: user.GetUserRequest{
				UserID: utilities.UserID(""),
			},
			setupMocks: func(repo *MockRepository) {
				// No mocks needed
			},
			expectError: true,
		},
		{
			name: "user retrieval failure - repository error",
			request: user.GetUserRequest{
				UserID: utilities.UserID("user456"),
			},
			setupMocks: func(repo *MockRepository) {
				repo.On("GetByID", mock.Anything, utilities.UserID("user456")).Return(user.User{}, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			repo := &MockRepository{}
			tt.setupMocks(repo)

			// Create service
			userRetriever := user.NewUserRetriever(repo)

			// Execute
			response, err := userRetriever.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, response.User)
				assert.Equal(t, tt.expectedID, response.User.ID)
			}

			// Verify mock expectations
			repo.AssertExpectations(t)
		})
	}
}

// Test User Updater Service
func TestUserUpdater_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     user.UpdateUserRequest
		setupMocks  func(*MockRepository, *MockValidator, *MockNotificationService)
		expectError bool
		errorType   string
	}{
		{
			name: "successful user update",
			request: user.UpdateUserRequest{
				UserID:    utilities.UserID("user123"),
				Name:      stringPtr("Updated User"),
				AvatarURL: stringPtr("https://example.com/new-avatar.jpg"),
			},
			setupMocks: func(repo *MockRepository, validator *MockValidator, notifier *MockNotificationService) {
				// Request validation succeeds
				validator.On("ValidateStruct", mock.AnythingOfType("user.UpdateUserRequest")).Return(nil)

				// Get existing user
				existingUser := user.User{
					ID:     utilities.UserID("user123"),
					Email:  "test@example.com",
					Name:   "Test User",
					Active: true,
					Status: user.StatusActiveUser,
				}

				repo.On("GetByID", mock.Anything, utilities.UserID("user123")).Return(existingUser, nil)

				// User entity validation succeeds
				validator.On("ValidateUser", mock.AnythingOfType("*user.User")).Return(nil)

				// Repository update succeeds
				repo.On("Update", mock.Anything, mock.AnythingOfType("user.User")).Return(nil)

				// Notification (called async)
				notifier.On("NotifyUserUpdated", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "user update failure - user not found",
			request: user.UpdateUserRequest{
				UserID: utilities.UserID("nonexistent"),
				Name:   stringPtr("Updated User"),
			},
			setupMocks: func(repo *MockRepository, validator *MockValidator, notifier *MockNotificationService) {
				// Request validation succeeds
				validator.On("ValidateStruct", mock.AnythingOfType("user.UpdateUserRequest")).Return(nil)

				repo.On("GetByID", mock.Anything, utilities.UserID("nonexistent")).Return(user.User{}, utilities.NewNotFoundError("user not found"))
			},
			expectError: true,
		},
		{
			name: "user update failure - validation error",
			request: user.UpdateUserRequest{
				UserID: utilities.UserID("user123"),
				Name:   stringPtr(""),
			},
			setupMocks: func(repo *MockRepository, validator *MockValidator, notifier *MockNotificationService) {
				// Request validation succeeds
				validator.On("ValidateStruct", mock.AnythingOfType("user.UpdateUserRequest")).Return(nil)

				// Get existing user
				existingUser := user.User{
					ID:     utilities.UserID("user123"),
					Email:  "test@example.com",
					Name:   "Test User",
					Active: true,
					Status: user.StatusActiveUser,
				}

				repo.On("GetByID", mock.Anything, utilities.UserID("user123")).Return(existingUser, nil)

				// User entity validation fails
				validator.On("ValidateUser", mock.AnythingOfType("*user.User")).Return(assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			repo := &MockRepository{}
			validator := &MockValidator{}
			notifier := &MockNotificationService{}

			tt.setupMocks(repo, validator, notifier)

			// Create service
			userUpdater := user.NewUserUpdater(repo, validator, notifier)

			// Execute
			response, err := userUpdater.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, response.User)

				if tt.request.Name != nil {
					assert.Equal(t, *tt.request.Name, response.User.Name)
				}
			}

			// Verify mock expectations
			repo.AssertExpectations(t)
			validator.AssertExpectations(t)
		})
	}
}

// Test User Lister Service
func TestUserLister_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     user.ListUsersRequest
		setupMocks  func(*MockRepository)
		expectError bool
		expectCount int
	}{
		{
			name: "successful user listing",
			request: user.ListUsersRequest{
				Page:   1,
				Limit:  10,
				Search: "",
			},
			setupMocks: func(repo *MockRepository) {
				users := []user.User{
					{
						ID:     utilities.UserID("user1"),
						Email:  "user1@example.com",
						Name:   "User One",
						Active: true,
					},
					{
						ID:     utilities.UserID("user2"),
						Email:  "user2@example.com",
						Name:   "User Two",
						Active: true,
					},
				}

				repo.On("List", mock.Anything, mock.AnythingOfType("utilities.PaginationParams"), mock.AnythingOfType("string")).Return(users, nil)
				repo.On("Count", mock.Anything, mock.AnythingOfType("string")).Return(2, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name: "successful user listing with search",
			request: user.ListUsersRequest{
				Page:   1,
				Limit:  10,
				Search: "john",
			},
			setupMocks: func(repo *MockRepository) {
				users := []user.User{
					{
						ID:     utilities.UserID("user1"),
						Email:  "john@example.com",
						Name:   "John Doe",
						Active: true,
					},
				}

				repo.On("List", mock.Anything, mock.AnythingOfType("utilities.PaginationParams"), mock.AnythingOfType("string")).Return(users, nil)
				repo.On("Count", mock.Anything, mock.AnythingOfType("string")).Return(1, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "user listing failure - repository error",
			request: user.ListUsersRequest{
				Page:  1,
				Limit: 10,
			},
			setupMocks: func(repo *MockRepository) {
				repo.On("List", mock.Anything, mock.AnythingOfType("utilities.PaginationParams"), mock.AnythingOfType("string")).Return([]user.User{}, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			repo := &MockRepository{}
			tt.setupMocks(repo)

			// Create service
			userLister := user.NewUserLister(repo)

			// Execute
			response, err := userLister.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.expectCount, len(response.Users))
				assert.Equal(t, tt.expectCount, response.Total)
			}

			// Verify mock expectations
			repo.AssertExpectations(t)
		})
	}
}

// Test User Deleter Service
func TestUserDeleter_Execute(t *testing.T) {
	tests := []struct {
		name        string
		request     user.DeleteUserRequest
		setupMocks  func(*MockRepository, *MockNotificationService)
		expectError bool
	}{
		{
			name: "successful user deletion",
			request: user.DeleteUserRequest{
				UserID: utilities.UserID("user123"),
			},
			setupMocks: func(repo *MockRepository, notifier *MockNotificationService) {
				// Get existing user
				existingUser := user.User{
					ID:     utilities.UserID("user123"),
					Email:  "test@example.com",
					Name:   "Test User",
					Active: true,
				}

				repo.On("GetByID", mock.Anything, utilities.UserID("user123")).Return(existingUser, nil)
				repo.On("Delete", mock.Anything, utilities.UserID("user123")).Return(nil)
				notifier.On("NotifyUserDeleted", mock.Anything, mock.AnythingOfType("utilities.UserID")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "user deletion failure - user not found",
			request: user.DeleteUserRequest{
				UserID: utilities.UserID("nonexistent"),
			},
			setupMocks: func(repo *MockRepository, notifier *MockNotificationService) {
				repo.On("GetByID", mock.Anything, utilities.UserID("nonexistent")).Return(user.User{}, utilities.NewNotFoundError("user not found"))
			},
			expectError: true,
		},
		{
			name: "user deletion failure - repository error",
			request: user.DeleteUserRequest{
				UserID: utilities.UserID("user123"),
			},
			setupMocks: func(repo *MockRepository, notifier *MockNotificationService) {
				existingUser := user.User{
					ID:     utilities.UserID("user123"),
					Email:  "test@example.com",
					Name:   "Test User",
					Active: true,
				}

				repo.On("GetByID", mock.Anything, utilities.UserID("user123")).Return(existingUser, nil)
				repo.On("Delete", mock.Anything, utilities.UserID("user123")).Return(assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			repo := &MockRepository{}
			notifier := &MockNotificationService{}

			tt.setupMocks(repo, notifier)

			// Create service
			userDeleter := user.NewUserDeleter(repo, notifier)

			// Execute
			response, err := userDeleter.Execute(context.Background(), tt.request)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.True(t, response.Success)
			}

			// Verify mock expectations
			repo.AssertExpectations(t)
			notifier.AssertExpectations(t)
		})
	}
}
