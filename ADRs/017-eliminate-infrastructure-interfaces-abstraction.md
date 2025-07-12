# ADR-017: Eliminate Infrastructure and Interfaces Abstraction

## Status
Proposed

## Context
Following the successful implementation of ADR-013 (Microservice-Focused Architecture with Package Consolidation), we've identified that the `internal/infrastructure` and `internal/interfaces` directories represent unnecessary abstraction layers that conflict with our consolidated service approach.

### Current Over-Abstraction Issues
1. **Infrastructure Directory** (`internal/infrastructure/`) splits service-specific adapters across artificial boundaries
2. **Interfaces Directory** (`internal/interfaces/`) separates HTTP/gRPC handlers from their domain services
3. **Package Coupling** creates unnecessary dependencies between services through shared infrastructure
4. **Development Friction** requires developers to navigate multiple directories for single features

### Current Structure Analysis
```
internal/
├── infrastructure/           # OVER-ABSTRACTED
│   ├── config/              # Service-specific configs - distribute to services
│   ├── database/            # Service-specific adapters
│   ├── llm/                 # AI service specific
│   ├── observability/       # Shared - should remain
│   └── validation/          # Service-specific - distribute to services
├── interfaces/              # OVER-ABSTRACTED
│   ├── http/                # Service-specific handlers
│   └── grpc/                # Service-specific handlers
├── domain/                  # OVER-ABSTRACTED (legacy from clean arch)
│   ├── ai/                  # Should be in internal/ai/
│   └── common/              # Should be in internal/utilities/
├── cache/                   # SHARED SERVICE
│   └── auth_cache.go        # Expand to shared cache service
├── database/                # SHARED UTILITY
│   ├── database.go          # Connection utilities - keep shared
│   └── postgres.go          # PostgreSQL utilities - keep shared
├── observability/           # SHARED UTILITY  
│   ├── logging.go           # Keep shared - unified logging format
│   ├── metrics.go           # Keep shared - centralized metrics
│   └── tracing.go           # Keep shared - distributed tracing
├── generation/              # SERVICE-SPECIFIC
│   ├── generation_handlers.go # Should be in internal/ai/
│   ├── redis_client.go      # Should be in internal/ai/
│   └── service.go           # Should be in internal/ai/
├── llm/                     # SERVICE-SPECIFIC
│   ├── types.go             # Should be in internal/ai/llm/
│   ├── vllm_client.go       # Should be in internal/ai/llm/
│   ├── vllm_helpers.go      # Should be in internal/ai/llm/
│   └── vllm_types.go        # Should be in internal/ai/llm/
├── middleware/              # GATEWAY-SPECIFIC
│   ├── auth_proxy.go        # Should be in internal/gateway/
│   ├── logging.go           # Should be in internal/gateway/
│   └── ratelimit.go         # Should be in internal/gateway/
├── proxy/                   # GATEWAY-SPECIFIC
│   └── proxy.go             # Should be in internal/gateway/
├── service/                 # UNCLEAR - needs analysis
│   └── service.go           # Determine ownership
├── tests/                   # MOVE TO ROOT
│   └── test_helpers.go      # Move to /tests/ (project root)
├── user/                    # GOOD - consolidated
├── ai/                      # GOOD - consolidated
├── auth/                    # GOOD - consolidated
└── utilities/               # GOOD - shared types only
```

## Decision
Eliminate over-abstracted directories by moving service-specific components into their respective service packages, while consolidating shared utilities and eliminating redundant domain abstractions.

### Migration Strategy

#### Phase 1: Move Service-Specific Infrastructure and Components
- `internal/infrastructure/database/project_repository.go` → `internal/user/repository.go`
- `internal/infrastructure/llm/` → `internal/ai/llm/` (merge with existing)
- `internal/infrastructure/config/` → service-specific config files in each service
- `internal/infrastructure/validation/` → service-specific validation in each service
- Expand `internal/cache/` into shared cache service (keep centralized for scalability)
- `internal/generation/` → `internal/ai/` (merge generation handlers and services)
- `internal/llm/` → `internal/ai/llm/` (consolidate all LLM functionality)
- Service-specific database adapters move to their service packages

#### Phase 2: Move Service-Specific Interfaces and Middleware
- `internal/interfaces/http/user_handler.go` → `internal/user/http_handler.go`
- `internal/interfaces/http/ai_handler.go` → `internal/ai/http_handler.go`
- `internal/interfaces/http/auth_handler.go` → `internal/auth/http_handler.go`
- `internal/interfaces/grpc/user_server.go` → `internal/user/grpc_server.go`
- `internal/interfaces/http/router.go` → `internal/gateway/router.go`
- `internal/middleware/` → `internal/gateway/` (flatten middleware into gateway)
- `internal/proxy/` → `internal/gateway/`

#### Phase 3: Eliminate Legacy Domain Abstractions
- `internal/domain/ai/` → `internal/ai/` (merge with existing AI service)
- `internal/domain/common/` → `internal/utilities/` (merge with existing utilities)
- Remove empty `internal/domain/` directory

