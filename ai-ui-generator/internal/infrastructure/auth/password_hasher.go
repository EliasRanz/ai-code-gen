// Package auth provides authentication infrastructure implementations
package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// BCryptPasswordHasher implements the PasswordHasher interface using bcrypt
type BCryptPasswordHasher struct {
	cost int
}

// NewBCryptPasswordHasher creates a new BCrypt password hasher
func NewBCryptPasswordHasher() *BCryptPasswordHasher {
	return &BCryptPasswordHasher{
		cost: bcrypt.DefaultCost,
	}
}

// NewBCryptPasswordHasherWithCost creates a new BCrypt password hasher with custom cost
func NewBCryptPasswordHasherWithCost(cost int) *BCryptPasswordHasher {
	return &BCryptPasswordHasher{
		cost: cost,
	}
}

// Hash hashes a password using bcrypt
func (h *BCryptPasswordHasher) Hash(password string) (string, error) {
	// bcrypt has a 72-byte limit
	if len(password) > 72 {
		return "", errors.New("password length exceeds 72 bytes")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Verify verifies a password against a hash using bcrypt
func (h *BCryptPasswordHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
