package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	appAuth "github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/database"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/observability"
	"github.com/EliasRanz/ai-code-gen/internal/interfaces/http"
	"github.com/EliasRanz/ai-code-gen/internal/service"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Override port for this service
	cfg.Server.Port = cfg.AuthService.Port

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Configuration validation failed")
	}

	// Initialize database connection
	db, err := database.NewDBConnection(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	userRepo, err := database.NewAuthUserRepository(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create user repository")
	}

	sessionRepo, err := database.NewPostgreSQLSessionRepository(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create session repository")
	}

	// Create token provider and password hasher
	tokenProvider := appAuth.NewJWTTokenProvider(cfg.Auth.JWTSecret, "ai-ui-generator")
	passwordHasher := appAuth.NewBCryptPasswordHasher()

	// Initialize use cases directly with infrastructure implementations
	loginUC := appAuth.NewLoginUseCase(userRepo, sessionRepo, passwordHasher, tokenProvider)
	logoutUC := appAuth.NewLogoutUseCase(sessionRepo)
	refreshTokenUC := appAuth.NewRefreshTokenUseCase(sessionRepo, tokenProvider, userRepo)

	// Initialize new centralized auth use cases
	validateTokenUC := appAuth.NewValidateToken(tokenProvider, userRepo)
	checkRoleUC := appAuth.NewCheckRole(userRepo)
	getSessionUC := appAuth.NewGetSession(sessionRepo, tokenProvider)
	getUserContextUC := appAuth.NewGetUserContextUseCase(userRepo)

	// Initialize logger
	logger := observability.NewLogger("info", "console")

	// Initialize handler with all use cases
	authHandler := http.NewAuthHandler(
		loginUC,
		logoutUC,
		refreshTokenUC,
		validateTokenUC,
		checkRoleUC,
		getSessionUC,
		getUserContextUC,
		logger,
	)

	// Setup router with auth endpoints
	router := setupAuthRouter(cfg, authHandler)

	// Setup HTTP server
	svc := service.New("auth-service", "1.0.0", cfg)
	svc.SetupHTTPServer(router)

	// Start service
	if err := svc.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start auth service")
	}
}

// setupAuthRouter configures all authentication routes
func setupAuthRouter(cfg *config.Config, handler *http.AuthHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "auth-service"})
	})

	// Register all authentication routes using the handler
	apiGroup := router.Group("/api")
	handler.RegisterRoutes(apiGroup)

	return router
}
