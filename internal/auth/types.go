package auth

import (
	"context"
	"errors"
	"time"
)

// Common auth errors
var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized access")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("resource conflict")
	ErrInternal     = errors.New("internal error")
)

// Core types moved from domain layer for simplicity
type UserID string
type SessionID string

// Session represents a user session
type Session struct {
	ID           SessionID
	UserID       UserID
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Status       SessionStatus
	CreatedAt    time.Time
	LastAccessed time.Time
	IPAddress    string
	UserAgent    string
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
	return s.ExpiresAt.IsZero() || time.Now().After(s.ExpiresAt)
}

// User represents a user for auth purposes
type User struct {
	ID       UserID
	Email    string
	Username string
	Name     string
	Roles    []string
	Active   bool
	Password string // For auth context only
}

// HasRole checks if user has a specific role
func (u User) HasRole(role string) bool {
	// Admin has access to everything
	for _, r := range u.Roles {
		if r == "admin" || r == "super_admin" {
			return true
		}
		if r == role {
			return true
		}
	}
	return false
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

// Repositories - simplified interfaces
type SessionRepository interface {
	Create(ctx context.Context, session Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (Session, error)
	GetByAccessToken(ctx context.Context, accessToken string) (Session, error)
	Update(ctx context.Context, session Session) error
	Delete(ctx context.Context, sessionID SessionID) error
	DeleteByUserID(ctx context.Context, userID UserID) error
	CleanExpired(ctx context.Context) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id UserID) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, u User) error
	Update(ctx context.Context, u User) error
	Delete(ctx context.Context, id UserID) error
}

// TokenProvider interface for JWT operations
type TokenProvider interface {
	GenerateAccessToken(userID UserID) (string, error)
	GenerateRefreshToken(userID UserID) (string, error)
	ValidateAccessToken(token string) (UserID, error)
	ValidateRefreshToken(token string) (UserID, error)
}

// PasswordHasher interface for password operations
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int32
	Limit int32
}

// Validate validates pagination parameters
func (p PaginationParams) Validate() error {
	if p.Page < 1 {
		return ErrInvalidInput
	}
	if p.Limit < 1 || p.Limit > 100 {
		return ErrInvalidInput
	}
	return nil
}

// Offset calculates the offset for pagination
func (p PaginationParams) Offset() int32 {
	return (p.Page - 1) * p.Limit
}

// DomainError represents a domain-specific error
type DomainError struct {
	Type    string
	Message string
	Cause   error
}

func (e DomainError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Error constructors
func NewNotFoundError(message string) error {
	return DomainError{
		Type:    "not_found",
		Message: message,
	}
}

func NewValidationError(message string, cause error) error {
	return DomainError{
		Type:    "validation",
		Message: message,
		Cause:   cause,
	}
}

func NewConflictError(message string) error {
	return DomainError{
		Type:    "conflict",
		Message: message,
	}
}

func NewUnauthorizedError(message string) error {
	return DomainError{
		Type:    "unauthorized",
		Message: message,
	}
}

// Error type checkers
func IsNotFoundError(err error) bool {
	var domainErr DomainError
	return errors.As(err, &domainErr) && domainErr.Type == "not_found"
}

func IsValidationError(err error) bool {
	var domainErr DomainError
	return errors.As(err, &domainErr) && domainErr.Type == "validation"
}

func IsConflictError(err error) bool {
	var domainErr DomainError
	return errors.As(err, &domainErr) && domainErr.Type == "conflict"
}

func IsUnauthorizedError(err error) bool {
	var domainErr DomainError
	return errors.As(err, &domainErr) && domainErr.Type == "unauthorized"
}
