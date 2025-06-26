package user

// MockRepository implements Repository for testing/stub purposes
type MockRepository struct {
	users map[string]*User
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		users: make(map[string]*User),
	}
}

// GetByID retrieves a user by ID
func (r *MockRepository) GetByID(id string) (*User, error) {
	if user, exists := r.users[id]; exists {
		return user, nil
	}
	return nil, nil // User not found
}

// GetByEmail retrieves a user by email
func (r *MockRepository) GetByEmail(email string) (*User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil // User not found
}

// Create creates a new user
func (r *MockRepository) Create(user *User) error {
	r.users[user.ID] = user
	return nil
}

// Update updates a user
func (r *MockRepository) Update(id string, updates map[string]interface{}) (*User, error) {
	if user, exists := r.users[id]; exists {
		// Apply updates (simplified for mock)
		if name, ok := updates["name"].(string); ok {
			user.Name = name
		}
		if email, ok := updates["email"].(string); ok {
			user.Email = email
		}
		if avatarURL, ok := updates["avatar_url"].(string); ok {
			user.AvatarURL = avatarURL
		}
		r.users[id] = user
		return user, nil
	}
	return nil, nil // User not found
}

// Delete deletes a user
func (r *MockRepository) Delete(id string) error {
	delete(r.users, id)
	return nil
}

// List lists users with pagination
func (r *MockRepository) List(limit, offset int) ([]*User, error) {
	var users []*User
	count := 0
	for _, user := range r.users {
		if count >= offset && len(users) < limit {
			users = append(users, user)
		}
		count++
	}
	return users, nil
}
