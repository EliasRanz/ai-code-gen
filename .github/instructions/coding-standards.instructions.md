# Coding Standards & Best Practices

## Purpose
Defines coding standards, best practices, and style guidelines for the AI UI Generator project to ensure consistent, maintainable, and high-quality code across Go microservices and TypeScript frontend.

## Function/Method Guidelines

### Size and Complexity
- **Function Length**: Keep functions under 50 lines of code (excluding comments/whitespace)
- **Method Complexity**: Maximum cyclomatic complexity of 10 per function
- **Parameters**: Limit function parameters to 5 or fewer; use structs for complex parameter sets
- **Single Responsibility**: Each function should do one thing well
- **Pure Functions**: Prefer pure functions where possible (no side effects)

### Examples
✅ **Good**: 
```go
func ValidateUser(user User) error {
    if user.Email == "" {
        return errors.New("email required")
    }
    return nil
}
```

❌ **Bad**: 
```go
func ProcessUserDataAndSendEmailAndLogAndUpdateDatabase(user User, db DB, logger Logger, emailSvc EmailService) error {
    // 150+ lines of mixed responsibilities
}
```

## Structure/Class Guidelines (Go Structs)

### Organization
- **Struct Size**: Keep structs focused with maximum 15 fields
- **Interface Segregation**: Small, focused interfaces (3-5 methods maximum)
- **Composition**: Favor composition over inheritance patterns
- **Naming**: Use clear, descriptive names following Go conventions (PascalCase for exported, camelCase for internal)

### Examples
✅ **Good**:
```go
type UserRepository interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
}
```

❌ **Bad**:
```go
type MegaUserManager interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    SendEmail(ctx context.Context, email Email) error
    LogActivity(ctx context.Context, activity Activity) error
    ProcessPayment(ctx context.Context, payment Payment) error
    // ... 10+ more unrelated methods
}
```

## File Organization

### Structure Standards
- **File Length**: Maximum 500 lines per file (excluding generated code)
- **Package Cohesion**: Group related functionality in packages
- **Import Groups**: Organize imports in standard groups (stdlib, external, internal)
- **Documentation**: All exported functions/types must have comments

### Import Organization
```go
// Standard library
import (
    "context"
    "fmt"
)

// External dependencies
import (
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

// Internal packages
import (
    "github.com/ai-code-gen/internal/user"
    "github.com/ai-code-gen/internal/auth"
)
```

## Language-Specific Standards

### Go Guidelines
- **Follow effective Go**: Use `gofmt`, respect Go idioms
- **Error Handling**: Go-style explicit error handling, avoid panic/recover except in truly exceptional cases
- **Context**: Always pass context.Context as first parameter for cancellation and timeouts
- **Goroutines**: Use goroutines responsibly with proper synchronization

### TypeScript Guidelines
- **Strict Mode**: Use TypeScript strict mode with explicit types
- **ESLint**: Follow ESLint rules configured in the project
- **Type Safety**: Avoid `any` type, use proper type definitions
- **React Patterns**: Follow React best practices and hooks patterns

## Naming Conventions

### Go Naming
- **Variables**: camelCase for unexported, PascalCase for exported
- **Functions**: PascalCase for exported, camelCase for unexported
- **Constants**: ALL_CAPS for package-level constants
- **Interfaces**: Often end with -er (e.g., `UserRepository`, `EmailSender`)

### TypeScript Naming
- **Variables/Functions**: camelCase
- **Types/Interfaces**: PascalCase
- **Constants**: SCREAMING_SNAKE_CASE
- **Files**: kebab-case for component files

## Documentation Standards

### Required Documentation
- All exported Go functions and types must have comments
- Complex algorithms should have inline comments explaining the approach
- API endpoints should have clear documentation
- README files for each major package

### Comment Style
```go
// GetUser retrieves a user by ID from the database.
// Returns an error if the user is not found or if there's a database connection issue.
func GetUser(ctx context.Context, id string) (*User, error) {
    // Implementation
}
```

## Code Quality Tools

### Mandatory Tools
- **Go**: `gofmt`, `golangci-lint`, `gosec`
- **TypeScript**: `eslint`, `prettier`, TypeScript compiler
- **Integration**: All tools configured in project and run via `make lint`

### Pre-commit Requirements
- All code must pass linting before commit
- Use `make ci` to validate complete pipeline locally
- Address all security warnings from `gosec`