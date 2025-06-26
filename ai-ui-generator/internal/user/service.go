package user

import (
	"fmt"
	"strings"
	"time"
)

const (
	// Pagination limits
	DefaultLimit = 10
	MaxLimit     = 100
	
	// Common roles
	RoleAdmin = "admin"
	RoleUser  = "user"
	RoleModerator = "moderator"
)

// User represents a user in the system
type User struct {
	ID            string     `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	Name          string     `json:"name" db:"name"`
	AvatarURL     string     `json:"avatar_url" db:"avatar_url"`
	Roles         []string   `json:"roles" db:"roles"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	EmailVerified bool       `json:"email_verified" db:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	PasswordHash  string     `json:"-" db:"password_hash"`
}

// Project represents a project in the system
type Project struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	UserID      string                 `json:"user_id" db:"user_id"`
	Status      string                 `json:"status" db:"status"`
	Tags        []string               `json:"tags" db:"tags"`
	Config      map[string]interface{} `json:"config" db:"config"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	IsPublic    bool                   `json:"is_public" db:"is_public"`
	TemplateID  *string                `json:"template_id,omitempty" db:"template_id"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Service provides user and project business logic
type Service struct {
	repo        Repository
	projectRepo ProjectRepository
}

// NewService creates a new user service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// NewServiceWithProjects creates a new user service with project repository
func NewServiceWithProjects(repo Repository, projectRepo ProjectRepository) *Service {
	return &Service{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	
	return user, nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(id string, updates map[string]interface{}) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	
	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}
	
	// Validate updates
	allowedFields := map[string]bool{
		"name":           true,
		"avatar_url":     true,
		"roles":          true,
		"is_active":      true,
		"email_verified": true,
		"last_login_at":  true,
	}
	
	// Create a copy of updates with only allowed fields
	validUpdates := make(map[string]interface{})
	for key, value := range updates {
		if !allowedFields[key] {
			return nil, fmt.Errorf("field '%s' cannot be updated", key)
		}
		validUpdates[key] = value
	}
	
	// Add timestamp
	validUpdates["updated_at"] = time.Now()
	
	// Check if user exists before updating
	existingUser, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, fmt.Errorf("user not found")
	}
	
	updatedUser, err := s.repo.Update(id, validUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	return updatedUser, nil
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(id string) error {
	if id == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	
	// Check if user exists before deleting
	existingUser, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return fmt.Errorf("user not found")
	}
	
	// Check if user has associated projects
	if s.projectRepo != nil {
		projects, err := s.projectRepo.ListByUserID(id, 1, 0)
		if err != nil {
			return fmt.Errorf("failed to check user projects: %w", err)
		}
		if len(projects) > 0 {
			return fmt.Errorf("cannot delete user with existing projects")
		}
	}
	
	err = s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	return nil
}

// ListUsers lists users with pagination
func (s *Service) ListUsers(limit, offset int) ([]*User, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	
	users, err := s.repo.List(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	
	return users, nil
}

// Project Service Methods

// GetProject retrieves a project by ID
func (s *Service) GetProject(id string) (*Project, error) {
	if s.projectRepo == nil {
		return nil, fmt.Errorf("project repository not initialized")
	}
	return s.projectRepo.GetByID(id)
}

// CreateProject creates a new project
func (s *Service) CreateProject(project *Project) error {
	if s.projectRepo == nil {
		return fmt.Errorf("project repository not initialized")
	}
	return s.projectRepo.Create(project)
}

// UpdateProject updates project information
func (s *Service) UpdateProject(id string, updates map[string]interface{}) (*Project, error) {
	if s.projectRepo == nil {
		return nil, fmt.Errorf("project repository not initialized")
	}
	return s.projectRepo.Update(id, updates)
}

// DeleteProject deletes a project
func (s *Service) DeleteProject(id string) error {
	if s.projectRepo == nil {
		return fmt.Errorf("project repository not initialized")
	}
	return s.projectRepo.Delete(id)
}

// ListProjects lists projects with pagination
func (s *Service) ListProjects(limit, offset int) ([]*Project, error) {
	if s.projectRepo == nil {
		return nil, fmt.Errorf("project repository not initialized")
	}
	return s.projectRepo.List(limit, offset)
}

// ListUserProjects lists projects for a specific user
func (s *Service) ListUserProjects(userID string, limit, offset int) ([]*Project, error) {
	if s.projectRepo == nil {
		return nil, fmt.Errorf("project repository not initialized")
	}
	return s.projectRepo.ListByUserID(userID, limit, offset)
}

// Additional Service Methods

// GetUserByEmail retrieves a user by email address
func (s *Service) GetUserByEmail(email string) (*User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	
	// Basic email validation
	if !strings.Contains(email, "@") {
		return nil, fmt.Errorf("invalid email format")
	}
	
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return user, nil
}

// ActivateUser activates a user account
func (s *Service) ActivateUser(id string) error {
	updates := map[string]interface{}{
		"is_active": true,
	}
	
	_, err := s.UpdateUser(id, updates)
	return err
}

// DeactivateUser deactivates a user account
func (s *Service) DeactivateUser(id string) error {
	updates := map[string]interface{}{
		"is_active": false,
	}
	
	_, err := s.UpdateUser(id, updates)
	return err
}

// VerifyUserEmail marks a user's email as verified
func (s *Service) VerifyUserEmail(id string) error {
	updates := map[string]interface{}{
		"email_verified": true,
	}
	
	_, err := s.UpdateUser(id, updates)
	return err
}

// UpdateLastLogin updates the user's last login timestamp
func (s *Service) UpdateLastLogin(id string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_login_at": &now,
	}
	
	_, err := s.UpdateUser(id, updates)
	return err
}

// IsUserActive checks if a user account is active
func (s *Service) IsUserActive(id string) (bool, error) {
	user, err := s.GetUser(id)
	if err != nil {
		return false, err
	}
	
	return user.IsActive, nil
}

// GetUserRoles retrieves the roles assigned to a user
func (s *Service) GetUserRoles(id string) ([]string, error) {
	user, err := s.GetUser(id)
	if err != nil {
		return nil, err
	}
	
	return user.Roles, nil
}

// HasRole checks if a user has a specific role
func (s *Service) HasRole(id string, role string) (bool, error) {
	roles, err := s.GetUserRoles(id)
	if err != nil {
		return false, err
	}
	
	for _, userRole := range roles {
		if userRole == role {
			return true, nil
		}
	}
	
	return false, nil
}

// AddRole adds a role to a user
func (s *Service) AddRole(id string, role string) error {
	if role == "" {
		return fmt.Errorf("role cannot be empty")
	}
	
	user, err := s.GetUser(id)
	if err != nil {
		return err
	}
	
	// Check if user already has the role
	for _, existingRole := range user.Roles {
		if existingRole == role {
			return nil // Role already exists, no error
		}
	}
	
	// Add the new role
	newRoles := append(user.Roles, role)
	updates := map[string]interface{}{
		"roles": newRoles,
	}
	
	_, err = s.UpdateUser(id, updates)
	return err
}

// RemoveRole removes a role from a user
func (s *Service) RemoveRole(id string, role string) error {
	if role == "" {
		return fmt.Errorf("role cannot be empty")
	}
	
	user, err := s.GetUser(id)
	if err != nil {
		return err
	}
	
	// Filter out the role to remove
	var newRoles []string
	roleFound := false
	for _, existingRole := range user.Roles {
		if existingRole != role {
			newRoles = append(newRoles, existingRole)
		} else {
			roleFound = true
		}
	}
	
	if !roleFound {
		return nil // Role didn't exist, no error
	}
	
	updates := map[string]interface{}{
		"roles": newRoles,
	}
	
	_, err = s.UpdateUser(id, updates)
	return err
}

// CreateUser creates a new user in the system
func (s *Service) CreateUser(user *User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	
	// Validate required fields
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	if user.Name == "" {
		return fmt.Errorf("name is required")
	}
	
	// Set defaults
	if user.ID == "" {
		user.ID = fmt.Sprintf("user_%d", time.Now().UnixNano())
	}
	if user.Roles == nil || len(user.Roles) == 0 {
		user.Roles = []string{RoleUser}
	}
	user.IsActive = true
	user.EmailVerified = false
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	// Create user in repository
	err := s.repo.Create(user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}
