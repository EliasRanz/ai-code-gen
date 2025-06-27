// Package auth contains auth domain interfaces
package auth

import (
	"context"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
)

// Repository defines auth domain data access
type Repository interface {
	SaveSession(ctx context.Context, session Session) error
	GetSession(ctx context.Context, refreshToken string) (Session, error)
	DeleteSession(ctx context.Context, refreshToken string) error
	CleanExpiredSessions(ctx context.Context) error
}

// SessionRepository defines session-specific data access
type SessionRepository interface {
	Create(ctx context.Context, session Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (Session, error)
	GetByAccessToken(ctx context.Context, accessToken string) (Session, error)
	Update(ctx context.Context, session Session) error
	Delete(ctx context.Context, sessionID common.SessionID) error
	DeleteByUserID(ctx context.Context, userID common.UserID) error
	CleanExpired(ctx context.Context) error
}

// TokenProvider defines token generation interface
type TokenProvider interface {
	GenerateAccessToken(userID common.UserID) (string, error)
	GenerateRefreshToken(userID common.UserID) (string, error)
	ValidateAccessToken(token string) (common.UserID, error)
	ValidateRefreshToken(token string) (common.UserID, error)
}

// TokenService defines token management interface
type TokenService interface {
	GenerateTokens(ctx context.Context, user AuthenticatedUser) (Token, error)
	ValidateAccessToken(ctx context.Context, token string) (AuthenticatedUser, error)
	RefreshTokens(ctx context.Context, refreshToken string) (Token, error)
}

// PasswordService defines password verification interface
type PasswordService interface {
	VerifyPassword(plaintext, hash string) bool
}

// UserService defines user lookup interface for auth
type UserService interface {
	GetByEmail(ctx context.Context, email string) (AuthenticatedUser, string, error) // returns user, password hash, error
	GetByID(ctx context.Context, userID common.UserID) (AuthenticatedUser, error)
}