#### Phase 4: Consolidate Shared Infrastructure and Move Tests
- Keep `internal/observability/` as shared (unified logging, metrics, tracing)
- Keep `internal/database/` as shared database utilities
- `internal/tests/` → `/tests/` (move to project root)
- Expand `internal/cache/` into comprehensive shared cache service
- Analyze and place `internal/service/` appropriately
- Remove `internal/config/` and `internal/validation/` (distributed to services)

### Target Structure
```
internal/
├── user/                    # Complete user service
│   ├── entities.go         # Domain entities
│   ├── services.go         # Business logic
│   ├── repository.go       # Database adapter (was infrastructure)
│   ├── validation.go       # User-specific validation (was infrastructure/validation)
│   ├── config.go           # User service configuration
│   ├── http_handler.go     # HTTP endpoints (was interfaces)
│   └── grpc_server.go      # gRPC endpoints (was interfaces)
├── ai/                      # Complete AI service
│   ├── entities.go         # Domain entities (+ merged from domain/ai)
│   ├── generate_code.go    # Business logic
│   ├── stream_code.go      # Streaming logic
│   ├── repository.go       # Database adapter
│   ├── validation.go       # AI-specific validation
│   ├── config.go           # AI service configuration (including LLM config)
│   ├── http_handler.go     # HTTP endpoints (was interfaces)
│   ├── generation_handlers.go # Generation logic (was internal/generation)
│   └── llm/                # Complete LLM functionality (consolidated from multiple sources)
│       ├── types.go        # LLM types and interfaces
│       ├── vllm_client.go  # vLLM client implementation
│       ├── vllm_helpers.go # vLLM utilities
│       ├── vllm_types.go   # vLLM-specific types
│       └── openai_client.go # OpenAI client (was infrastructure/llm)
├── auth/                    # Complete auth service
│   ├── middleware.go       # Auth middleware
│   ├── services.go         # Auth business logic
│   ├── repository.go       # Session/token storage
│   ├── validation.go       # Auth-specific validation
│   ├── config.go           # Auth service configuration
│   └── http_handler.go     # Auth endpoints (was interfaces)
├── gateway/                 # API Gateway service
│   ├── router.go           # Main routing (was interfaces/http)
│   ├── proxy.go            # Proxy logic (was internal/proxy)
│   ├── auth_proxy.go       # Auth proxy middleware (was middleware/)
│   ├── logging.go          # Request logging (was middleware/)
│   ├── ratelimit.go        # Rate limiting (was middleware/)
│   └── config.go           # Gateway service configuration
├── utilities/               # Shared types only (+ merged from domain/common)
├── cache/                   # Shared cache service (expanded from auth-only)
├── database/                # Shared database utilities (keep existing)
└── observability/           # Shared logging/metrics/tracing (keep as shared)
```

### Root Level Changes
```
/tests/                      # Test utilities (moved from internal/tests)
```
```

## Consequences

### Positive
- **Complete Service Ownership**: Each service owns its entire stack (domain, adapters, handlers, config, validation)
- **Reduced Coupling**: Services are truly independent with clear boundaries
- **Simplified Development**: All related code lives in one place
- **Better Testing**: Service-specific tests can cover the complete stack
- **Clearer Dependencies**: Import paths clearly show service relationships
- **Shared Infrastructure Benefits**: Centralized cache, database connections, and observability
- **Lower Operational Cost**: Shared database instance with schema isolation
- **Service-Specific Configuration**: Each service manages its own config without global coupling
- **Focused Validation**: Domain-specific validation logic co-located with business rules

### Negative
- **Some Code Duplication**: Database connection patterns may be repeated
- **Migration Effort**: Requires updating imports across the codebase
- **Temporary Disruption**: Build will break during migration phases

### Mitigation
- **Shared Utilities**: Keep truly common patterns in dedicated packages
- **Interface Standards**: Establish clear interface patterns for HTTP/gRPC handlers
- **Migration Scripts**: Automate import path updates where possible

## Implementation Plan

### Week 1: Phase 1 - Infrastructure Migration
1. Move database adapters to service packages
2. Move LLM clients to AI service
3. Update imports and test compilation

### Week 2: Phase 2 - Interfaces Migration  
1. Move HTTP handlers to service packages
2. Move gRPC servers to service packages
3. Consolidate routing in gateway package
4. Update all handler references

### Week 3: Phase 3 - Shared Infrastructure
1. Move shared infrastructure to top-level packages
2. Update configuration management
3. Ensure observability works across services
4. Final testing and documentation updates

### Week 4: Cleanup and Documentation
1. Remove empty directories
2. Update ADRs and documentation
3. Update build scripts and CI/CD
4. Create migration guide for developers

## Acceptance Criteria
- [ ] No `internal/infrastructure` directory exists
- [ ] No `internal/interfaces` directory exists  
- [ ] Each service package contains its complete implementation
- [ ] Shared utilities remain in dedicated packages
- [ ] All tests pass and imports are resolved
- [ ] Documentation reflects new structure
- [ ] CI/CD pipeline works with new structure

## Related ADRs
- ADR-013: Microservice-Focused Architecture with Package Consolidation (builds upon)
- ADR-016: Remove Auth from User Service (alignment with service boundaries)

## Date
2025-01-12
