# Phase C: Domain Cleanup (ADR-017 Phase 3)
*Priority: MEDIUM | Estimated Time: 3-4 days*

## **C.1 Domain Layer Elimination**
**Goal**: Flatten domain abstractions into service packages

### Implementation Steps:
- [ ] **Domain migration**:
  - [ ] Move `internal/domain/ai/` contents → `internal/ai/` (merge with existing entities)
  - [ ] Move `internal/domain/common/` contents → `internal/utilities/` (merge with existing)
- [ ] **Resolve conflicts**: Between domain entities and existing service entities
- [ ] **Update imports**: Across all services
- [ ] **Remove directory**: Empty `internal/domain/`
- [ ] **🏗️ IMPLEMENT ENTITY INTERFACE PATTERN**: Create consistent domain entity interface across all services
  - [ ] Design `DomainEntity` interface following Java-style pattern for consistent entity behavior
  - [ ] Implement interface for User, Project, Generation, and future entities
  - [ ] Create entity factory pattern for dynamic entity creation and validation
  - [ ] Ensure all entities follow same interface contract with validation and serialization
  - [ ] Add entity-agnostic validation, auditing, and change tracking

### Entity Interface Pattern Design:
```go
// Core interface that all domain entities must implement
type DomainEntity interface {
    // Entity identification and metadata
    GetID() string
    GetType() EntityType
    GetVersion() int64
    GetCreatedAt() time.Time
    GetUpdatedAt() time.Time
    
    // Validation and business rules
    Validate() error
    IsValid() bool
    GetValidationRules() []ValidationRule
    
    // Serialization and persistence
    ToJSON() ([]byte, error)
    FromJSON(data []byte) error
    ToMap() map[string]interface{}
    
    // Change tracking and auditing
    MarkDirty(field string)
    GetDirtyFields() []string
    ClearDirtyFields()
    
    // Lifecycle events
    BeforeSave() error
    AfterSave() error
    BeforeDelete() error
}

// Factory pattern for entity instantiation
type EntityFactory interface {
    CreateEntity(entityType EntityType, data map[string]interface{}) (DomainEntity, error)
    CreateFromJSON(entityType EntityType, jsonData []byte) (DomainEntity, error)
    ListEntityTypes() []EntityType
}

// Validation interface for consistent entity validation
type EntityValidator interface {
    ValidateEntity(entity DomainEntity) error
    ValidateField(entity DomainEntity, field string, value interface{}) error
    GetFieldRules(entityType EntityType, field string) []ValidationRule
}

// Service-specific entity interfaces
type User interface {
    DomainEntity
    GetEmail() string
    GetUsername() string
    SetPassword(password string) error
    ValidatePassword(password string) bool
    GetRoles() []string
    HasPermission(permission string) bool
}

type Project interface {
    DomainEntity
    GetOwnerID() string
    GetName() string
    GetStatus() ProjectStatus
    SetStatus(status ProjectStatus) error
    GetGenerations() []Generation
    AddGeneration(generation Generation) error
}

type Generation interface {
    DomainEntity
    GetProjectID() string
    GetPrompt() string
    GetContent() string
    GetProvider() string
    GetTokensUsed() int
    SetContent(content string) error
}
```

### Entity Interface Benefits:
- **Consistent Behavior**: All entities follow same validation, serialization, and lifecycle patterns
- **Change Tracking**: Automatic dirty field tracking for optimized database updates
- **Validation**: Standardized validation rules and error handling across entities
- **Serialization**: Consistent JSON/map conversion for API and storage
- **Auditing**: Built-in change tracking and lifecycle event hooks
- **Testing**: Easy mock entity implementations for comprehensive testing

