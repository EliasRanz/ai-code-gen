package auth

import (
	"context"
	"fmt"
)

// ValidateTokenService handles token validation and user context resolution
type ValidateTokenService struct {
	tokenProvider TokenProvider
	userRepo      UserRepository
}

// NewValidateToken creates a new instance of ValidateTokenService
func NewValidateToken(
	tokenProvider TokenProvider,
	userRepo UserRepository,
) *ValidateTokenService {
	return &ValidateTokenService{
		tokenProvider: tokenProvider,
		userRepo:      userRepo,
	}
}

// ValidateTokenRequest represents the input for token validation
type ValidateTokenRequest struct {
	AccessToken string `json:"access_token" validate:"required"`
}

// ValidateTokenResponse represents the output of token validation
type ValidateTokenResponse struct {
	Valid       bool             `json:"valid"`
	UserContext *UserContextData `json:"user_context,omitempty"`
	Error       string           `json:"error,omitempty"`
}

// UserContextData represents user context information
type UserContextData struct {
	UserID      UserID   `json:"user_id"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	Active      bool     `json:"active"`
}

// Execute performs the token validation use case
func (s *ValidateTokenService) Execute(ctx context.Context, req ValidateTokenRequest) (*ValidateTokenResponse, error) {
	// Validate request input
	if err := s.validateRequest(req); err != nil {
		return s.createErrorResponse(err.Error()), nil
	}

	// Extract user ID from token
	userID, err := s.validateAndExtractUserID(req.AccessToken)
	if err != nil {
		return s.createErrorResponse(err.Error()), nil
	}

	// Get and validate user data
	userData, err := s.getUserAndValidate(ctx, userID)
	if err != nil {
		if isValidationError(err) {
			return s.createErrorResponse(err.Error()), nil
		}
		return nil, fmt.Errorf("failed to get user details: %w", err)
	}

	// Build successful response with user context
	userContext := s.buildUserContext(userData)
	return s.createSuccessResponse(userContext), nil
}

// validateRequest validates the incoming request
func (s *ValidateTokenService) validateRequest(req ValidateTokenRequest) error {
	if req.AccessToken == "" {
		return fmt.Errorf("access token is required")
	}
	return nil
}

// validateAndExtractUserID validates the token and extracts user ID
func (s *ValidateTokenService) validateAndExtractUserID(token string) (UserID, error) {
	userID, err := s.tokenProvider.ValidateAccessToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid or expired token")
	}
	return userID, nil
}

// getUserAndValidate retrieves user data and validates user status
func (s *ValidateTokenService) getUserAndValidate(ctx context.Context, userID UserID) (*User, error) {
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !userData.Active {
		return nil, fmt.Errorf("user account is inactive")
	}

	return &userData, nil
}

// buildUserContext creates user context data from user entity
func (s *ValidateTokenService) buildUserContext(userData *User) *UserContextData {
	userContext := &UserContextData{
		UserID:   userData.ID,
		Email:    userData.Email,
		Username: userData.Username,
		Name:     userData.Name,
		Active:   userData.Active,
		Roles:    userData.Roles,
	}

	// Set primary role (first role if available, default to "user")
	if len(userData.Roles) > 0 {
		userContext.Role = userData.Roles[0]
	} else {
		userContext.Role = "user"
	}

	// TODO: Add permissions based on roles (future enhancement)
	userContext.Permissions = []string{} // Will be implemented later

	return userContext
}

// createErrorResponse creates a validation error response
func (s *ValidateTokenService) createErrorResponse(errorMsg string) *ValidateTokenResponse {
	return &ValidateTokenResponse{
		Valid: false,
		Error: errorMsg,
	}
}

// createSuccessResponse creates a successful validation response
func (s *ValidateTokenService) createSuccessResponse(userContext *UserContextData) *ValidateTokenResponse {
	return &ValidateTokenResponse{
		Valid:       true,
		UserContext: userContext,
	}
}

// isValidationError checks if the error is a validation error (not a system error)
func isValidationError(err error) bool {
	errMsg := err.Error()
	return errMsg == "user not found" || errMsg == "user account is inactive"
}
