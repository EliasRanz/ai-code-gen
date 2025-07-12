package authtest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
)

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		session  auth.Session
		expected bool
	}{
		{
			name: "not expired session",
			session: auth.Session{
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "expired session",
			session: auth.Session{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "session expiring now",
			session: auth.Session{
				ExpiresAt: time.Now(),
			},
			expected: true, // Equal to now is considered expired
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.session.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		user     auth.User
		role     string
		expected bool
	}{
		{
			name: "user has exact role",
			user: auth.User{
				Roles: []string{"user", "editor"},
			},
			role:     "editor",
			expected: true,
		},
		{
			name: "user does not have role",
			user: auth.User{
				Roles: []string{"user"},
			},
			role:     "admin",
			expected: false,
		},
		{
			name: "admin user has any role",
			user: auth.User{
				Roles: []string{"admin"},
			},
			role:     "editor",
			expected: true,
		},
		{
			name: "super_admin user has any role",
			user: auth.User{
				Roles: []string{"super_admin"},
			},
			role:     "user",
			expected: true,
		},
		{
			name: "user with no roles",
			user: auth.User{
				Roles: []string{},
			},
			role:     "user",
			expected: false,
		},
		{
			name: "admin and other roles",
			user: auth.User{
				Roles: []string{"user", "admin", "editor"},
			},
			role:     "manager",
			expected: true, // Admin can access any role
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.HasRole(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		token    auth.Token
		expected bool
	}{
		{
			name: "not expired token",
			token: auth.Token{
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "expired token",
			token: auth.Token{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationParams_Validate(t *testing.T) {
	tests := []struct {
		name     string
		params   auth.PaginationParams
		hasError bool
	}{
		{
			name: "valid pagination params",
			params: auth.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			hasError: false,
		},
		{
			name: "zero page",
			params: auth.PaginationParams{
				Page:  0,
				Limit: 10,
			},
			hasError: true,
		},
		{
			name: "negative page",
			params: auth.PaginationParams{
				Page:  -1,
				Limit: 10,
			},
			hasError: true,
		},
		{
			name: "zero limit",
			params: auth.PaginationParams{
				Page:  1,
				Limit: 0,
			},
			hasError: true,
		},
		{
			name: "limit too large",
			params: auth.PaginationParams{
				Page:  1,
				Limit: 101,
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaginationParams_Offset(t *testing.T) {
	tests := []struct {
		name     string
		params   auth.PaginationParams
		expected int32
	}{
		{
			name: "first page",
			params: auth.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			expected: 0,
		},
		{
			name: "second page",
			params: auth.PaginationParams{
				Page:  2,
				Limit: 10,
			},
			expected: 10,
		},
		{
			name: "third page with larger limit",
			params: auth.PaginationParams{
				Page:  3,
				Limit: 25,
			},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.Offset()
			assert.Equal(t, tt.expected, result)
		})
	}
}
