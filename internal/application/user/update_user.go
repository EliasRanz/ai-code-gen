package user

import (
	"context"
	"fmt"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
)

// UpdateUserUseCase handles user update business logic
type UpdateUserUseCase struct {
	userRepo  user.Repository
	validator user.Validator
	notifier  user.NotificationService
}

// NewUpdateUserUseCase creates a new instance of UpdateUserUseCase
func NewUpdateUserUseCase(
	userRepo user.Repository,
	validator user.Validator,
	notifier user.NotificationService,
) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		userRepo:  userRepo,
		validator: validator,
		notifier:  notifier,
	}
}

// UpdateUserRequest represents the input for updating a user
type UpdateUserRequest struct {
	UserID    common.UserID `json:"user_id" validate:"required"`
	Email     *string       `json:"email,omitempty" validate:"omitempty,email"`
	Name      *string       `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	AvatarURL *string       `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Roles     *[]string     `json:"roles,omitempty" validate:"omitempty,dive,oneof=admin user viewer"`
	Active    *bool         `json:"active,omitempty"`
}

// UpdateUserResponse represents the output of user update
type UpdateUserResponse struct {
	User *user.User `json:"user"`
}

// Execute performs the user update use case
func (uc *UpdateUserUseCase) Execute(ctx context.Context, req UpdateUserRequest) (*UpdateUserResponse, error) {
	// Validate input
	if err := uc.validator.ValidateStruct(req); err != nil {
		return nil, common.NewValidationError("invalid update data", err)
	}

	// Get existing user
	existingUser, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil, common.NewNotFoundError("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Apply updates
	if req.Email != nil {
		existingUser.Email = *req.Email
	}
	if req.Name != nil {
		existingUser.Name = *req.Name
	}
	if req.AvatarURL != nil {
		existingUser.AvatarURL = *req.AvatarURL
	}
	if req.Roles != nil {
		existingUser.Roles = *req.Roles
	}
	if req.Active != nil {
		existingUser.Active = *req.Active
		if *req.Active {
			existingUser.Status = user.StatusActiveUser
		} else {
			existingUser.Status = user.StatusInactiveUser
		}
	}

	// Validate updated user entity
	if err := uc.validator.ValidateUser(&existingUser); err != nil {
		return nil, common.NewValidationError("invalid user entity", err)
	}

	// Update timestamps
	existingUser.Touch()

	// Save user
	if err := uc.userRepo.Update(ctx, existingUser); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Send notification (async, don't block on failure)
	go func() {
		if err := uc.notifier.NotifyUserUpdated(context.Background(), &existingUser); err != nil {
			// Log error but don't fail the use case
			// Logger would be injected in production
		}
	}()

	return &UpdateUserResponse{
		User: &existingUser,
	}, nil
}
