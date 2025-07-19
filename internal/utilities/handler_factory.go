// Package utilities provides shared handler factory implementation
package utilities

import (
	"fmt"
)

// handlerFactory implements HandlerFactory interface
type handlerFactory struct {
	handlers map[string]HTTPHandler
}

// NewHandlerFactory creates a new handler factory
func NewHandlerFactory() HandlerFactory {
	return &handlerFactory{
		handlers: make(map[string]HTTPHandler),
	}
}

// CreateHandler creates a handler based on service type
func (f *handlerFactory) CreateHandler(serviceType string, config ServiceConfig) (HTTPHandler, error) {
	if serviceType == "" {
		return nil, NewValidationError("service type cannot be empty", nil)
	}

	if config.ServiceName == "" {
		return nil, NewValidationError("service name cannot be empty", nil)
	}

	// Handler creation logic would be implemented by concrete factories
	// This is a skeleton implementation
	return nil, fmt.Errorf("handler creation not implemented for service type: %s", serviceType)
}

// RegisterHandler registers a handler with the router
func (f *handlerFactory) RegisterHandler(router Router, handler HTTPHandler) error {
	if router == nil {
		return NewValidationError("router cannot be nil", nil)
	}

	if handler == nil {
		return NewValidationError("handler cannot be nil", nil)
	}

	// Validate handler routes before registration
	if err := handler.ValidateRoutes(); err != nil {
		return fmt.Errorf("handler route validation failed: %w", err)
	}

	// Register routes
	if err := handler.RegisterRoutes(router); err != nil {
		return fmt.Errorf("failed to register handler routes: %w", err)
	}

	return nil
}

// ListAvailableHandlers returns list of available handler types
func (f *handlerFactory) ListAvailableHandlers() []string {
	handlers := make([]string, 0, len(f.handlers))
	for handlerType := range f.handlers {
		handlers = append(handlers, handlerType)
	}
	return handlers
}
