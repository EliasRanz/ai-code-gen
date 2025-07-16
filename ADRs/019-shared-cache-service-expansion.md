# ADR-018: Shared Cache Service Expansion (Phase A.1)

**Date**: 2025-01-12  
**Status**: Accepted  
**Context**: Infrastructure Consolidation (ADR-017 Phase 1)

## Context

As part of the infrastructure consolidation effort (ADR-017), we needed to expand the existing auth cache infrastructure into a comprehensive, shared cache service that supports all microservices while implementing robust design patterns for scalability and reliability.

The existing `internal/cache/auth_cache.go` provided a foundation but lacked:
- Standardized interfaces across cache providers
- Circuit breaker protection for resilience
- Service-specific cache managers
- Consistent error handling and metrics

## Decision

We have implemented a comprehensive cache service expansion with the following design patterns:

### 1. Cache Interface Pattern
- **Core Interface**: `CacheProvider` defining consistent operations across all providers
- **Factory Pattern**: `CacheFactory` for provider-agnostic instantiation
- **Service Managers**: Service-specific cache managers in each microservice package

### 2. Circuit Breaker Pattern
- **Resilience**: Automatic failure detection and recovery for Redis connections
- **Fallback Strategy**: Graceful degradation to in-memory cache when Redis unavailable
- **Configurable Thresholds**: Tunable failure and recovery parameters

### 3. Multi-tier Architecture
- **Primary Provider**: Redis with connection pooling and circuit breaker protection
- **Fallback Provider**: In-memory cache for development and Redis failures
- **Multi Provider**: Combines Redis primary with memory fallback

### 4. Service Ownership Model
Following ADR-017 target architecture:
- **Shared Infrastructure**: `internal/cache/` for Redis client, pooling, interfaces
- **Service-specific Managers**: Each service owns its cache manager:
  - `internal/auth/cache.go` - Authentication context and session caching
  - `internal/user/cache.go` - User profiles, projects, and chat sessions
  - `internal/ai/cache.go` - AI generations, model responses, and rate limiting

## Implementation Details

### Core Cache Infrastructure (`internal/cache/`)
```
├── interfaces.go       # Cache Interface Pattern definitions
├── factory.go         # Factory Pattern implementation
├── circuit_breaker.go # Circuit Breaker Pattern
├── redis_provider.go  # Redis implementation with circuit breaker
├── memory_provider.go # In-memory fallback implementation
├── multi_provider.go  # Multi-tier provider (Redis + Memory)
└── service.go         # Unified cache service entry point
```

### Service-specific Cache Managers
```
├── internal/auth/cache.go    # Auth-specific caching (tokens, sessions)
├── internal/user/cache.go    # User-specific caching (profiles, projects)
└── internal/ai/cache.go      # AI-specific caching (generations, rate limits)
```

### Design Pattern Benefits

**Cache Interface Pattern**:
- Provider-agnostic operations (Redis, memory, future providers)
- Consistent API across all services
- Easy testing with mock implementations
- Future extensibility without service changes

**Circuit Breaker Pattern**:
- Prevents cascade failures when Redis is down
- Automatic recovery with configurable timeouts
- Performance protection during failures
- Detailed failure and recovery metrics

**Factory Pattern**:
- Dynamic provider creation based on configuration
- Centralized provider registration and management
- Extensible for future cache providers

**Service Ownership**:
- Each service owns its complete cache implementation
- Type-safe cache operations with service-specific data structures
- Independent cache configuration and TTL policies
- Clear service boundaries following microservice architecture

## Configuration

Enhanced Redis configuration with circuit breaker and pooling:

```go
type RedisConfig struct {
    // Connection settings
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Password string `json:"password"`
    DB       int    `json:"db"`
    
    // Connection pooling
    MaxConnections     int           `json:"max_connections"`
    MaxIdleConnections int           `json:"max_idle_connections"`
    ConnectionTimeout  time.Duration `json:"connection_timeout"`
    
    // Circuit breaker settings
    FailureThreshold       int           `json:"failure_threshold"`
    RequestVolumeThreshold int           `json:"request_volume_threshold"`
    RecoveryTimeout        time.Duration `json:"recovery_timeout"`
    
    // Service-specific TTL settings
    AuthCacheTTL string `json:"auth_cache_ttl"`
    UserCacheTTL string `json:"user_cache_ttl"`
    AICacheTTL   string `json:"ai_cache_ttl"`
}
```

## Migration Path

1. **Existing auth cache**: Migrated to new `internal/auth/cache.go` using shared infrastructure
2. **Service integration**: Each service creates cache manager using shared `CacheProvider`
3. **Backward compatibility**: Existing auth cache operations maintain same interface
4. **Gradual rollout**: Services can adopt new cache managers incrementally

## Testing Strategy

- **Unit Tests**: 90%+ coverage across all cache components
- **Integration Tests**: Real cache operations with memory provider
- **Performance Tests**: Benchmark cache operations under load
- **Circuit Breaker Tests**: Failure simulation and recovery validation
- **Pattern Tests**: Dedicated tests for each design pattern implementation

## Consequences

### Positive
- **Robust Architecture**: Circuit breaker prevents cascade failures
- **Scalable Design**: Interface patterns enable easy extension
- **Service Autonomy**: Each service owns its cache implementation
- **Development Friendly**: Memory provider works without Redis
- **Production Ready**: Redis with connection pooling and resilience
- **Consistent APIs**: Unified caching interface across all services

### Negative
- **Initial Complexity**: More sophisticated than simple Redis client
- **Learning Curve**: Teams need to understand circuit breaker patterns
- **Configuration Overhead**: More configuration options to manage

### Neutral
- **Migration Required**: Existing services need to adopt new cache managers
- **Monitoring**: Circuit breaker states and cache metrics need observability integration

## Compliance

This implementation follows all coding standards:
- **File Size**: All files under 300 lines (max 500)
- **Function Size**: All functions under 30 lines (max 50)
- **SOLID Principles**: Clear separation of concerns and single responsibility
- **Error Handling**: Explicit error handling with proper error types
- **Testing**: High test coverage with comprehensive validation
- **Input Validation**: All inputs validated with descriptive errors

## Success Criteria

✅ **All Design Patterns Implemented**: Cache Interface, Circuit Breaker, Factory Pattern  
✅ **Service-specific Cache Managers**: Auth, User, AI cache managers completed  
✅ **Redis Connection Pooling**: Centralized Redis configuration with pooling  
✅ **Circuit Breaker Protection**: Automatic failover to memory cache  
✅ **Comprehensive Testing**: 90%+ test coverage achieved  
✅ **Performance Validated**: Benchmark tests pass under load  
✅ **Configuration Enhanced**: Extended Redis config with all required settings  
✅ **Migration Ready**: Backward compatibility with existing auth cache

## Next Steps

1. **Phase A.2**: LLM Functionality Consolidation with rate limiting integration
2. **Phase B**: Interface & Middleware Migration to service packages
3. **Production Deployment**: Redis configuration for production environment
4. **Observability Integration**: Cache metrics and circuit breaker monitoring
