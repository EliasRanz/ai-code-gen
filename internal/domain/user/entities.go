// Package user contains user domain entities and business rules
package user

import (
	"strings"
	"unicode"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
)

// User represents a user entity
type User struct {
	ID           common.UserID
	Email        string
	Username     string
	Name         string
	AvatarURL    string
	PasswordHash string // Password hash - never expose this in JSON
	Roles        []string
	Role         Role
	Active       bool
	Status       UserStatus
	common.Timestamps
}

// SetPassword sets the password hash for the user
func (u *User) SetPassword(hasher PasswordHasher, password string) error {
	hash, err := hasher.Hash(password)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return nil
}

// VerifyPassword verifies the provided password against the stored hash
func (u *User) VerifyPassword(hasher PasswordHasher, password string) bool {
	return hasher.Verify(password, u.PasswordHash)
}

// UserStatus represents the status of a user
type UserStatus string

const (
	StatusActiveUser    UserStatus = "active"
	StatusInactiveUser  UserStatus = "inactive"
	StatusSuspendedUser UserStatus = "suspended"
)

// Role represents user roles
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// IsAdmin returns true if the user is an admin
func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanAccessProject returns true if user can access the project
func (u User) CanAccessProject(projectUserID common.UserID) bool {
	return u.IsAdmin() || u.ID == projectUserID
}

// Project represents a project entity
type Project struct {
	ID          common.ProjectID
	Name        string
	Description string
	UserID      common.UserID
	Status      ProjectStatus
	common.Timestamps
}

// ProjectStatus represents project status
type ProjectStatus string

const (
	StatusActive   ProjectStatus = "active"
	StatusInactive ProjectStatus = "inactive"
	StatusArchived ProjectStatus = "archived"
)

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email    string
	Username string
	Password string
	Role     Role
}

// Validate validates the create user request
func (r CreateUserRequest) Validate() error {
	if r.Email == "" || !isValidEmail(r.Email) {
		return common.ErrInvalidInput
	}
	if r.Username == "" || len(r.Username) < 3 {
		return common.ErrInvalidInput
	}
	if !isValidPassword(r.Password) {
		return common.ErrInvalidInput
	}
	if r.Role != RoleUser && r.Role != RoleAdmin {
		return common.ErrInvalidInput
	}
	return nil
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	ID       common.UserID
	Email    *string
	Username *string
	Role     *Role
	Active   *bool
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name        string
	Description string
	UserID      common.UserID
}

// Validate validates the create project request
func (r CreateProjectRequest) Validate() error {
	if r.Name == "" || len(r.Name) < 3 {
		return common.ErrInvalidInput
	}
	if r.UserID.IsEmpty() {
		return common.ErrInvalidInput
	}
	return nil
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// isValidPassword validates password strength
func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}