### Test Requirements:
- [ ] **Organization of tests**: All tests must be appropriately packaged in the `test` directory at the root of the workspace.
- [ ] **Entity tests**: `internal/ai/entities_test.go`, `internal/utilities/types_test.go`
- [ ] **95%+ coverage**: entity creation, validation, business rules, type conversions
- [ ] **Integration tests**: Entity behavior and validation
- [ ] **Serialization tests**: Entity marshaling/unmarshaling
- [ ] **🏗️ IMPLEMENT MOCK INTEGRATION PATTERN**: Follow ADR-024 mock integration strategy
  - [ ] Generate mocks for entity interfaces using `scripts/generate-mocks.sh`
  - [ ] Apply `mocks.NewMockDomainEntity(ctrl)` pattern for entity testing
  - [ ] Use generated mocks for factory and validator testing with clean dependency injection
  - [ ] Reference mock integration patterns from user service tests in `tests/unit/user/`

### Coding Standards Validation:
- [ ] **File size limits**: Keep all entity files under 300 lines (refactor at 300+, never exceed 500)
- [ ] **Function size limits**: Keep entity methods under 30 lines (refactor at 30+, never exceed 50)
- [ ] **Single responsibility**: Each entity method handles one business rule or operation
- [ ] **Immutable design**: Prefer immutable entities where possible
- [ ] **Clear validation**: Explicit validation rules for all entity properties
- [ ] **Business logic**: Keep domain logic within entities, avoid anemic models

### Success Criteria:
✅ Clean structure for frontend auth integration  
✅ No domain layer abstractions  
✅ Service entities properly consolidated  
✅ Entity interface pattern implemented across all domain objects  
✅ Consistent validation, serialization, and change tracking  

### Version Control:
- [ ] **Commit changes**: `git add . && git commit -m "feat: eliminate domain layer and flatten abstractions"`
- [ ] **Validate build**: Ensure all tests pass and services compile before committing

---

## **C.2 Final Infrastructure Moves**
**Goal**: Complete shared infrastructure organization

### Implementation Steps:
- [ ] **Configuration management**:
  - [ ] Move `internal/infrastructure/config/` → `internal/config/`
  - [ ] Update config imports across services
  - [ ] Ensure shared configuration accessibility
- [ ] **Keep shared components**:
  - [ ] Keep `internal/observability/` as shared infrastructure
  - [ ] Keep `internal/database/` for shared connection utilities
- [ ] **Test migration**:
  - [ ] Move `internal/tests/` → `/tests/` (project root)
  - [ ] Update test import paths across packages
- [ ] **🏗️ IMPLEMENT OBSERVABILITY INTERFACE PATTERN**: Create consistent monitoring and observability interface
  - [ ] Design `ObservabilityProvider` interface following Java-style pattern for consistent metrics and monitoring
  - [ ] Implement interface for Prometheus, Grafana, OpenTelemetry, and future monitoring solutions
  - [ ] Create observability factory pattern for dynamic monitoring provider selection
  - [ ] Ensure all services follow same observability contract with metrics, tracing, and logging
  - [ ] Add provider-agnostic health checks, alerts, and performance monitoring
- [ ] **🏗️ IMPLEMENT DECORATOR PATTERN**: Add enhanced monitoring capabilities
  - [ ] Design `MonitoringDecorator` interface for wrapping existing components with monitoring
  - [ ] Implement decorators for repository, cache, and service monitoring
  - [ ] Create composable monitoring layers for different levels of detail
  - [ ] Enable dynamic monitoring enhancement without changing core business logic

### Observability Interface Pattern Design:
```go
// Core interface that all observability providers must implement
type ObservabilityProvider interface {
    // Metrics collection
    RecordMetric(name string, value float64, labels map[string]string) error
    IncrementCounter(name string, labels map[string]string) error
    RecordHistogram(name string, value float64, labels map[string]string) error
    RecordGauge(name string, value float64, labels map[string]string) error
    
    // Tracing
    StartSpan(ctx context.Context, name string) (context.Context, Span) 
    CreateTracer(serviceName string) Tracer
    
    // Health checking
    RegisterHealthCheck(name string, check HealthCheckFunc) error
    GetHealthStatus() HealthStatus
    
    // Configuration and lifecycle
    Configure(config ObservabilityConfig) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}

// Factory pattern for observability provider instantiation
type ObservabilityFactory interface {
    CreateProvider(providerType string, config ObservabilityConfig) (ObservabilityProvider, error)
    CreateMultiProvider(providers []ObservabilityProvider) ObservabilityProvider
    ListAvailableProviders() []string
}

// Service-specific observability managers
type ServiceObservability struct {
    provider    ObservabilityProvider
    serviceName string
    tracer      Tracer
    metrics     MetricRegistry
}

// Metrics interface for consistent metric collection
type MetricRegistry interface {
    Counter(name string) Counter
    Histogram(name string) Histogram
    Gauge(name string) Gauge
    Timer(name string) Timer
}

// Tracing interface for distributed tracing
type Tracer interface {
    StartSpan(ctx context.Context, operationName string) (context.Context, Span)
    InjectHeaders(span Span, headers map[string]string) error
    ExtractSpan(ctx context.Context, headers map[string]string) (context.Context, Span)
}

// Health check interface
type HealthCheckFunc func(ctx context.Context) error
```

