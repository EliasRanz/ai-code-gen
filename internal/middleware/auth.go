package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware validates JWT tokens with full user context (for services with database access)
func AuthMiddleware(tokenManager *auth.TokenManager, userRepo user.Repository) gin.HandlerFunc {
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not provided or invalid"})
			c.Abort()
			return
		}

		// Validate JWT token
		userID, err := tokenManager.ValidateAccessToken(token)
		if err != nil {
			log.Debug().
				Str("token_prefix", token[:min(10, len(token))]).
				Err(err).
				Msg("JWT token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Get user details from repository
		userData, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			log.Debug().
				Str("user_id", string(userID)).
				Err(err).
				Msg("Failed to get user details")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Check if user is active
		if !userData.Active {
			log.Debug().
				Str("user_id", string(userID)).
				Msg("Inactive user attempted authentication")
			c.JSON(http.StatusForbidden, gin.H{"error": "User account is inactive"})
			c.Abort()
			return
		}

		// Determine user role (take first role if multiple, default to "user")
		userRole := "user"
		if len(userData.Roles) > 0 {
			userRole = userData.Roles[0]
		}

		// Set user context from validated JWT claims
		c.Set("user_id", string(userID))
		c.Set("user_email", userData.Email)
		c.Set("user_role", userRole)
		c.Set("authenticated", true)

		log.Debug().
			Str("user_id", string(userID)).
			Str("user_email", userData.Email).
			Str("user_role", userRole).
			Msg("User authenticated successfully")

		c.Next()
	}
}

// LightweightAuthMiddleware validates JWT tokens without database access (for API gateways)
func LightweightAuthMiddleware(tokenManager *auth.TokenManager) gin.HandlerFunc {
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not provided or invalid"})
			c.Abort()
			return
		}

		// Validate JWT token
		userID, err := tokenManager.ValidateAccessToken(token)
		if err != nil {
			log.Debug().
				Str("token_prefix", token[:min(10, len(token))]).
				Err(err).
				Msg("JWT token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set minimal user context from validated JWT claims
		c.Set("user_id", string(userID))
		c.Set("authenticated", true)

		log.Debug().
			Str("user_id", string(userID)).
			Msg("User authenticated successfully (lightweight)")

		c.Next()
	}
}

// OptionalAuth middleware that doesn't fail if no auth is provided
func OptionalAuth(tokenManager *auth.TokenManager, userRepo user.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Try to validate token but don't fail if invalid
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != "" {
				// Validate token and set user context if valid
				userID, err := tokenManager.ValidateAccessToken(token)
				if err == nil {
					// Get user details from repository
					userData, err := userRepo.GetByID(c.Request.Context(), userID)
					if err == nil && userData.Active {
						// Determine user role (take first role if multiple, default to "user")
						userRole := "user"
						if len(userData.Roles) > 0 {
							userRole = userData.Roles[0]
						}

						// Set user context from validated JWT claims
						c.Set("user_id", string(userID))
						c.Set("user_email", userData.Email)
						c.Set("user_role", userRole)
						c.Set("authenticated", true)

						// Log successful auth for debugging
						uid, _ := strconv.ParseUint(string(userID), 10, 64)
						log.Debug().
							Str("method", c.Request.Method).
							Str("path", c.Request.URL.Path).
							Uint64("user_id", uid).
							Str("user_role", userRole).
							Msg("Authenticated request (optional)")
					} else {
						// Log failed auth for debugging
						uid, _ := strconv.ParseUint(string(userID), 10, 64)
						log.Debug().
							Str("method", c.Request.Method).
							Str("path", c.Request.URL.Path).
							Uint64("user_id", uid).
							Msg("User inactive or not found (optional auth)")
					}
				} else {
					log.Debug().
						Str("token_prefix", token[:min(10, len(token))]).
						Err(err).
						Msg("Optional auth: JWT token validation failed")
				}
			}
		}

		c.Next()
	}
}

// AdminRequired middleware ensures user has admin privileges
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check user role from context for admin access
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// Validate user has admin role
		if userRole != "admin" {
			userID, _ := c.Get("user_id")
			log.Debug().
				Str("user_id", userID.(string)).
				Str("user_role", userRole.(string)).
				Msg("Non-admin user attempted admin access")
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		log.Debug().
			Str("user_id", string(c.GetString("user_id"))).
			Msg("Admin access granted")
		c.Next()
	}
}

// RequireRole middleware ensures user has a specific role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check user role from context
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// Validate user has required role
		if userRole != requiredRole {
			userID, _ := c.Get("user_id")
			log.Debug().
				Str("user_id", userID.(string)).
				Str("user_role", userRole.(string)).
				Str("required_role", requiredRole).
				Msg("User lacks required role for access")
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAuthentication middleware ensures user is authenticated
func RequireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserContext extracts user information from context
func GetUserContext(c *gin.Context) (userID common.UserID, email, role string, authenticated bool) {
	if auth, exists := c.Get("authenticated"); exists && auth.(bool) {
		if uid, exists := c.Get("user_id"); exists {
			userID = common.UserID(uid.(string))
		}
		if em, exists := c.Get("user_email"); exists {
			email = em.(string)
		}
		if r, exists := c.Get("user_role"); exists {
			role = r.(string)
		}
		return userID, email, role, true
	}
	return "", "", "", false
}

// IsAdmin checks if the current user has admin role
func IsAdmin(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	return exists && userRole == "admin"
}

// IsAuthenticated checks if the current user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	authenticated, exists := c.Get("authenticated")
	return exists && authenticated.(bool)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LightweightAuthMiddlewareWithClaims validates JWT tokens and extracts claims (for services needing claim data)
func LightweightAuthMiddlewareWithClaims(tokenManager *auth.TokenManager, userRepo user.Repository, roles ...string) gin.HandlerFunc {
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not provided or invalid"})
			c.Abort()
			return
		}

		// Validate token
		userID, err := tokenManager.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Get user details from repository
		userData, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Check if user is active
		if !userData.Active {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is inactive"})
			c.Abort()
			return
		}

		// Check user role
		userRole := "user"
		if len(userData.Roles) > 0 {
			userRole = userData.Roles[0]
		}

		roleMatch := false
		if len(roles) == 0 {
			roleMatch = true // No specific roles required
		} else {
			for _, r := range roles {
				if userRole == r {
					roleMatch = true
					break
				}
			}
		}

		if !roleMatch {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", string(userID))
		c.Set("user_email", userData.Email)
		c.Set("user_role", userRole)
		c.Set("authenticated", true)

		c.Next()
	}
}
