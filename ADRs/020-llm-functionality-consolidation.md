# ADR-020: LLM Functionality Consolidation with Design Patterns

## Status
**Accepted** - 2025-01-16

## Context
As part of Phase A.2 Infrastructure Consolidation (ADR-017), we needed to consolidate all LLM functionality into the AI service while implementing robust design patterns for maintainability and extensibility. The scattered LLM components across `internal/llm/` and `internal/infrastructure/llm/` needed to be unified with proper rate limiting integration.

## Decision
We consolidate all LLM functionality into `internal/ai/llm/` with three key design patterns:

### 1. Interface Pattern for LLM Providers
```go
type LLMProvider interface {
    GenerateCode(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error)
    HealthCheck(ctx context.Context) error
    GetProviderInfo() ProviderInfo
    GetLimits() ProviderLimits
    Close() error
}
```

**Benefits:**
- Provider-agnostic orchestration (OpenAI, vLLM, future providers)
- Consistent API contracts across all implementations
- Easy testing with mock implementations
- Future extensibility without changing orchestrator

### 2. Builder Pattern for Request Configuration
```go
type GenerationRequestBuilder interface {
    SetUserID(userID string) GenerationRequestBuilder
    SetPrompt(prompt string) GenerationRequestBuilder
    SetMaxTokens(tokens int) GenerationRequestBuilder
    Build() (*GenerationRequest, error)
}
```

**Benefits:**
- Fluent API for complex request construction
- Built-in validation at each step
- Automatic free tier enforcement and defaults
- Immutable request objects for thread safety

### 3. Factory Pattern for Provider Management
```go
type LLMFactory interface {
    CreateProvider(providerType string, config ProviderConfig) (LLMProvider, error)
    ListAvailableProviders() []string
}
```

**Benefits:**
- Dynamic provider instantiation based on configuration
- Consistent provider configuration and validation
- Centralized free tier enforcement
- Easy addition of new providers

## Free Tier Enforcement
**Critical Security Requirement:** All LLM providers are configured for FREE TIER ONLY:
- OpenAI: Mock responses for development/testing
- vLLM: Local/self-hosted instances only
- Factory enforces `FreeTierOnly: true` validation
- Builder pattern validates prompt length and token limits

## Rate Limiting Integration
Seamless integration with existing `internal/ai/rate_limit.go`:
- `LLMOrchestrator` uses adapter pattern for rate limiter and quota manager
- Request validation before LLM calls
- Token usage tracking after successful generation
- Consistent error handling for rate limit and quota exceeded

## File Organization
```
internal/ai/llm/
├── types.go           # Core interfaces and types
├── builder.go         # Builder pattern implementation
├── factory.go         # Factory pattern implementation
├── orchestrator.go    # Multi-provider orchestration
├── openai_client.go   # OpenAI provider implementation
├── vllm_client.go     # vLLM provider implementation
└── legacy/           # Moved original files
```

## Testing Strategy
- **90%+ test coverage** with unit and integration tests
- **Mock implementations** avoid external API dependencies
- **Provider interface tests** ensure consistent implementation
- **Builder pattern tests** validate fluent API and validation
- **Factory pattern tests** verify provider creation and configuration

## Consequences

### Positive
- **Unified LLM Management**: All LLM functionality owned by AI service
- **Provider Extensibility**: Easy to add new LLM providers (Anthropic, etc.)
- **Type Safety**: Strong typing with proper validation
- **Rate Limiting**: Integrated protection against API abuse
- **Free Tier Safety**: Prevents accidental paid API usage
- **Maintainability**: Clear separation of concerns with design patterns
- **Testability**: High coverage with mock implementations

### Negative
- **Complexity**: Additional abstraction layers
- **Migration Effort**: Required moving and refactoring existing LLM code
- **Learning Curve**: Team needs to understand new patterns

### Risks
- **Over-Engineering**: Design patterns add complexity for simple use cases
- **Abstraction Penalty**: Small performance overhead from interfaces

## Compliance
- **Coding Standards**: All files <300 lines, functions <30 lines
- **Error Handling**: Explicit error handling throughout
- **Testing**: Comprehensive test coverage in `tests/ai/` directory
- **Free Tier**: Enforced at factory and client levels

## Migration Notes
- Original LLM files moved to `internal/ai/llm/legacy/`
- Existing AI service updated to use new orchestrator
- Legacy compatibility maintained through adapter patterns
- Configuration consolidated in `internal/ai/config.go`

---

This ADR establishes the foundation for scalable, maintainable LLM functionality with proper design patterns and free tier enforcement.
