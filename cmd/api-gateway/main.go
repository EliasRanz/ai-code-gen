package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/middleware"
	"github.com/EliasRanz/ai-code-gen/internal/proxy"
	"github.com/EliasRanz/ai-code-gen/internal/service"
)

// Metrics tracking
var (
	requestCount int64
	startTime    = time.Now()
)

const (
	serviceName = "api-gateway"
	version     = "1.0.0"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Override port for this service
	cfg.Server.Port = cfg.APIGateway.Port

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Configuration validation failed")
	}

	// Create service instance
	svc := service.New(serviceName, version, cfg)

	// Initialize observability
	if err := svc.Initialize(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize service")
	}

	// Setup router
	router := setupRouter(cfg)

	// Setup HTTP server
	svc.SetupHTTPServer(router)

	// Start service
	if err := svc.Start(); err != nil {
		log.Fatal().Err(err).Msg("Service failed to start")
	}
}

// setupRouter configures all routes and middleware for the API Gateway
func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Initialize auth cache for middleware
	authCache, err := cache.NewAuthCache(
		fmt.Sprintf("redis://%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		5*time.Minute, // 5 minute TTL for cached auth results
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize auth cache")
	}

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.TracingMiddleware(serviceName))
	router.Use(middleware.MetricsMiddleware())

	// Add request counting middleware
	router.Use(func(c *gin.Context) {
		incrementRequestCount()
		c.Next()
	})
	router.Use(middleware.ErrorHandler())

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:3001"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Rate limiting (100 requests per minute per IP)
	router.Use(middleware.CreateRateLimitMiddleware(100, 10))

	// Health and metrics endpoints
	router.GET("/health", healthHandler)
	router.GET("/metrics", metricsHandler)

	// Service configurations
	authService := proxy.ServiceConfig{
		Name:       "auth-service",
		BaseURL:    fmt.Sprintf("http://localhost:%d", cfg.AuthService.Port),
		HealthPath: "/health",
	}

	userService := proxy.ServiceConfig{
		Name:       "user-service",
		BaseURL:    fmt.Sprintf("http://localhost:%d", cfg.UserService.Port),
		HealthPath: "/health",
	}

	aiService := proxy.ServiceConfig{
		Name:       "ai-service",
		BaseURL:    fmt.Sprintf("http://localhost:%d", cfg.AIService.Port),
		HealthPath: "/health",
	}

	// Service health checks
	healthGroup := router.Group("/health")
	{
		healthGroup.GET("/auth", proxy.HealthCheck(authService))
		healthGroup.GET("/users", proxy.HealthCheck(userService))
		healthGroup.GET("/ai", proxy.HealthCheck(aiService))
	}

	// API routes with authentication
	api := router.Group("/api")
	{
		// Public authentication routes
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", proxy.ReverseProxy(authService))
			authGroup.POST("/callback", proxy.ReverseProxy(authService))
			authGroup.POST("/refresh", proxy.ReverseProxy(authService))
			authGroup.POST("/logout", proxy.ReverseProxy(authService))
		}

		// Protected user routes
		userGroup := api.Group("/users")
		userGroup.Use(middleware.AuthServiceProxy(authService.BaseURL, authCache))
		{
			userGroup.GET("/profile", proxy.ReverseProxy(userService))
			userGroup.PUT("/profile", proxy.ReverseProxy(userService))
			userGroup.GET("/projects", proxy.ReverseProxy(userService))
			userGroup.POST("/projects", proxy.ReverseProxy(userService))
			userGroup.GET("/projects/:id", proxy.ReverseProxy(userService))
			userGroup.PUT("/projects/:id", proxy.ReverseProxy(userService))
			userGroup.DELETE("/projects/:id", proxy.ReverseProxy(userService))
		}

		// Protected AI generation routes
		generateGroup := api.Group("/generate")
		generateGroup.Use(middleware.AuthServiceProxy(authService.BaseURL, authCache))
		{
			generateGroup.POST("/ui", proxy.ReverseProxy(aiService))
			generateGroup.POST("/component", proxy.ReverseProxy(aiService))
			generateGroup.GET("/templates", proxy.ReverseProxy(aiService))
			generateGroup.POST("/analyze", proxy.ReverseProxy(aiService))
		}

		// Admin routes
		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.AuthServiceRoleProxy(authService.BaseURL, authCache, "admin"))
		{
			adminGroup.GET("/users", proxy.ReverseProxy(userService))
			adminGroup.GET("/projects", proxy.ReverseProxy(userService))
			adminGroup.GET("/metrics", metricsHandler)
		}
	}

	return router
}

// healthHandler provides overall gateway health status
func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": serviceName,
		"version": version,
	})
}

// metricsHandler provides basic metrics (placeholder)
func metricsHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"service": serviceName,
		"metrics": gin.H{
			"requests_total":     getRequestCount(),
			"request_duration":   getAverageRequestDuration(),
			"active_connections": getActiveConnections(),
			"uptime_seconds":     getUptimeSeconds(),
		},
	})
}

// getRequestCount returns the total number of requests processed
func getRequestCount() int64 {
	return atomic.LoadInt64(&requestCount)
}

// getAverageRequestDuration returns a placeholder for request duration
func getAverageRequestDuration() string {
	return "~50ms" // Placeholder - would need actual request timing
}

// getActiveConnections returns a placeholder for active connections
func getActiveConnections() int {
	return 42 // Placeholder - would need actual connection tracking
}

// getUptimeSeconds returns the service uptime in seconds
func getUptimeSeconds() int64 {
	return int64(time.Since(startTime).Seconds())
}

// incrementRequestCount increments the request counter
func incrementRequestCount() {
	atomic.AddInt64(&requestCount, 1)
}
