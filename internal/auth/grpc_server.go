package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/EliasRanz/ai-code-gen/api/proto/auth"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// AuthGRPCServer implements the AuthService gRPC interface and utilities.AuthGRPCService
type AuthGRPCServer struct {
	auth.UnimplementedAuthServiceServer
	loginUC       *LoginUseCase
	logoutUC      *LogoutUseCase
	validateTokUC *ValidateTokenService
	refreshTokUC  *RefreshTokenUseCase
	serviceInfo   utilities.ServiceInfo
}

// NewAuthGRPCServer creates a new gRPC server instance
func NewAuthGRPCServer(
	loginUC *LoginUseCase,
	logoutUC *LogoutUseCase,
	validateTokUC *ValidateTokenService,
	refreshTokUC *RefreshTokenUseCase,
) *AuthGRPCServer {
	return &AuthGRPCServer{
		loginUC:       loginUC,
		logoutUC:      logoutUC,
		validateTokUC: validateTokUC,
		refreshTokUC:  refreshTokUC,
		serviceInfo: utilities.ServiceInfo{
			Name:        "auth-service",
			Version:     "1.0.0",
			Description: "Authentication and authorization gRPC service",
			Endpoints: []string{
				"Login", "Logout", "ValidateToken", "RefreshToken",
			},
		},
	}
}

// RegisterService registers the service with gRPC server
func (s *AuthGRPCServer) RegisterService(server *grpc.Server) error {
	auth.RegisterAuthServiceServer(server, s)
	log.Info().
		Str("service", s.serviceInfo.Name).
		Str("version", s.serviceInfo.Version).
		Msg("Auth gRPC service registered")
	return nil
}

// GetServiceInfo returns service metadata
func (s *AuthGRPCServer) GetServiceInfo() utilities.ServiceInfo {
	return s.serviceInfo
}

// GetInterceptors returns list of interceptors for this service
func (s *AuthGRPCServer) GetInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		s.loggingInterceptor(),
		s.validationInterceptor(),
	}
}

// HealthCheck validates service health
func (s *AuthGRPCServer) HealthCheck(ctx context.Context) error {
	if s.loginUC == nil || s.validateTokUC == nil {
		return fmt.Errorf("auth service dependencies not initialized")
	}
	return nil
}

// ValidateService validates service configuration
func (s *AuthGRPCServer) ValidateService() error {
	if s.loginUC == nil || s.logoutUC == nil ||
		s.validateTokUC == nil || s.refreshTokUC == nil {
		return fmt.Errorf("one or more auth service dependencies not initialized")
	}
	return nil
}

// Start initializes the service
func (s *AuthGRPCServer) Start(ctx context.Context) error {
	return s.ValidateService()
}

// Stop gracefully shuts down the service
func (s *AuthGRPCServer) Stop(ctx context.Context) error {
	log.Info().Msg("Auth gRPC service stopping")
	return nil
}

// loggingInterceptor provides request/response logging
func (s *AuthGRPCServer) loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		log.Info().
			Str("method", info.FullMethod).
			Msg("gRPC request received")

		resp, err := handler(ctx, req)
		if err != nil {
			log.Error().Err(err).
				Str("method", info.FullMethod).
				Msg("gRPC request failed")
		}
		return resp, err
	}
}

// validationInterceptor provides request validation
func (s *AuthGRPCServer) validationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		// Basic validation can be added here
		return handler(ctx, req)
	}
}

// Login authenticates a user
func (s *AuthGRPCServer) Login(ctx context.Context,
	req *auth.LoginRequest) (*auth.LoginResponse, error) {
	log.Info().
		Str("email", req.Email).
		Msg("gRPC Login called")

	// Validate request
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	// Authenticate user
	loginReq := LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := s.loginUC.Execute(ctx, loginReq)
	if err != nil {
		log.Error().Err(err).Msg("Login failed")
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return &auth.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    int64(time.Until(resp.ExpiresAt).Seconds()),
	}, nil
}

// Logout invalidates user session
func (s *AuthGRPCServer) Logout(ctx context.Context,
	req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	log.Info().Msg("gRPC Logout called")

	// Validate request
	if req.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Logout user
	logoutReq := LogoutRequest{AccessToken: req.Token}
	_, err := s.logoutUC.Execute(ctx, logoutReq)
	if err != nil {
		log.Error().Err(err).Msg("Logout failed")
		return nil, fmt.Errorf("logout failed: %w", err)
	}

	return &auth.LogoutResponse{Success: true}, nil
}

// ValidateToken validates JWT token
func (s *AuthGRPCServer) ValidateToken(ctx context.Context,
	req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	log.Debug().Msg("gRPC ValidateToken called")

	// Validate request
	if req.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Validate token
	validateReq := ValidateTokenRequest{AccessToken: req.Token}
	resp, err := s.validateTokUC.Execute(ctx, validateReq)
	if err != nil {
		log.Warn().Err(err).Msg("Token validation failed")
		return &auth.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	userID := ""
	if resp.UserContext != nil {
		userID = string(resp.UserContext.UserID)
	}

	return &auth.ValidateTokenResponse{
		Valid:  resp.Valid,
		UserId: userID,
	}, nil
}

// RefreshToken refreshes JWT access token
func (s *AuthGRPCServer) RefreshToken(ctx context.Context,
	req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	log.Info().Msg("gRPC RefreshToken called")

	// Validate request
	if req.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	// Refresh token
	refreshReq := RefreshTokenRequest{RefreshToken: req.RefreshToken}
	resp, err := s.refreshTokUC.Execute(ctx, refreshReq)
	if err != nil {
		log.Error().Err(err).Msg("Token refresh failed")
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	return &auth.RefreshTokenResponse{
		AccessToken: resp.AccessToken,
		ExpiresIn:   int64(time.Until(resp.ExpiresAt).Seconds()),
	}, nil
}
