# Phase B: Interface & Middleware Migration (ADR-017 Phase 2)
*Priority: HIGH | Estimated Time: 1 week*

## **B.1 HTTP Handlers Consolidation**
**Goal**: Move handlers to service packages

### Implementation Steps:
- [x] **Move handlers to services**:
  - [x] `internal/interfaces/http/user_handler.go` → `internal/user/http_handler.go`
  - [x] `internal/interfaces/http/auth_handler.go` → `internal/auth/http_handler.go`
  - [x] `internal/interfaces/http/ai_handler.go` → `internal/ai/http_handler.go`
- [x] **Update constructors**: Fix dependency injection
- [x] **Update imports**: Router and main function imports
- [x] **🏗️ IMPLEMENT HANDLER INTERFACE PATTERN**: Create consistent HTTP handler interface across services
  - [x] Design `HTTPHandler` interface following Java-style pattern for consistent request handling
  - [x] Implement interface for REST, GraphQL, and future API paradigms
  - [x] Create handler factory pattern for dynamic handler registration
  - [x] Ensure all HTTP operations follow same interface contract with middleware support
  - [x] Add handler-agnostic request validation, response formatting, and error handling

### HTTP Handler Interface Pattern Design:
```go
// Core interface that all HTTP handlers must implement
type HTTPHandler interface {
    // Register routes with the router
    RegisterRoutes(router Router) error
    
    // Middleware management
    GetMiddleware() []Middleware
    SetMiddleware(middleware []Middleware)
    
    // Health and validation
    HealthCheck() error
    ValidateRoutes() error
}

// Router interface for consistent routing across different frameworks
type Router interface {
    GET(path string, handler HandlerFunc, middleware ...Middleware)
    POST(path string, handler HandlerFunc, middleware ...Middleware)
    PUT(path string, handler HandlerFunc, middleware ...Middleware)
    DELETE(path string, handler HandlerFunc, middleware ...Middleware)
    Group(prefix string) RouterGroup
}

// Factory pattern for handler instantiation
type HandlerFactory interface {
    CreateHandler(serviceType string, config ServiceConfig) (HTTPHandler, error)
    RegisterHandler(router Router, handler HTTPHandler) error
}

// Service-specific handler interfaces
type AuthHandler interface {
    HTTPHandler
    Login(ctx Context) error
    Logout(ctx Context) error
    RefreshToken(ctx Context) error
    ValidateToken(ctx Context) error
}

type UserHandler interface {
    HTTPHandler
    GetUser(ctx Context) error
    UpdateUser(ctx Context) error
    GetProjects(ctx Context) error
}

type AIHandler interface {
    HTTPHandler
    GenerateCode(ctx Context) error
    GetGenerationHistory(ctx Context) error
    StreamGeneration(ctx Context) error
}
```

### HTTP Handler Interface Benefits:
- **Framework Agnostic**: Switch between Gin, Echo, or future HTTP frameworks
- **Consistent API**: All HTTP handlers follow same registration and middleware patterns
- **Middleware Integration**: Standardized middleware application across handlers
- **Easy Testing**: Mock handler implementations for automated testing
- **Route Validation**: Compile-time route validation and conflict detection
- **Future Extensibility**: Add new API paradigms without changing service logic

### Test Requirements:
- [x] **Organization of tests**: All tests must be appropriately packaged in the `test` directory at the root of the workspace.
- [x] **Handler tests**: Move to service packages (`http_handler_test.go`)
- [x] **90%+ coverage**: HTTP endpoints, request validation, response formatting, error handling
- [x] **Integration tests**: Test server with real requests
- [x] **Middleware tests**: Authentication flows
- [x] **🏗️ IMPLEMENT MOCK INTEGRATION PATTERN**: Follow ADR-024 mock integration strategy
  - [x] Use generated mocks from `tests/mocks/` instead of manual mocks
  - [x] Follow `gomock.Controller` pattern for test setup
  - [x] Apply progressive mock migration: manual mocks → generated mocks → pattern validation
  - [x] Reference ADR-024 for mock generation script usage and conflict resolution

### Coding Standards Validation:
- [x] **File size limits**: Keep all handler files under 300 lines (refactor at 300+, never exceed 500)
- [x] **Function size limits**: Keep handler functions under 30 lines (refactor at 30+, never exceed 50)
- [x] **Single responsibility**: Each handler function handles one HTTP endpoint
- [x] **Clear separation**: Keep request validation, business logic, and response formatting separate
- [x] **Error handling**: Consistent error response formatting across all handlers
- [x] **Input validation**: Validate all HTTP inputs explicitly

