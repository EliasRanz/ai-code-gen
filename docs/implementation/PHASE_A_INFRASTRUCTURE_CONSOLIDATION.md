# Phase A: Infrastructure Consolidation (ADR-017 Phase 1)
*Priority: HIGH | Estimated Time: 1-2 weeks*

## **A.1 Shared Cache Service Expansion** ⭐ COMPLETED ✅
**Goal**: Build comprehensive cache service on existing auth Redis infrastructure

### Implementation Steps:
- [x] **Expand existing cache**: Built comprehensive cache infrastructure on existing foundation
- [x] **Create cache interfaces**: Core cache interface and provider management in `interfaces.go`
- [x] **Service-specific managers**: Implemented cache managers in service packages:
  - [x] `internal/auth/cache.go` - Auth cache manager (migrated and enhanced)  
  - [x] `internal/user/cache.go` - User data caching (profiles, projects, sessions)
  - [x] `internal/ai/cache.go` - AI generation result caching with rate limiting
- [x] **Implement connection pooling**: Centralized Redis configuration with pooling
- [x] **Add observability**: Cache metrics and error handling integration
- [x] **Core infrastructure**: Complete cache provider implementations:
  - [x] `redis_provider.go` - Redis with circuit breaker protection
  - [x] `memory_provider.go` - In-memory fallback implementation  
  - [x] `multi_provider.go` - Multi-tier provider (Redis + Memory)
  - [x] `factory.go` - Factory pattern for provider creation
  - [x] `circuit_breaker.go` - Circuit breaker implementation
  - [x] `service.go` - Unified cache service entry point
- [x] **🏗️ IMPLEMENTED CACHE INTERFACE PATTERN**: Consistent caching interface for all cache types
  - [x] Designed `CacheProvider` interface following consistent pattern for cache operations
  - [x] Implemented interface for Redis, in-memory, and multi-tier providers
  - [x] Created cache factory pattern for dynamic cache selection per service
  - [x] All cache operations follow same interface contract
  - [x] Added cache-agnostic error handling and configuration
- [x] **🏗️ IMPLEMENTED CIRCUIT BREAKER PATTERN**: Resilience for Redis connections
  - [x] Implemented circuit breaker for cache operations to handle Redis failures gracefully
  - [x] Added automatic fallback to in-memory cache when Redis is unavailable
  - [x] Included configurable failure thresholds and recovery mechanisms
  - [x] Added proper state management for circuit breaker patterns

### Cache Interface Pattern Design:
```go
// Core interface that all cache providers must implement
type CacheProvider interface {
    // Basic cache operations
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string, ttl time.Duration) error
    Del(ctx context.Context, key string) error
    
    // Advanced operations
    Exists(ctx context.Context, key string) (bool, error)
    TTL(ctx context.Context, key string) (time.Duration, error)
    
    // Health and metrics
    HealthCheck(ctx context.Context) error
    GetMetrics() CacheMetrics
    
    // Cleanup and management
    Close() error
}

// Factory pattern for cache provider instantiation
type CacheFactory interface {
    CreateProvider(providerType string, config CacheConfig) (CacheProvider, error)
    ListAvailableProviders() []string
}

// Service-specific cache managers
type AuthCacheManager struct {
    provider CacheProvider
    prefix   string
}

type UserCacheManager struct {
    provider CacheProvider
    prefix   string
}

type AICacheManager struct {
    provider CacheProvider
    prefix   string
}
```

### Cache Interface Benefits:
- **Provider Agnostic**: Switch between Redis, in-memory, or future cache solutions
- **Consistent API**: All services use same cache interface contract
- **Easy Testing**: Mock implementations for automated testing
- **Future Extensibility**: Add new cache providers without changing service code
- **Metrics Integration**: Standardized cache metrics across all providers
- **Configuration Flexibility**: Per-service cache configuration and TTL policies

