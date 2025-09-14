package utilities_handler_factory_test

import (
	"errors"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

// MockHTTPHandler implements HTTPHandler interface for testing
type MockHTTPHandler struct {
	mock.Mock
}

func (m *MockHTTPHandler) RegisterRoutes(router utilities.Router) error {
	args := m.Called(router)
	return args.Error(0)
}

func (m *MockHTTPHandler) ValidateRoutes() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHTTPHandler) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

// MockRouter implements Router interface for testing
type MockRouter struct {
	mock.Mock
}

func (m *MockRouter) GET(path string, handlers ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(path, handlers)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) POST(path string, handlers ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(path, handlers)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) PUT(path string, handlers ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(path, handlers)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) DELETE(path string, handlers ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(path, handlers)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) Group(prefix string) utilities.RouterGroup {
	args := m.Called(prefix)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouter) Use(middlewares ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(middlewares)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) Engine() interface{} {
	args := m.Called()
	return args.Get(0)
}

// MockRouterGroup implements RouterGroup interface for testing
type MockRouterGroup struct {
	mock.Mock
}

func (m *MockRouterGroup) GET(path string, handlers ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(path, handlers)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) POST(path string, handlers ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(path, handlers)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) PUT(path string, handlers ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(path, handlers)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) DELETE(path string, handlers ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(path, handlers)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) Group(relativePath string) utilities.RouterGroup {
	args := m.Called(relativePath)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) Use(middlewares ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(middlewares)
	return args.Get(0).(utilities.RouterGroup)
}

// Tests for HandlerFactory creation
func TestNewHandlerFactory(t *testing.T) {
	t.Run("create new handler factory", func(t *testing.T) {
		factory := utilities.NewHandlerFactory()

		assert.NotNil(t, factory)
		assert.Implements(t, (*utilities.HandlerFactory)(nil), factory)

		// Test that it starts with empty handlers list
		handlers := factory.ListAvailableHandlers()
		assert.Empty(t, handlers)
	})
}

// Tests for CreateHandler method
func TestHandlerFactory_CreateHandler(t *testing.T) {
	factory := utilities.NewHandlerFactory()

	t.Run("empty service type error", func(t *testing.T) {
		config := utilities.ServiceConfig{
			ServiceName: "test-service",
			Timeout:     30,
			Version:     "v1",
		}

		handler, err := factory.CreateHandler("", config)

		assert.Nil(t, handler)
		assert.Error(t, err)

		// Check if it's a ValidationError
		assert.True(t, utilities.IsValidationError(err))
		assert.Contains(t, err.Error(), "service type cannot be empty")
	})

	t.Run("empty service name error", func(t *testing.T) {
		config := utilities.ServiceConfig{
			ServiceName: "",
			Timeout:     30,
			Version:     "v1",
		}

		handler, err := factory.CreateHandler("auth", config)

		assert.Nil(t, handler)
		assert.Error(t, err)

		// Check if it's a ValidationError
		assert.True(t, utilities.IsValidationError(err))
		assert.Contains(t, err.Error(), "service name cannot be empty")
	})

	t.Run("handler creation not implemented", func(t *testing.T) {
		config := utilities.ServiceConfig{
			ServiceName: "test-service",
			Timeout:     30,
			Version:     "v1",
			BaseURL:     "http://localhost:8080",
		}

		handler, err := factory.CreateHandler("auth", config)

		assert.Nil(t, handler)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler creation not implemented for service type: auth")
	})
}

// Tests for RegisterHandler method
func TestHandlerFactory_RegisterHandler(t *testing.T) {
	factory := utilities.NewHandlerFactory()

	t.Run("successful handler registration", func(t *testing.T) {
		mockRouter := &MockRouter{}
		mockHandler := &MockHTTPHandler{}

		// Set up expectations
		mockHandler.On("ValidateRoutes").Return(nil)
		mockHandler.On("RegisterRoutes", mockRouter).Return(nil)

		err := factory.RegisterHandler(mockRouter, mockHandler)

		assert.NoError(t, err)

		// Verify all expectations were met
		mockRouter.AssertExpectations(t)
		mockHandler.AssertExpectations(t)
	})

	t.Run("nil router error", func(t *testing.T) {
		mockHandler := &MockHTTPHandler{}

		err := factory.RegisterHandler(nil, mockHandler)

		assert.Error(t, err)
		assert.True(t, utilities.IsValidationError(err))
		assert.Contains(t, err.Error(), "router cannot be nil")
	})

	t.Run("nil handler error", func(t *testing.T) {
		mockRouter := &MockRouter{}

		err := factory.RegisterHandler(mockRouter, nil)

		assert.Error(t, err)
		assert.True(t, utilities.IsValidationError(err))
		assert.Contains(t, err.Error(), "handler cannot be nil")
	})

	t.Run("handler route validation failure", func(t *testing.T) {
		mockRouter := &MockRouter{}
		mockHandler := &MockHTTPHandler{}

		validationError := errors.New("invalid route configuration")
		mockHandler.On("ValidateRoutes").Return(validationError)

		err := factory.RegisterHandler(mockRouter, mockHandler)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler route validation failed")
		assert.Contains(t, err.Error(), "invalid route configuration")

		mockHandler.AssertExpectations(t)
	})

	t.Run("handler route registration failure", func(t *testing.T) {
		mockRouter := &MockRouter{}
		mockHandler := &MockHTTPHandler{}

		registrationError := errors.New("failed to bind route")
		mockHandler.On("ValidateRoutes").Return(nil)
		mockHandler.On("RegisterRoutes", mockRouter).Return(registrationError)

		err := factory.RegisterHandler(mockRouter, mockHandler)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to register handler routes")
		assert.Contains(t, err.Error(), "failed to bind route")

		mockHandler.AssertExpectations(t)
	})
}

