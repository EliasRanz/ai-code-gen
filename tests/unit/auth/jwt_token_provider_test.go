package auth_test

import (
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestTokenRefresh(t *testing.T) {
	jwtProvider := auth.NewJWTTokenProvider("testsecret", "testissuer")
	assert.NotNil(t, jwtProvider)

	// Test that we can generate a refresh token
	refreshToken, err := jwtProvider.GenerateRefreshToken(auth.UserID("user123"))
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)
}