### Circuit Breaker Pattern Design:
```go
// Circuit breaker for cache operations and external dependencies
type CircuitBreaker interface {
    Execute(operation func() (interface{}, error)) (interface{}, error)
    GetState() CircuitState
    GetMetrics() CircuitMetrics
    Reset() error
}

type CircuitState string
const (
    StateClosed   CircuitState = "closed"    // Normal operation
    StateOpen     CircuitState = "open"      // Failing, requests rejected
    StateHalfOpen CircuitState = "half_open" // Testing recovery
)

// Cache with circuit breaker and fallback
type ResilientCacheProvider struct {
    primary    CacheProvider  // Redis cache
    fallback   CacheProvider  // In-memory cache
    breaker    CircuitBreaker
    config     CircuitConfig
    metrics    *CircuitMetrics
}

func (r *ResilientCacheProvider) Get(ctx context.Context, key string) (string, error) {
    // Try primary cache through circuit breaker
    result, err := r.breaker.Execute(func() (interface{}, error) {
        return r.primary.Get(ctx, key)
    })
    
    // Fallback to in-memory cache on circuit open
    if err != nil && r.breaker.GetState() == StateOpen {
        return r.fallback.Get(ctx, key)
    }
    
    return result.(string), err
}

// Circuit breaker configuration
type CircuitConfig struct {
    FailureThreshold       int           // Number of failures before opening
    RecoveryTimeout        time.Duration // Time before trying half-open
    SuccessThreshold       int           // Successes needed to close circuit
    RequestVolumeThreshold int           // Min requests before checking failure rate
}
```

### Circuit Breaker Benefits:
- **Resilience**: Prevents cascade failures when dependencies are down
- **Automatic Recovery**: Self-healing behavior with configurable recovery periods
- **Fallback Strategy**: Graceful degradation to backup systems
- **Performance Protection**: Prevents resource exhaustion during failures
- **Metrics**: Detailed failure and recovery tracking for monitoring
- **Configurable**: Tunable thresholds for different failure scenarios

### Test Requirements:
- [x] **Organization of tests**: All tests must be appropriately packaged in the `test` directory at the root of the workspace.
- [x] Move cache tests to service-specific files: `cache_test.go`, `auth_test.go`, `user_test.go`, `ai_test.go`
- [x] **90%+ coverage**: connection pooling, Redis operations, error handling, metrics
- [x] **Integration tests**: Real Redis instance testing
- [x] **Performance tests**: Cache operations under load

### Coding Standards Validation:
- [x] **File size limits**: Keep all files under 300 lines (refactor at 300+, never exceed 500)
- [x] **Function size limits**: Keep functions under 30 lines (refactor at 30+, never exceed 50)
- [x] **Single responsibility**: Each function does one thing and does it well
- [x] **Testability**: All functions are easily testable with clear inputs/outputs
- [x] **Error handling**: Explicit error handling, fail fast approach
- [x] **Descriptive naming**: Clear, consistent naming for variables, functions, and files

### Success Criteria:
✅ All services use unified cache service  
✅ Redis connection pooling implemented  
✅ Cache metrics integrated with observability  
✅ Performance tests pass under load  
✅ Cache interface pattern implemented across all providers  
✅ Service-specific cache managers follow consistent interface  
✅ Circuit breaker pattern implemented for Redis resilience  
✅ Automatic fallback to in-memory cache on Redis failures  

### Version Control:
- [x] **Commit changes**: `git add . && git commit -m "feat: expand shared cache service with Redis pooling and observability"`
- [x] **Validate build**: Ensure all tests pass and services compile before committing

---

## **A.2 LLM Functionality Consolidation + Rate Limiting Integration** ⭐ COMPLETED ✅
**Goal**: Complete AI service ownership with integrated rate limiting

### Implementation Steps:
- [x] **Move LLM components to AI service**:
  - [x] `internal/infrastructure/llm/` → `internal/ai/llm/`
  - [x] `internal/llm/types.go` → `internal/ai/llm/types.go`
  - [x] `internal/llm/vllm_client.go` → `internal/ai/llm/vllm_client.go`
  - [x] `internal/llm/vllm_helpers.go` → `internal/ai/llm/vllm_helpers.go`
  - [x] `internal/llm/vllm_types.go` → `internal/ai/llm/vllm_types.go`
  - [x] `internal/infrastructure/llm/openai_service.go` → `internal/ai/llm/openai_client.go`
- [x] **🚨 CRITICAL - Free Tier Configuration**:
  - [x] Configure all LLM clients for **FREE TIER ONLY**
  - [x] OpenAI: Use free tier limits, avoid paid API calls during development/testing
  - [x] vLLM: Use local/self-hosted instances, no cloud paid services
  - [x] Add configuration flags to prevent paid API usage
  - [x] Document free tier limitations and usage patterns
- [x] **Integrate rate limiting**: Connect with committed `internal/ai/rate_limit.go`
  - [x] Integrate RateLimiter with LLM client calls
  - [x] Integrate QuotaManager with daily usage tracking
  - [x] Add rate limiting middleware to generation endpoints
