package auth_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	appAuth "github.com/EliasRanz/ai-code-gen/internal/auth"
)

func TestBCryptPasswordHasher(t *testing.T) {
	hasher := appAuth.NewBCryptPasswordHasher()

	t.Run("should hash password successfully", func(t *testing.T) {
		password := "testpassword123"

		hash, err := hasher.Hash(password)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)
		assert.True(t, strings.HasPrefix(hash, "$2a$"))
	})

	t.Run("should verify correct password", func(t *testing.T) {
		password := "testpassword123"

		hash, err := hasher.Hash(password)
		assert.NoError(t, err)

		isValid := hasher.Verify(password, hash)
		assert.True(t, isValid)
	})

	t.Run("should reject incorrect password", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"

		hash, err := hasher.Hash(password)
		assert.NoError(t, err)

		isValid := hasher.Verify(wrongPassword, hash)
		assert.False(t, isValid)
	})

	t.Run("should handle empty password", func(t *testing.T) {
		hash, err := hasher.Hash("")

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)

		isValid := hasher.Verify("", hash)
		assert.True(t, isValid)
	})

	t.Run("should handle very long password", func(t *testing.T) {
		// bcrypt has a 72-byte limit
		longPassword := strings.Repeat("a", 100)

		hash, err := hasher.Hash(longPassword)

		assert.Error(t, err)
		assert.Empty(t, hash)
		assert.Contains(t, err.Error(), "password length exceeds")
	})

	t.Run("should handle invalid hash in verify", func(t *testing.T) {
		password := "testpassword123"
		invalidHash := "not-a-valid-hash"

		isValid := hasher.Verify(password, invalidHash)
		assert.False(t, isValid)
	})
}

func TestBCryptPasswordHasherWithCustomCost(t *testing.T) {
	customCost := 6
	hasher := appAuth.NewBCryptPasswordHasherWithCost(customCost)

	t.Run("should use custom cost", func(t *testing.T) {
		password := "testpassword123"

		hash, err := hasher.Hash(password)

		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		// Verify the hash works
		isValid := hasher.Verify(password, hash)
		assert.True(t, isValid)
	})
}
