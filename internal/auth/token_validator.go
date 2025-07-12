package auth

// TokenValidator provides token validation functionality for middleware
type TokenValidator interface {
	ValidateAccessToken(token string) (UserID, error)
}

// Ensure JWTTokenProvider implements TokenValidator
var _ TokenValidator = (*JWTTokenProvider)(nil)