- [x] **Consolidate configuration**: All LLM settings in `internal/ai/config.go`
- [x] **Update imports**: Fix all AI service dependencies
- [x] **🏗️ IMPLEMENT INTERFACE PATTERN**: Create consistent LLM interface for all providers
  - [x] Design `LLMProvider` interface following Java-style pattern for consistent orchestration
  - [x] Implement interface for OpenAI, vLLM, and future providers
  - [x] Create provider factory pattern for dynamic model selection
  - [x] Ensure all providers follow same request/response structure
  - [x] Add provider-agnostic error handling and retry logic
- [x] **🏗️ IMPLEMENT BUILDER PATTERN**: Add complex configuration management for LLM requests
  - [x] Design `GenerationRequestBuilder` interface for fluent API construction of complex requests
  - [x] Implement validation at each step to prevent invalid configurations
  - [x] Add automatic free tier enforcement and defaults
  - [x] Enable chain-able method calls for readable request construction
- [x] **Legacy Cleanup**: Remove all legacy files and directories
  - [x] Remove `internal/ai/llm/legacy/` directory
  - [x] Remove empty `internal/llm/` directory
  - [x] Remove empty `internal/infrastructure/llm/` directory
  - [x] Update all import references to use new consolidated structure
  - [x] Remove outdated test files that reference legacy structure

### LLM Interface Pattern Design:
```go
// Core interface that all LLM providers must implement
type LLMProvider interface {
    // Generate code using the provider's model
    GenerateCode(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error)
    
    // Validate provider configuration and connectivity
    HealthCheck(ctx context.Context) error
    
    // Get provider-specific metadata (model name, version, capabilities)
    GetProviderInfo() ProviderInfo
    
    // Get provider-specific rate limits and quotas
    GetLimits() ProviderLimits
}

// Factory pattern for provider instantiation
type LLMFactory interface {
    CreateProvider(providerType string, config ProviderConfig) (LLMProvider, error)
    ListAvailableProviders() []string
}

// Standard request structure across all providers
type GenerationRequest struct {
    UserID      string            `json:"user_id"`
    Prompt      string            `json:"prompt"`
    Language    string            `json:"language"`
    MaxTokens   int               `json:"max_tokens"`
    Temperature float64           `json:"temperature"`
    Metadata    map[string]string `json:"metadata"`
}

// Standard response structure across all providers
type GenerationResponse struct {
    Content     string            `json:"content"`
    TokensUsed  int               `json:"tokens_used"`
    Provider    string            `json:"provider"`
    Model       string            `json:"model"`
    Latency     time.Duration     `json:"latency"`
    Metadata    map[string]string `json:"metadata"`
}

// Orchestrator manages multiple providers consistently
type LLMOrchestrator struct {
    providers    map[string]LLMProvider
    factory      LLMFactory
    rateLimiter  *RateLimiter
    quotaManager *QuotaManager
}
```

### Interface Implementation Benefits:
- **Consistent API**: All LLM providers follow same interface contract
- **Provider Agnostic**: Orchestrator works identically across OpenAI, vLLM, etc.
- **Easy Testing**: Mock implementations for automated testing
- **Future Extensibility**: Add new providers without changing orchestrator
- **Error Handling**: Standardized error types and retry logic
- **Configuration**: Unified configuration pattern across providers

