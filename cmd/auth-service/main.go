package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	appAuth "github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/interfaces/http"
	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/EliasRanz/ai-code-gen/internal/utilities/database"
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

	userRepo, err := appAuth.NewAuthUserRepository(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create user repository")
	}

	sessionRepo, err := appAuth.NewPostgreSQLSessionRepository(db)
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

	// Initialize gRPC server
	authGRPCServer := appAuth.NewAuthGRPCServer(
		loginUC,
		logoutUC,
		validateTokenUC,
		refreshTokenUC,
	)

	// Setup router with auth endpoints
	router := setupAuthRouter(cfg, authHandler)

	// Setup HTTP server
	svc := utilities.New("auth-service", "1.0.0", cfg)
	svc.SetupHTTPServer(router)

	// Start gRPC server in a goroutine
	go func() {
		if err := startGRPCServer(cfg, authGRPCServer); err != nil {
			log.Fatal().Err(err).Msg("Failed to start gRPC server")
		}
	}()

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

// startGRPCServer initializes and starts the gRPC server for the auth service
func startGRPCServer(cfg *config.Config, authServer *appAuth.AuthGRPCServer) error {
	// Use GRPC_PORT if set, otherwise use AuthService.Port + 2
	grpcPort := cfg.AuthService.Port + 2
	if grpcPortEnv := os.Getenv("GRPC_PORT"); grpcPortEnv != "" {
		if port, err := strconv.Atoi(grpcPortEnv); err == nil {
			grpcPort = port
		}
	}

	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	log.Info().Int("port", grpcPort).Msg("Starting gRPC server for auth service")

	// Create listener
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	// Create gRPC server
	server := grpc.NewServer()

	// Register the auth service
	if err := authServer.RegisterService(server); err != nil {
		return fmt.Errorf("failed to register auth service: %w", err)
	}

	log.Info().Str("address", grpcAddr).Msg("Auth gRPC server listening")
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}