### Decorator Pattern Design for Enhanced Monitoring:
```go
// Decorator pattern for adding monitoring to existing components
type MonitoringDecorator interface {
    WrapComponent(component interface{}) interface{}
    GetMetrics() MonitoringMetrics
    Configure(config MonitoringConfig) error
}

// Repository monitoring decorator
type RepositoryMonitoringDecorator struct {
    repository Repository
    metrics    MetricsCollector
    tracer     Tracer
    logger     Logger
}

func (r *RepositoryMonitoringDecorator) Create(ctx context.Context, entity interface{}) error {
    // Start span for tracing
    ctx, span := r.tracer.StartSpan(ctx, "repository.create")
    defer span.End()
    
    // Record metrics
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        r.metrics.RecordHistogram("repository_operation_duration", duration.Seconds(), map[string]string{
            "operation": "create",
            "entity_type": getEntityType(entity),
        })
    }()
    
    // Execute operation with error tracking
    err := r.repository.Create(ctx, entity)
    if err != nil {
        r.metrics.IncrementCounter("repository_errors", map[string]string{
            "operation": "create",
            "error_type": getErrorType(err),
        })
        r.logger.ErrorContext(ctx, "Repository create failed", "error", err)
    } else {
        r.metrics.IncrementCounter("repository_operations", map[string]string{
            "operation": "create",
        })
    }
    
    return err
}

// Cache monitoring decorator
type CacheMonitoringDecorator struct {
    cache   CacheProvider
    metrics MetricsCollector
    tracer  Tracer
}

func (c *CacheMonitoringDecorator) Get(ctx context.Context, key string) (string, error) {
    ctx, span := c.tracer.StartSpan(ctx, "cache.get")
    defer span.End()
    
    start := time.Now()
    value, err := c.cache.Get(ctx, key)
    duration := time.Since(start)
    
    // Record cache hit/miss metrics
    if err != nil {
        c.metrics.IncrementCounter("cache_misses", map[string]string{
            "operation": "get",
        })
    } else {
        c.metrics.IncrementCounter("cache_hits", map[string]string{
            "operation": "get",
        })
    }
    
    c.metrics.RecordHistogram("cache_operation_duration", duration.Seconds(), map[string]string{
        "operation": "get",
    })
    
    return value, err
}

// Service monitoring decorator
type ServiceMonitoringDecorator struct {
    service interface{}
    metrics MetricsCollector
    tracer  Tracer
    config  MonitoringConfig
}

func (s *ServiceMonitoringDecorator) WrapMethod(methodName string, originalMethod func(ctx context.Context, args ...interface{}) (interface{}, error)) func(ctx context.Context, args ...interface{}) (interface{}, error) {
    return func(ctx context.Context, args ...interface{}) (interface{}, error) {
        // Start distributed trace
        ctx, span := s.tracer.StartSpan(ctx, fmt.Sprintf("service.%s", methodName))
        defer span.End()
        
        // Record request metrics
        s.metrics.IncrementCounter("service_requests", map[string]string{
            "method": methodName,
        })
        
        start := time.Now()
        result, err := originalMethod(ctx, args...)
        duration := time.Since(start)
        
        // Record response metrics
        if err != nil {
            s.metrics.IncrementCounter("service_errors", map[string]string{
                "method": methodName,
                "error_type": getErrorType(err),
            })
        } else {
            s.metrics.IncrementCounter("service_success", map[string]string{
                "method": methodName,
            })
        }
        
        s.metrics.RecordHistogram("service_duration", duration.Seconds(), map[string]string{
            "method": methodName,
        })
        
        return result, err
    }
}

// Factory for creating monitoring decorators
type MonitoringDecoratorFactory struct {
    observability ObservabilityProvider
    config        MonitoringConfig
}

func (f *MonitoringDecoratorFactory) CreateRepositoryDecorator(repository Repository) Repository {
    return &RepositoryMonitoringDecorator{
        repository: repository,
        metrics:    f.observability.GetMetrics(),
        tracer:     f.observability.CreateTracer("repository"),
        logger:     f.observability.GetLogger(),
    }
}

func (f *MonitoringDecoratorFactory) CreateCacheDecorator(cache CacheProvider) CacheProvider {
    return &CacheMonitoringDecorator{
        cache:   cache,
        metrics: f.observability.GetMetrics(),
        tracer:  f.observability.CreateTracer("cache"),
    }
}
```

