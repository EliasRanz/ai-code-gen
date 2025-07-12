# Architecture Decision Record: Centralized Authentication Server

**Status**: Proposed  
**Date**: 2025-01-11  
**Context**: Multiple services currently have their own auth implementations

## Context

Currently, the system has distributed authentication logic across multiple services:

### Current Auth Distribution:
1. **Auth Service** (`/cmd/auth-service/`): Dedicated auth service with login/logout/refresh
2. **API Gateway** (`/cmd/api-gateway/`): JWT validation using `LightweightAuthMiddleware`
3. **AI Generation Service** (`/cmd/ai-generation-service/`): JWT validation using `TokenManager`
4. **User Service** (`/cmd/user-service/`): User management with auth context
5. **Frontend** (`/web/`): NextAuth.js with custom providers
6. **Shared Auth Components**:
   - `/internal/auth/`: Token managers and middleware
   - `/internal/middleware/auth.go`: Multiple auth middleware variants
   - `/internal/application/auth/`: Auth use cases
   - `/internal/domain/auth/`: Auth domain logic

### Problems Identified:
- **Auth Logic Duplication**: Multiple services implement JWT validation
- **Inconsistent Security**: Different auth middleware implementations
- **Complex Token Management**: Each service manages tokens differently
- **Multiple Sources of Truth**: User context scattered across services
- **Maintenance Overhead**: Auth updates require changes in multiple services

## Decision

Implement a **Single Central Auth Server** architecture where:

1. **Auth Service becomes the ONLY source of authentication/authorization**
2. **All other services delegate auth decisions to the central auth server**
3. **API Gateway uses auth service for validation instead of local JWT validation**
4. **Services never perform local auth logic or token validation**

## Proposed Architecture

### Central Auth Server Responsibilities:
- User authentication (login/logout)
- Token generation and validation
- Session management
- Role-based access control (RBAC)
- User context resolution
- OAuth provider integration
- Password management

### Service Communication Pattern:
```
Fronetnd -> API Gateway -> Auth Service (for validation) -> Target Service
```

### Auth Flow:
1. **Authentication**: Client authenticates with Auth Service directly
2. **Authorization**: All requests include auth token
3. **Validation**: API Gateway forwards auth validation to Auth Service
4. **Context Propagation**: Auth Service returns user context to requesting service

### New Service Structure:
- **Auth Service**: Complete auth logic, user management, token validation
- **API Gateway**: Thin auth proxy - forwards validation requests to Auth Service
- **Other Services**: Auth-agnostic - receive validated user context from gateway
- **Frontend**: Single auth provider pointing to Auth Service

## Implementation Plan

### Phase 1: Enhanced Auth Service
1. Expand auth service to handle all auth scenarios
2. Add user context resolution endpoints
3. Implement comprehensive RBAC
4. Add OAuth provider support

### Phase 2: Gateway Restructure  
1. Replace local JWT validation with auth service calls
2. Implement auth context forwarding
3. Add auth caching for performance

### Phase 3: Service Cleanup
1. Remove auth logic from AI Generation Service
2. Remove auth logic from User Service (keep user data management)
3. Consolidate auth middleware

### Phase 4: Frontend Integration
1. Update NextAuth.js to use centralized auth service
2. Implement proper token refresh flows
3. Update all auth API calls

## Consequences

### Positive:
- **Single Source of Truth**: All auth logic centralized
- **Consistent Security**: Uniform auth policies across all services  
- **Easier Maintenance**: Auth updates in one place
- **Better Auditability**: Centralized auth logging and monitoring
- **Improved Scalability**: Auth service can be scaled independently

### Negative:
- **Single Point of Failure**: Auth service becomes critical dependency
- **Network Latency**: Additional network calls for auth validation
- **Migration Effort**: Requires refactoring multiple services

### Mitigations:
- **High Availability**: Deploy auth service with redundancy
- **Caching**: Implement auth result caching in API Gateway
- **Circuit Breaker**: Graceful degradation when auth service unavailable
- **Monitoring**: Comprehensive auth service monitoring and alerting
