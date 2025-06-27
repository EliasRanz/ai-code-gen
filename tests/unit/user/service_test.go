package user

import (
	"fmt"
	"testing"
	
	"github.com/EliasRanz/ai-code-gen/internal/user"
)

// MockRepository implements the user.Repository interface for testing
type MockRepository struct {
	users map[string]*user.User
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users: make(map[string]*user.User),
	}
}

func (m *MockRepository) GetByID(id string) (*user.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *MockRepository) GetByEmail(email string) (*user.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) Create(user *user.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockRepository) Update(id string, updates map[string]interface{}) (*user.User, error) {
	if user, exists := m.users[id]; exists {
		// Apply updates to user (simplified for testing)
		if name, ok := updates["name"]; ok {
			user.Name = name.(string)
		}
		if email, ok := updates["email"]; ok {
			user.Email = email.(string)
		}
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockRepository) Delete(id string) error {
	if _, exists := m.users[id]; exists {
		delete(m.users, id)
		return nil
	}
	return fmt.Errorf("user not found")
}

func (m *MockRepository) List(limit, offset int) ([]*user.User, error) {
	users := make([]*user.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func TestService_GetUser(t *testing.T) {
	repo := NewMockRepository()
	service := user.NewService(repo)
	
	// Setup test data
	testUser := &user.User{
		ID:    "user1",
		Email: "test@example.com",
		Name:  "Test User",
	}
	repo.users["user1"] = testUser
	
	tests := []struct {
		name        string
		userID      string
		wantError   bool
		errorString string
	}{
		{
			name:   "successful get user",
			userID: "user1",
		},
		{
			name:        "empty user ID",
			userID:      "",
			wantError:   true,
			errorString: "user ID cannot be empty",
		},
		{
			name:        "user not found",
			userID:      "nonexistent",
			wantError:   true,
			errorString: "user not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.GetUser(tt.userID)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorString)
					return
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if user.ID != testUser.ID {
				t.Errorf("Expected user ID %s, got %s", testUser.ID, user.ID)
			}
		})
	}
}

func TestService_UpdateUser(t *testing.T) {
	repo := NewMockRepository()
	service := user.NewService(repo)
	
	// Setup test data
	testUser := &user.User{
		ID:    "user1",
		Email: "test@example.com",
		Name:  "Test user.User",
		Roles: []string{"user"},
	}
	repo.users["user1"] = testUser
	
	tests := []struct {
		name        string
		userID      string
		updates     map[string]interface{}
		wantError   bool
		errorString string
	}{
		{
			name:   "successful update",
			userID: "user1",
			updates: map[string]interface{}{
				"name": "Updated Name",
			},
		},
		{
			name:        "empty user ID",
			userID:      "",
			updates:     map[string]interface{}{"name": "test"},
			wantError:   true,
			errorString: "user ID cannot be empty",
		},
		{
			name:        "no updates provided",
			userID:      "user1",
			updates:     map[string]interface{}{},
			wantError:   true,
			errorString: "no updates provided",
		},
		{
			name:   "invalid field",
			userID: "user1",
			updates: map[string]interface{}{
				"email": "new@example.com", // email updates not allowed
			},
			wantError:   true,
			errorString: "field 'email' cannot be updated",
		},
		{
			name:        "user not found",
			userID:      "nonexistent",
			updates:     map[string]interface{}{"name": "test"},
			wantError:   true,
			errorString: "user not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.UpdateUser(tt.userID, tt.updates)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorString)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestService_ListUsers(t *testing.T) {
	repo := NewMockRepository()
	service := user.NewService(repo)
	
	// Setup test data
	for i := 1; i <= 5; i++ {
		user := &user.User{
			ID:    fmt.Sprintf("user%d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
		}
		repo.users[user.ID] = user
	}
	
	tests := []struct {
		name   string
		limit  int
		offset int
	}{
		{
			name:   "valid pagination",
			limit:  2,
			offset: 0,
		},
		{
			name:   "zero limit uses default",
			limit:  0,
			offset: 0,
		},
		{
			name:   "negative offset uses zero",
			limit:  10,
			offset: -1,
		},
		{
			name:   "limit exceeds maximum",
			limit:  200,
			offset: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := service.ListUsers(tt.limit, tt.offset)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if users == nil {
				t.Error("Expected users slice, got nil")
			}
		})
	}
}

func TestService_RoleManagement(t *testing.T) {
	repo := NewMockRepository()
	service := user.NewService(repo)
	
	testUser := &user.User{
		ID:    "user1",
		Email: "test@example.com",
		Roles: []string{"user"},
	}
	repo.users["user1"] = testUser
	
	t.Run("HasRole", func(t *testing.T) {
		hasRole, err := service.HasRole("user1", "user")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !hasRole {
			t.Error("Expected user to have 'user' role")
		}
		
		hasAdmin, err := service.HasRole("user1", "admin")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if hasAdmin {
			t.Error("Expected user to not have 'admin' role")
		}
	})
	
	t.Run("ActivateUser", func(t *testing.T) {
		err := service.ActivateUser("user1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
	
	t.Run("DeactivateUser", func(t *testing.T) {
		err := service.DeactivateUser("user1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
	
	t.Run("VerifyUserEmail", func(t *testing.T) {
		err := service.VerifyUserEmail("user1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
	
	t.Run("UpdateLastLogin", func(t *testing.T) {
		err := service.UpdateLastLogin("user1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