### Builder Pattern Design for Complex LLM Configuration:
```go
// Builder pattern for complex LLM request configuration
type GenerationRequestBuilder interface {
    SetUserID(userID string) GenerationRequestBuilder
    SetPrompt(prompt string) GenerationRequestBuilder
    SetLanguage(language string) GenerationRequestBuilder
    SetMaxTokens(tokens int) GenerationRequestBuilder
    SetTemperature(temp float64) GenerationRequestBuilder
    AddMetadata(key, value string) GenerationRequestBuilder
    SetProvider(provider string) GenerationRequestBuilder
    SetModel(model string) GenerationRequestBuilder
    EnableStreaming() GenerationRequestBuilder
    SetTimeout(timeout time.Duration) GenerationRequestBuilder
    Build() (*GenerationRequest, error)
    Validate() error
}

// Concrete builder implementation
type LLMRequestBuilder struct {
    request *GenerationRequest
    config  *LLMConfig
    errors  []error
}

func NewGenerationRequestBuilder() GenerationRequestBuilder {
    return &LLMRequestBuilder{
        request: &GenerationRequest{
            Metadata: make(map[string]string),
        },
    }
}

func (b *LLMRequestBuilder) SetUserID(userID string) GenerationRequestBuilder {
    if userID == "" {
        b.errors = append(b.errors, errors.New("user ID cannot be empty"))
    }
    b.request.UserID = userID
    return b
}

func (b *LLMRequestBuilder) SetPrompt(prompt string) GenerationRequestBuilder {
    if len(prompt) == 0 {
        b.errors = append(b.errors, errors.New("prompt cannot be empty"))
    }
    if len(prompt) > 8000 { // FREE TIER limit
        b.errors = append(b.errors, errors.New("prompt exceeds free tier limit"))
    }
    b.request.Prompt = prompt
    return b
}

func (b *LLMRequestBuilder) Build() (*GenerationRequest, error) {
    if len(b.errors) > 0 {
        return nil, fmt.Errorf("validation errors: %v", b.errors)
    }
    
    // Apply free tier defaults
    if b.request.MaxTokens == 0 {
        b.request.MaxTokens = 1000 // Free tier default
    }
    if b.request.Temperature == 0 {
        b.request.Temperature = 0.7 // Balanced default
    }
    
    return b.request, nil
}

// Usage example for complex generation requests
func (s *AIService) GenerateWithBuilder(ctx context.Context, userID, prompt string) (*GenerationResponse, error) {
    request, err := NewGenerationRequestBuilder().
        SetUserID(userID).
        SetPrompt(prompt).
        SetLanguage("go").
        SetMaxTokens(500).
        SetTemperature(0.7).
        AddMetadata("request_id", uuid.New().String()).
        SetProvider("openai"). // FREE TIER only
        EnableStreaming().
        Build()
    
    if err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    return s.llmProvider.GenerateCode(ctx, request)
}
```

### Builder Pattern Benefits:
- **Complex Configuration**: Simplifies building complex LLM requests with many optional parameters
- **Validation**: Built-in validation at each step prevents invalid configurations
- **Free Tier Enforcement**: Automatic application of free tier limits and defaults
- **Immutable Results**: Builder creates immutable request objects for thread safety
- **Fluent API**: Chain-able method calls for readable request construction
- **Error Accumulation**: Collects all validation errors before failing

### Rate Limiting Integration Details:
```go
// Use existing RateLimiter from internal/ai/rate_limit.go
type AIService struct {
    rateLimiter *RateLimiter
    quotaManager *QuotaManager
    llmClient LLMClient
}

func (s *AIService) GenerateCode(ctx context.Context, req GenerationRequest) error {
    // Apply rate limiting before LLM call
    if !s.rateLimiter.Allow(req.UserID) {
        return ErrRateLimitExceeded
    }
    
    // Check daily quota
    if s.quotaManager.IsQuotaExceeded(req.UserID) {
        return ErrQuotaExceeded
    }
    
    // Proceed with LLM generation...
}
```

### Test Requirements:
- [x] **Organization of tests**: All tests must be appropriately packaged in the `test` directory at the root of the workspace.
- [x] **Consolidate tests**: `vllm_client_test.go`, `openai_client_test.go`, `types_test.go`
- [x] **Interface pattern tests**: `llm_provider_test.go`, `llm_factory_test.go`, `llm_orchestrator_test.go`
- [x] **Builder pattern tests**: `generation_request_builder_test.go`
- [x] **90%+ coverage**: LLM operations, response parsing, timeout handling, retry logic
- [x] **Provider interface tests**: Test all providers implement interface consistently
- [x] **Mock LLM tests**: Avoid external API dependencies in automated tests
- [x] **Integration tests**: Manual validation with FREE TIER services only
- [x] **Rate limiting tests**: Integration with quota and rate limit functionality
- [x] **Factory pattern tests**: Provider creation, configuration validation, error scenarios
- [x] **Legacy cleanup tests**: Remove outdated test files and update references

### Coding Standards Validation:
- [x] **File size limits**: Keep all files under 300 lines (refactor at 300+, never exceed 500)
- [x] **Function size limits**: Keep functions under 30 lines (refactor at 30+, never exceed 50)
- [x] **Single responsibility**: Each function does one thing and does it well
- [x] **Interface compliance**: All LLM providers implement interface contract consistently
- [x] **Avoid nested logic**: Split complex conditions into smaller functions
- [x] **Clear separation**: Business logic, data access, and presentation layers distinct
- [x] **Error handling**: Explicit error handling for all LLM operations and rate limiting
- [x] **Dependency injection**: Use interface-based dependency injection for testability

