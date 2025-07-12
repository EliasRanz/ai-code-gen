# ADR-015: Remove Local Auth Logic from AI Generation Service

**Status**: Accepted  
**Date**: 2024-12-28

## Context

The AI Generation Service currently implements local JWT validation logic, which contradicts the centralized authentication architecture established in ADR-012 and implemented in Phase 2. This creates several issues:

1. **Architectural Inconsistency**: Service performs local auth validation instead of trusting the API Gateway
2. **Code Duplication**: Auth logic is duplicated across services instead of centralized
3. **Maintenance Overhead**: Multiple services need updates when auth logic changes
4. **Security Concerns**: Multiple JWT validation points increase attack surface

According to the centralized auth plan Phase 3, services should be auth-agnostic and trust the authentication context provided by the API Gateway.

## Decision

Remove all local authentication logic from the AI Generation Service and make it auth-agnostic:

1. **Remove Local JWT Validation**: Eliminate `tokenProvider`, `tokenValidator`, and `MinimalUserRepository` from service startup
2. **Remove Auth Middleware**: Replace `auth.AuthMiddleware` with trust-based user context extraction
3. **Trust Gateway Context**: Service assumes requests are pre-authenticated by the API Gateway
4. **Simplified Route Registration**: Remove auth dependencies from `RegisterRoutes` function
5. **Context-Based User Data**: Extract user information from request context set by gateway

## Consequences

### Positive
- **Simplified Service**: Removes auth complexity from generation service
- **Consistent Architecture**: Aligns with centralized auth design
- **Reduced Dependencies**: Eliminates auth-related dependencies and repositories
- **Better Separation of Concerns**: Service focuses solely on AI generation logic
- **Easier Testing**: Service can be tested without auth setup

### Negative
- **Gateway Dependency**: Service security depends entirely on API Gateway auth proxy
- **Context Trust Model**: Must trust that gateway provides correct user context
- **Migration Complexity**: Requires coordinated deployment with gateway changes

## Implementation Plan

1. Update route registration to remove auth middleware and dependencies
2. Replace auth context extraction with gateway-provided context
3. Remove local auth infrastructure from service startup
4. Update service to trust pre-authenticated requests
5. Ensure comprehensive testing covers gateway integration scenarios
