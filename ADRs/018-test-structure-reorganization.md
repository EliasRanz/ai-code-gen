# ADR-018: Test Structure Reorganization for Microservice Architecture

## Status
Accepted

## Context
Following the microservice-focused architecture migration (ADR-013) and infrastructure consolidation (ADR-017), our test structure needs to be reorganized to match the new service boundaries and package structure.

## Current Issues
- Tests reference old domain/application packages that have been consolidated
- Test organization follows old layered architecture instead of service boundaries
- Some tests reference moved infrastructure components
- Missing test coverage for new consolidated packages

## Decision
Reorganize tests to align with microservice boundaries:

### New Test Structure
```
tests/
├── unit/
│   ├── auth/           # Auth service tests
│   ├── user/           # User service tests  
│   ├── ai/             # AI service tests
│   ├── utilities/      # Shared utilities tests
│   └── integration/    # Cross-service integration tests
├── performance/
└── fixtures/
```

### Migration Plan
1. **Auth Service Tests**: Move and update auth-related tests
2. **User Service Tests**: Move and consolidate user-related tests  
3. **AI Service Tests**: Move AI-related tests
4. **Infrastructure Tests**: Move to appropriate service or utilities
5. **Remove Legacy**: Delete obsolete test files

## Consequences
- **Positive**: Test structure matches service boundaries
- **Positive**: Easier to run service-specific test suites
- **Positive**: Better test isolation between services
- **Negative**: Requires updating CI/CD pipelines
- **Negative**: Some test reorganization effort required

## Date
July 12, 2025
