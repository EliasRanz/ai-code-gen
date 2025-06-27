package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/user"
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

// Service provides authentication business logic
type Service struct {
	userRepo       user.Repository
	TokenManager   *TokenManager
	passwordHasher PasswordHasher
}

// NewService creates a new auth service
func NewService(userRepo user.Repository, tokenManager *TokenManager) *Service {
	return &Service{
		userRepo:       userRepo,
		TokenManager:   tokenManager,
		passwordHasher: nil, // Will be set via SetPasswordHasher
	}
}

// NewServiceWithPasswordHasher creates a new auth service with password hasher
func NewServiceWithPasswordHasher(userRepo user.Repository, tokenManager *TokenManager, passwordHasher PasswordHasher) *Service {
	return &Service{
		userRepo:       userRepo,
		TokenManager:   tokenManager,
		passwordHasher: passwordHasher,
	}
}

// SetPasswordHasher sets the password hasher for the service
func (s *Service) SetPasswordHasher(passwordHasher PasswordHasher) {
	s.passwordHasher = passwordHasher
}

// Login authenticates a user
func (s *Service) Login(email, password string) (string, error) {
	// Validate input
	if email == "" {
		return "", errors.New("email cannot be empty")
	}
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return "", ErrUserInactive
	}

	// Verify password
	if s.passwordHasher != nil {
		if !s.passwordHasher.Verify(password, user.PasswordHash) {
			return "", ErrInvalidCredentials
		}
	} else {
		// Fallback to basic bcrypt verification for backward compatibility
		if err := bcryptVerify(password, user.PasswordHash); err != nil {
			return "", ErrInvalidCredentials
		}
	}

	// Generate access token (15 minutes expiry)
	accessToken, err := s.TokenManager.GenerateToken(user.ID, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return accessToken, nil
}

// ValidateToken validates a JWT token
func (s *Service) ValidateToken(token string) (string, error) {
	// Validate input
	if token == "" {
		return "", errors.New("token cannot be empty")
	}

	// Validate token using token manager
	userID, err := s.TokenManager.ValidateToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	// Verify user still exists and is active
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	if !user.IsActive {
		return "", ErrUserInactive
	}

	return userID, nil
}

// RefreshToken generates a new access and refresh token from a valid refresh token
func (s *Service) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := s.TokenManager.ParseToken(refreshToken)
	if err != nil {
		return "", "", err
	}
	typ, ok := claims["typ"].(string)
	if !ok || typ != "refresh" {
		return "", "", ErrInvalidTokenType
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", "", ErrInvalidToken
	}
	// Validate signature and expiry
	_, err = s.TokenManager.ValidateToken(refreshToken)
	if err != nil {
		return "", "", err
	}
	accessToken, err := s.TokenManager.GenerateToken(userID, 15*60) // 15 min expiry
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err := s.TokenManager.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}
	return accessToken, newRefreshToken, nil
}

// Logout invalidates a refresh token (stub, stateless)
func (s *Service) Logout(refreshToken string) error {
	// In a real implementation, you would blacklist the token or remove it from a store
	return nil
}

var (
	ErrInvalidTokenType   = errors.New("invalid token type")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user account is not active")
)

// bcryptVerify is a fallback function for password verification
func bcryptVerify(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
