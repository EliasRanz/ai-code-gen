# Architecture Patterns & Project Guidelines

## Purpose
Project-specific architecture patterns, API design guidelines, and microservice interaction patterns for the AI UI Generator project, focusing on maintainable and extensible system design.

## Microservices Architecture

### Service Boundaries & Responsibilities
The AI UI Generator follows a well-decomposed microservices architecture:

- **`ai/`**: AI generation and LLM orchestration, code generation workflows
- **`auth/`**: Authentication, authorization, JWT token management, session handling
- **`user/`**: User management, profile data, user preferences
- **`gateway/`**: API gateway, routing, rate limiting, request/response transformation
- **`cache/`**: Caching strategies, Redis integration, cache invalidation
- **`observability/`**: Monitoring, logging, metrics collection, distributed tracing
- **`utilities/`**: Shared utilities, common interfaces, helper functions

### Service Communication Patterns
- **gRPC**: Primary communication between internal microservices
- **REST/HTTP**: External API endpoints for frontend and third-party integrations
- **SSE (Server-Sent Events)**: Real-time streaming for AI generation progress
- **Event-Driven**: Asynchronous communication for non-critical operations

**Example Service Communication**:
```go
// gRPC service definition
service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

// HTTP endpoint for external access
func (h *HTTPHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    user, err := h.grpcClient.GetUser(ctx, &pb.GetUserRequest{Id: userID})
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, user)
}
```

## API Design & Extensibility

### Core Framework Integrity
- **Stable APIs**: Maintain stable, well-defined APIs that protect core functionality while enabling extensions
- **Backward Compatibility**: Ensure API changes maintain backward compatibility
- **Interface Segregation**: Design focused interfaces that serve specific purposes
- **Extension Points**: Identify and document specific points where community plugins can integrate

### Plugin Architecture Support
- **Plugin APIs**: Design APIs to support third-party plugins and community-driven extensions
- **API Boundaries**: Clear separation between internal service APIs (gRPC) and external extension APIs (REST/HTTP)
- **Versioning Strategy**: Implement API versioning to ensure backward compatibility for external integrations
- **Security Model**: Ensure plugin APIs include proper authentication, authorization, and sandboxing

### Documentation Standards
- **API Documentation**: Comprehensive API documentation for both internal and external consumers
- **OpenAPI Specs**: Maintain OpenAPI specifications for REST endpoints
- **gRPC Documentation**: Document gRPC services and message types
- **Integration Examples**: Provide clear examples for API integration

**Example API Versioning**:
```go
// Version-aware routing
func setupRoutes(r *gin.Engine) {
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users/:id", handlers.GetUserV1)
        v1.POST("/users", handlers.CreateUserV1)
    }
    
    v2 := r.Group("/api/v2")
    {
        v2.GET("/users/:id", handlers.GetUserV2)  // Enhanced response format
        v2.POST("/users", handlers.CreateUserV2)  // Additional validation
    }
}
```

## Framework & Library Strategy

### Library-First Philosophy
- **Existing Solutions**: Research and use established libraries for common functionality (validation, serialization, HTTP handling, etc.)
- **Framework Benefits**: Leverage framework testing, documentation, and community support rather than building from scratch
- **Business Value Focus**: Focus custom development on features that provide unique business value
- **Technical Debt Avoidance**: Avoid creating maintenance burden through custom implementations of solved problems

### Evaluation Criteria for Libraries
Choose libraries based on:
- **Active Maintenance**: Strong community support and ongoing development
- **Test Coverage**: Comprehensive testing and documentation
- **Compatibility**: Works well with Go ecosystem and project requirements
- **Security Track Record**: Good security history and vulnerability response
- **Performance**: Meets performance requirements for the use case

### Integration Testing Strategy
- **Test Integration Points**: Test integration points with libraries rather than reimplementing their functionality
- **Mock External Dependencies**: Mock library interfaces in unit tests
- **Version Compatibility**: Test library upgrades in controlled environments

## Repository & Interface Patterns

