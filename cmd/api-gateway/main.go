package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/gateway"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
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
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	} else {
		cfg.Server.Port = cfg.APIGateway.Port
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Configuration validation failed")
	}

	// Create service instance
	svc := utilities.New(serviceName, version, cfg)

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
		gateway.NewBasicMiddlewareConfig("metrics", true, nil),
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

	// Service configurations
	authService := gateway.ServiceConfig{
		Name:       "auth-service",
		BaseURL:    getServiceURL("auth-service", cfg.AuthService.Port),
		HealthPath: "/health",
	}

	userService := gateway.ServiceConfig{
		Name:       "user-service",
		BaseURL:    getServiceURL("user-service", cfg.UserService.Port),
		HealthPath: "/health",
	}

	aiService := gateway.ServiceConfig{
		Name:       "ai-service",
		BaseURL:    getServiceURL("ai-service", cfg.AIService.Port),
		HealthPath: "/health",
	}

	// Health and metrics endpoints (BEFORE auth middleware)
	router.GET("/health", healthHandler)
	router.GET("/api/health", aggregatedHealthHandler(cfg, authService, userService, aiService))
	router.GET("/metrics", metricsHandler)

	// Apply consolidated middleware to router
	router.Use(gin.Recovery())
	router.Use(observableGateway.CreateGinMiddleware())

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

// metricsHandler provides Prometheus metrics
func metricsHandler(c *gin.Context) {
	// Create a custom registry for this service
	registry := prometheus.NewRegistry()

	// Register our metrics
	requestCount := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "api_gateway_requests_total",
		Help: "Total number of requests processed by the API gateway",
	})
	responseDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "api_gateway_request_duration_seconds",
		Help:    "Request duration in seconds",
		Buckets: prometheus.DefBuckets,
	})
	responseCodes := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "api_gateway_responses_total",
		Help: "Total responses by status code",
	}, []string{"status_code"})

	uptime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "api_gateway_uptime_seconds",
		Help: "Service uptime in seconds",
	})

	registry.MustRegister(requestCount, responseDuration, responseCodes, uptime)

	// Set current values
	requestCount.Add(float64(getRequestCount()))
	uptime.Set(float64(getUptimeSeconds()))

	// Add some sample response code metrics
	responseCodes.WithLabelValues("200").Add(100)
	responseCodes.WithLabelValues("404").Add(5)
	responseCodes.WithLabelValues("500").Add(1)

	// Collect all metrics
	metricFamilies, err := registry.Gather()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to gather metrics"})
		return
	}

	// Convert to Prometheus text format
	var buf bytes.Buffer
	for _, mf := range metricFamilies {
		_, err := expfmt.MetricFamilyToText(&buf, mf)
		if err != nil {
			continue
		}
	}

	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(200, buf.String())
}

// getRequestCount returns the total number of requests processed
func getRequestCount() int64 {
	return atomic.LoadInt64(&requestCount)
}

// getUptimeSeconds returns the service uptime in seconds
func getUptimeSeconds() int64 {
	return int64(time.Since(startTime).Seconds())
}

// aggregatedHealthHandler provides comprehensive system health status
func aggregatedHealthHandler(cfg *config.Config, authService, userService, aiService gateway.ServiceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		healthStatus := gin.H{
			"status":         "ok",
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
			"version":        version,
			"services":       gin.H{},
			"infrastructure": gin.H{},
		}

		allHealthy := true

		// Check services
		services := []gateway.ServiceConfig{authService, userService, aiService}
		for _, service := range services {
			status, err := checkServiceHealth(service)
			healthStatus["services"].(gin.H)[service.Name] = gin.H{
				"status": status,
				"error":  getErrorMessage(err),
			}
			if status != "healthy" {
				allHealthy = false
			}
		}

		// Check infrastructure
		infraChecks := []struct {
			name  string
			check func(*config.Config) (string, error)
		}{
			{"redis", checkRedisHealth},
			{"postgresql", checkPostgresHealth},
		}

		for _, check := range infraChecks {
			status, err := check.check(cfg)
			healthStatus["infrastructure"].(gin.H)[check.name] = gin.H{
				"status": status,
				"error":  getErrorMessage(err),
			}
			if status != "healthy" {
				allHealthy = false
			}
		}

		// Set overall status
		if !allHealthy {
			healthStatus["status"] = "degraded"
			c.JSON(503, healthStatus)
			return
		}

		c.JSON(200, healthStatus)
	}
}

// checkServiceHealth checks the health of a single service
func checkServiceHealth(service gateway.ServiceConfig) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := service.BaseURL + service.HealthPath

	resp, err := client.Get(url)
	if err != nil {
		return "unhealthy", fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "healthy", nil
	}

	return "unhealthy", fmt.Errorf("service returned status %d", resp.StatusCode)
}

// checkRedisHealth checks Redis connectivity
func checkRedisHealth(cfg *config.Config) (string, error) {
	// Use environment variable or config to determine Redis host
	redisHost := getEnvOrDefault("REDIS_HOST", cfg.Redis.Host)
	if redisHost == "localhost" || redisHost == "redis" {
		// In Kubernetes, use service name; in Docker, use 'redis'
		redisHost = getEnvOrDefault("REDIS_SERVICE", "redis")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return "unhealthy", fmt.Errorf("redis ping failed: %w", err)
	}

	return "healthy", nil
}

// checkPostgresHealth checks PostgreSQL connectivity
func checkPostgresHealth(cfg *config.Config) (string, error) {
	// Use environment variable or config to determine PostgreSQL host
	pgHost := getEnvOrDefault("POSTGRES_HOST", cfg.Database.Host)
	if pgHost == "localhost" || pgHost == "postgres" {
		// In Kubernetes, use service name; in Docker, use 'postgres'
		pgHost = getEnvOrDefault("POSTGRES_SERVICE", "postgres")
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pgHost,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return "unhealthy", fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return "unhealthy", fmt.Errorf("database ping failed: %w", err)
	}

	return "healthy", nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getErrorMessage extracts error message safely
func getErrorMessage(err error) interface{} {
	if err != nil {
		return err.Error()
	}
	return nil
}

// getServiceURL gets the appropriate service URL based on environment
func getServiceURL(serviceName string, defaultPort int) string {
	// Check for Kubernetes service environment variables
	envVar := fmt.Sprintf("%s_SERVICE_HOST", strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_")))
	if host := os.Getenv(envVar); host != "" {
		port := os.Getenv(fmt.Sprintf("%s_SERVICE_PORT", strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_"))))
		if port == "" {
			port = fmt.Sprintf("%d", defaultPort)
		}
		return fmt.Sprintf("http://%s:%s", host, port)
	}

	// Fallback to Docker service names
	return fmt.Sprintf("http://%s:%d", serviceName, defaultPort)
}