### Success Criteria:
✅ AI service has complete LLM ownership  
✅ Rate limiting integrated and tested  
✅ FREE TIER configuration enforced  
✅ No external LLM dependencies in automated tests  
✅ LLM interface pattern implemented across all providers  
✅ Builder pattern implemented for complex request configuration  
✅ Free tier enforcement and validation at build time  
✅ Fluent API for readable request construction  

### Version Control:
- [x] **Commit changes**: `git add . && git commit -m "feat: consolidate LLM functionality with rate limiting in AI service"`
- [x] **Validate build**: Ensure all tests pass and services compile before committing

---

## **A.3 Service-Specific Configuration Distribution**
**Goal**: Each service owns its configuration

### Implementation Steps:
- [ ] **Create service configs**:
  - [ ] `internal/user/config.go` - User service configuration
  - [ ] `internal/ai/config.go` - AI service configuration (including LLM settings)
  - [ ] `internal/auth/config.go` - Auth service configuration (JWT, OAuth, sessions)
  - [ ] `internal/gateway/config.go` - Gateway configuration (routing, proxy settings)
- [ ] **Distribute configuration**: Move relevant sections from `internal/infrastructure/config/`
- [ ] **Remove global dependencies**: Services load only required configuration
- [ ] **Environment handling**: Service-specific environment variable management
- [ ] **🏗️ IMPLEMENT CONFIGURATION INTERFACE PATTERN**: Create consistent configuration interface across services
  - [ ] Design `ConfigProvider` interface following Java-style pattern for consistent configuration loading
  - [ ] Implement interface for YAML, JSON, environment variables, and future config sources
  - [ ] Create configuration factory pattern for dynamic config source selection
  - [ ] Ensure all services follow same configuration validation and loading pattern
  - [ ] Add config-agnostic validation and environment handling

### Configuration Interface Pattern Design:
```go
// Core interface that all configuration providers must implement
type ConfigProvider interface {
    // Load configuration from source
    Load(ctx context.Context, source string) error
    
    // Get configuration values with type safety
    GetString(key string) (string, error)
    GetInt(key string) (int, error)
    GetBool(key string) (bool, error)
    GetDuration(key string) (time.Duration, error)
    
    // Validation and health
    Validate() error
    GetSource() string
    
    // Watch for changes (for hot reloading)
    Watch(ctx context.Context, callback func()) error
}

// Factory pattern for configuration provider instantiation
type ConfigFactory interface {
    CreateProvider(providerType string, source string) (ConfigProvider, error)
    ListAvailableProviders() []string
}

// Service-specific configuration managers
type AuthConfig struct {
    provider ConfigProvider
    JWT      JWTConfig      `yaml:"jwt"`
    OAuth    OAuthConfig    `yaml:"oauth"`
    Sessions SessionConfig  `yaml:"sessions"`
}

type AIConfig struct {
    provider ConfigProvider
    LLM      LLMConfig      `yaml:"llm"`
    Rate     RateConfig     `yaml:"rate_limiting"`
    Cache    CacheConfig    `yaml:"cache"`
}
```

### Configuration Interface Benefits:
- **Source Agnostic**: Load from YAML, JSON, env vars, or future config sources
- **Type Safety**: Consistent type conversion and validation across services
- **Hot Reloading**: Watch for configuration changes without service restart
- **Environment Flexibility**: Easy switching between dev, staging, production configs
- **Validation**: Standardized configuration validation patterns
- **Testing**: Mock configurations for automated testing

### Test Requirements:
- [ ] **Organization of tests**: All tests must be appropriately packaged in the `test` directory at the root of the workspace.
- [ ] **Service config tests**: `config_test.go` for each service
- [ ] **85%+ coverage**: config loading, validation, environment handling
- [ ] **Validation tests**: Invalid configuration scenarios
- [ ] **Environment tests**: Defaults and overrides

### Coding Standards Validation:
- [ ] **File size limits**: Keep all config files under 300 lines (refactor at 300+, never exceed 500)
- [ ] **Function size limits**: Keep config functions under 30 lines (refactor at 30+, never exceed 50)
- [ ] **Single responsibility**: Each config function handles one configuration aspect
- [ ] **Input validation**: Validate all configuration inputs explicitly
- [ ] **Clear naming**: Descriptive configuration variable and function names
- [ ] **No hardcoded values**: All configuration values externally configurable

