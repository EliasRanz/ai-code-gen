package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
)

// JWTTokenProvider implements TokenProvider interface using JWT tokens
type JWTTokenProvider struct {
	secretKey         string
	issuer            string
	accessTokenExpiry time.Duration
	refreshTokenExpiry time.Duration
}

// NewJWTTokenProvider creates a new JWT token provider
func NewJWTTokenProvider(secretKey, issuer string) *JWTTokenProvider {
	return &JWTTokenProvider{
		secretKey:          secretKey,
		issuer:             issuer,
		accessTokenExpiry:  15 * time.Minute,     // Access tokens expire in 15 minutes
		refreshTokenExpiry: 7 * 24 * time.Hour,   // Refresh tokens expire in 7 days
	}
}

// GenerateAccessToken generates a new access token for the given user ID
func (p *JWTTokenProvider) GenerateAccessToken(userID common.UserID) (string, error) {
	claims := jwt.MapClaims{
		"sub":  string(userID),
		"iss":  p.issuer,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(p.accessTokenExpiry).Unix(),
		"type": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.secretKey))
}

// GenerateRefreshToken generates a new refresh token for the given user ID
func (p *JWTTokenProvider) GenerateRefreshToken(userID common.UserID) (string, error) {
	claims := jwt.MapClaims{
		"sub":  string(userID),
		"iss":  p.issuer,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(p.refreshTokenExpiry).Unix(),
		"type": "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.secretKey))
}

// ValidateAccessToken validates an access token and returns the user ID
func (p *JWTTokenProvider) ValidateAccessToken(tokenString string) (common.UserID, error) {
	return p.validateToken(tokenString, "access")
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (p *JWTTokenProvider) ValidateRefreshToken(tokenString string) (common.UserID, error) {
	return p.validateToken(tokenString, "refresh")
}

// validateToken validates a token and returns the user ID
func (p *JWTTokenProvider) validateToken(tokenString, expectedType string) (common.UserID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(p.secretKey), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Verify token type
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != expectedType {
		return "", fmt.Errorf("invalid token type")
	}

	// Verify issuer
	iss, ok := claims["iss"].(string)
	if !ok || iss != p.issuer {
		return "", fmt.Errorf("invalid issuer")
	}

	// Extract user ID
	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("invalid subject")
	}

	return common.UserID(sub), nil
}
