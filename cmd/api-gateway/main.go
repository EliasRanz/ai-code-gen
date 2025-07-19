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
	"github.com/EliasRanz/ai-code-gen/internal/gateway"
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

	// Create gateway factory with auth service URL and cache
	authServiceURL := fmt.Sprintf("http://localhost:%d", cfg.AuthService.Port)
	factory := gateway.NewMiddlewareFactory(authServiceURL, authCache)

	// Create observable gateway with middleware and observers
	observableGateway := gateway.NewObservableGateway(factory)

	// Setup consolidated middleware using our new gateway
	middlewareConfigs := []gateway.MiddlewareConfig{
		gateway.NewBasicMiddlewareConfig("logging", true, nil),
		gateway.NewBasicMiddlewareConfig("rate-limit", true, map[string]interface{}{
			"requests_per_second": 100,
			"burst":               10,
		}),
		gateway.NewBasicMiddlewareConfig("auth-proxy", true, nil),
	}

	// Setup middleware chain
	if err := observableGateway.SetupMiddleware(middlewareConfigs); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup gateway middleware")
	}

	// Apply consolidated middleware to router
	router.Use(gin.Recovery())
	router.Use(observableGateway.CreateGinMiddleware())

	// Add request counting middleware
	router.Use(func(c *gin.Context) {
		incrementRequestCount()
		c.Next()
	})

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:3001"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Health and metrics endpoints
	router.GET("/health", healthHandler)
	router.GET("/metrics", metricsHandler)

	// Service configurations
	authService := gateway.ServiceConfig{
		Name:       "auth-service",
		BaseURL:    fmt.Sprintf("http://localhost:%d", cfg.AuthService.Port),
		HealthPath: "/health",
	}

	userService := gateway.ServiceConfig{
		Name:       "user-service",
		BaseURL:    fmt.Sprintf("http://localhost:%d", cfg.UserService.Port),
		HealthPath: "/health",
	}

	aiService := gateway.ServiceConfig{
		Name:       "ai-service",
		BaseURL:    fmt.Sprintf("http://localhost:%d", cfg.AIService.Port),
		HealthPath: "/health",
	}

	// Service health checks
	healthGroup := router.Group("/health")
	{
		healthGroup.GET("/auth", gateway.HealthCheckHandler(authService))
		healthGroup.GET("/users", gateway.HealthCheckHandler(userService))
		healthGroup.GET("/ai", gateway.HealthCheckHandler(aiService))
	}

	// API routes with authentication
	api := router.Group("/api")
	{
		// Public authentication routes
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", gateway.ReverseProxy(authService))
			authGroup.POST("/callback", gateway.ReverseProxy(authService))
			authGroup.POST("/refresh", gateway.ReverseProxy(authService))
			authGroup.POST("/logout", gateway.ReverseProxy(authService))
		}

		// Protected user routes (auth handled by global gateway middleware)
		userGroup := api.Group("/users")
		{
			userGroup.GET("/profile", gateway.ReverseProxy(userService))
			userGroup.PUT("/profile", gateway.ReverseProxy(userService))
			userGroup.GET("/projects", gateway.ReverseProxy(userService))
			userGroup.POST("/projects", gateway.ReverseProxy(userService))
			userGroup.GET("/projects/:id", gateway.ReverseProxy(userService))
			userGroup.PUT("/projects/:id", gateway.ReverseProxy(userService))
			userGroup.DELETE("/projects/:id", gateway.ReverseProxy(userService))
		}

		// Protected AI generation routes (auth handled by global gateway middleware)
		generateGroup := api.Group("/generate")
		{
			generateGroup.POST("/ui", gateway.ReverseProxy(aiService))
			generateGroup.POST("/component", gateway.ReverseProxy(aiService))
			generateGroup.GET("/templates", gateway.ReverseProxy(aiService))
			generateGroup.POST("/analyze", gateway.ReverseProxy(aiService))
		}

		// Admin routes (auth handled by global gateway middleware)
		adminGroup := api.Group("/admin")
		{
			adminGroup.GET("/users", gateway.ReverseProxy(userService))
			adminGroup.GET("/projects", gateway.ReverseProxy(userService))
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