### Success Criteria:
✅ Clean handler interfaces for auth integration  
✅ Service packages own their HTTP interfaces  
✅ Router imports from service packages  
✅ HTTP handler interface pattern implemented across all services  
✅ Framework-agnostic handler registration and middleware support  

### Version Control:
- [x] **Commit changes**: `git add . && git commit -m "feat: consolidate HTTP handlers into service packages"`
- [x] **Validate build**: Ensure all tests pass and services compile before committing

---

## **B.2 Gateway Consolidation** ⭐ BUILD ON EXISTING AUTH PROXY
**Goal**: Consolidate gateway components while preserving auth integration

### Implementation Steps:
- [ ] **Move components to gateway package**:
  - [ ] `internal/interfaces/http/router.go` → `internal/gateway/router.go`
  - [ ] `internal/middleware/auth_proxy.go` → `internal/gateway/auth_proxy.go` 
  - [ ] `internal/middleware/logging.go` → `internal/gateway/logging.go`
  - [ ] `internal/middleware/ratelimit.go` → `internal/gateway/ratelimit.go`
  - [ ] `internal/proxy/proxy.go` → `internal/gateway/proxy.go`
- [ ] **⚠️ PRESERVE AUTH INTEGRATION**: Keep existing auth service proxy working
- [ ] **Update router**: Import handlers from service packages
- [ ] **Update main gateway**: Use consolidated components
- [ ] **🏗️ IMPLEMENT MIDDLEWARE INTERFACE PATTERN**: Create consistent middleware interface for all gateway functions
  - [ ] Design `Middleware` interface following Java-style pattern for consistent request processing
  - [ ] Implement interface for auth, logging, rate limiting, and future middleware types
  - [ ] Create middleware factory pattern for dynamic middleware composition
  - [ ] Ensure all middleware follows same interface contract with proper chain handling
  - [ ] Add middleware-agnostic configuration and metrics integration
- [ ] **🏗️ IMPLEMENT OBSERVER PATTERN**: Add comprehensive gateway event monitoring
  - [ ] Design `GatewayEventObserver` interface for decoupled monitoring and alerting
  - [ ] Implement observers for metrics collection, security monitoring, and error tracking
  - [ ] Create event notifier system for real-time gateway event broadcasting
  - [ ] Enable configurable observer registration for different monitoring needs

### Middleware Interface Pattern Design:
```go
// Core interface that all middleware implementations must implement
type Middleware interface {
    // Process the request and call next middleware in chain
    Process(ctx Context, next Next) error
    
    // Get middleware configuration and metadata
    GetConfig() MiddlewareConfig
    GetName() string
    GetOrder() int
    
    // Health and validation
    HealthCheck() error
    ValidateConfig() error
}

// Chain interface for middleware composition
type MiddlewareChain interface {
    Add(middleware Middleware) MiddlewareChain
    Execute(ctx Context) error
    GetMiddleware() []Middleware
}

// Factory pattern for middleware instantiation
type MiddlewareFactory interface {
    CreateMiddleware(middlewareType string, config MiddlewareConfig) (Middleware, error)
    CreateChain(middlewares []Middleware) MiddlewareChain
    ListAvailableMiddleware() []string
}

// Specific middleware interfaces
type AuthMiddleware interface {
    Middleware
    ValidateToken(ctx Context) (*UserContext, error)
    CheckPermissions(ctx Context, permissions []string) error
}

type RateLimitMiddleware interface {
    Middleware
    CheckLimit(ctx Context, identifier string) error
    GetLimitInfo(ctx Context, identifier string) (*LimitInfo, error)
}

type LoggingMiddleware interface {
    Middleware
    LogRequest(ctx Context) error
    LogResponse(ctx Context) error
}
```

