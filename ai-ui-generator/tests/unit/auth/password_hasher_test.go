// Package auth provides authentication infrastructure implementations
package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBCryptPasswordHasher_Hash(t *testing.T) {
	hasher := NewBCryptPasswordHasher()

	t.Run("should hash password successfully", func(t *testing.T) {
		password := "testpassword123"
		hash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)
		assert.True(t, len(hash) > 50) // bcrypt hashes are typically 60 characters
	})

	t.Run("should generate different hashes for same password", func(t *testing.T) {
		password := "testpassword123"
		hash1, err1 := hasher.Hash(password)
		hash2, err2 := hasher.Hash(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2) // bcrypt includes salt, so hashes should differ
	})

	t.Run("should handle empty password", func(t *testing.T) {
		hash, err := hasher.Hash("")

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})
}

func TestBCryptPasswordHasher_Verify(t *testing.T) {
	hasher := NewBCryptPasswordHasher()

	t.Run("should verify correct password", func(t *testing.T) {
		password := "testpassword123"
		hash, err := hasher.Hash(password)
		require.NoError(t, err)

		result := hasher.Verify(password, hash)
		assert.True(t, result)
	})

	t.Run("should reject incorrect password", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"
		hash, err := hasher.Hash(password)
		require.NoError(t, err)

		result := hasher.Verify(wrongPassword, hash)
		assert.False(t, result)
	})

	t.Run("should handle empty password verification", func(t *testing.T) {
		password := ""
		hash, err := hasher.Hash(password)
		require.NoError(t, err)

		result := hasher.Verify("", hash)
		assert.True(t, result)

		result = hasher.Verify("nonempty", hash)
		assert.False(t, result)
	})

	t.Run("should handle invalid hash", func(t *testing.T) {
		result := hasher.Verify("password", "invalid_hash")
		assert.False(t, result)
	})
}

func TestBCryptPasswordHasher_WithCustomCost(t *testing.T) {
	customCost := 6 // Lower cost for faster tests
	hasher := NewBCryptPasswordHasherWithCost(customCost)

	t.Run("should use custom cost", func(t *testing.T) {
		password := "testpassword123"
		hash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		// Verify the password still works
		result := hasher.Verify(password, hash)
		assert.True(t, result)
	})
}

func TestBCryptPasswordHasher_RoundTrip(t *testing.T) {
	hasher := NewBCryptPasswordHasher()
	passwords := []string{
		"simple",
		"complex!Password123",
		"with spaces and symbols @#$%",
		"reasonable_length_unicode_测试", // Shorter unicode password
		"1234567890",
		"!@#$%^&*()",
	}

	for _, password := range passwords {
		t.Run("password: "+password, func(t *testing.T) {
			// Hash the password
			hash, err := hasher.Hash(password)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)

			// Verify the correct password
			assert.True(t, hasher.Verify(password, hash))

			// Verify an incorrect password fails
			assert.False(t, hasher.Verify(password+"wrong", hash))
		})
	}
}

func TestBCryptPasswordHasher_LongPassword(t *testing.T) {
	hasher := NewBCryptPasswordHasher()

	t.Run("should handle very long passwords", func(t *testing.T) {
		// bcrypt has a 72-byte limit, test with a password that exceeds this
		longPassword := "this_is_a_very_long_password_that_exceeds_the_72_byte_limit_of_bcrypt_with_unicode_测试_更多字符"

		// Should return an error for very long passwords
		_, err := hasher.Hash(longPassword)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password length exceeds 72 bytes")
	})
}
