// Package user contains consolidated user business logic and entities
package user

import (
	"context"
	"strings"
	"unicode"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// User represents a user entity
type User struct {
	ID        utilities.UserID `json:"id" gorm:"type:varchar(36);primaryKey"`
	Email     string           `json:"email" gorm:"uniqueIndex;not null"`
	Username  string           `json:"username" gorm:"uniqueIndex"`
	Name      string           `json:"name"`
	AvatarURL string           `json:"avatar_url"`
	// PasswordHash field removed - password handling centralized in Auth Service
	Roles  []string   `json:"roles" gorm:"type:text[]"`
	Role   Role       `json:"role" gorm:"type:varchar(20);default:'user'"`
	Active bool       `json:"active" gorm:"default:true"`
	Status UserStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	utilities.Timestamps
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
func (u User) CanAccessProject(projectUserID utilities.UserID) bool {
	return u.IsAdmin() || u.ID == projectUserID
}

// Project represents a project entity
type Project struct {
	ID          utilities.ProjectID `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string              `json:"name" gorm:"not null"`
	Description string              `json:"description"`
	UserID      utilities.UserID    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Status      ProjectStatus       `json:"status" gorm:"type:varchar(20);default:'active'"`
	utilities.Timestamps
}

// ProjectStatus represents project status
type ProjectStatus string

const (
	StatusActive   ProjectStatus = "active"
	StatusInactive ProjectStatus = "inactive"
	StatusArchived ProjectStatus = "archived"
)

// Validation functions
func isValidEmail(email string) bool {
	// Basic email validation
	parts := strings.Split(email, "@")
	return len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0 && strings.Contains(parts[1], ".")
}

func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

// Repository interfaces (consolidated from domain)
type Repository interface {
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, id utilities.UserID) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id utilities.UserID) error
	List(ctx context.Context, params utilities.PaginationParams, search string) ([]User, error)
	Count(ctx context.Context, search string) (int, error)
}

// ProjectRepository defines project repository interface
type ProjectRepository interface {
	Create(ctx context.Context, project Project) error
	GetByID(ctx context.Context, id utilities.ProjectID) (Project, error)
	Update(ctx context.Context, project Project) error
	Delete(ctx context.Context, id utilities.ProjectID) error
	List(ctx context.Context, params utilities.PaginationParams, search string, status ProjectStatus) ([]Project, error)
	ListByUserID(ctx context.Context, userID utilities.UserID, params utilities.PaginationParams) ([]Project, error)
}

// EventPublisher defines event publishing interface
type EventPublisher interface {
	PublishUserCreated(ctx context.Context, user User) error
	PublishUserUpdated(ctx context.Context, user User) error
	PublishProjectCreated(ctx context.Context, project Project) error
}

// Validator defines validation interface
type Validator interface {
	ValidateStruct(s interface{}) error
	ValidateUser(user *User) error
}

// NotificationService defines notification interface
type NotificationService interface {
	NotifyUserCreated(ctx context.Context, user *User) error
	NotifyUserUpdated(ctx context.Context, user *User) error
	NotifyUserDeleted(ctx context.Context, userID utilities.UserID) error
}
