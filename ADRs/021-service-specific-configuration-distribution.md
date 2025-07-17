# ADR-021: Service-Specific Configuration Distribution

**Status:** Accepted  
**Date:** 2024-01-15  
**Relates to:** [ADR-017: Eliminate Infrastructure Interfaces Abstraction](017-eliminate-infrastructure-interfaces-abstraction.md)

## Context

As part of Phase A.3 of the infrastructure consolidation effort (ADR-017), we need to migrate from centralized global configuration management to service-specific configuration ownership. This addresses several issues:

1. **Tight Coupling**: Services were dependent on a central configuration system that knew about all service configurations
2. **Single Point of Failure**: Changes to any service configuration could affect other services
3. **Configuration Complexity**: The global configuration structure was becoming increasingly complex
4. **Service Autonomy**: Services should own their configuration concerns independently

## Decision

We will implement a **Configuration Interface Pattern** with the following architectural components:

### 1. Core Configuration Interfaces

- **ConfigProvider**: Abstracts configuration sources (environment, YAML, JSON)
- **ConfigFactory**: Creates and manages configuration providers
- **ConfigManager**: Manages configuration data and validation for services
- **Service-Specific Managers**: Each service owns its configuration management

### 2. Provider Architecture

- **EnvironmentProvider**: Parses environment variables with automatic type conversion
- **YAMLProvider**: Handles YAML file configuration with hot reloading
- **JSONProvider**: Handles JSON file configuration with hot reloading
- **Extensible Registration**: New providers can be registered with the factory

### 3. Configuration Features

- **Source-Agnostic Loading**: Services can use any provider without changing code
- **Comprehensive Validation**: Type checking, range validation, pattern matching
- **Hot Reloading**: File watchers for dynamic configuration updates
- **Default Value Management**: Service-specific default application
- **Environment Variable Mapping**: Automatic conversion from env vars to config keys

### 4. Service-Specific Implementation

Each service implements its own configuration manager:
- **AuthServiceConfigManager**: Handles JWT, OAuth, security configurations
- **UserServiceConfigManager**: Manages database, pagination, validation settings
- **AIServiceConfigManager**: Controls LLM providers, rate limiting, quota management
- **GatewayServiceConfigManager**: Manages routing, CORS, auth proxy, load balancing

## Implementation Details

### Configuration Loading Flow

1. Service creates provider (environment/YAML/JSON)
2. Provider loads and parses configuration data
3. Service-specific manager maps raw data to typed configuration
4. Validation rules are applied to ensure correctness
5. Default values are applied where needed
6. Configuration is available for service use

### Environment Variable Conventions

- Service prefix: `{SERVICE}_` (e.g., `AUTH_`, `USER_`, `AI_`, `GATEWAY_`)
- Nested structure: underscores converted to dots (e.g., `AUTH_JWT_SECRET_KEY` → `jwt.secret.key`)
- Type conversion: automatic parsing of booleans, integers, durations, comma-separated lists

### Validation System

```go
type ValidationRule struct {
    Key      string
    Type     string
    Required bool
    MinValue interface{}
    MaxValue interface{}
    Pattern  string
    Custom   func(interface{}) error
}
```

### Service Manager Pattern

```go
type ServiceConfigManager interface {
    LoadConfig(ctx context.Context) error
    GetConfig() interface{}
    Watch(ctx context.Context, callback func()) error
    Reload(ctx context.Context) error
}
```

## Architectural Benefits

### 1. **Service Autonomy**
- Each service owns its configuration schema and validation
- Independent deployment and configuration changes
- Service-specific default values and business logic

### 2. **Source Flexibility**
- Services can use environment variables in containers
- Development can use YAML files for complex configurations
- Production can mix sources as needed

### 3. **Type Safety**
- Strongly typed configuration structures
- Compile-time checking of configuration usage
- Comprehensive validation with meaningful error messages

### 4. **Operational Excellence**
- Hot reloading for development and production
- Health checks for configuration validity
- Monitoring and observability support

### 5. **Testability**
- Mock providers for unit testing
- Configuration scenarios easily tested
- Validation logic independently testable

## Migration Strategy

### Phase 1: Interface Implementation ✅
- Created core configuration interfaces
- Implemented provider architecture
- Built factory and manager systems

### Phase 2: Service Integration ✅
- Migrated auth service configuration
- Migrated user service configuration
- Migrated AI service configuration
- Migrated gateway service configuration

### Phase 3: Testing & Validation ✅
- Comprehensive test suite for all providers
- Service-specific configuration testing
- Integration testing with real environment variables

### Phase 4: Documentation ✅
- Service configuration examples
- Provider usage documentation
- Migration guides for new services

## Examples

### Environment-Based Configuration

```bash
# Auth Service
export AUTH_JWT_SECRET_KEY="super-secret-key"
export AUTH_JWT_EXPIRATION_TIME="24h"
export AUTH_OAUTH_GOOGLE_CLIENT_ID="google-client-id"

# AI Service
export AI_LLM_OPENAI_API_KEY="openai-key"
export AI_LLM_OPENAI_MODEL="gpt-4"
export AI_RATE_LIMIT_REQUESTS_PER_HOUR="1000"

# Gateway Service
export GATEWAY_CORS_ALLOWED_ORIGINS="http://localhost:3000,https://example.com"
export GATEWAY_RATE_LIMIT_REQUESTS_PER_SECOND="100"
```

### Service Configuration Usage

```go
// Create and load configuration
provider := config.NewEnvironmentProvider("AUTH_")
manager := auth.NewAuthServiceConfigManager(provider)
if err := manager.LoadConfig(ctx); err != nil {
    return err
}

// Use configuration
cfg := manager.GetConfig()
jwtSecret := cfg.JWT.SecretKey
```

## Consequences

### Positive

- **Reduced Coupling**: Services are independent in configuration concerns
- **Improved Maintainability**: Configuration logic is localized to services
- **Enhanced Flexibility**: Multiple configuration sources supported
- **Better Testing**: Service configurations can be tested in isolation
- **Operational Improvements**: Hot reloading and validation enhance operations

### Negative

- **Code Duplication**: Some configuration patterns repeated across services
- **Learning Curve**: Developers need to understand the provider pattern
- **Migration Effort**: Existing services need configuration updates

### Neutral

- **Configuration Complexity**: Moved from global to distributed complexity
- **Interface Overhead**: Additional abstractions for configuration management

## Implementation Notes

### Key Technical Decisions

1. **Type Conversion Strategy**: Environment provider automatically converts types based on value patterns
2. **Default Value Timing**: Applied after configuration loading to avoid overriding valid values
3. **Validation Architecture**: Rule-based system allows for complex validation scenarios
4. **Error Handling**: Comprehensive error messages with context for debugging

### Critical Bug Fixes

- **String Slice Handling**: Fixed `GetStringSlice()` to handle both `[]string` and `[]interface{}` types
- **Environment Key Mapping**: Ensured consistent conversion from underscore to dot notation
- **Default Value Logic**: Prevented defaults from overriding parsed environment values

## Future Considerations

1. **Configuration Versioning**: Support for configuration schema evolution
2. **Distributed Configuration**: Integration with external configuration services (e.g., Consul, etcd)
3. **Configuration Encryption**: Secure handling of sensitive configuration values
4. **Dynamic Reconfiguration**: Runtime configuration updates without service restart

## Conclusion

The service-specific configuration distribution pattern successfully decouples configuration concerns while maintaining flexibility and operational excellence. The implementation provides a solid foundation for service autonomy and scalable configuration management.
