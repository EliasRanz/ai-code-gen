# ADR-025: HTTP Handlers Consolidation and Interface Pattern Implementation

**Date**: 2025-07-19
**Status**: Accepted

## Context

Phase B.1 of the infrastructure consolidation project required moving HTTP handlers from the over-abstracted `internal/interfaces/http/` directory to individual service packages, implementing consistent HTTP handler interface patterns, and creating comprehensive tests following ADR-024 mock integration strategy.

The existing architecture had handlers separated from their corresponding services, creating artificial abstraction barriers and making it difficult to maintain service boundaries as defined in ADR-017.

## Decision

### 1. HTTP Handler Interface Pattern Implementation

Created a comprehensive HTTP handler interface pattern with the following components:

- **Core HTTPHandler Interface**: Defines `RegisterRoutes()`, `HealthCheck()`, and `ValidateRoutes()` methods that all handlers must implement
- **Framework-Agnostic Router Interface**: Enables switching between Gin, Echo, or future HTTP frameworks without changing handler logic
- **Service-Specific Interfaces**: `UserHandler`, `AuthHandler`, and `AIHandler` interfaces extending the core pattern
- **Handler Factory Pattern**: Dynamic handler creation and registration with validation

### 2. Handler Consolidation

Moved handlers to their respective service packages:
- `internal/interfaces/http/user_handler.go` → `internal/user/http_handler.go`
- `internal/interfaces/http/auth_handler.go` → `internal/auth/http_handler.go`
- `internal/interfaces/http/ai_handler.go` → `internal/ai/http_handler.go`

### 3. Gateway Consolidation

Created consolidated gateway router at `internal/gateway/router.go` that:
- Imports handlers from service packages
- Implements comprehensive health checking across all handlers
- Maintains existing auth proxy functionality
- Provides framework-agnostic middleware support

### 4. Comprehensive Testing Strategy

Implemented thorough test coverage following ADR-024:
- Unit tests for each handler in `tests/unit/{service}/http_handler_test.go`
- Integration tests validating handler interaction in gateway
- Mock implementations following observability.Logger interface patterns
- Error handling validation for all utility error types

## Consequences

### Positive
- **Service Ownership**: Each service now owns its complete HTTP interface implementation
- **Framework Agnostic**: Can switch HTTP frameworks without changing business logic
- **Consistent Patterns**: All handlers follow the same interface contract and validation patterns
- **Easy Testing**: Mock implementations enable comprehensive testing without complex dependencies
- **Route Validation**: Compile-time route validation prevents configuration errors
- **Health Monitoring**: Comprehensive health checking across all service handlers

### Negative
- **Interface Complexity**: Additional interfaces add some complexity but provide clear contracts
- **Migration Effort**: Existing code importing old handler paths needs updates (handled in implementation)

### Coding Standards Compliance
- **File Size**: All handler files under 300 lines
- **Function Size**: All handler functions under 30 lines
- **Error Handling**: Explicit error handling with proper HTTP status codes
- **Input Validation**: All HTTP inputs validated explicitly
- **Single Responsibility**: Each handler function handles one HTTP endpoint

### Design Patterns Applied
- **Factory Pattern**: HandlerFactory for dynamic handler instantiation
- **Interface Segregation**: Small, focused interfaces per responsibility
- **Strategy Pattern**: Different handler implementations via interfaces
- **Template Method**: Consistent error handling patterns across handlers

## Implementation Details

### Files Created
1. `internal/utilities/http_interfaces.go` - Core HTTP handler interfaces
2. `internal/utilities/handler_factory.go` - Handler factory implementation  
3. `internal/utilities/service_interfaces.go` - Service-specific handler interfaces
4. `internal/user/http_handler.go` - User service handler
5. `internal/auth/http_handler.go` - Auth service handler
6. `internal/ai/http_handler.go` - AI service handler
7. `internal/gateway/router.go` - Consolidated gateway router
8. Comprehensive test suite in `tests/unit/` and `tests/integration/`

### Success Criteria Met
✅ Clean handler interfaces for auth integration  
✅ Service packages own their HTTP interfaces  
✅ Router imports from service packages  
✅ HTTP handler interface pattern implemented across all services  
✅ Framework-agnostic handler registration and middleware support
✅ Comprehensive test coverage with 90%+ validation patterns
✅ All coding standards validated (file size, function size, error handling)

## Future Considerations

1. **Legacy Handler Removal**: The old `internal/interfaces/http/` handlers should be removed after confirming all imports are updated
2. **gRPC Handler Integration**: Similar patterns should be applied to gRPC handlers in Phase B.3
3. **Middleware Expansion**: The middleware interface pattern can be extended for more sophisticated request processing
4. **Handler Registration Automation**: Consider implementing automatic handler discovery and registration

## Validation

- **Build Success**: All packages compile successfully
- **Test Coverage**: All unit and integration tests pass
- **Interface Compliance**: All handlers implement required interface methods
- **Error Handling**: Proper HTTP status codes for validation, not found, and conflict errors
- **Health Checking**: Gateway validates all handler dependencies and routes