### Success Criteria:
✅ Each service has independent configuration  
✅ No global config dependencies  
✅ Environment variable handling per service  
✅ Configuration validation comprehensive  
✅ Configuration interface pattern implemented across all providers  
✅ Hot reloading and source-agnostic configuration loading  

### Version Control:
- [ ] **Commit changes**: `git add . && git commit -m "feat: implement service-specific configuration distribution"`
- [ ] **Validate build**: Ensure all tests pass and services compile before committing

---

## **A.4 Database Adapters & Generation Consolidation**
**Goal**: Move service-specific components to their services

### Implementation Steps:
- [ ] **User service database**:
  - [ ] Move `internal/infrastructure/database/project_repository.go` → `internal/user/repository.go`
  - [ ] Create unified `internal/user/repository.go` with PostgreSQL implementation
- [ ] **Generation service consolidation**:
  - [ ] Move `internal/generation/generation_handlers.go` → `internal/ai/generation_handlers.go`
  - [ ] Move `internal/generation/redis_client.go` → `internal/ai/redis_client.go`
  - [ ] Move `internal/generation/service.go` → `internal/ai/generation_service.go`
- [ ] **Update imports**: Ensure no conflicts with existing AI service files
- [ ] **🏗️ IMPLEMENT REPOSITORY INTERFACE PATTERN**: Create consistent data access interface across services
  - [ ] Design `Repository` interface following Java-style pattern for consistent data operations
  - [ ] Implement interface for PostgreSQL, future database providers, and testing mocks
  - [ ] Create repository factory pattern for dynamic database selection
  - [ ] Ensure all data access follows same interface contract with proper transaction management
  - [ ] Add database-agnostic error handling and connection management
- [ ] **🏗️ IMPLEMENT TEMPLATE METHOD PATTERN**: Standardize database operation workflows
  - [ ] Design `RepositoryTemplate` interface for consistent operation algorithms
  - [ ] Implement template methods for transaction handling, batch operations, and error recovery
  - [ ] Add customizable hook methods for service-specific behavior
  - [ ] Ensure standardized logging, metrics, and error handling across all database operations

### Repository Interface Pattern Design:
```go
// Core interface that all repository implementations must implement
type Repository interface {
    // Basic CRUD operations
    Create(ctx context.Context, entity interface{}) error
    GetByID(ctx context.Context, id string, entity interface{}) error
    Update(ctx context.Context, entity interface{}) error
    Delete(ctx context.Context, id string) error
    
    // Query operations
    List(ctx context.Context, filter QueryFilter, entities interface{}) error
    Count(ctx context.Context, filter QueryFilter) (int64, error)
    
    // Transaction management
    BeginTx(ctx context.Context) (Transaction, error)
    
    // Health and maintenance
    HealthCheck(ctx context.Context) error
    Close() error
}

// Transaction interface for consistent transaction handling
type Transaction interface {
    Commit() error
    Rollback() error
    Repository() Repository
}

// Factory pattern for repository instantiation
type RepositoryFactory interface {
    CreateRepository(entityType string, config DatabaseConfig) (Repository, error)
    CreateUserRepository(config DatabaseConfig) (UserRepository, error)
    CreateProjectRepository(config DatabaseConfig) (ProjectRepository, error)
}

// Service-specific repository interfaces
type UserRepository interface {
    Repository
    GetByEmail(ctx context.Context, email string) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
    UpdateLastLogin(ctx context.Context, userID string) error
}

type ProjectRepository interface {
    Repository
    GetByUserID(ctx context.Context, userID string) ([]*Project, error)
    GetByStatus(ctx context.Context, status ProjectStatus) ([]*Project, error)
}
```