// Tests for ListAvailableHandlers method
func TestHandlerFactory_ListAvailableHandlers(t *testing.T) {
	factory := utilities.NewHandlerFactory()

	t.Run("empty handlers list initially", func(t *testing.T) {
		handlers := factory.ListAvailableHandlers()

		assert.Empty(t, handlers)
		assert.IsType(t, []string{}, handlers)
	})

	// Note: Since the current implementation doesn't provide a way to add handlers
	// to the internal map, we can only test the empty case. In a real implementation,
	// there would be methods to register handler types.
}

// Tests for HandlerFactory interface compliance
func TestHandlerFactory_InterfaceCompliance(t *testing.T) {
	factory := utilities.NewHandlerFactory()

	// Verify the factory implements HandlerFactory interface
	assert.Implements(t, (*utilities.HandlerFactory)(nil), factory)

	// Test all methods are callable with proper error handling
	assert.NotPanics(t, func() {
		config := utilities.ServiceConfig{ServiceName: "test"}
		_, _ = factory.CreateHandler("test", config)
		_ = factory.ListAvailableHandlers()
	})
}

// Tests for error handling patterns
func TestHandlerFactory_ErrorHandling(t *testing.T) {
	factory := utilities.NewHandlerFactory()

	t.Run("validation errors are properly typed", func(t *testing.T) {
		config := utilities.ServiceConfig{ServiceName: ""}

		_, err := factory.CreateHandler("test", config)

		assert.Error(t, err)
		assert.True(t, utilities.IsValidationError(err))
	})

	t.Run("wrapped errors preserve original context", func(t *testing.T) {
		mockRouter := &MockRouter{}
		mockHandler := &MockHTTPHandler{}

		originalError := errors.New("original error")
		mockHandler.On("ValidateRoutes").Return(originalError)

		err := factory.RegisterHandler(mockRouter, mockHandler)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler route validation failed")
		assert.True(t, errors.Is(err, originalError))

		mockHandler.AssertExpectations(t)
	})
}

// Tests for concurrent usage
func TestHandlerFactory_ConcurrentUsage(t *testing.T) {
	factory := utilities.NewHandlerFactory()

	t.Run("concurrent access to ListAvailableHandlers", func(t *testing.T) {
		done := make(chan bool, 10)

		// Start multiple goroutines
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				// This should not panic or cause data races
				handlers := factory.ListAvailableHandlers()
				assert.NotNil(t, handlers)
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("concurrent CreateHandler calls", func(t *testing.T) {
		done := make(chan bool, 10)

		// Start multiple goroutines
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				config := utilities.ServiceConfig{ServiceName: "test"}
				_, err := factory.CreateHandler("test", config)
				assert.Error(t, err) // Expected since implementation is not complete
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// Integration tests
func TestHandlerFactory_Integration(t *testing.T) {
	t.Run("complete workflow simulation", func(t *testing.T) {
		factory := utilities.NewHandlerFactory()
		mockRouter := &MockRouter{}
		mockHandler := &MockHTTPHandler{}

		// Set up handler expectations
		mockHandler.On("ValidateRoutes").Return(nil)
		mockHandler.On("RegisterRoutes", mockRouter).Return(nil)

		// Test complete workflow
		// 1. Create factory (already done)
		assert.NotNil(t, factory)

		// 2. List available handlers (should be empty initially)
		handlers := factory.ListAvailableHandlers()
		assert.Empty(t, handlers)

		// 3. Attempt to create a handler (will fail with not implemented)
		config := utilities.ServiceConfig{ServiceName: "test", Version: "v1"}
		_, err := factory.CreateHandler("auth", config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler creation not implemented")

		// 4. Register an existing handler (should succeed)
		err = factory.RegisterHandler(mockRouter, mockHandler)
		assert.NoError(t, err)

		// Verify expectations
		mockHandler.AssertExpectations(t)
		mockRouter.AssertExpectations(t)
	})
}
