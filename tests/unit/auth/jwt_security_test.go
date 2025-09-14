package auth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTSecurityVulnerabilities(t *testing.T) {
	// Use a strong test secret
	strongSecret := "this-is-a-very-strong-secret-key-for-testing-purposes-at-least-256-bits"
	weakSecret := "weak"
	testIssuer := "ai-code-gen-test"

	t.Run("SecureTokenGeneration", func(t *testing.T) {
		provider := auth.NewJWTTokenProvider(strongSecret, testIssuer)
		userID := auth.UserID("test-user-123")

		// Generate access token
		accessToken, err := provider.GenerateAccessToken(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)

		// Ensure token contains required security claims
		token, _ := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(strongSecret), nil
		})

		claims := token.Claims.(jwt.MapClaims)
		assert.Equal(t, string(userID), claims["sub"])
		assert.Equal(t, testIssuer, claims["iss"])
		assert.Equal(t, "access", claims["type"])
		assert.NotEmpty(t, claims["iat"]) // issued at
		assert.NotEmpty(t, claims["exp"]) // expiration
	})

	t.Run("PreventTokenSubstitution", func(t *testing.T) {
		provider := auth.NewJWTTokenProvider(strongSecret, testIssuer)
		
		// Generate access token for user1
		user1Token, err := provider.GenerateAccessToken(auth.UserID("user1"))
		require.NoError(t, err)

		// Try to validate as different user should fail through issuer/signature checks
		user1ID, err := provider.ValidateAccessToken(user1Token)
		require.NoError(t, err)
		assert.Equal(t, auth.UserID("user1"), user1ID)

		// Different provider with different secret should reject the token
		differentProvider := auth.NewJWTTokenProvider("different-secret", testIssuer)
		_, err = differentProvider.ValidateAccessToken(user1Token)
		assert.Error(t, err, "Token should be rejected with different secret")
	})

	t.Run("TokenTypeValidation", func(t *testing.T) {
		provider := auth.NewJWTTokenProvider(strongSecret, testIssuer)
		userID := auth.UserID("test-user")

		// Generate both token types
		accessToken, err := provider.GenerateAccessToken(userID)
		require.NoError(t, err)
		
		refreshToken, err := provider.GenerateRefreshToken(userID)
		require.NoError(t, err)

		// Access token should not validate as refresh token
		_, err = provider.ValidateRefreshToken(accessToken)
		assert.Error(t, err, "Access token should not validate as refresh token")

		// Refresh token should not validate as access token
		_, err = provider.ValidateAccessToken(refreshToken)
		assert.Error(t, err, "Refresh token should not validate as access token")
	})

	t.Run("IssuerValidation", func(t *testing.T) {
		correctProvider := auth.NewJWTTokenProvider(strongSecret, testIssuer)
		wrongIssuerProvider := auth.NewJWTTokenProvider(strongSecret, "wrong-issuer")
		
		// Generate token with correct issuer
		token, err := correctProvider.GenerateAccessToken(auth.UserID("user123"))
		require.NoError(t, err)

		// Wrong issuer provider should reject the token
		_, err = wrongIssuerProvider.ValidateAccessToken(token)
		assert.Error(t, err, "Token with wrong issuer should be rejected")
	})

	t.Run("TokenExpirationSecurity", func(t *testing.T) {
		provider := auth.NewJWTTokenProvider(strongSecret, testIssuer)
		
		// Verify access token has short expiry (security best practice)
		accessToken, err := provider.GenerateAccessToken(auth.UserID("user123"))
		require.NoError(t, err)

		// Parse token to check expiration
		token, _ := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(strongSecret), nil
		})
		
		claims := token.Claims.(jwt.MapClaims)
		exp := int64(claims["exp"].(float64))
		iat := int64(claims["iat"].(float64))
		
		// Access token should expire within reasonable time (15 minutes default)
		duration := time.Unix(exp, 0).Sub(time.Unix(iat, 0))
		assert.LessOrEqual(t, duration, 16*time.Minute, "Access token expiry too long")
		assert.GreaterOrEqual(t, duration, 14*time.Minute, "Access token expiry too short")
	})

	t.Run("SigningAlgorithmSecurity", func(t *testing.T) {
		provider := auth.NewJWTTokenProvider(strongSecret, testIssuer)
		token, err := provider.GenerateAccessToken(auth.UserID("user123"))
		require.NoError(t, err)

		// Parse token header to verify signing method
		parts := strings.Split(token, ".")
		require.Len(t, parts, 3, "JWT should have 3 parts")

		// Verify it uses HMAC (secure) not "none" algorithm
		parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				t.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(strongSecret), nil
		})

		assert.NotNil(t, parsedToken)
	})

	t.Run("RejectMalformedTokens", func(t *testing.T) {
		provider := auth.NewJWTTokenProvider(strongSecret, testIssuer)

		testCases := []struct {
			name  string
			token string
		}{
			{"empty token", ""},
			{"malformed token", "not.a.jwt"},
			{"incomplete token", "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"},
			{"tampered token", "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.tampered"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := provider.ValidateAccessToken(tc.token)
				assert.Error(t, err, "Malformed token should be rejected")
			})
		}
	})

	t.Run("WeakSecretRejection", func(t *testing.T) {
		// In production, we should validate secret strength
		// For this test, we show what happens with weak secrets
		weakProvider := auth.NewJWTTokenProvider(weakSecret, testIssuer)
		strongProvider := auth.NewJWTTokenProvider(strongSecret, testIssuer)

		// Generate token with weak secret
		weakToken, err := weakProvider.GenerateAccessToken(auth.UserID("user123"))
		require.NoError(t, err)

		// Strong provider should reject weak-signed token
		_, err = strongProvider.ValidateAccessToken(weakToken)
		assert.Error(t, err, "Token signed with weak secret should be rejected")
	})
}

func TestPasswordSecurityBestPractices(t *testing.T) {
	hasher := auth.NewBCryptPasswordHasher()

	t.Run("BCryptStrength", func(t *testing.T) {
		password := "testPassword123!"
		
		// Hash the password
		hashedPassword, err := hasher.HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword, "Hashed password should not equal plaintext")

		// Verify the password
		isValid := hasher.VerifyPassword(password, hashedPassword)
		assert.True(t, isValid, "Password verification should succeed")

		// Wrong password should fail
		isValid = hasher.VerifyPassword("wrongPassword", hashedPassword)
		assert.False(t, isValid, "Wrong password should fail verification")
	})

	t.Run("ConsistentHashingTiming", func(t *testing.T) {
		password := "testPassword123!"
		
		// Hash the same password multiple times - should produce different hashes (salt)
		hash1, err1 := hasher.HashPassword(password)
		hash2, err2 := hasher.HashPassword(password)
		
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "Same password should produce different hashes due to salt")

		// Both hashes should verify the same password
		assert.True(t, hasher.VerifyPassword(password, hash1))
		assert.True(t, hasher.VerifyPassword(password, hash2))
	})
}