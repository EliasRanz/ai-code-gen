package main

import (
	"context"
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	pb "github.com/EliasRanz/ai-code-gen/api/proto/user"
	appUser "github.com/EliasRanz/ai-code-gen/internal/application/user"
	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/domain/common"
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/database"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/observability"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/validation"
	grpc_iface "github.com/EliasRanz/ai-code-gen/internal/interfaces/grpc"
	http_iface "github.com/EliasRanz/ai-code-gen/internal/interfaces/http"
	"github.com/EliasRanz/ai-code-gen/internal/service"
)

// Dummy notification service
type dummyNotifier struct{}

func (n *dummyNotifier) NotifyUserCreated(ctx context.Context, user *user.User) error { return nil }
func (n *dummyNotifier) NotifyUserUpdated(ctx context.Context, user *user.User) error { return nil }
func (n *dummyNotifier) NotifyUserDeleted(ctx context.Context, userID common.UserID) error {
	return nil
}

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Override port for this service
	cfg.Server.Port = cfg.UserService.Port

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Configuration validation failed")
	}

	// Create service
	svc := service.New("user-service", "1.0.0", cfg)

	// Initialize observability
	if err := svc.Initialize(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize service")
	}

	// Initialize database connection
	db, err := database.NewDBConnection(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Initialize repositories
	userRepo, err := database.NewPostgreSQLUserRepository(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create user repository")
	}

	// Initialize validator and notifier
	validator := validation.NewValidator()
	notifier := &dummyNotifier{}

	// Initialize user service (use cases)
	createUserUC := appUser.NewCreateUserUseCase(userRepo, validator, notifier)
	getUserUC := appUser.NewGetUserUseCase(userRepo)
	updateUserUC := appUser.NewUpdateUserUseCase(userRepo, validator, notifier)
	deleteUserUC := appUser.NewDeleteUserUseCase(userRepo, notifier)
	listUsersUC := appUser.NewListUsersUseCase(userRepo)

	// Initialize gRPC server
	userGRPCServer := grpc_iface.NewUserServer(
		createUserUC,
		getUserUC,
		updateUserUC,
		deleteUserUC,
		listUsersUC,
	)

	// Setup HTTP router with user and project endpoints
	logger := observability.NewLogger(cfg.Logging.Level, cfg.Logging.Format)
	userHandler := http_iface.NewUserHandler(
		createUserUC,
		getUserUC,
		updateUserUC,
		listUsersUC,
		deleteUserUC,
		logger,
	)
	router := setupUserRouter(cfg, userHandler)

	// Setup HTTP server
	svc.SetupHTTPServer(router)

	// Start gRPC server in a goroutine
	go func() {
		if err := startGRPCServer(cfg, userGRPCServer); err != nil {
			log.Fatal().Err(err).Msg("Failed to start gRPC server")
		}
	}()

	// Start service
	if err := svc.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start user service")
	}
}

// setupUserRouter configures all user and project routes
func setupUserRouter(cfg *config.Config, handler *http_iface.UserHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})

	// Register all user and project routes using the handler
	apiGroup := router.Group("/api")
	handler.RegisterUserRoutes(apiGroup)

	return router
}

// startGRPCServer initializes and starts the gRPC server for the user service
func startGRPCServer(cfg *config.Config, userServer *grpc_iface.UserServer) error {
	grpcAddr := fmt.Sprintf(":%d", cfg.UserService.Port)
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userServer)

	log.Info().Msgf("gRPC server listening on %s", grpcAddr)
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}
