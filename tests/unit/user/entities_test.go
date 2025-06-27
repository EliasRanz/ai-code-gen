// Package user contains tests for user domain entities
package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
)

// MockPasswordHasher for testing
type MockPasswordHasher struct {
	hashFunc   func(string) (string, error)
	verifyFunc func(string, string) bool
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	if m.hashFunc != nil {
		return m.hashFunc(password)
	}
	return "hashed_" + password, nil
}

func (m *MockPasswordHasher) Verify(password, hash string) bool {
	if m.verifyFunc != nil {
		return m.verifyFunc(password, hash)
	}
	return hash == "hashed_"+password
}

func TestUser_SetPassword(t *testing.T) {
	user := &user.User{
		Email: "test@example.com",
	}
	hasher := &MockPasswordHasher{}

	t.Run("should set password hash successfully", func(t *testing.T) {
		password := "testpassword123"
		err := user.SetPassword(hasher, password)

		require.NoError(t, err)
		assert.Equal(t, "hashed_"+password, user.PasswordHash)
	})

	t.Run("should handle hasher error", func(t *testing.T) {
		hasher := &MockPasswordHasher{
			hashFunc: func(string) (string, error) {
				return "", assert.AnError
			},
		}

		err := user.SetPassword(hasher, "password")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("should handle empty password", func(t *testing.T) {
		err := user.SetPassword(hasher, "")

		require.NoError(t, err)
		assert.Equal(t, "hashed_", user.PasswordHash)
	})
}

func TestUser_VerifyPassword(t *testing.T) {
	user := &user.User{
		Email:        "test@example.com",
		PasswordHash: "hashed_correctpassword",
	}

	t.Run("should verify correct password", func(t *testing.T) {
		hasher := &MockPasswordHasher{}
		result := user.VerifyPassword(hasher, "correctpassword")
		assert.True(t, result)
	})

	t.Run("should reject incorrect password", func(t *testing.T) {
		hasher := &MockPasswordHasher{}
		result := user.VerifyPassword(hasher, "wrongpassword")
		assert.False(t, result)
	})

	t.Run("should handle custom verify logic", func(t *testing.T) {
		hasher := &MockPasswordHasher{
			verifyFunc: func(password, hash string) bool {
				return password == "special" && hash == "hashed_correctpassword"
			},
		}

		result := user.VerifyPassword(hasher, "special")
		assert.True(t, result)

		result = user.VerifyPassword(hasher, "notspecial")
		assert.False(t, result)
	})
}

func TestUser_PasswordWorkflow(t *testing.T) {
	user := &user.User{
		Email: "test@example.com",
	}
	hasher := &MockPasswordHasher{}

	t.Run("complete password workflow", func(t *testing.T) {
		password := "mypassword123"

		// Set password
		err := user.SetPassword(hasher, password)
		require.NoError(t, err)
		assert.NotEmpty(t, user.PasswordHash)

		// Verify correct password
		assert.True(t, user.VerifyPassword(hasher, password))

		// Verify incorrect password
		assert.False(t, user.VerifyPassword(hasher, "wrongpassword"))

		// Change password
		newPassword := "newpassword456"
		err = user.SetPassword(hasher, newPassword)
		require.NoError(t, err)

		// Old password should not work
		assert.False(t, user.VerifyPassword(hasher, password))

		// New password should work
		assert.True(t, user.VerifyPassword(hasher, newPassword))
	})
}