### Template Method Pattern Design for Database Operations:
```go
// Template method pattern for standardized repository operations
type RepositoryTemplate interface {
    // Template methods that define the algorithm structure
    ExecuteWithTransaction(ctx context.Context, operation TransactionOperation) error
    ExecuteQuery(ctx context.Context, query QueryOperation) (interface{}, error)
    ExecuteBatch(ctx context.Context, operations []BatchOperation) error
    
    // Hook methods for customization
    BeforeOperation(ctx context.Context, operation OperationType) error
    AfterOperation(ctx context.Context, operation OperationType, result interface{}) error
    OnError(ctx context.Context, operation OperationType, err error) error
}

// Abstract base repository with template method implementation
type BaseRepository struct {
    db      Database
    logger  Logger
    metrics MetricsCollector
}

func (b *BaseRepository) ExecuteWithTransaction(ctx context.Context, operation TransactionOperation) error {
    // Template method algorithm
    if err := b.BeforeOperation(ctx, OperationTypeTransaction); err != nil {
        return err
    }
    
    tx, err := b.db.BeginTx(ctx)
    if err != nil {
        b.OnError(ctx, OperationTypeTransaction, err)
        return err
    }
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        } else if err != nil {
            tx.Rollback()
        } else {
            err = tx.Commit()
        }
    }()
    
    // Execute the operation
    result, err := operation.Execute(ctx, tx)
    
    if err != nil {
        b.OnError(ctx, OperationTypeTransaction, err)
        return err
    }
    
    return b.AfterOperation(ctx, OperationTypeTransaction, result)
}

// Concrete repository implementations can override hook methods
type UserRepositoryImpl struct {
    BaseRepository
}

func (u *UserRepositoryImpl) BeforeOperation(ctx context.Context, operation OperationType) error {
    // User-specific pre-operation logic
    u.logger.InfoContext(ctx, "Starting user repository operation", "type", operation)
    u.metrics.IncrementCounter("user_repository_operations", map[string]string{
        "operation": string(operation),
    })
    return nil
}

func (u *UserRepositoryImpl) OnError(ctx context.Context, operation OperationType, err error) error {
    // User-specific error handling
    u.logger.ErrorContext(ctx, "User repository operation failed", "error", err, "operation", operation)
    u.metrics.IncrementCounter("user_repository_errors", map[string]string{
        "operation": string(operation),
        "error_type": getErrorType(err),
    })
    
    // Transform database-specific errors to domain errors
    return u.transformError(err)
}
```

### Repository Interface Benefits:
- **Database Agnostic**: Switch between PostgreSQL, MySQL, or future databases
- **Consistent API**: All data access follows same interface patterns
- **Transaction Safety**: Standardized transaction management across repositories
- **Easy Testing**: Mock repository implementations for automated testing
- **Future Extensibility**: Add new databases without changing service logic
- **Connection Management**: Centralized connection pooling and health checking

### Template Method Benefits:
- **Consistent Workflow**: All database operations follow same algorithm structure
- **Customizable Behavior**: Services can override hook methods for specific needs
- **Error Handling**: Standardized error handling and recovery patterns
- **Metrics Integration**: Automatic metrics collection for all operations
- **Transaction Safety**: Guaranteed transaction cleanup and error recovery
- **Logging**: Consistent operation logging across all repositories

### Test Requirements:
- [ ] **Organization of tests**: All tests must be appropriately packaged in the `test` directory at the root of the workspace.
- [ ] **Repository tests**: `internal/user/repository_test.go`
- [ ] **Template method tests**: `base_repository_test.go`
- [ ] **95%+ coverage**: CRUD operations, error cases, transaction handling
- [ ] **Integration tests**: Test database with real connections
- [ ] **Generation tests**: `generation_handlers_test.go`, `generation_service_test.go`
- [ ] **95%+ coverage**: generation workflows, streaming, error handling, rate limiting

### Coding Standards Validation:
- [ ] **File size limits**: Keep all files under 300 lines (refactor at 300+, never exceed 500)
- [ ] **Function size limits**: Keep functions under 30 lines (refactor at 30+, never exceed 50)
- [ ] **Single responsibility**: Each repository method handles one data operation
- [ ] **Clear separation**: Keep business logic separate from data access layer
- [ ] **Error handling**: Explicit error handling for all database operations
- [ ] **Transaction management**: Proper transaction scoping and cleanup

### Success Criteria:
✅ User service owns its database operations  
✅ AI service owns generation functionality  
✅ No conflicts between existing and moved components  
✅ Database connection handling robust  
✅ Repository interface pattern implemented across all data access  
✅ Transaction management standardized and database-agnostic  
✅ Template method pattern implemented for consistent operation workflows  
✅ Hook methods allow service-specific customization while maintaining standards  

### Version Control:
- [ ] **Commit changes**: `git add . && git commit -m "feat: consolidate database adapters and generation functionality"`
- [ ] **Validate build**: Ensure all tests pass and services compile before committing
