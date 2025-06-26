// Package auth contains auth domain entities and business rules
package auth

import (
	"time"

	"github.com/ai-code-gen/ai-ui-generator/internal/domain/common"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string
	Password string
}

// Validate validates the login request
func (r LoginRequest) Validate() error {
	if r.Email == "" {
		return common.ErrInvalidInput
	}
	if r.Password == "" {
		return common.ErrInvalidInput
	}
	return nil
}

// Token represents an authentication token
type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
}

// IsExpired returns true if the token is expired
func (t Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// AuthenticatedUser represents an authenticated user
type AuthenticatedUser struct {
	UserID   common.UserID
	Email    string
	Role     string
	IsActive bool
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string
}

// Validate validates the refresh token request
func (r RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return common.ErrInvalidInput
	}
	return nil
}

// Session represents a user session
type Session struct {
	ID           common.SessionID
	UserID       common.UserID
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Status       SessionStatus
	CreatedAt    time.Time
}

// SessionStatus represents the status of a session
type SessionStatus string

const (
	StatusActive  SessionStatus = "active"
	StatusExpired SessionStatus = "expired"
	StatusRevoked SessionStatus = "revoked"
)

// IsExpired returns true if the session is expired
func (s Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
