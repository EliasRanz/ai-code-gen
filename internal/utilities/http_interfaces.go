// Package utilities provides shared interfaces and types
package utilities

import "github.com/gin-gonic/gin"

// Context represents HTTP request context (framework-agnostic)
type Context interface {
	JSON(code int, obj interface{})
	Param(key string) string
	Query(key string) string
	DefaultQuery(key, defaultValue string) string
	GetHeader(key string) string
	Set(key string, value interface{})
	ShouldBindJSON(obj interface{}) error
	Request() interface{}
	Writer() interface{}
}

// HTTPHandler defines the core interface all HTTP handlers must implement
type HTTPHandler interface {
	// RegisterRoutes registers handler routes with the router
	RegisterRoutes(router Router) error

	// Health and validation
	HealthCheck() error
	ValidateRoutes() error
}

// HandlerFunc represents a framework-agnostic HTTP handler function
type HandlerFunc func(Context)

// RouterGroup represents a grouped set of routes
type RouterGroup interface {
	GET(relativePath string, handlers ...HandlerFunc) RouterGroup
	POST(relativePath string, handlers ...HandlerFunc) RouterGroup
	PUT(relativePath string, handlers ...HandlerFunc) RouterGroup
	DELETE(relativePath string, handlers ...HandlerFunc) RouterGroup
	Group(relativePath string) RouterGroup
	Use(middleware ...HandlerFunc) RouterGroup
}

// Router interface for consistent routing across different frameworks
type Router interface {
	GET(relativePath string, handlers ...HandlerFunc) Router
	POST(relativePath string, handlers ...HandlerFunc) Router
	PUT(relativePath string, handlers ...HandlerFunc) Router
	DELETE(relativePath string, handlers ...HandlerFunc) Router
	Group(relativePath string) RouterGroup
	Use(middleware ...HandlerFunc) Router
	Engine() interface{}
}

// ServiceConfig holds configuration for service handlers
type ServiceConfig struct {
	ServiceName string
	Version     string
	BaseURL     string
	Timeout     int
}

// HandlerFactory creates handlers dynamically
type HandlerFactory interface {
	CreateHandler(serviceType string, config ServiceConfig) (HTTPHandler, error)
	RegisterHandler(router Router, handler HTTPHandler) error
	ListAvailableHandlers() []string
}

// GinContextAdapter adapts gin.Context to our Context interface
type GinContextAdapter struct {
	ctx *gin.Context
}

// NewGinContextAdapter creates a new Gin context adapter
func NewGinContextAdapter(ctx *gin.Context) Context {
	return &GinContextAdapter{ctx: ctx}
}

// JSON implements Context interface
func (g *GinContextAdapter) JSON(code int, obj interface{}) {
	g.ctx.JSON(code, obj)
}

// Param implements Context interface
func (g *GinContextAdapter) Param(key string) string {
	return g.ctx.Param(key)
}

// Query implements Context interface
func (g *GinContextAdapter) Query(key string) string {
	return g.ctx.Query(key)
}

// DefaultQuery implements Context interface
func (g *GinContextAdapter) DefaultQuery(key, defaultValue string) string {
	return g.ctx.DefaultQuery(key, defaultValue)
}

// GetHeader implements Context interface
func (g *GinContextAdapter) GetHeader(key string) string {
	return g.ctx.GetHeader(key)
}

// Set implements Context interface
func (g *GinContextAdapter) Set(key string, value interface{}) {
	g.ctx.Set(key, value)
}

// ShouldBindJSON implements Context interface
func (g *GinContextAdapter) ShouldBindJSON(obj interface{}) error {
	return g.ctx.ShouldBindJSON(obj)
}

// Request implements Context interface
func (g *GinContextAdapter) Request() interface{} {
	return g.ctx.Request
}

// Writer implements Context interface
func (g *GinContextAdapter) Writer() interface{} {
	return g.ctx.Writer
}
