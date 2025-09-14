package gateway

import (
	"fmt"
	"sort"
	"sync"

	"github.com/EliasRanz/ai-code-gen/internal/cache"
)

// BasicMiddlewareConfig provides basic middleware configuration
type BasicMiddlewareConfig struct {
	name       string
	enabled    bool
	parameters map[string]interface{}
}

// GetName returns the middleware name
func (c *BasicMiddlewareConfig) GetName() string {
	return c.name
}

// IsEnabled returns whether middleware is enabled
func (c *BasicMiddlewareConfig) IsEnabled() bool {
	return c.enabled
}

// GetParameters returns middleware parameters
func (c *BasicMiddlewareConfig) GetParameters() map[string]interface{} {
	return c.parameters
}

// NewBasicMiddlewareConfig creates a basic middleware configuration
func NewBasicMiddlewareConfig(name string, enabled bool, parameters map[string]interface{}) *BasicMiddlewareConfig {
	if parameters == nil {
		parameters = make(map[string]interface{})
	}

	return &BasicMiddlewareConfig{
		name:       name,
		enabled:    enabled,
		parameters: parameters,
	}
}

// MiddlewareChainImpl implements the middleware chain
type MiddlewareChainImpl struct {
	middlewares []Middleware
	mu          sync.RWMutex
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain() *MiddlewareChainImpl {
	return &MiddlewareChainImpl{
		middlewares: make([]Middleware, 0),
	}
}

// Add adds middleware to the chain in order priority
func (c *MiddlewareChainImpl) Add(middleware Middleware) MiddlewareChain {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.middlewares = append(c.middlewares, middleware)

	// Sort by order priority
	sort.Slice(c.middlewares, func(i, j int) bool {
		return c.middlewares[i].GetOrder() < c.middlewares[j].GetOrder()
	})

	return c
}

// Execute runs all middleware in the chain
func (c *MiddlewareChainImpl) Execute(ctx Context) error {
	c.mu.RLock()
	middlewares := make([]Middleware, len(c.middlewares))
	copy(middlewares, c.middlewares)
	c.mu.RUnlock()

	return c.executeChain(ctx, middlewares, 0)
}

// executeChain recursively executes middleware chain
func (c *MiddlewareChainImpl) executeChain(ctx Context, middlewares []Middleware, index int) error {
	if index >= len(middlewares) {
		return nil // Chain completed successfully
	}

	currentMiddleware := middlewares[index]

	// Skip disabled middleware
	if !currentMiddleware.GetConfig().IsEnabled() {
		return c.executeChain(ctx, middlewares, index+1)
	}

	// Create next function for current middleware
	next := func() error {
		return c.executeChain(ctx, middlewares, index+1)
	}

	return currentMiddleware.Process(ctx, next)
}

// GetMiddleware returns all middleware in the chain
func (c *MiddlewareChainImpl) GetMiddleware() []Middleware {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Middleware, len(c.middlewares))
	copy(result, c.middlewares)
	return result
}

// DefaultMiddlewareFactory creates middleware instances
type DefaultMiddlewareFactory struct {
	authServiceURL string
	authCache      *cache.AuthCache
	mu             sync.RWMutex
}

// NewMiddlewareFactory creates a new middleware factory
func NewMiddlewareFactory(authServiceURL string, authCache *cache.AuthCache) *DefaultMiddlewareFactory {
	return &DefaultMiddlewareFactory{
		authServiceURL: authServiceURL,
		authCache:      authCache,
	}
}

// CreateMiddleware creates a middleware instance by type
func (f *DefaultMiddlewareFactory) CreateMiddleware(middlewareType string, config MiddlewareConfig) (Middleware, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	switch middlewareType {
	case "auth-proxy":
		return NewAuthProxyMiddleware(f.authServiceURL, f.authCache), nil
	case "logging":
		return NewLoggingMiddleware(), nil
	case "metrics":
		return NewMetricsMiddleware(), nil
	case "rate-limit":
		// Get parameters from config
		params := config.GetParameters()
		rps := 10
		burst := 5

		if r, ok := params["requests_per_second"].(int); ok {
			rps = r
		}
		if b, ok := params["burst"].(int); ok {
			burst = b
		}

		return NewRateLimitMiddleware(rps, burst), nil
	default:
		return nil, fmt.Errorf("unknown middleware type: %s", middlewareType)
	}
}

// CreateChain creates a middleware chain from a list of middleware
func (f *DefaultMiddlewareFactory) CreateChain(middlewares []Middleware) MiddlewareChain {
	chain := NewMiddlewareChain()

	for _, middleware := range middlewares {
		chain.Add(middleware)
	}

	return chain
}

// ListAvailableMiddleware returns list of available middleware types
func (f *DefaultMiddlewareFactory) ListAvailableMiddleware() []string {
	return []string{
		"auth-proxy",
		"logging",
		"rate-limit",
	}
}
