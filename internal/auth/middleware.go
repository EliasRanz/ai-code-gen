package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/EliasRanz/ai-code-gen/internal/user"
)

// JWTMiddleware validates JWT tokens for protected routes
func JWTMiddleware(authService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check for Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			c.Abort()
			return
		}

		// Validate JWT token using auth service
		userID, err := authService.ValidateToken(token)
		if err != nil {
			log.Warn().
				Err(err).
				Str("token_prefix", token[:min(10, len(token))]).
				Msg("JWT token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Get user details to set in context
		user, err := authService.userRepo.GetByID(userID)
		if err != nil || user == nil {
			log.Warn().
				Err(err).
				Str("user_id", userID).
				Msg("Failed to get user details for validated token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user context from validated token claims
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user_roles", user.Roles)
		c.Set("user", user) // Set full user object for convenience

		log.Debug().
			Str("user_id", user.ID).
			Str("user_email", user.Email).
			Strs("user_roles", user.Roles).
			Msg("JWT token validated successfully")

		c.Next()
	}
}

// AuthMiddleware provides JWT authentication middleware (legacy function for compatibility)
func AuthMiddleware(authService *Service) gin.HandlerFunc {
	return JWTMiddleware(authService)
}

// RequireAuth ensures user is authenticated
func RequireAuth(authService *Service) gin.HandlerFunc {
	return JWTMiddleware(authService)
}

// RequireAdmin ensures user has admin privileges
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user roles from context (set by JWT middleware)
		roles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User roles not found"})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid user roles"})
			c.Abort()
			return
		}

		// Check if user has admin role
		hasAdmin := false
		for _, role := range userRoles {
			if role == "admin" {
				hasAdmin = true
				break
			}
		}

		if !hasAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		log.Debug().
			Str("user_id", c.GetString("user_id")).
			Msg("Admin access granted")

		c.Next()
	}
}

// Helper functions for extracting user information from context

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

// GetUserEmail extracts the user email from the Gin context
func GetUserEmail(c *gin.Context) (string, bool) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		return "", false
	}
	email, ok := userEmail.(string)
	return email, ok
}

// GetUserRoles extracts the user roles from the Gin context
func GetUserRoles(c *gin.Context) ([]string, bool) {
	userRoles, exists := c.Get("user_roles")
	if !exists {
		return nil, false
	}
	roles, ok := userRoles.([]string)
	return roles, ok
}

// GetUser extracts the full user object from the Gin context
func GetUser(c *gin.Context) (*user.User, bool) {
	userObj, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	user, ok := userObj.(*user.User)
	return user, ok
}

// HasRole checks if the current user has a specific role
func HasRole(c *gin.Context, role string) bool {
	roles, exists := GetUserRoles(c)
	if !exists {
		return false
	}

	for _, userRole := range roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// IsAdmin checks if the current user has admin role
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, "admin")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
