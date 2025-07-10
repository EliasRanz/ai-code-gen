package auth

import (
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/golang-jwt/jwt/v5"
)

// TokenManager handles JWT token operations
type TokenManager struct {
	secretKey []byte
	issuer    string
}

// NewTokenManager creates a new token manager
func NewTokenManager(secretKey string, issuer string) *TokenManager {
	return &TokenManager{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

// GenerateAccessToken generates a new JWT access token
func (tm *TokenManager) GenerateAccessToken(userID common.UserID) (string, error) {
	return tm.generateToken(string(userID), 15*time.Minute, "access")
}

// GenerateRefreshToken generates a new JWT refresh token
func (tm *TokenManager) GenerateRefreshToken(userID common.UserID) (string, error) {
	return tm.generateToken(string(userID), 7*24*time.Hour, "refresh")
}

func (tm *TokenManager) generateToken(userID string, expiresIn time.Duration, tokenType string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iss": tm.issuer,
		"exp": time.Now().Add(expiresIn).Unix(),
		"iat": time.Now().Unix(),
		"typ": tokenType,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// ValidateAccessToken validates an access token and returns the user ID
func (tm *TokenManager) ValidateAccessToken(tokenStr string) (common.UserID, error) {
	return tm.validateToken(tokenStr, "access")
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (tm *TokenManager) ValidateRefreshToken(tokenStr string) (common.UserID, error) {
	return tm.validateToken(tokenStr, "refresh")
}

func (tm *TokenManager) validateToken(tokenStr string, expectedType string) (common.UserID, error) {
	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return tm.secretKey, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		if tokenType, ok := claims["typ"].(string); !ok || tokenType != expectedType {
			return "", jwt.ErrTokenInvalidClaims
		}
		userID, ok := claims["sub"].(string)
		if !ok {
			return "", jwt.ErrTokenMalformed
		}
		return common.UserID(userID), nil
	}

	return "", jwt.ErrTokenMalformed
}
