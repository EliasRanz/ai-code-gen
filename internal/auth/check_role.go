package auth

import (
	"context"
	"fmt"
)

// CheckRoleService handles role and permission validation
type CheckRoleService struct {
	userRepo UserRepository
}

// NewCheckRole creates a new instance of CheckRoleService
func NewCheckRole(userRepo UserRepository) *CheckRoleService {
	return &CheckRoleService{
		userRepo: userRepo,
	}
}

// CheckRoleRequest represents the input for role validation
type CheckRoleRequest struct {
	UserID       UserID `json:"user_id" validate:"required"`
	RequiredRole string `json:"required_role" validate:"required"`
	Resource     string `json:"resource,omitempty"`
	Action       string `json:"action,omitempty"`
}

// CheckRoleResponse represents the output of role validation
type CheckRoleResponse struct {
	Authorized  bool     `json:"authorized"`
	UserRoles   []string `json:"user_roles"`
	Permissions []string `json:"permissions,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}

// Execute performs the role validation use case
func (s *CheckRoleService) Execute(ctx context.Context, req CheckRoleRequest) (*CheckRoleResponse, error) {
	// Validate request input
	if err := s.validateRequest(req); err != nil {
		return s.createUnauthorizedResponse(err.Error(), nil), nil
	}

	// Get user data for role validation
	userData, err := s.getUserData(ctx, req.UserID)
	if err != nil {
		if isUserValidationError(err) {
			// For inactive users, userData might still be available for roles
			userRoles := []string{}
			if userData != nil {
				userRoles = userData.Roles
			}
			return s.createUnauthorizedResponse(err.Error(), userRoles), nil
		}
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}

	// Check if user has required role
	// Empty required role means no role requirement - authorize any active user
	if req.RequiredRole == "" {
		return s.createAuthorizedResponse(userData.Roles), nil
	}

	authorized := s.checkUserRole(userData.Roles, req.RequiredRole)

	if authorized {
		return s.createAuthorizedResponse(userData.Roles), nil
	}

	return s.createUnauthorizedResponse(fmt.Sprintf("user does not have required role: %s", req.RequiredRole), userData.Roles), nil
}

// validateRequest validates the incoming request
func (s *CheckRoleService) validateRequest(req CheckRoleRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	// Empty required role is allowed - means no role requirement
	return nil
}

// getUserData retrieves and validates user data
func (s *CheckRoleService) getUserData(ctx context.Context, userID UserID) (*User, error) {
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, fmt.Errorf("user not found")
		}
		// For other database errors, return a generic message
		return nil, fmt.Errorf("failed to retrieve user information")
	}

	if !userData.Active {
		// Return user data even if inactive, but with an error to indicate inactive status
		return &userData, fmt.Errorf("user account is inactive")
	}

	return &userData, nil
}

// checkUserRole checks if user has the required role
func (s *CheckRoleService) checkUserRole(userRoles []string, requiredRole string) bool {
	// Admin role has access to everything
	if s.hasAdminRole(userRoles) {
		return true
	}

	// Check for exact role match
	for _, role := range userRoles {
		if role == requiredRole {
			return true
		}
	}

	return false
}

// hasAdminRole checks if user has admin role
func (s *CheckRoleService) hasAdminRole(userRoles []string) bool {
	for _, role := range userRoles {
		if role == "admin" || role == "super_admin" {
			return true
		}
	}
	return false
}

// createAuthorizedResponse creates a successful authorization response
func (s *CheckRoleService) createAuthorizedResponse(userRoles []string) *CheckRoleResponse {
	return &CheckRoleResponse{
		Authorized: true,
		UserRoles:  userRoles,
		// TODO: Add permissions based on roles (future enhancement)
		Permissions: []string{}, // Will be implemented later
	}
}

// createUnauthorizedResponse creates an unauthorized response
func (s *CheckRoleService) createUnauthorizedResponse(reason string, userRoles []string) *CheckRoleResponse {
	return &CheckRoleResponse{
		Authorized: false,
		UserRoles:  userRoles,
		Reason:     reason,
	}
}

// isUserValidationError checks if the error is a user validation error
func isUserValidationError(err error) bool {
	errMsg := err.Error()
	return errMsg == "user not found" || errMsg == "user account is inactive" || errMsg == "failed to retrieve user information"
}
