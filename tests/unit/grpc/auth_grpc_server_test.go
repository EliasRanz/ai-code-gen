package grpc_test

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestAuthGRPCServer_ServiceInfo(t *testing.T) {
	server := auth.NewAuthGRPCServer(nil, nil, nil, nil)

	// Test that service info is properly set
	info := server.GetServiceInfo()
	assert.Equal(t, "auth-service", info.Name)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "Authentication and authorization gRPC service", info.Description)
	assert.Contains(t, info.Endpoints, "Login")
	assert.Contains(t, info.Endpoints, "Logout")
	assert.Contains(t, info.Endpoints, "ValidateToken")
	assert.Contains(t, info.Endpoints, "RefreshToken")
}

func TestAuthGRPCServer_HealthCheck(t *testing.T) {
	server := auth.NewAuthGRPCServer(nil, nil, nil, nil)

	ctx := context.Background()
	err := server.HealthCheck(ctx)
	assert.Error(t, err, "Should fail with nil dependencies")
	assert.Contains(t, err.Error(), "auth service dependencies not initialized")
}

func TestAuthGRPCServer_ValidateService(t *testing.T) {
	// Test with nil dependencies
	server := auth.NewAuthGRPCServer(nil, nil, nil, nil)
	err := server.ValidateService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestAuthGRPCServer_GetInterceptors(t *testing.T) {
	server := auth.NewAuthGRPCServer(nil, nil, nil, nil)

	interceptors := server.GetInterceptors()
	assert.Len(t, interceptors, 2, "Should have logging and validation interceptors")
}

func TestAuthGRPCServer_RegisterService(t *testing.T) {
	server := auth.NewAuthGRPCServer(nil, nil, nil, nil)
	grpcServer := grpc.NewServer()

	err := server.RegisterService(grpcServer)
	assert.NoError(t, err, "Service registration should succeed")
}

func TestAuthGRPCServer_Lifecycle(t *testing.T) {
	server := auth.NewAuthGRPCServer(nil, nil, nil, nil)
	ctx := context.Background()

	// Test Stop (should always succeed)
	err := server.Stop(ctx)
	assert.NoError(t, err)
}
