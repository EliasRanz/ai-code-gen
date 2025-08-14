// Package gateway provides consolidated gateway components
package gateway

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// Router wraps gin.Engine with consolidated gateway functionality
type Router struct {
	engine        *gin.Engine
	userHandler   *user.HTTPHandler
	authHandler   *auth.HTTPHandler
	aiHandler     *ai.HTTPHandler
	logger        observability.Logger
	tokenProvider auth.TokenProvider
}

// NewRouter creates a new consolidated gateway router
func NewRouter(
	userHandler *user.HTTPHandler,
	authHandler *auth.HTTPHandler,
	aiHandler *ai.HTTPHandler,
	tokenProvider auth.TokenProvider,
	logger observability.Logger,
) *Router {
	// Set gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	// Add middleware
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())
	engine.Use(loggingMiddleware(logger))

	router := &Router{
		engine:        engine,
		userHandler:   userHandler,
		authHandler:   authHandler,
		aiHandler:     aiHandler,
		tokenProvider: tokenProvider,
		logger:        logger,
	}

	router.setupRoutes()
	return router
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes() {
	// Health check
	r.engine.GET("/health", r.healthCheck)

	// API v1 routes
	v1 := r.engine.Group("/api/v1")

	// Public auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", r.authHandler.Login)
		auth.POST("/refresh", r.authHandler.RefreshToken)
	}

	// Protected routes (require authentication)
	protected := v1.Group("/")
	protected.Use(r.authMiddleware())
	{
		// Auth routes
		protected.POST("/auth/logout", r.authHandler.Logout)
		protected.POST("/auth/validate", r.authHandler.ValidateToken)
		protected.POST("/auth/check-role", r.authHandler.CheckRole)
		protected.GET("/auth/session", r.authHandler.GetSession)
		protected.GET("/auth/user/:id", r.authHandler.GetUserContext)

		// User routes
		users := protected.Group("/users")
		{
			users.POST("", r.userHandler.CreateUser)
			users.GET("", r.userHandler.ListUsers)
			users.GET("/:id", r.userHandler.GetUser)
			users.PUT("/:id", r.userHandler.UpdateUser)
			users.DELETE("/:id", r.userHandler.DeleteUser)
		}

		// AI routes
		ai := protected.Group("/ai")
		{
			ai.POST("/generate", r.aiHandler.GenerateCode)
			ai.POST("/stream", r.aiHandler.StreamCode)
		}
	}
}

// healthCheck returns service health status
func (r *Router) healthCheck(c *gin.Context) {
	// Check handlers health
	errors := []string{}

	if err := r.userHandler.HealthCheck(); err != nil {
		errors = append(errors, "user handler: "+err.Error())
	}

	if err := r.authHandler.HealthCheck(); err != nil {
		errors = append(errors, "auth handler: "+err.Error())
	}

	if err := r.aiHandler.HealthCheck(); err != nil {
		errors = append(errors, "ai handler: "+err.Error())
	}

	status := "healthy"
	statusCode := http.StatusOK
	if len(errors) > 0 {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := gin.H{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"service":   "ai-ui-generator-gateway",
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(statusCode, response)
}

// Engine returns the underlying gin.Engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// corsMiddleware configures CORS
func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger observability.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log request details
		duration := time.Since(start)
		logger.Info("HTTP request", map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
			"duration_ms": duration.Milliseconds(),
			"user_agent":  c.Request.UserAgent(),
		})
	}
}

// authMiddleware validates JWT tokens
func (r *Router) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check for Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract token (remove "Bearer " prefix)
		token := authHeader[7:]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
			c.Abort()
			return
		}

		// Validate token using TokenProvider
		userID, err := r.tokenProvider.ValidateAccessToken(token)
		if err != nil {
			r.logger.Warn("Invalid access token", map[string]interface{}{
				"error": err.Error(),
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Add user ID to context for use in handlers
		c.Set("user_id", userID)
		c.Set("authenticated_user_id", userID) // Alternative key for consistency

		c.Next()
	}
}

// ValidateHandlers validates all registered handlers
func (r *Router) ValidateHandlers() error {
	if err := r.userHandler.ValidateRoutes(); err != nil {
		return utilities.NewValidationError("user handler validation failed", err)
	}

	if err := r.authHandler.ValidateRoutes(); err != nil {
		return utilities.NewValidationError("auth handler validation failed", err)
	}

	if err := r.aiHandler.ValidateRoutes(); err != nil {
		return utilities.NewValidationError("ai handler validation failed", err)
	}

	return nil
}
