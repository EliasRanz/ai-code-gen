// Package common contains shared domain concepts and types
package common

import (
	"errors"
	"fmt"
	"time"
)

// Common domain errors
var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized access")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("resource conflict")
	ErrInternal     = errors.New("internal error")
)

// DomainError represents a domain-specific error
type DomainError struct {
	Type    string
	Message string
	Cause   error
}

func (e DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e DomainError) Unwrap() error {
	return e.Cause
}

// Error constructors
func NewValidationError(message string, cause error) error {
	return DomainError{
		Type:    "validation",
		Message: message,
		Cause:   cause,
	}
}

func NewNotFoundError(message string) error {
	return DomainError{
		Type:    "not_found",
		Message: message,
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

// UserID represents a unique user identifier
type UserID string

// String returns the string representation of UserID
func (u UserID) String() string {
	return string(u)
}

// IsEmpty returns true if the UserID is empty
func (u UserID) IsEmpty() bool {
	return string(u) == ""
}

// ProjectID represents a unique project identifier
type ProjectID string

// String returns the string representation of ProjectID
func (p ProjectID) String() string {
	return string(p)
}

// SessionID represents a unique session identifier
type SessionID string

// String returns the string representation of SessionID
func (s SessionID) String() string {
	return string(s)
}

// Timestamps provides common timestamp fields
type Timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Touch updates the UpdatedAt timestamp
func (t *Timestamps) Touch() {
	t.UpdatedAt = time.Now()
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
