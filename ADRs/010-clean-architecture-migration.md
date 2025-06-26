# Architecture Decision Record: Clean Architecture Migration

## Status
Accepted

## Context
The current project structure violates several coding standards and architecture principles:

1. **File Size Violations**: Multiple files exceed 300+ line limits
2. **Function Size Violations**: Functions exceed 40-50 line limits  
3. **Clean Architecture Violations**: Missing proper layer separation
4. **Mixed Concerns**: Handlers contain business logic, DTOs, and HTTP concerns
5. **Missing Domain Layer**: No clear domain entities or business rules
6. **Infrastructure Dependencies**: Services directly depend on external libraries

## Decision
Migrate to Clean Architecture with proper layer separation:

### New Structure
```
internal/
├── domain/          # Domain entities and business rules (no external deps)
│   ├── ai/         # AI domain logic and entities
│   ├── user/       # User domain logic and entities  
│   ├── auth/       # Auth domain logic and entities
│   └── common/     # Shared domain concepts
├── application/     # Use cases and application services
│   ├── ai/         # AI use cases
│   ├── user/       # User use cases
│   └── auth/       # Auth use cases
├── infrastructure/ # External dependencies and adapters
│   ├── database/   # Database adapters
│   ├── llm/        # LLM client adapters
│   ├── observability/ # Logging, metrics adapters
│   └── config/     # Configuration
└── interfaces/     # HTTP handlers, gRPC, CLI (thin layer)
    ├── http/       # HTTP handlers and DTOs
    ├── grpc/       # gRPC handlers
    └── middleware/ # HTTP middleware
```

### Layer Rules
1. **Domain Layer**: No external dependencies, pure business logic
2. **Application Layer**: Orchestrates domain objects, implements use cases
3. **Infrastructure Layer**: Implements interfaces defined in domain/application
4. **Interfaces Layer**: Thin adapters for external communication

### File Size Limits
- Maximum 300 lines per file (refactor at 250+)
- Maximum 50 lines per function (refactor at 40+)
- Each file has single responsibility

## Consequences
- **Positive**: Better testability, maintainability, clear separation of concerns
- **Positive**: Follows Clean Architecture principles and coding standards
- **Positive**: Easier to add new features without modifying existing code
- **Negative**: Initial refactoring effort required
- **Negative**: More files and folders to navigate

## Implementation
1. Create new directory structure
2. Migrate domain entities first
3. Create application use cases
4. Migrate infrastructure adapters
5. Thin HTTP handlers in interfaces layer
6. Update all imports and dependencies
7. Ensure all tests pass after migration
