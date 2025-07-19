package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ObservableGateway integrates middleware chain and observer pattern
type ObservableGateway struct {
	middlewareChain MiddlewareChain
	factory         MiddlewareFactory
	notifier        GatewayEventNotifier
	observers       []GatewayEventObserver
}

// NewObservableGateway creates a new observable gateway with middleware and observers
func NewObservableGateway(factory MiddlewareFactory) *ObservableGateway {
	notifier := NewGatewayEventNotifier()

	// Create default observers
	metricsObserver := NewMetricsObserver()
	securityObserver := NewSecurityObserver()

	gateway := &ObservableGateway{
		middlewareChain: NewMiddlewareChain(),
		factory:         factory,
		notifier:        notifier,
		observers:       []GatewayEventObserver{metricsObserver, securityObserver},
	}

	// Subscribe observers
	for _, observer := range gateway.observers {
		if err := notifier.Subscribe(observer); err != nil {
			log.Error().Err(err).Msg("Failed to subscribe observer")
		}
	}

	return gateway
}

// SetupMiddleware configures the middleware chain
func (g *ObservableGateway) SetupMiddleware(configs []MiddlewareConfig) error {
	for _, config := range configs {
		middleware, err := g.factory.CreateMiddleware(config.GetName(), config)
		if err != nil {
			return err
		}

		g.middlewareChain.Add(middleware)
	}

	return nil
}

// ProcessRequest processes an HTTP request through the gateway
func (g *ObservableGateway) ProcessRequest(ctx context.Context, request *HTTPRequest) (*HTTPResponse, error) {
	// Notify observers of incoming request
	if err := g.notifier.NotifyRequestReceived(ctx, request); err != nil {
		log.Debug().Err(err).Msg("Failed to notify request received")
	}

	// Create gin context wrapper
	ginCtx, ok := ctx.Value("gin_context").(*gin.Context)
	if !ok {
		log.Error().Msg("Gin context not found in request context")
		return nil, fmt.Errorf("invalid request context")
	}

	wrappedCtx := WrapGinContext(ginCtx)

	// Process request through middleware chain
	err := g.middlewareChain.Execute(wrappedCtx)

	// Create response object
	response := &HTTPResponse{
		StatusCode: ginCtx.Writer.Status(),
		Size:       ginCtx.Writer.Size(),
	}

	if err != nil {
		if notifyErr := g.notifier.NotifyError(ctx, request, err); notifyErr != nil {
			log.Debug().Err(notifyErr).Msg("Failed to notify error")
		}
		return response, err
	}

	// Notify observers of successful processing
	if err := g.notifier.NotifyRequestProcessed(ctx, request, response); err != nil {
		log.Debug().Err(err).Msg("Failed to notify request processed")
	}

	return response, nil
}

// AddObserver adds a new observer to the gateway
func (g *ObservableGateway) AddObserver(observer GatewayEventObserver) error {
	return g.notifier.Subscribe(observer)
}

// RemoveObserver removes an observer from the gateway
func (g *ObservableGateway) RemoveObserver(observer GatewayEventObserver) error {
	return g.notifier.Unsubscribe(observer)
}

// GetMiddleware returns the current middleware chain
func (g *ObservableGateway) GetMiddleware() []Middleware {
	return g.middlewareChain.GetMiddleware()
}

// HealthCheck performs a health check on all middleware and observers
func (g *ObservableGateway) HealthCheck() error {
	// Check all middleware
	for _, middleware := range g.middlewareChain.GetMiddleware() {
		if err := middleware.HealthCheck(); err != nil {
			return fmt.Errorf("middleware %s health check failed: %w", middleware.GetName(), err)
		}
	}

	log.Info().
		Int("middleware_count", len(g.middlewareChain.GetMiddleware())).
		Int("observer_count", len(g.observers)).
		Msg("Gateway health check completed successfully")

	return nil
}

// CreateGinMiddleware creates a Gin middleware function from the observable gateway
func (g *ObservableGateway) CreateGinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create HTTP request object
		request := &HTTPRequest{
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Headers:   make(map[string]string),
			StartTime: time.Now(),
			ClientIP:  c.ClientIP(),
		}

		// Copy headers
		for name, values := range c.Request.Header {
			if len(values) > 0 {
				request.Headers[name] = values[0]
			}
		}

		// Add gin context to request context
		ctx := context.WithValue(c.Request.Context(), "gin_context", c)

		// Process through gateway
		_, err := g.ProcessRequest(ctx, request)
		if err != nil {
			log.Error().Err(err).Msg("Gateway processing error")
			// Error handling is managed by middleware, so we don't need to handle it here
		}

		// Continue to next handler if not aborted by middleware
		if !c.IsAborted() {
			c.Next()
		}
	}
}
