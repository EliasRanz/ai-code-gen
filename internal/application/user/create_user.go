package user

import (
	"context"
	"fmt"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
)

// CreateUserUseCase handles user creation business logic
type CreateUserUseCase struct {
	userRepo  user.Repository
	validator user.Validator
	notifier  user.NotificationService
}

// NewCreateUserUseCase creates a new instance of CreateUserUseCase
func NewCreateUserUseCase(
	userRepo user.Repository,
	validator user.Validator,
	notifier user.NotificationService,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:  userRepo,
		validator: validator,
		notifier:  notifier,
	}
}

// CreateUserRequest represents the input for creating a user
type CreateUserRequest struct {
	Email     string   `json:"email" validate:"required,email"`
	Name      string   `json:"name" validate:"required,min=2,max=100"`
	AvatarURL string   `json:"avatar_url" validate:"omitempty,url"`
	Roles     []string `json:"roles" validate:"dive,oneof=admin user viewer"`
}

// CreateUserResponse represents the output of user creation
type CreateUserResponse struct {
	User *user.User `json:"user"`
}

// Execute performs the user creation use case
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	// Validate input
	if err := uc.validator.ValidateStruct(req); err != nil {
		return nil, common.NewValidationError("invalid user data", err)
	}

	// Check if user already exists
	existing, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !common.IsNotFoundError(err) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if !existing.ID.IsEmpty() {
		return nil, common.NewConflictError("user with this email already exists")
	}

	// Create user entity
	newUser := user.User{
		Email:     req.Email,
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
		Roles:     req.Roles,
		Status:    user.StatusActiveUser,
		Active:    true,
	}

	// Validate user entity
	if err := uc.validator.ValidateUser(&newUser); err != nil {
		return nil, common.NewValidationError("invalid user entity", err)
	}

	// Save user
	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Send notification (async, don't block on failure)
	go func() {
		if err := uc.notifier.NotifyUserCreated(context.Background(), &newUser); err != nil {
			// Log error but don't fail the use case
			// Logger would be injected in production
		}
	}()

	return &CreateUserResponse{
		User: &newUser,
	}, nil
}
