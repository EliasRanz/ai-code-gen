package user

// GetGRPCClient returns the global gRPC client
func GetGRPCClient() UserGRPCClient {
	return grpcClient
}
