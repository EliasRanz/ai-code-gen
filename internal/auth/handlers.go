package auth

import (
	"net/http"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// LoginHandler handles user login with email/password
func LoginHandler(c *gin.Context) {
	log.Info().Msg("Login attempt")
	type LoginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}
	// Lookup user
	service := c.MustGet("authService").(*Service)
	user, err := service.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// Generate tokens using service's token manager
	accessToken, err := service.TokenManager.GenerateToken(user.ID, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}
	refreshToken, err := service.TokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600,
	})
}

// GoogleOAuthHandler initiates Google OAuth flow
func GoogleOAuthHandler(c *gin.Context) {
	log.Info().Msg("Google OAuth initiation")

	// Get auth service from context
	service := c.MustGet("authService").(*Service)

	// Get config from context (will be injected by the setup)
	cfg, exists := c.Get("config")
	if !exists {
		log.Error().Msg("Config not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
		return
	}
	config := cfg.(*config.Config)

	// Generate state parameter for CSRF protection
	state, err := service.TokenManager.GenerateToken("oauth-state", time.Minute*10)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate OAuth state")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate OAuth"})
		return
	}

	// Build Google OAuth URL
	authURL := "https://accounts.google.com/o/oauth2/auth" +
		"?client_id=" + config.Auth.OAuth.Google.ClientID +
		"&redirect_uri=" + config.Auth.OAuth.Google.RedirectURL +
		"&scope=openid%20email%20profile" +
		"&response_type=code" +
		"&state=" + state

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GoogleCallbackHandler handles Google OAuth callback
func GoogleCallbackHandler(c *gin.Context) {
	log.Info().Msg("Google OAuth callback")

	// Get code and state from query parameters
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State parameter not provided"})
		return
	}

	// Get auth service from context
	service := c.MustGet("authService").(*Service)

	// Validate state parameter (basic validation)
	_, err := service.TokenManager.ValidateToken(state)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid OAuth state parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Exchange code for tokens (simplified implementation)
	// In a production environment, this should use proper OAuth2 library
	// and exchange the code for an access token, then get user info

	// For now, we'll create a placeholder user based on the code
	// This should be replaced with actual Google API calls
	userEmail := "oauth-user-" + code[:8] + "@example.com" // Placeholder
	userName := "OAuth User"

	// Check if user exists, create if not
	existingUser, err := service.userRepo.GetByEmail(userEmail)
	if err != nil {
		log.Error().Err(err).Msg("Database error during OAuth callback")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var currentUser *user.User
	if existingUser == nil {
		// Create new user
		newUser := &user.User{
			Email:         userEmail,
			Name:          userName,
			IsActive:      true,
			EmailVerified: true, // OAuth users are considered verified
		}
		if err := service.userRepo.Create(newUser); err != nil {
			log.Error().Err(err).Msg("Failed to create OAuth user")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		currentUser = newUser
	} else {
		currentUser = existingUser
	}

	// Generate JWT tokens
	accessToken, err := service.TokenManager.GenerateToken(currentUser.ID, time.Hour)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate access token for OAuth user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := service.TokenManager.GenerateRefreshToken(currentUser.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate refresh token for OAuth user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600,
		"user": gin.H{
			"id":    currentUser.ID,
			"email": currentUser.Email,
			"name":  currentUser.Name,
		},
	})
}

// RefreshTokenHandler handles token refresh
func RefreshTokenHandler(c *gin.Context) {
	log.Info().Msg("Token refresh attempt")

	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	// Get auth service from context
	service := c.MustGet("authService").(*Service)

	// Use service method to refresh tokens
	accessToken, newRefreshToken, err := service.RefreshToken(req.RefreshToken)
	if err != nil {
		log.Warn().Err(err).Msg("Token refresh failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    900, // 15 minutes
	})
}

// LogoutHandler handles user logout
func LogoutHandler(c *gin.Context) {
	log.Info().Msg("Logout attempt")

	type LogoutRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	// Get auth service from context
	service := c.MustGet("authService").(*Service)

	// Use service method to logout (invalidate refresh token)
	err := service.Logout(req.RefreshToken)
	if err != nil {
		log.Warn().Err(err).Msg("Logout failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// ValidateTokenHandler validates an access token
func ValidateTokenHandler(c *gin.Context) {
	log.Info().Msg("Token validation attempt")

	type ValidateRequest struct {
		Token string `json:"token" binding:"required"`
	}

	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	// Get auth service from context
	service := c.MustGet("authService").(*Service)

	// Use service method to validate token
	userID, err := service.ValidateToken(req.Token)
	if err != nil {
		log.Warn().Err(err).Msg("Token validation failed")
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": "Invalid or expired token",
		})
		return
	}

	// Get user details
	user, err := service.userRepo.GetByID(userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("Failed to get user details")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"user_id":    user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"is_active":  user.IsActive,
		"roles":      user.Roles,
		"expires_at": time.Now().Add(time.Hour).Unix(), // Approximate expiry
	})
}


// GetUserHandler returns current user information
func GetUserHandler(c *gin.Context) {
	log.Info().Msg("Get user info request")

	// Extract user ID from JWT context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get auth service from context
	service := c.MustGet("authService").(*Service)

	// Fetch user from database
	user, err := service.userRepo.GetByID(userID.(string))
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.(string)).Msg("Failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":        user.ID,
		"email":          user.Email,
		"name":           user.Name,
		"avatar_url":     user.AvatarURL,
		"roles":          user.Roles,
		"is_active":      user.IsActive,
		"email_verified": user.EmailVerified,
		"last_login_at":  user.LastLoginAt,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
	})
}

// ChangePasswordHandler handles password change requests
func ChangePasswordHandler(c *gin.Context) {
	log.Info().Msg("Password change attempt")

	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password and new password are required"})
		return
	}

	// Extract user ID from JWT context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Validate new password strength
	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 8 characters long"})
		return
	}

	// Get auth service from context
	service := c.MustGet("authService").(*Service)

	// Get user from database
	userEntity, err := service.userRepo.GetByID(userID.(string))
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.(string)).Msg("Failed to get user for password change")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if userEntity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Validate current password
	err = bcrypt.CompareHashAndPassword([]byte(userEntity.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		log.Warn().Str("user_id", userID.(string)).Msg("Invalid current password in change password attempt")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash new password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process new password"})
		return
	}

	// Update password in database
	updates := map[string]interface{}{
		"password_hash": string(newPasswordHash),
		"updated_at":    time.Now(),
	}

	_, err = service.userRepo.Update(userID.(string), updates)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.(string)).Msg("Failed to update password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	log.Info().Str("user_id", userID.(string)).Msg("Password changed successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}
