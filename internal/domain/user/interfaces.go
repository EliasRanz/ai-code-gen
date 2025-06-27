// Package user contains user domain interfaces
package user

import (
	"context"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
)

// Repository defines user domain data access
type Repository interface {
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, id common.UserID) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id common.UserID) error
	List(ctx context.Context, params common.PaginationParams, search string) ([]User, error)
	Count(ctx context.Context, search string) (int, error)
}

// ProjectRepository defines project domain data access
type ProjectRepository interface {
	Create(ctx context.Context, project Project) error
	GetByID(ctx context.Context, id common.ProjectID) (Project, error)
	Update(ctx context.Context, project Project) error
	Delete(ctx context.Context, id common.ProjectID) error
	List(ctx context.Context, params common.PaginationParams, search string, status ProjectStatus) ([]Project, error)
	ListByUserID(ctx context.Context, userID common.UserID, params common.PaginationParams) ([]Project, error)
}

// PasswordHasher defines password hashing interface
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
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
	NotifyUserDeleted(ctx context.Context, userID common.UserID) error
}
