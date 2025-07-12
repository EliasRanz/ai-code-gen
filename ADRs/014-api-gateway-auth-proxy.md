# ADR-014: API Gateway Auth Proxy Implementation

## Status
Accepted - Implemented (2025-07-11)

## Context
Following the completion of Phase 1 (centralized auth service), Phase 2 required integrating the API Gateway with the centralized auth service to eliminate local JWT validation and achieve true authentication centralization.

The existing API Gateway used local JWT validation via `LightweightAuthMiddleware`, which contradicted the centralized auth goals outlined in ADR-012.

## Decision
Implemented Auth Service Proxy pattern in the API Gateway:

### 1. Auth Service Proxy Middleware
- Created `AuthServiceProxy` middleware that forwards token validation to auth service
- Implemented `AuthServiceRoleProxy` for combined token validation and role checking
- Removed local JWT validation from API Gateway completely

### 2. Request Flow Changes
```
Before: Client → Gateway [Local JWT] → Service
After:  Client → Gateway → Auth Service [Validate] → Service (with context)
```

### 3. User Context Enrichment
Gateway now enriches requests with validated user context:
- `user_id`, `user_email`, `user_role`, `authenticated` flags
- Downstream services receive clean context without auth logic

### 4. Implementation Details
- **File**: `/internal/middleware/auth_proxy.go`
- **Functions**: `AuthServiceProxy()`, `AuthServiceRoleProxy()`
- **HTTP Client**: 5-second timeout for auth service calls
- **Error Handling**: Comprehensive error logging and proper HTTP status codes

## Implementation Changes

### API Gateway (`cmd/api-gateway/main.go`)
```go
// REMOVED: Local token manager
// tokenManager := auth.NewTokenManager(cfg.Auth.JWTSecret, "ai-ui-generator")

// REPLACED: Local auth middleware
// userGroup.Use(middleware.LightweightAuthMiddleware(tokenManager))
userGroup.Use(middleware.AuthServiceProxy(authService.BaseURL))

// REPLACED: Local admin middleware  
// adminGroup.Use(middleware.LightweightAuthMiddleware(tokenManager))
// adminGroup.Use(middleware.AdminRequired())
adminGroup.Use(middleware.AuthServiceRoleProxy(authService.BaseURL, "admin"))
```

### Auth Proxy Middleware
- **Auth Validation**: HTTP POST to `/api/auth/validate`
- **Role Checking**: HTTP POST to `/api/auth/check-role`
- **Context Setting**: Enriches Gin context with user information
- **Error Handling**: Proper 401/403 responses with logging

## Test Coverage
Comprehensive test suite in `tests/unit/middleware/auth_proxy_test.go`:
- Invalid authorization headers (missing, malformed, empty)
- Valid token validation with mock auth service
- Role-based access control validation
- Error handling scenarios

## Consequences

### Positive
1. **True Centralization**: All auth decisions now made by auth service
2. **Service Simplification**: API Gateway becomes pure proxy/router
3. **Consistency**: Uniform auth behavior across all routes
4. **Maintainability**: Auth changes only require auth service updates
5. **Testability**: Clear separation enables focused testing

### Trade-offs
1. **Network Latency**: Auth service HTTP calls add ~5-50ms per request
2. **Dependency**: Gateway depends on auth service availability
3. **Complexity**: More network calls in request flow

### Mitigations
1. **Performance**: Future caching layer can reduce latency
2. **Reliability**: Auth service designed for high availability
3. **Monitoring**: Request tracing and error logging implemented

## Compliance with Goals
✅ **Single Auth Source**: All auth decisions via auth service  
✅ **No Auth Duplication**: Zero local JWT validation in gateway  
✅ **Consistent Security**: Uniform auth policies system-wide  
✅ **Clean Architecture**: Clear separation of concerns

## Next Steps
1. **Phase 3**: Remove auth logic from other services (AI, User)
2. **Performance**: Add Redis caching layer for auth results
3. **Monitoring**: Implement auth service performance metrics
4. **Testing**: End-to-end integration testing with real services

## Related
- ADR-012: Centralized Auth Server Decision
- ADR-010: Clean Architecture Migration
- Phase 1: Auth Service Implementation (Complete)
- Phase 3: Service Cleanup (Next Priority)
