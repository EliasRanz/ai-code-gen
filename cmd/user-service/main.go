package main

import (
	"context"
	"fmt"
	"net"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	http_iface "github.com/EliasRanz/ai-code-gen/internal/interfaces/http"
	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/EliasRanz/ai-code-gen/internal/utilities/database"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// Dummy notification service
type dummyNotifier struct{}

func (n *dummyNotifier) NotifyUserCreated(ctx context.Context, user *user.User) error { return nil }
func (n *dummyNotifier) NotifyUserUpdated(ctx context.Context, user *user.User) error { return nil }
func (n *dummyNotifier) NotifyUserDeleted(ctx context.Context, userID utilities.UserID) error {
	return nil
}

// userValidator implements the user.Validator interface using observability validation
type userValidator struct {
	validator *observability.Validator
}

func newUserValidator() *userValidator {
	return &userValidator{
		validator: observability.NewValidator(),
	}
}

func (v *userValidator) ValidateStruct(s interface{}) error {
	return v.validator.ValidateStruct(s)
}

func (v *userValidator) ValidateUser(user *user.User) error {
	return v.validator.ValidateStruct(user)
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
	svc := utilities.New("user-service", "1.0.0", cfg)

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
	userRepo, err := user.NewPostgreSQLUserRepository(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create user repository")
	}

	// Initialize validator and notifier
	validator := newUserValidator()
	notifier := &dummyNotifier{}

	// Initialize user service (business logic services)
	createUserUC := user.NewUserCreator(userRepo, validator, notifier)
	getUserUC := user.NewUserRetriever(userRepo)
	updateUserUC := user.NewUserUpdater(userRepo, validator, notifier)
	deleteUserUC := user.NewUserDeleter(userRepo, notifier)
	listUsersUC := user.NewUserLister(userRepo)

	// Initialize gRPC server
	userGRPCServer := user.NewUserGRPCServer(
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
func startGRPCServer(cfg *config.Config, userServer *user.UserGRPCServer) error {
	grpcAddr := fmt.Sprintf(":%d", cfg.UserService.Port)
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	grpcServer := grpc.NewServer()
	if err := userServer.RegisterService(grpcServer); err != nil {
		return fmt.Errorf("failed to register user service: %w", err)
	}

	log.Info().Msgf("gRPC server listening on %s", grpcAddr)
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}
