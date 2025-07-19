package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Context represents a request context with common fields
type Context interface {
	Request() *http.Request
	Writer() http.ResponseWriter
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	ClientIP() string
	GetHeader(key string) string
}

// Next represents the next middleware in the chain
type Next func() error

// MiddlewareConfig holds configuration for middleware
type MiddlewareConfig interface {
	GetName() string
	IsEnabled() bool
	GetParameters() map[string]interface{}
}

// Middleware defines the core interface for all middleware
type Middleware interface {
	Process(ctx Context, next Next) error
	GetConfig() MiddlewareConfig
	GetName() string
	GetOrder() int
	HealthCheck() error
	ValidateConfig() error
}

// MiddlewareChain manages the execution chain of middleware
type MiddlewareChain interface {
	Add(middleware Middleware) MiddlewareChain
	Execute(ctx Context) error
	GetMiddleware() []Middleware
}

// MiddlewareFactory creates middleware instances
type MiddlewareFactory interface {
	CreateMiddleware(middlewareType string, config MiddlewareConfig) (Middleware, error)
	CreateChain(middlewares []Middleware) MiddlewareChain
	ListAvailableMiddleware() []string
}

// AuthMiddleware provides authentication functionality
type AuthMiddleware interface {
	Middleware
	ValidateToken(ctx Context) (*UserContext, error)
	CheckPermissions(ctx Context, permissions []string) error
}

// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware interface {
	Middleware
	CheckLimit(ctx Context, identifier string) error
	GetLimitInfo(ctx Context, identifier string) (*LimitInfo, error)
}

// LoggingMiddleware provides request/response logging
type LoggingMiddleware interface {
	Middleware
	LogRequest(ctx Context) error
	LogResponse(ctx Context) error
}

// HTTPRequest represents an incoming HTTP request
type HTTPRequest struct {
	Method    string
	Path      string
	Headers   map[string]string
	Body      []byte
	StartTime time.Time
	ClientIP  string
}

// HTTPResponse represents an outgoing HTTP response
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Size       int
}

// RequestMetrics contains request processing metrics
type RequestMetrics struct {
	Path       string
	Method     string
	StatusCode int
	Duration   time.Duration
	Size       int
}

// UserContext represents authenticated user information
type UserContext struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Active bool   `json:"active"`
}

// LimitInfo contains rate limiting information
type LimitInfo struct {
	Remaining   int
	ResetTime   time.Time
	WindowStart time.Time
}

// GatewayEventObserver observes gateway events
type GatewayEventObserver interface {
	OnRequestReceived(ctx context.Context, request *HTTPRequest) error
	OnRequestProcessed(ctx context.Context, request *HTTPRequest, response *HTTPResponse) error
	OnError(ctx context.Context, request *HTTPRequest, err error) error
	OnMetricsUpdate(ctx context.Context, metrics *RequestMetrics) error
}

// GatewayEventNotifier manages event notifications
type GatewayEventNotifier interface {
	Subscribe(observer GatewayEventObserver) error
	Unsubscribe(observer GatewayEventObserver) error
	NotifyRequestReceived(ctx context.Context, request *HTTPRequest) error
	NotifyRequestProcessed(ctx context.Context, request *HTTPRequest, response *HTTPResponse) error
	NotifyError(ctx context.Context, request *HTTPRequest, err error) error
}

// GinContextWrapper adapts gin.Context to our Context interface
type GinContextWrapper struct {
	*gin.Context
}

// Request returns the underlying HTTP request
func (g *GinContextWrapper) Request() *http.Request {
	return g.Context.Request
}

// Writer returns the HTTP response writer
func (g *GinContextWrapper) Writer() http.ResponseWriter {
	return g.Context.Writer
}

// Status returns the HTTP status code
func (g *GinContextWrapper) Status() int {
	return g.Context.Writer.Status()
}

// Size returns the response size
func (g *GinContextWrapper) Size() int {
	return g.Context.Writer.Size()
}

// Get retrieves a value from the context
func (g *GinContextWrapper) Get(key string) (interface{}, bool) {
	return g.Context.Get(key)
}

// Set stores a value in the context
func (g *GinContextWrapper) Set(key string, value interface{}) {
	g.Context.Set(key, value)
}

// ClientIP returns the client IP address
func (g *GinContextWrapper) ClientIP() string {
	return g.Context.ClientIP()
}

// GetHeader returns a header value
func (g *GinContextWrapper) GetHeader(key string) string {
	return g.Context.GetHeader(key)
}

// WrapGinContext creates a Context from gin.Context
func WrapGinContext(c *gin.Context) Context {
	return &GinContextWrapper{Context: c}
}