### Observability Interface Benefits:
- **Provider Agnostic**: Switch between Prometheus, Grafana, OpenTelemetry, or future solutions
- **Consistent Metrics**: All services use same metric collection and naming patterns
- **Distributed Tracing**: Standardized trace propagation across service boundaries
- **Health Monitoring**: Centralized health check registration and status reporting
- **Multi-Provider**: Support multiple observability providers simultaneously
- **Testing**: Mock observability implementations for testing without external dependencies

### Decorator Pattern Benefits:
- **Non-Intrusive**: Add monitoring to existing components without modifying their code
- **Composable**: Stack multiple decorators for different monitoring concerns
- **Configurable**: Enable/disable monitoring features dynamically
- **Consistent**: Same monitoring patterns across all component types
- **Performance**: Minimal overhead when monitoring is disabled
- **Separation of Concerns**: Keep monitoring logic separate from business logic

### Test Requirements:
- [ ] **Organization of tests**: All tests must be appropriately packaged in the `test` directory at the root of the workspace.
- [ ] **Config tests**: `internal/config/config_test.go`
- [ ] **Decorator tests**: `monitoring_decorator_test.go`
- [ ] **85%+ coverage**: shared config behavior, environment handling
- [ ] **Observability tests**: Cross-service functionality
- [ ] **Database tests**: Connection management, migration utilities
- [ ] **🏗️ IMPLEMENT MOCK INTEGRATION PATTERN**: Follow ADR-024 mock integration strategy
  - [ ] Generate mocks for config/database interfaces using `scripts/generate-mocks.sh`
  - [ ] Apply `mocks.NewMockConfigProvider(ctrl)` pattern for configuration testing
  - [ ] Use generated mocks for decorator and observability component testing
  - [ ] Reference existing config mock patterns from cache service tests
- [ ] **Monitoring integration tests**: Verify decorator behavior with real components

### Coding Standards Validation:
- [ ] **File size limits**: Keep all shared files under 300 lines (refactor at 300+, never exceed 500)
- [ ] **Function size limits**: Keep utility functions under 30 lines (refactor at 30+, never exceed 50)
- [ ] **Single responsibility**: Each utility function handles one specific task
- [ ] **Reusability**: Design for reuse across multiple services
- [ ] **Clear interfaces**: Well-defined interfaces for shared components
- [ ] **Documentation**: Clear documentation for shared utility functions

### Success Criteria:
✅ Shared infrastructure properly organized  
✅ Test organization follows Go conventions  
✅ Configuration accessible to all services  
✅ Observability interface pattern implemented across all monitoring  
✅ Provider-agnostic metrics, tracing, and health checking  
✅ Decorator pattern implemented for non-intrusive monitoring enhancement  
✅ Composable monitoring layers with configurable detail levels  

### Version Control:
- [ ] **Commit changes**: `git add . && git commit -m "feat: complete final infrastructure moves and test organization"`
- [ ] **Validate build**: Ensure all tests pass and services compile before committing