### Observer Pattern Design for Gateway Events:
```go
// Observer pattern for gateway request/response monitoring
type GatewayEventObserver interface {
    OnRequestReceived(ctx Context, request *HTTPRequest) error
    OnRequestProcessed(ctx Context, request *HTTPRequest, response *HTTPResponse) error
    OnError(ctx Context, request *HTTPRequest, err error) error
    OnMetricsUpdate(ctx Context, metrics *RequestMetrics) error
}

// Event subject that notifies observers
type GatewayEventNotifier interface {
    Subscribe(observer GatewayEventObserver) error
    Unsubscribe(observer GatewayEventObserver) error
    NotifyRequestReceived(ctx Context, request *HTTPRequest) error
    NotifyRequestProcessed(ctx Context, request *HTTPRequest, response *HTTPResponse) error
    NotifyError(ctx Context, request *HTTPRequest, err error) error
}

// Concrete observer implementations
type MetricsObserver struct {
    metricsCollector *MetricsCollector
}

func (m *MetricsObserver) OnRequestReceived(ctx Context, request *HTTPRequest) error {
    m.metricsCollector.IncrementRequestCount(request.Path, request.Method)
    return nil
}

func (m *MetricsObserver) OnRequestProcessed(ctx Context, request *HTTPRequest, response *HTTPResponse) error {
    duration := time.Since(request.StartTime)
    m.metricsCollector.RecordLatency(request.Path, duration)
    m.metricsCollector.IncrementResponseCode(response.StatusCode)
    return nil
}

type SecurityObserver struct {
    alertManager *AlertManager
    logger       Logger
}

func (s *SecurityObserver) OnError(ctx Context, request *HTTPRequest, err error) error {
    if IsSecurityError(err) {
        s.alertManager.SendSecurityAlert(ctx, request, err)
        s.logger.LogSecurityEvent(ctx, request, err)
    }
    return nil
}

// Gateway with observer pattern integration
type ObservableGateway struct {
    router    Router
    notifier  GatewayEventNotifier
    observers []GatewayEventObserver
}

func (g *ObservableGateway) ProcessRequest(ctx Context, request *HTTPRequest) (*HTTPResponse, error) {
    // Notify observers of incoming request
    g.notifier.NotifyRequestReceived(ctx, request)
    
    // Process request through middleware chain
    response, err := g.router.ProcessRequest(ctx, request)
    
    if err != nil {
        g.notifier.NotifyError(ctx, request, err)
        return nil, err
    }
    
    // Notify observers of successful processing
    g.notifier.NotifyRequestProcessed(ctx, request, response)
    return response, nil
}
```

### Middleware Interface Benefits:
- **Composable**: Easy middleware composition and reordering
- **Consistent API**: All middleware follows same processing patterns
- **Configuration**: Standardized middleware configuration and validation
- **Easy Testing**: Mock middleware implementations for automated testing
- **Metrics Integration**: Centralized middleware performance monitoring
- **Future Extensibility**: Add new middleware types without changing gateway logic

### Observer Pattern Benefits:
- **Decoupled Monitoring**: Separate monitoring concerns from core gateway logic
- **Extensible**: Easy to add new observers for different monitoring needs
- **Real-time**: Immediate notification of gateway events for responsive monitoring
- **Configurable**: Enable/disable specific observers based on environment
- **Multi-Purpose**: Same event stream feeds metrics, logging, alerting, and analytics
- **Testable**: Mock observers for testing gateway behavior without side effects

### Test Requirements:
- [ ] **Organization of tests**: All tests must be appropriately packaged in the `test` directory at the root of the workspace.
- [ ] **Gateway tests**: `router_test.go`, `auth_proxy_test.go`, `logging_test.go`, etc.
- [ ] **Observer tests**: `gateway_observer_test.go`, `metrics_observer_test.go`
- [ ] **90%+ coverage**: routing, proxy forwarding, auth middleware, rate limiting
- [ ] **Integration tests**: Multiple service backends
- [ ] **Failure scenarios**: Circuit breaker behavior
- [ ] **Observer integration tests**: Event notification and handling
- [ ] **🏗️ IMPLEMENT MOCK INTEGRATION PATTERN**: Follow ADR-024 mock integration strategy
  - [ ] Generate mocks for gateway interfaces: `middleware.go`, `proxy.go`, `router.go`
  - [ ] Use `mocks.NewMockMiddleware(ctrl)` pattern for middleware testing
  - [ ] Apply `gomock.EXPECT()` chains for gateway request/response validation
  - [ ] Reference mock integration examples from cache/ai service tests

### Coding Standards Validation:
- [ ] **File size limits**: Keep all gateway files under 300 lines (refactor at 300+, never exceed 500)
- [ ] **Function size limits**: Keep middleware functions under 30 lines (refactor at 30+, never exceed 50)
- [ ] **Single responsibility**: Each middleware function handles one concern (auth, logging, etc.)
- [ ] **Avoid nested logic**: Split complex routing and proxy logic into smaller functions
- [ ] **Error handling**: Consistent error handling across all middleware
- [ ] **Clear separation**: Keep routing, middleware, and proxy logic distinct

