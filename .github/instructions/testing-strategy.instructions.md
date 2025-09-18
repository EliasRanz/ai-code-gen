# Testing Strategy & CI/CD Requirements

## Purpose
Comprehensive testing guidelines for the AI UI Generator microservices project, covering unit testing, integration testing, performance testing, and CI/CD pipeline requirements.

## Testing Definitions & Scope

### Unit Tests
**Scope**: Test individual functions/methods in isolation with all external dependencies mocked

**Requirements**:
- **Location**: `tests/**` directory structure matching `internal/**`
- **Mocking**: Mock ALL external dependencies (databases, HTTP clients, other services) - no in-memory substitutes
- **Speed**: Fast execution (< 100ms per test), suitable for TDD and rapid feedback
- **Coverage**: Minimum 80% coverage for new code, focus on business logic and edge cases
- **CI/CD**: Must pass in CI pipeline, no external dependencies required
- **Code Design**: Business logic must be designed with proper abstractions to enable complete mocking

**Examples**:
✅ **Good Unit Test**:
```go
func TestUserService_CreateUser(t *testing.T) {
    mockRepo := &MockUserRepository{}
    mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)
    
    service := NewUserService(mockRepo)
    err := service.CreateUser(context.Background(), &User{Email: "test@example.com"})
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

❌ **Bad Unit Test**:
```go
func TestUserService_CreateUser(t *testing.T) {
    // Using real database connection in unit test
    db := setupRealPostgresConnection()
    service := NewUserService(NewUserRepository(db))
    // ...
}
```

### Integration Tests
**Scope**: Test service interactions and data flow between components

**Requirements**:
- **Real Systems**: Use testcontainers for databases (PostgreSQL, Redis)
- **Service Boundaries**: Test actual gRPC communication between services
- **Environment**: Isolated, repeatable environments using Docker containers
- **Data**: Test with realistic data sets and scenarios
- **CI/CD**: Must be fully automated and reproducible in CI environment

**Examples**:
✅ **Good Integration Test**:
```go
func TestUserRepository_Integration(t *testing.T) {
    // Using testcontainers for real PostgreSQL
    container := testcontainers.PostgreSQLContainer(ctx, "postgres:15")
    defer container.Terminate(ctx)
    
    repo := NewUserRepository(container.ConnectionString())
    user, err := repo.CreateUser(ctx, &User{Email: "test@example.com"})
    
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### Contract Tests
**Scope**: Verify API contracts between services

**Requirements**:
- **API Compatibility**: Ensure backward compatibility of REST and gRPC APIs
- **Schema Validation**: Validate request/response schemas and protobuf definitions
- **Version Compatibility**: Test API versioning and migration strategies
- **Consumer-Driven**: Test from consumer perspective to catch breaking changes

### End-to-End Tests
**Scope**: Test complete user workflows across all services

**Requirements**:
- **Full System**: All services running with real databases (via testcontainers)
- **User Scenarios**: Complete workflows (auth → AI generation → response streaming)
- **Performance**: Include basic performance validation
- **CI/CD**: Automated in staging environment, can be triggered in CI for major changes

## CI/CD Testing Requirements

### Local Development Commands
- **Quick Feedback**: `make test` runs unit tests only (< 2 minutes)
- **Integration**: `make test-integration` runs integration tests with testcontainers
- **Performance**: `make perf-test` runs performance benchmarks locally
- **Full Suite**: `make ci` runs complete test pipeline matching CI/CD

### Continuous Integration Pipeline
1. **Stage 1 - Unit Tests**: Fast feedback with mocked dependencies
2. **Stage 2 - Integration Tests**: Services with testcontainer databases
3. **Stage 3 - Contract Tests**: API compatibility and schema validation
4. **Stage 4 - Security Tests**: `make security` vulnerability scanning
5. **Stage 5 - Performance Tests**: Benchmark regression testing
6. **Stage 6 - End-to-End**: Full system validation (staging deployment)

## Mocking Strategy

### Always Mock in Unit Tests
- **External APIs**: LLM providers (OpenAI, vLLM), third-party services
- **Databases**: PostgreSQL, Redis - use mocks (not in-memory databases) to ensure code is properly abstracted
- **Inter-Service Calls**: gRPC calls between microservices
- **File System**: File operations, configuration loading
- **Time/Random**: Deterministic testing with fixed values
- **Design Requirement**: Code must be designed with proper abstractions to enable complete mocking

### Use Real Systems in Integration Tests
- **Databases**: PostgreSQL and Redis via testcontainers
- **Service Communication**: Actual gRPC between test service instances
- **Message Queues**: If added, use testcontainer instances
- **Configuration**: Real config files and environment variables

## Testcontainers Usage

### Implementation Guidelines
- **Database Testing**: PostgreSQL with realistic schemas and data
- **Cache Testing**: Redis with actual caching scenarios
- **Isolation**: Each test gets fresh container instances
- **Cleanup**: Automatic container cleanup after tests
- **CI Compatibility**: Works in Docker-based CI environments

### Example Usage
```go
func TestWithPostgreSQL(t *testing.T) {
    ctx := context.Background()
    
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "postgres:15",
            Env: map[string]string{
                "POSTGRES_PASSWORD": "test",
                "POSTGRES_DB":       "testdb",
            },
            ExposedPorts: []string{"5432/tcp"},
            WaitingFor:   wait.ForListeningPort("5432/tcp"),
        },
        Started: true,
    })
    require.NoError(t, err)
    defer postgres.Terminate(ctx)
    
    // Use real PostgreSQL for integration test
}
```

## Performance Testing Requirements

### Baseline Metrics
- **Establish baselines**: Performance baselines for all critical paths
- **Regression Testing**: Fail builds if performance degrades > 20%
- **Load Testing**: Controlled load testing with realistic user scenarios
- **Resource Monitoring**: Track memory, CPU, and database connection usage
- **AI/LLM Optimization**: Specific benchmarks for AI generation performance

### Performance Test Types
- **Benchmark Tests**: `make perf-benchmark` - Go benchmark tests for critical functions
- **Load Tests**: `make perf-load` - Realistic user load scenarios
- **Stress Tests**: `make perf-stress` - System limits and breaking points

## Test Organization & Best Practices

### Test Structure
```
tests/
├── unit/
│   ├── ai/
│   ├── auth/
│   ├── user/
│   └── gateway/
├── integration/
│   ├── database/
│   ├── grpc/
│   └── end-to-end/
└── performance/
    ├── benchmarks/
    └── load/
```

### Naming Conventions
- **Test Files**: `*_test.go` for Go, `*.test.ts` for TypeScript
- **Test Functions**: `TestFunctionName_Scenario` (Go), `describe/it` blocks (TypeScript)
- **Mock Files**: `mock_*.go` generated via `make generate-mocks`

### Coverage Requirements
- **Minimum Coverage**: 80% for new code
- **Focus Areas**: Business logic, error handling, edge cases
- **Exclusions**: Generated code, simple getters/setters
- **Reporting**: Coverage reports generated via `make test-coverage`