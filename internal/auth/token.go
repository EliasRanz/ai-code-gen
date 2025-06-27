package auth

import (
	"time"

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

// GenerateToken generates a new JWT token
func (tm *TokenManager) GenerateToken(userID string, expiresIn time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iss": tm.issuer,
		"exp": time.Now().Add(expiresIn).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// ValidateToken validates a JWT token and returns user ID
func (tm *TokenManager) ValidateToken(tokenStr string) (string, error) {
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
		userID, ok := claims["sub"].(string)
		if !ok {
			return "", jwt.ErrTokenMalformed
		}
		return userID, nil
	}
	return "", jwt.ErrTokenMalformed
}

// ParseToken parses a JWT token without validation
func (tm *TokenManager) ParseToken(tokenStr string) (map[string]interface{}, error) {
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}
	return nil, jwt.ErrTokenMalformed
}

// GenerateRefreshToken generates a refresh token
func (tm *TokenManager) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iss": tm.issuer,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days expiry
		"iat": time.Now().Unix(),
		"typ": "refresh",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}
