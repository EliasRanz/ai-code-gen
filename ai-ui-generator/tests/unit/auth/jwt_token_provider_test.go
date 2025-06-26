package auth

import (
	"testing"

	"github.com/ai-code-gen/ai-ui-generator/internal/domain/common"
)

func TestJWTTokenProvider_GenerateAndValidateTokens(t *testing.T) {
	provider := NewJWTTokenProvider("test-secret-key", "test-issuer")
	userID := common.UserID("user123")

	t.Run("GenerateAccessToken", func(t *testing.T) {
		token, err := provider.GenerateAccessToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate access token: %v", err)
		}
		if token == "" {
			t.Fatal("Generated token is empty")
		}
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		token, err := provider.GenerateRefreshToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate refresh token: %v", err)
		}
		if token == "" {
			t.Fatal("Generated token is empty")
		}
	})

	t.Run("ValidateAccessToken", func(t *testing.T) {
		// Generate a valid access token
		token, err := provider.GenerateAccessToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate access token: %v", err)
		}

		// Validate the token
		validatedUserID, err := provider.ValidateAccessToken(token)
		if err != nil {
			t.Fatalf("Failed to validate access token: %v", err)
		}

		if validatedUserID != userID {
			t.Fatalf("Expected user ID %s, got %s", userID, validatedUserID)
		}
	})

	t.Run("ValidateRefreshToken", func(t *testing.T) {
		// Generate a valid refresh token
		token, err := provider.GenerateRefreshToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate refresh token: %v", err)
		}

		// Validate the token
		validatedUserID, err := provider.ValidateRefreshToken(token)
		if err != nil {
			t.Fatalf("Failed to validate refresh token: %v", err)
		}

		if validatedUserID != userID {
			t.Fatalf("Expected user ID %s, got %s", userID, validatedUserID)
		}
	})

	t.Run("ValidateInvalidToken", func(t *testing.T) {
		_, err := provider.ValidateAccessToken("invalid-token")
		if err == nil {
			t.Fatal("Expected error for invalid token, got nil")
		}
	})

	t.Run("ValidateWrongTokenType", func(t *testing.T) {
		// Generate a refresh token but validate as access token
		refreshToken, err := provider.GenerateRefreshToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate refresh token: %v", err)
		}

		_, err = provider.ValidateAccessToken(refreshToken)
		if err == nil {
			t.Fatal("Expected error for wrong token type, got nil")
		}
	})
}

func TestJWTTokenProvider_InvalidSignature(t *testing.T) {
	provider1 := NewJWTTokenProvider("secret1", "issuer")
	provider2 := NewJWTTokenProvider("secret2", "issuer")
	userID := common.UserID("user123")

	// Generate token with provider1
	token, err := provider1.GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Try to validate with provider2 (different secret)
	_, err = provider2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("Expected error for token with wrong signature, got nil")
	}
}

func TestJWTTokenProvider_InvalidIssuer(t *testing.T) {
	provider1 := NewJWTTokenProvider("secret", "issuer1")
	provider2 := NewJWTTokenProvider("secret", "issuer2")
	userID := common.UserID("user123")

	// Generate token with provider1
	token, err := provider1.GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Try to validate with provider2 (different issuer)
	_, err = provider2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("Expected error for token with wrong issuer, got nil")
	}
}