### Success Criteria:
✅ All middleware consolidated in gateway  
✅ Auth service integration preserved  
✅ Gateway owns routing and proxy logic  
✅ Clean service boundaries maintained  
✅ Middleware interface pattern implemented for all gateway functions  
✅ Composable middleware chain with standardized configuration  
✅ Observer pattern implemented for comprehensive gateway monitoring  
✅ Decoupled event-driven monitoring and alerting system  

### Version Control:
- [ ] **Commit changes**: `git add . && git commit -m "feat: consolidate gateway components while preserving auth integration"`
- [ ] **Validate build**: Ensure all tests pass and services compile before committing

---

## **B.3 gRPC & Service Migration**
**Goal**: Complete interface consolidation

### Implementation Steps:
- [ ] **gRPC consolidation**:
  - [ ] Move `internal/interfaces/grpc/user_server.go` → `internal/user/grpc_server.go`
  - [ ] Update gRPC service registration
  - [ ] Update imports in server startup code
- [ ] **Service analysis**:
  - [ ] Analyze `internal/service/service.go` for ownership
  - [ ] Move to appropriate service package or gateway
- [ ] **🏗️ IMPLEMENT gRPC INTERFACE PATTERN**: Create consistent gRPC service interface across all services
  - [ ] Design `GRPCService` interface following Java-style pattern for consistent RPC handling
  - [ ] Implement interface for user, auth, AI, and future gRPC services
  - [ ] Create service factory pattern for dynamic gRPC service registration
  - [ ] Ensure all gRPC operations follow same interface contract with interceptor support
  - [ ] Add service-agnostic error handling, validation, and metrics integration

### gRPC Service Interface Pattern Design:
```go
// Core interface that all gRPC service implementations must implement
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

// Factory pattern for gRPC service instantiation
type GRPCServiceFactory interface {
    CreateService(serviceType string, config ServiceConfig) (GRPCService, error)
    RegisterAllServices(server *grpc.Server) error
    ListAvailableServices() []string
}

// Service-specific gRPC interfaces
type UserGRPCService interface {
    GRPCService
    GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error)
    UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error)
    ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error)
}

type AuthGRPCService interface {
    GRPCService
    ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error)
    RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error)
}

// Interceptor interface for consistent middleware
type GRPCInterceptor interface {
    UnaryInterceptor() grpc.UnaryServerInterceptor
    StreamInterceptor() grpc.StreamServerInterceptor
    GetConfig() InterceptorConfig
}
```

### gRPC Interface Benefits:
- **Protocol Agnostic**: Easy migration to gRPC-Web, Connect, or future RPC protocols
- **Consistent Registration**: All services follow same registration and interceptor patterns
- **Interceptor Management**: Standardized middleware application across gRPC services
- **Easy Testing**: Mock gRPC service implementations for automated testing
- **Service Discovery**: Automatic service registration and health checking
- **Future Extensibility**: Add new RPC protocols without changing service logic

### Test Requirements:
- [ ] **Organization of tests**: All tests must be appropriately packaged in the `test` directory at the root of the workspace.
- [ ] **gRPC tests**: `grpc_server_test.go` in service packages
- [ ] **90%+ coverage**: gRPC endpoints, protobuf validation, error status codes
- [ ] **Integration tests**: Test client connections
- [ ] **Service tests**: Startup and shutdown procedures

### Coding Standards Validation:
- [ ] **File size limits**: Keep all gRPC files under 300 lines (refactor at 300+, never exceed 500)
- [ ] **Function size limits**: Keep gRPC methods under 30 lines (refactor at 30+, never exceed 50)
- [ ] **Single responsibility**: Each gRPC method handles one service operation
- [ ] **Error handling**: Proper gRPC error status codes and messages
- [ ] **Input validation**: Validate all protobuf inputs explicitly
- [ ] **Clear separation**: Keep gRPC transport separate from business logic

### Success Criteria:
✅ gRPC servers owned by services  
✅ Service components properly distributed  
✅ Clean compilation and imports  
✅ gRPC interface pattern implemented across all services  
✅ Protocol-agnostic service registration and interceptor support  

### Version Control:
- [ ] **Commit changes**: `git add . && git commit -m "feat: complete gRPC and service component migration"`
- [ ] **Validate build**: Ensure all tests pass and services compile before committing
