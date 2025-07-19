// Package utilities provides service-specific HTTP handler interfaces
package utilities

// AuthHandler defines authentication-specific HTTP operations
type AuthHandler interface {
	HTTPHandler
	Login(ctx Context) error
	Logout(ctx Context) error
	RefreshToken(ctx Context) error
	ValidateToken(ctx Context) error
}

// UserHandler defines user management-specific HTTP operations
type UserHandler interface {
	HTTPHandler
	GetUser(ctx Context) error
	UpdateUser(ctx Context) error
	GetProjects(ctx Context) error
}

// AIHandler defines AI service-specific HTTP operations
type AIHandler interface {
	HTTPHandler
	GenerateCode(ctx Context) error
	GetGenerationHistory(ctx Context) error
	StreamGeneration(ctx Context) error
}
