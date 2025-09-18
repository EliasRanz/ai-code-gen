# Performance Guidelines & Optimization

## Purpose
Performance optimization guidelines and testing requirements for the AI UI Generator microservices project, covering database optimization, caching strategies, microservice performance, and systematic performance testing.

## Performance Testing & Requirements

### Local Performance Testing
- **Controlled Environment**: Use testcontainers for isolated, reproducible performance tests
- **Baseline Metrics**: Establish performance baselines for all critical paths
- **Impact Assessment**: Test performance impact of every significant change before merge
- **Automated Benchmarks**: Include Go benchmark tests (`go test -bench`) for critical functions
- **Load Testing**: Use controlled load testing to identify bottlenecks and capacity limits

### Performance Testing Tools & Commands
- **Local Benchmarks**: `make perf-test` - Run performance benchmarks locally
- **Benchmark Tests**: `make perf-benchmark` - Go benchmark tests for critical functions
- **Load Testing**: `make perf-load` - Controlled load testing scenarios
- **Stress Testing**: `make perf-stress` - System limits and breaking points
- **Testcontainers**: Isolated database and service instances for consistent test environments

### Performance Requirements
- **Baseline Metrics**: Establish performance baselines for all critical paths
- **Regression Testing**: Fail builds if performance degrades > 20%
- **Load Testing**: Controlled load testing with realistic user scenarios
- **Resource Monitoring**: Track memory, CPU, and database connection usage
- **AI/LLM Optimization**: Specific benchmarks for AI generation performance

**Example Benchmark Test**:
```go
func BenchmarkUserService_CreateUser(b *testing.B) {
    service := setupBenchmarkUserService()
    user := &User{Email: "test@example.com"}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := service.CreateUser(context.Background(), user)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Database Performance Optimization

### Query Optimization
- **Indexing Strategy**: Use proper indexing for frequently queried columns
- **Query Analysis**: Analyze query execution plans and optimize slow queries
- **Batch Operations**: Use batch operations for bulk data processing
- **Pagination**: Implement efficient pagination for large result sets
- **Query Caching**: Cache frequently executed queries when appropriate

**Examples**:
✅ **Good Database Query**:
```go
// Using proper indexing and parameterized queries
func GetUsersByStatus(ctx context.Context, status string, limit, offset int) ([]*User, error) {
    query := `
        SELECT id, email, status, created_at 
        FROM users 
        WHERE status = $1 
        ORDER BY created_at DESC 
        LIMIT $2 OFFSET $3
    `
    rows, err := db.QueryContext(ctx, query, status, limit, offset)
    // ... handle results
}
```

❌ **Bad Database Query**:
```go
// Inefficient query without proper indexing or limits
func GetAllUsers(ctx context.Context) ([]*User, error) {
    query := "SELECT * FROM users"  // No WHERE clause, no LIMIT
    rows, err := db.QueryContext(ctx, query)
    // ... could return millions of records
}
```

### Connection Management
- **Connection Pooling**: Implement proper database connection pooling
- **Connection Limits**: Set appropriate connection limits based on load
- **Connection Timeout**: Configure appropriate connection timeouts
- **Health Checks**: Regular database health checks and monitoring

### Database-Specific Optimizations
- **PostgreSQL**: Utilize PostgreSQL-specific features like JSONB, arrays, and advanced indexing
- **Transactions**: Use transactions appropriately, keep them short
- **Bulk Operations**: Use COPY for large data imports
- **Vacuum**: Regular maintenance and vacuum operations

## Caching Strategy

### Redis Caching Implementation
- **Strategic Caching**: Cache frequently accessed but rarely changing data
- **Cache Keys**: Use consistent, hierarchical cache key naming
- **TTL Management**: Set appropriate time-to-live (TTL) for cached data
- **Cache Invalidation**: Implement proper cache invalidation strategies
- **Cache-Aside Pattern**: Use cache-aside pattern for most operations

**Examples**:
✅ **Good Caching Strategy**:
```go
func GetUser(ctx context.Context, userID string) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", userID)
    
    // Try cache first
    if cached, err := redis.Get(ctx, cacheKey).Result(); err == nil {
        var user User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            return &user, nil
        }
    }
    
    // Cache miss - get from database
    user, err := db.GetUser(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    if data, err := json.Marshal(user); err == nil {
        redis.Set(ctx, cacheKey, data, 15*time.Minute)
    }
    
    return user, nil
}
```

### Cache Performance Patterns
- **Read-Through**: Cache automatically loads data on cache miss
- **Write-Behind**: Asynchronously write to database after updating cache
- **Cache Warming**: Pre-populate cache with frequently accessed data
- **Multi-Level Caching**: Use multiple cache layers for optimal performance

## Microservices Performance

### Inter-Service Communication
- **gRPC Optimization**: Use gRPC connection pooling and keep-alive settings
- **Circuit Breakers**: Implement circuit breakers for external service calls
- **Timeout Management**: Set appropriate timeouts for service calls
- **Retry Strategies**: Implement exponential backoff for retries
- **Load Balancing**: Use appropriate load balancing strategies

**Example gRPC Optimization**:
```go
func NewGRPCClient(address string) (*grpc.ClientConn, error) {
    opts := []grpc.DialOption{
        grpc.WithInsecure(),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:                10 * time.Second,
            Timeout:             3 * time.Second,
            PermitWithoutStream: true,
        }),
        grpc.WithDefaultCallOptions(
            grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB
            grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB
        ),
    }
    return grpc.Dial(address, opts...)
}
```

### Memory Management
- **Go Memory Patterns**: Follow Go memory management best practices
- **Goroutine Management**: Proper goroutine lifecycle management
- **Memory Pooling**: Use object pooling for frequently allocated objects
- **Garbage Collection**: Monitor and optimize garbage collection performance
- **Memory Leaks**: Regular memory leak detection and prevention

### Concurrency Optimization
- **Goroutine Usage**: Use goroutines efficiently with proper synchronization
- **Channel Patterns**: Use appropriate channel patterns for communication
- **Worker Pools**: Implement worker pools for CPU-intensive tasks
- **Context Cancellation**: Proper context usage for cancellation and timeouts

## AI/LLM Integration Performance

### LLM API Optimization
- **Connection Pooling**: Reuse HTTP connections for LLM API calls
- **Request Batching**: Batch multiple requests when possible
- **Streaming Responses**: Use streaming for real-time user experience
- **Caching**: Cache LLM responses for identical prompts
- **Rate Limiting**: Implement proper rate limiting for LLM APIs

**Example LLM Client Optimization**:
```go
type LLMClient struct {
    httpClient *http.Client
    cache      *redis.Client
}

