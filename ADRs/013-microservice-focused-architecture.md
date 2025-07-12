# ADR-013: Adopt Microservice-Focused Architecture with Package Consolidation

## Status
Proposed

## Context
During the implementation of the centralized auth system (ADR-012), we observed that the strict Clean Architecture domain layer separation has led to unnecessary complexity and over-abstraction in our codebase. The current domain package structure creates several issues:

### Current Problems
1. **Over-Abstraction**: Domain entities and interfaces are often thin wrappers around simple data structures
2. **Development Friction**: Developers spend significant time navigating between domain, application, and infrastructure layers for simple operations
3. **Maintenance Overhead**: Changes require updates across multiple layers even for simple modifications
4. **Code Duplication**: Similar validation and business logic scattered across domain and application layers
5. **Testing Complexity**: Complex mock hierarchies required for testing simple business operations

### Architecture Analysis
Our microservices architecture already provides natural domain boundaries at the service level:
- **Auth Service**: Authentication and authorization domain
- **User Service**: User management domain  
- **AI Service**: Code generation domain
- **API Gateway**: Routing and aggregation

The additional domain package layer within each service creates unnecessary internal complexity without significant architectural benefits.

## Decision
We will **remove the `internal/domain/` package** and **simplify the architecture** by consolidating business logic directly into service-specific packages under `internal/`.

### New Structure
```
internal/
├── auth/           # Auth business logic (consolidated from domain/auth + application/auth)
├── user/           # User business logic (consolidated from domain/user + application/user)  
├── generation/     # AI generation business logic
├── infrastructure/ # External adapters (DB, LLM, etc.)
└── interfaces/     # HTTP/gRPC handlers, middleware
```

### Business Logic Consolidation
- **Domain entities** → Move to service-specific packages with simplified structs
- **Domain interfaces** → Move to service-specific packages as needed
- **Application use cases** → Merge into service-specific business logic
- **Domain validation** → Consolidate into service business logic

## Implementation Plan

### Phase 1: Auth Package (✅ Completed)
- Consolidated `domain/auth` and `application/auth` into `internal/auth`
- Removed unnecessary abstractions while maintaining business logic separation
- Achieved significant reduction in complexity with improved testability

### Phase 2: User Package
- Consolidate `domain/user` and `application/user` into enhanced `internal/user`
- Merge user entities and use cases into cohesive business logic
- Simplify user operations and validation

### Phase 3: Generation Package  
- Consolidate `domain/generation` and `application/generation` into `internal/generation`
- Merge AI generation entities and use cases
- Simplify code generation workflow

### Phase 4: Cleanup
- Remove empty `domain/` and `application/` directories
- Update imports throughout codebase
- Update documentation and architectural diagrams

## Consequences

### Positive
1. **Reduced Complexity**: Fewer layers to navigate and maintain
2. **Improved Developer Experience**: Faster development and easier debugging
3. **Better Testability**: Simpler test setup with fewer mocks
4. **Cleaner Codebase**: Less boilerplate and duplicate code
5. **Maintainability**: Changes localized to relevant service packages
6. **Performance**: Reduced indirection and faster compilation

### Negative
1. **Less Theoretical Purity**: Deviates from strict Clean Architecture principles
2. **Potential Coupling**: Risk of mixing business logic with infrastructure concerns
3. **Refactoring Effort**: Significant work to consolidate existing code

### Mitigation Strategies
- **Clear Package Boundaries**: Maintain separation between business logic, infrastructure, and interfaces
- **Code Reviews**: Ensure business logic stays separate from infrastructure concerns
- **Testing Standards**: Maintain comprehensive test coverage during consolidation
- **Documentation**: Update architectural guidance to reflect simplified approach

## Alternatives Considered

### 1. Keep Strict Clean Architecture
**Rejected**: Current implementation proves too complex for our use case, creating more problems than it solves.

### 2. Partial Domain Removal
**Rejected**: Half-measures would create inconsistency and confusion across the codebase.

### 3. Domain-Driven Design (DDD) Approach
**Rejected**: Our microservices already provide domain boundaries; internal DDD adds unnecessary complexity.

## Success Criteria
1. **Reduced Code Complexity**: Measurable reduction in lines of code and package count
2. **Improved Test Coverage**: Simpler test setup leading to better coverage
3. **Faster Development**: Reduced time for implementing new features
4. **Maintainability**: Easier bug fixes and feature modifications
5. **Developer Satisfaction**: Improved developer experience based on team feedback

## Related ADRs
- **ADR-010**: Clean Architecture Migration (superseded by this decision)
- **ADR-012**: Centralized Auth Server (example of successful consolidation)

## Microservice-Focused Architecture

### Service-Level Domain Boundaries
This decision reinforces our **microservice-first approach** where domain boundaries are established at the service level rather than through internal package hierarchies. Each service becomes a focused unit responsible for specific business capabilities:

- **Auth Service**: Complete authentication/authorization domain
- **User Service**: User management and profile domain  
- **AI Service**: Code generation and LLM integration domain
- **API Gateway**: Request routing and aggregation

### Extending the Approach
The consolidation principle should be applied to other areas of the application:

#### Infrastructure Directory Consolidation
Similar to the auth consolidation, the `internal/infrastructure/` directory could benefit from this approach:

**Current Structure** (Over-abstracted):
```
internal/infrastructure/
├── database/
│   ├── user_repository.go
│   ├── session_repository.go
│   └── generation_repository.go
├── llm/
│   └── client.go
└── cache/
    └── redis_client.go
```

**Proposed Structure** (Service-focused):
```
internal/
├── auth/
│   ├── auth_logic.go
│   └── auth_repository.go    # Auth-specific DB operations
├── user/
│   ├── user_logic.go
│   └── user_repository.go    # User-specific DB operations
├── generation/
│   ├── generation_logic.go
│   ├── generation_repository.go
│   └── llm_client.go         # Generation-specific LLM operations
└── shared/
    ├── database.go           # Common DB connection
    └── cache.go              # Shared cache utilities
```

#### Benefits of Service-Focused Infrastructure
1. **Cohesion**: Repository logic lives with the business logic it serves
2. **Autonomy**: Each service owns its complete data access patterns
3. **Simplified Testing**: Business logic and data access tested together
4. **Reduced Coupling**: No shared repository interfaces across domains

### Microservice Principles Applied Internally
- **Single Responsibility**: Each package has one clear business purpose
- **Focused Teams**: Developers can work on complete features within one package
- **Independent Evolution**: Packages can evolve independently with minimal cross-cutting changes
- **Clear Boundaries**: Business logic, data access, and external integrations grouped by domain

## Notes
This decision represents a pragmatic approach to software architecture, prioritizing developer productivity and maintainability over theoretical purity. The auth package consolidation (ADR-012) serves as a successful proof of concept for this approach.

**Microservice Philosophy**: By treating internal packages like micro-components within each service, we achieve the benefits of microservice architecture (autonomy, focused responsibility, clear boundaries) while maintaining the simplicity of a monolithic deployment when needed.

---
**Date**: 2025-07-11  
**Proposed by**: Development Team  
**Reason**: Adopt microservice-focused architecture to reduce over-abstraction and improve developer experience
