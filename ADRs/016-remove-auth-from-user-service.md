# ADR-016: User Service Cleanup - Auth Removal + Package Consolidation

**Status**: Accepted  
**Date**: 2024-12-28

## Context

Following Phase 3.1 completion (AI Generation Service auth cleanup), Phase 3.2 requires removing authentication logic from the User Service while preserving user data management capabilities. Additionally, ADR-013 (Microservice-Focused Architecture) requires consolidating the user domain and application layers into a unified `internal/user` package.

### Current State Issues:
1. **Password Management**: `SetPassword()` and `VerifyPassword()` methods in User entity
2. **PasswordHasher Interface**: Domain interface for password hashing operations  
3. **Password Hash Storage**: `PasswordHash` field in User entity
4. **Over-Abstraction**: Separate `domain/user` and `application/user` packages create unnecessary complexity
5. **Legacy Server**: `/cmd/server/main.go` contains outdated auth infrastructure

According to the centralized auth architecture and microservice-focused approach, authentication should be handled exclusively by the Auth Service, while the User Service should focus on consolidated user data management.

## Decision

**Combined Approach**: Remove authentication logic AND consolidate user packages following ADR-013:

### Package Consolidation (ADR-013 Phase 2):
1. **Consolidate Packages**: Merge `internal/domain/user` and `internal/application/user` into enhanced `internal/user`
2. **Simplify Structure**: Remove domain/application layer separation within user service
3. **Unified Business Logic**: Combine entities, use cases, and business rules in single package

### Auth Logic Removal (Phase 3.2):
4. **Remove Password Methods**: Eliminate `SetPassword()` and `VerifyPassword()` during consolidation
5. **Remove PasswordHasher Interface**: Remove from consolidated user package
6. **Remove Password Hash Field**: Eliminate `PasswordHash` from User entity (keep in database for migration safety)
7. **Clean Legacy Code**: Remove or update legacy `/cmd/server/main.go` if needed
8. **Trust Gateway Context**: Service assumes requests are pre-authenticated by API Gateway

## Consequences

### Positive
- **Architectural Consistency**: Aligns with ADR-013 microservice-focused architecture
- **Reduced Complexity**: Eliminates both auth logic and unnecessary layer separation
- **Clear Domain Separation**: User package focuses purely on user data management
- **Simplified Development**: Single package for all user-related operations
- **Better Testability**: Unified testing approach without complex layer mocking
- **Improved Maintainability**: Single source of truth for user logic
- **Faster Development**: Reduced navigation between domain/application layers

### Negative
- **Database Migration Consideration**: Password field removal requires careful migration
- **Dependency on Auth Service**: User password changes must go through Auth Service
- **Breaking Change Potential**: Package consolidation affects existing imports
- **Refactoring Effort**: Significant consolidation work required

## Implementation Plan

### Phase 1: Package Structure Analysis
1. Analyze current `internal/domain/user` and `internal/application/user` contents
2. Identify consolidation opportunities and dependencies
3. Plan import updates across codebase

### Phase 2: Consolidation + Auth Removal  
4. Move domain entities to `internal/user` (without auth methods)
5. Move application use cases to `internal/user`
6. Merge interfaces and remove auth-related ones
7. Update user service to use consolidated package
8. Remove empty domain/application directories

### Phase 3: Import Updates & Testing
9. Update all imports from `domain/user` and `application/user` to `internal/user`
10. Verify user service builds and tests pass
11. Test integration with API Gateway auth proxy
12. Clean up legacy `/cmd/server/main.go` if needed
