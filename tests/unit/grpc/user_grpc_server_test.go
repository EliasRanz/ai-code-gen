package grpc_test

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/user"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestUserGRPCServer_ServiceInfo(t *testing.T) {
	server := user.NewUserGRPCServer(nil, nil, nil, nil, nil)

	// Test that service info is properly set
	info := server.GetServiceInfo()
	assert.Equal(t, "user-service", info.Name)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "User management gRPC service", info.Description)
	assert.Contains(t, info.Endpoints, "GetUser")
	assert.Contains(t, info.Endpoints, "CreateUser")
	assert.Contains(t, info.Endpoints, "UpdateUser")
	assert.Contains(t, info.Endpoints, "DeleteUser")
	assert.Contains(t, info.Endpoints, "ListUsers")
}

func TestUserGRPCServer_HealthCheck(t *testing.T) {
	server := user.NewUserGRPCServer(nil, nil, nil, nil, nil)

	ctx := context.Background()
	err := server.HealthCheck(ctx)
	assert.Error(t, err, "Should fail with nil dependencies")
	assert.Contains(t, err.Error(), "user retriever not initialized")
}

func TestUserGRPCServer_ValidateService(t *testing.T) {
	// Test with nil dependencies
	server := user.NewUserGRPCServer(nil, nil, nil, nil, nil)
	err := server.ValidateService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestUserGRPCServer_GetInterceptors(t *testing.T) {
	server := user.NewUserGRPCServer(nil, nil, nil, nil, nil)

	interceptors := server.GetInterceptors()
	assert.Len(t, interceptors, 2, "Should have logging and validation interceptors")
}

func TestUserGRPCServer_RegisterService(t *testing.T) {
	server := user.NewUserGRPCServer(nil, nil, nil, nil, nil)
	grpcServer := grpc.NewServer()

	err := server.RegisterService(grpcServer)
	assert.NoError(t, err, "Service registration should succeed")
}

func TestUserGRPCServer_Lifecycle(t *testing.T) {
	server := user.NewUserGRPCServer(nil, nil, nil, nil, nil)
	ctx := context.Background()

	// Test Stop (should always succeed)
	err := server.Stop(ctx)
	assert.NoError(t, err)
}
