package utilities

import (
	"context"

	"google.golang.org/grpc"
)

// ServiceInfo contains metadata about a gRPC service
type ServiceInfo struct {
	Name        string
	Version     string
	Description string
	Endpoints   []string
}

// InterceptorConfig contains configuration for gRPC interceptors
type InterceptorConfig struct {
	EnableLogging   bool
	EnableMetrics   bool
	EnableTracing   bool
	EnableAuth      bool
	MaxRequestSize  int64
	MaxResponseSize int64
}

// GRPCServiceConfig contains configuration for a gRPC service
type GRPCServiceConfig struct {
	Port               int
	MaxConcurrentRPC   int
	ConnectionTimeout  int
	InterceptorConfig  InterceptorConfig
	HealthCheckEnabled bool
}

// GRPCService is the core interface that all gRPC service implementations must implement
type GRPCService interface {
	// Register service with gRPC server
	RegisterService(server *grpc.Server) error

	// Get service metadata and configuration
	GetServiceInfo() ServiceInfo
	GetInterceptors() []grpc.UnaryServerInterceptor

	// Health and validation
	HealthCheck(ctx context.Context) error
	ValidateService() error

	// Lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// GRPCServiceFactory provides factory pattern for gRPC service instantiation
type GRPCServiceFactory interface {
	CreateService(serviceType string, config GRPCServiceConfig) (GRPCService, error)
	RegisterAllServices(server *grpc.Server) error
	ListAvailableServices() []string
}

// GRPCInterceptor provides interface for consistent middleware
type GRPCInterceptor interface {
	UnaryInterceptor() grpc.UnaryServerInterceptor
	StreamInterceptor() grpc.StreamServerInterceptor
	GetConfig() InterceptorConfig
}

// GRPCServerManager manages gRPC server lifecycle
type GRPCServerManager interface {
	CreateServer(config GRPCServiceConfig) (*grpc.Server, error)
	RegisterService(server *grpc.Server, service GRPCService) error
	StartServer(server *grpc.Server, address string) error
	StopServer(server *grpc.Server) error
}
