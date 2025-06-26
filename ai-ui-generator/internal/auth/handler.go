package auth

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/EliasRanz/ai-code-gen/ai-ui-generator/internal/user"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers authentication routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		// OAuth login endpoints
		auth.POST("/login", h.Login)
		auth.GET("/login/google", h.GoogleOAuth)
		auth.GET("/callback/google", h.GoogleCallback)

		// JWT token management
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
		auth.POST("/validate", h.ValidateToken)

		// User registration (if not using OAuth)
		auth.POST("/register", h.Register)
	}

	// Protected routes requiring authentication
	protected := r.Group("/")
	protected.Use(JWTMiddleware(h.service))
	{
		protected.GET("/user", h.GetCurrentUser)
		protected.POST("/change-password", h.ChangePassword)
	}
}

// Login handles form-based login (alternative to OAuth)
func (h *Handler) Login(c *gin.Context) {
	type LoginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Email and password are required"})
		return
	}
	
	// Lookup user
	user, err := h.service.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if user == nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}
	
	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}
	
	// Generate tokens
	accessToken, err := h.service.TokenManager.GenerateToken(user.ID, time.Hour)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate access token"})
		return
	}
	refreshToken, err := h.service.TokenManager.GenerateRefreshToken(user.ID)
		if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600,
	})
}



// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	type LogoutRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "refresh_token is required"})
		return
	}
	err := h.service.Logout(req.RefreshToken)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Logged out successfully"})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "refresh_token is required"})
		return
	}
	accessToken, refreshToken, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// GetCurrentUser returns current user info
func (h *Handler) GetCurrentUser(c *gin.Context) {
	// Extract user ID from JWT context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}

	// Fetch user from database
	user, err := h.service.userRepo.GetByID(userID.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get user"})
		return
	}

	if user == nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	c.JSON(200, gin.H{
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



// GoogleOAuth initiates Google OAuth flow
func (h *Handler) GoogleOAuth(c *gin.Context) {
	// Generate state parameter for CSRF protection
	state, err := h.service.TokenManager.GenerateToken("oauth-state", time.Minute*10)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to initiate OAuth"})
		return
	}

	// For now, return a placeholder OAuth URL
	// In production, this should use proper OAuth configuration
	authURL := "https://accounts.google.com/o/oauth2/auth" +
		"?client_id=your-client-id" +
		"&redirect_uri=http://localhost:3000/api/auth/callback/google" +
		"&scope=openid%20email%20profile" +
		"&response_type=code" +
		"&state=" + state

	c.JSON(200, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GoogleCallback handles Google OAuth callback
func (h *Handler) GoogleCallback(c *gin.Context) {
	// Get code and state from query parameters
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" {
		c.JSON(400, gin.H{"error": "Authorization code not provided"})
		return
	}

	if state == "" {
		c.JSON(400, gin.H{"error": "State parameter not provided"})
		return
	}

	// Validate state parameter
	_, err := h.service.TokenManager.ValidateToken(state)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid state parameter"})
		return
	}

	// For now, create a placeholder user based on the code
	// In production, this should exchange the code for tokens and get user info from OAuth provider
	userEmail := "oauth-user-" + code[:8] + "@example.com"
	userName := "OAuth User"

	// Check if user exists, create if not
	existingUser, err := h.service.userRepo.GetByEmail(userEmail)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	var currentUser *user.User
	if existingUser == nil {
		// Create new user
		newUser := &user.User{
			Email:         userEmail,
			Name:          userName,
			IsActive:      true,
			EmailVerified: true,
		}
		if err := h.service.userRepo.Create(newUser); err != nil {
			c.JSON(500, gin.H{"error": "Failed to create user"})
			return
		}
		currentUser = newUser
	} else {
		currentUser = existingUser
	}

	// Generate JWT tokens
	accessToken, err := h.service.TokenManager.GenerateToken(currentUser.ID, time.Hour)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := h.service.TokenManager.GenerateRefreshToken(currentUser.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(200, gin.H{
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

// ValidateToken validates an access token
func (h *Handler) ValidateToken(c *gin.Context) {
	type ValidateRequest struct {
		Token string `json:"token" binding:"required"`
	}

	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "token is required"})
		return
	}

	// Use service method to validate token
	userID, err := h.service.ValidateToken(req.Token)
	if err != nil {
		c.JSON(401, gin.H{
			"valid": false,
			"error": "Invalid or expired token",
		})
		return
	}

	// Get user details
	user, err := h.service.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get user details"})
		return
	}

	c.JSON(200, gin.H{
		"valid":      true,
		"user_id":    user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"is_active":  user.IsActive,
		"roles":      user.Roles,
		"expires_at": time.Now().Add(time.Hour).Unix(), // Approximate expiry
	})
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	type RegisterRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Email and password are required"})
		return
	}
	if !isValidEmail(req.Email) {
		c.JSON(400, gin.H{"error": "Invalid email format"})
		return
	}
	if len(req.Password) < 8 {
		c.JSON(400, gin.H{"error": "Password must be at least 8 characters"})
		return
	}

	// Check if user already exists
	existingUser, err := h.service.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if existingUser != nil {
		c.JSON(409, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}
	
	// Create user
	newUser := &user.User{
		Email:    req.Email,
		Name:     "", // Optionally parse from request
		IsActive: true,
	}
	newUser.PasswordHash = string(hash)
	if err := h.service.userRepo.Create(newUser); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}
	
	// Generate tokens
	accessToken, err := h.service.TokenManager.GenerateToken(newUser.ID, time.Hour)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate access token"})
		return
	}
	refreshToken, err := h.service.TokenManager.GenerateRefreshToken(newUser.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	c.JSON(201, gin.H{
		"message":       "Registration successful",
		"user_id":       newUser.ID,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// ChangePassword handles password change requests
func (h *Handler) ChangePassword(c *gin.Context) {
	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Current password and new password are required"})
		return
	}

	// Extract user ID from JWT context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}

	// Validate new password strength
	if len(req.NewPassword) < 8 {
		c.JSON(400, gin.H{"error": "New password must be at least 8 characters long"})
		return
	}

	// Get user from database
	userEntity, err := h.service.userRepo.GetByID(userID.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	if userEntity == nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Validate current password
	err = bcrypt.CompareHashAndPassword([]byte(userEntity.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		c.JSON(401, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to process new password"})
		return
	}

	// Update password in database
	updates := map[string]interface{}{
		"password_hash": string(newPasswordHash),
		"updated_at":    time.Now(),
	}

	_, err = h.service.userRepo.Update(userID.(string), updates)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Password changed successfully",
	})
}

// isValidEmail checks if the email has a basic valid format
func isValidEmail(email string) bool {
	// Simple regex for demonstration (not RFC compliant)
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	at := strings.Index(email, "@")
	if at < 1 || at == len(email)-1 {
		return false
	}
	return true
}
