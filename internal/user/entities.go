// Package user contains consolidated user business logic and entities
package user

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// User represents a user entity
type User struct {
	utilities.BaseEntity
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

// Validate validates the user entity
func (u User) Validate() error {
	if u.Email == "" {
		return utilities.NewValidationError("email is required", nil)
	}
	if !isValidEmail(u.Email) {
		return utilities.NewValidationError("invalid email format", nil)
	}
	if u.Username == "" {
		return utilities.NewValidationError("username is required", nil)
	}
	if !isValidUsername(u.Username) {
		return utilities.NewValidationError("invalid username format", nil)
	}
	if u.Role != RoleUser && u.Role != RoleAdmin {
		return utilities.NewValidationError("invalid role", nil)
	}
	return nil
}

// IsValid returns true if the user passes validation
func (u User) IsValid() bool {
	return u.Validate() == nil
}

// GetValidationRules returns validation rules for user
func (u User) GetValidationRules() []utilities.ValidationRule {
	return []utilities.ValidationRule{
		{Field: "email", Rule: "required", Message: "Email is required"},
		{Field: "email", Rule: "email", Message: "Invalid email format"},
		{Field: "username", Rule: "min_length", Message: "Username must be at least 3 characters"},
	}
}

// ToMap converts user to map
func (u User) ToMap() map[string]interface{} {
	// Start with BaseEntity map
	baseMap := u.BaseEntity.ToMap()
	// Add User specific fields
	baseMap["id"] = string(u.ID)
	baseMap["email"] = u.Email
	baseMap["username"] = u.Username
	baseMap["name"] = u.Name
	baseMap["avatar_url"] = u.AvatarURL
	baseMap["roles"] = u.Roles
	baseMap["role"] = string(u.Role)
	baseMap["active"] = u.Active
	baseMap["status"] = string(u.Status)
	baseMap["created_at"] = u.CreatedAt
	baseMap["updated_at"] = u.UpdatedAt
	return baseMap
}

// ToJSON serializes user to JSON
func (u User) ToJSON() ([]byte, error) {
	data := u.ToMap()
	return json.Marshal(data)
}

// FromJSON deserializes user from JSON
func (u *User) FromJSON(data []byte) error {
	var mapData map[string]interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		return err
	}
	if email, ok := mapData["email"].(string); ok {
		u.Email = email
	}
	if username, ok := mapData["username"].(string); ok {
		u.Username = username
	}
	if name, ok := mapData["name"].(string); ok {
		u.Name = name
	}
	return nil
}

// User interface implementation methods
func (u User) GetEmail() string {
	return u.Email
}

func (u User) GetUsername() string {
	return u.Username
}

func (u *User) SetPassword(password string) error {
	// Password handling is centralized in Auth Service
	return utilities.NewValidationError("password handling centralized in auth service", nil)
}

func (u User) ValidatePassword(password string) bool {
	// Password validation is centralized in Auth Service
	return false
}

func (u User) GetRoles() []string {
	return u.Roles
}

func (u User) HasPermission(permission string) bool {
	// Basic permission check based on role
	if u.IsAdmin() {
		return true
	}
	// Add more specific permission checks here
	return false
}

// NewUser creates a new user with entity properties
func NewUser(id utilities.UserID, email, username, name string) *User {
	now := time.Now()
	return &User{
		BaseEntity: utilities.NewBaseEntity(string(id), utilities.EntityTypeUser),
		ID:         id,
		Email:      email,
		Username:   username,
		Name:       name,
		Roles:      []string{},
		Role:       RoleUser,
		Active:     true,
		Status:     StatusActiveUser,
		Timestamps: utilities.Timestamps{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
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
	utilities.BaseEntity
	ID          utilities.ProjectID `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string              `json:"name" gorm:"not null"`
	Description string              `json:"description"`
	UserID      utilities.UserID    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Status      ProjectStatus       `json:"status" gorm:"type:varchar(20);default:'active'"`
	utilities.Timestamps
}

// Validate validates the project entity
func (p Project) Validate() error {
	if p.Name == "" {
		return utilities.NewValidationError("name is required", nil)
	}
	if len(p.Name) < 3 {
		return utilities.NewValidationError("name must be at least 3 characters", nil)
	}
	if p.UserID.IsEmpty() {
		return utilities.NewValidationError("user ID is required", nil)
	}
	return nil
}

// IsValid returns true if the project passes validation
func (p Project) IsValid() bool {
	return p.Validate() == nil
}

// GetValidationRules returns validation rules for project
func (p Project) GetValidationRules() []utilities.ValidationRule {
	return []utilities.ValidationRule{
		{Field: "name", Rule: "required", Message: "Name is required"},
		{Field: "name", Rule: "min_length", Message: "Name must be at least 3 characters"},
		{Field: "user_id", Rule: "required", Message: "User ID is required"},
	}
}

// ToMap converts project to map
func (p Project) ToMap() map[string]interface{} {
	// Start with BaseEntity map
	baseMap := p.BaseEntity.ToMap()
	// Add Project specific fields
	baseMap["id"] = string(p.ID)
	baseMap["name"] = p.Name
	baseMap["description"] = p.Description
	baseMap["user_id"] = string(p.UserID)
	baseMap["status"] = string(p.Status)
	baseMap["created_at"] = p.CreatedAt
	baseMap["updated_at"] = p.UpdatedAt
	return baseMap
}

// ToJSON serializes project to JSON
func (p Project) ToJSON() ([]byte, error) {
	data := p.ToMap()
	return json.Marshal(data)
}

// FromJSON deserializes project from JSON
func (p *Project) FromJSON(data []byte) error {
	var mapData map[string]interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		return err
	}
	if name, ok := mapData["name"].(string); ok {
		p.Name = name
	}
	if description, ok := mapData["description"].(string); ok {
		p.Description = description
	}
	return nil
}

// Project interface implementation methods
func (p Project) GetOwnerID() string {
	return string(p.UserID)
}

func (p Project) GetName() string {
	return p.Name
}

func (p Project) GetStatus() utilities.ProjectStatus {
	return utilities.ProjectStatus(p.Status)
}

func (p *Project) SetStatus(status utilities.ProjectStatus) error {
	p.Status = ProjectStatus(status)
	p.MarkDirty("status")
	return nil
}

func (p Project) GetGenerations() []utilities.Generation {
	// This would be populated from repository when needed
	return []utilities.Generation{}
}

func (p *Project) AddGeneration(generation utilities.Generation) error {
	// This would be handled by repository
	return nil
}

// NewProject creates a new project with entity properties
func NewProject(id utilities.ProjectID, name, description string, userID utilities.UserID) *Project {
	now := time.Now()
	return &Project{
		BaseEntity:  utilities.NewBaseEntity(string(id), utilities.EntityTypeProject),
		ID:          id,
		Name:        name,
		Description: description,
		UserID:      userID,
		Status:      StatusActive,
		Timestamps: utilities.Timestamps{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
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