func NewLLMClient() *LLMClient {
    return &LLMClient{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        cache: redis.NewClient(&redis.Options{...}),
    }
}
```

### Timeout & Retry Strategies
- **Request Timeouts**: Set appropriate timeouts for LLM API calls
- **Progressive Timeouts**: Increase timeouts for longer operations
- **Exponential Backoff**: Implement exponential backoff for retries
- **Circuit Breakers**: Use circuit breakers for LLM service failures

## Monitoring & Observability

### Performance Metrics
- **Response Times**: Track API response times and percentiles
- **Throughput**: Monitor requests per second and concurrent users
- **Error Rates**: Track error rates and failure patterns
- **Resource Usage**: Monitor CPU, memory, and network usage
- **Database Performance**: Track query times and connection usage

### Alerting & Dashboards
- **Performance Alerts**: Set up alerts for performance degradation
- **Real-time Dashboards**: Create dashboards for performance monitoring
- **Capacity Planning**: Use metrics for capacity planning and scaling
- **SLA Monitoring**: Monitor against service level agreements

### Integration with Development Workflow
- **Pre-commit**: Run quick performance smoke tests locally
- **CI/CD Pipeline**: Include performance regression tests in automated builds
- **Local Development**: Easy access to performance testing via `make perf-*` commands
- **Production Monitoring**: Set up performance alerts and dashboards for production insights
- **Capacity Planning**: Use performance test results to inform scaling decisions

## Performance Best Practices

### General Guidelines
- **Measure First**: Always measure before optimizing
- **Profile Regularly**: Use profiling tools to identify bottlenecks
- **Optimize Hotpaths**: Focus optimization efforts on critical paths
- **Load Testing**: Regular load testing with realistic scenarios
- **Documentation**: Document performance characteristics and decisions

### Go-Specific Performance
- **Profiling Tools**: Use `go tool pprof` for performance analysis
- **Benchmark Tests**: Write comprehensive benchmark tests
- **Memory Allocation**: Minimize memory allocations in hot paths
- **Interface Usage**: Be mindful of interface boxing overhead
- **String Operations**: Use `strings.Builder` for string concatenation

### Frontend Performance
- **Code Splitting**: Implement proper code splitting for React components
- **Lazy Loading**: Use lazy loading for non-critical components
- **Bundle Optimization**: Optimize webpack bundles for production
- **Caching**: Implement proper browser caching strategies