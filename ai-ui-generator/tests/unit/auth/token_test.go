package authtest

import (
	"testing"

	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestTokenRefresh(t *testing.T) {
	tm := auth.NewTokenManager("testsecret", "testissuer")
	assert.NotNil(t, tm)

	// Test that we can generate a refresh token
	refreshToken, err := tm.GenerateRefreshToken("user123")
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)
}