### Repository Pattern Implementation
- **Interface Definition**: Clear repository interfaces for data access
- **Dependency Injection**: Use dependency injection for repository implementations
- **Testability**: Design repositories to be easily mockable for testing
- **Error Handling**: Consistent error handling across repository implementations

**Example Repository Pattern**:
```go
// Repository interface
type UserRepository interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    UpdateUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, id string) error
}

// PostgreSQL implementation
type PostgreSQLUserRepository struct {
    db *sql.DB
}

func (r *PostgreSQLUserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    query := "SELECT id, email, created_at FROM users WHERE id = $1"
    row := r.db.QueryRowContext(ctx, query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Email, &user.CreatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return &user, nil
}

// Service using repository
type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### Factory Pattern for Dependencies
- **Configuration-Based**: Use factory patterns for creating configured dependencies
- **Environment-Aware**: Support different implementations for different environments
- **Testing Support**: Easy switching between real and mock implementations

## Configuration Management

### Environment-Based Configuration
- **Environment Variables**: Primary configuration through environment variables
- **Configuration Files**: YAML/JSON configuration files for complex settings
- **Default Values**: Sensible defaults for development environments
- **Validation**: Configuration validation at startup

### Service-Specific Configuration
- **Scoped Configuration**: Each service has its own configuration scope
- **Shared Configuration**: Common configuration shared across services
- **Hot Reloading**: Support for configuration updates without restart where appropriate

**Example Configuration Structure**:
```yaml
# configs/config.yaml
server:
  port: 8080
  host: "0.0.0.0"
  timeout: "30s"

database:
  host: "${DB_HOST:localhost}"
  port: "${DB_PORT:5432}"
  name: "${DB_NAME:ai_generator}"
  ssl_mode: "${DB_SSL_MODE:disable}"

redis:
  host: "${REDIS_HOST:localhost}"
  port: "${REDIS_PORT:6379}"
  db: "${REDIS_DB:0}"

llm:
  provider: "${LLM_PROVIDER:openai}"
  api_key: "${LLM_API_KEY}"
  timeout: "${LLM_TIMEOUT:60s}"
```

## Error Handling & Observability

### Consistent Error Handling
- **Error Types**: Define specific error types for different failure scenarios
- **Error Wrapping**: Use error wrapping to maintain context
- **Logging**: Structured logging with appropriate context
- **Recovery**: Graceful error recovery where possible

### Observability Patterns
- **Distributed Tracing**: Implement tracing across service boundaries
- **Metrics Collection**: Collect relevant metrics for monitoring
- **Health Checks**: Implement health check endpoints for all services
- **Alerting**: Set up appropriate alerting for critical failures

## Build & Deployment Patterns

### Makefile Integration
- **Consistent Commands**: Use Makefile for consistent build and deployment commands
- **Environment Setup**: Makefile targets for environment setup and configuration
- **Testing Integration**: Integrated testing commands for different test types
- **CI/CD Support**: Makefile targets that support CI/CD pipelines

### Containerization
- **Docker Support**: All services containerized for consistent deployment
- **Multi-stage Builds**: Use multi-stage Docker builds for optimization
- **Health Checks**: Container health checks for orchestration platforms
- **Resource Limits**: Appropriate resource limits and requests

## Development Workflow

### Code Generation
- **Automated Generation**: Use `make generate` for automated code generation
- **Mocks**: Generate mocks for testing using project tooling
- **Protocol Buffers**: Generate gRPC code from proto definitions
- **Database Migrations**: Automated database migration generation

### Quality Gates
- **Pre-commit**: Local quality checks before committing code
- **CI Pipeline**: Comprehensive CI pipeline validation
- **Code Review**: Structured code review process
- **Performance Testing**: Integrated performance validation

### Documentation & Knowledge Sharing
- **ADRs**: Document architectural decisions in `ADRs/` directory
- **API Documentation**: Maintain up-to-date API documentation
- **Setup Guides**: Clear setup and development guides
- **Troubleshooting**: Common troubleshooting guides and solutions