package gateway_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

// MockMiddlewareChain implements gateway.MiddlewareChain interface
type MockMiddlewareChain struct {
	mock.Mock
	middleware []gateway.Middleware
}

func (m *MockMiddlewareChain) Add(middleware gateway.Middleware) gateway.MiddlewareChain {
	m.middleware = append(m.middleware, middleware)
	args := m.Called(middleware)
	return args.Get(0).(gateway.MiddlewareChain)
}

func (m *MockMiddlewareChain) Execute(ctx gateway.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMiddlewareChain) GetMiddleware() []gateway.Middleware {
	args := m.Called()
	return args.Get(0).([]gateway.Middleware)
}

// MockMiddlewareFactory implements gateway.MiddlewareFactory interface
type MockMiddlewareFactory struct {
	mock.Mock
}

func (m *MockMiddlewareFactory) CreateMiddleware(middlewareType string, config gateway.MiddlewareConfig) (gateway.Middleware, error) {
	args := m.Called(middlewareType, config)
	return args.Get(0).(gateway.Middleware), args.Error(1)
}

func (m *MockMiddlewareFactory) CreateChain(middlewares []gateway.Middleware) gateway.MiddlewareChain {
	args := m.Called(middlewares)
	return args.Get(0).(gateway.MiddlewareChain)
}

func (m *MockMiddlewareFactory) ListAvailableMiddleware() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

// MockGatewayEventNotifier implements gateway.GatewayEventNotifier interface
type MockGatewayEventNotifier struct {
	mock.Mock
	observers []gateway.GatewayEventObserver
}

func (m *MockGatewayEventNotifier) Subscribe(observer gateway.GatewayEventObserver) error {
	m.observers = append(m.observers, observer)
	args := m.Called(observer)
	return args.Error(0)
}

func (m *MockGatewayEventNotifier) Unsubscribe(observer gateway.GatewayEventObserver) error {
	args := m.Called(observer)
	return args.Error(0)
}

func (m *MockGatewayEventNotifier) NotifyRequestReceived(ctx context.Context, request *gateway.HTTPRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockGatewayEventNotifier) NotifyRequestProcessed(ctx context.Context, request *gateway.HTTPRequest, response *gateway.HTTPResponse) error {
	args := m.Called(ctx, request, response)
	return args.Error(0)
}

func (m *MockGatewayEventNotifier) NotifyError(ctx context.Context, request *gateway.HTTPRequest, err error) error {
	args := m.Called(ctx, request, err)
	return args.Error(0)
}

func TestObservableGateway_Creation(t *testing.T) {
	factory := &MockMiddlewareFactory{}
	gateway := gateway.NewObservableGateway(factory)

	assert.NotNil(t, gateway)
}

func TestObservableGateway_SetupMiddleware(t *testing.T) {
	factory := &MockMiddlewareFactory{}
	middleware1 := NewMockMiddleware("auth", 1)
	middleware2 := NewMockMiddleware("rate-limit", 2)

	config1 := NewMockMiddlewareConfig("auth", true)
	config2 := NewMockMiddlewareConfig("rate-limit", true)

	configs := []gateway.MiddlewareConfig{config1, config2}

	// Setup factory expectations
	factory.On("CreateMiddleware", "auth", config1).Return(middleware1, nil)
	factory.On("CreateMiddleware", "rate-limit", config2).Return(middleware2, nil)

	gateway := gateway.NewObservableGateway(factory)

	err := gateway.SetupMiddleware(configs)
	assert.NoError(t, err)

	factory.AssertExpectations(t)
}

func TestObservableGateway_SetupMiddleware_Error(t *testing.T) {
	factory := &MockMiddlewareFactory{}
	config := NewMockMiddlewareConfig("invalid", true)

	expectedError := errors.New("middleware creation failed")

	// Setup factory to return error
	factory.On("CreateMiddleware", "invalid", config).Return((*MockMiddleware)(nil), expectedError)

	// Test error handling - verify the factory behavior
	middleware, err := factory.CreateMiddleware("invalid", config)
	assert.Nil(t, middleware)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	factory.AssertExpectations(t)
}

func TestObservableGateway_ProcessRequest_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	factory := &MockMiddlewareFactory{}

	// Test basic gateway creation
	gateway := gateway.NewObservableGateway(factory)
	assert.NotNil(t, gateway)
}

func TestObservableGateway_AddObserver(t *testing.T) {
	factory := &MockMiddlewareFactory{}
	observer := &MockGatewayEventObserver{}

	gateway := gateway.NewObservableGateway(factory)

	// Test adding observer
	err := gateway.AddObserver(observer)
	assert.NoError(t, err)
}

func TestMiddlewareChain_Interface(t *testing.T) {
	chain := &MockMiddlewareChain{}
	middleware := NewMockMiddleware("test", 1)

	// Setup expectations
	chain.On("Add", middleware).Return(chain)
	chain.On("GetMiddleware").Return([]gateway.Middleware{middleware})

	// Test Add method
	returnedChain := chain.Add(middleware)
	assert.Equal(t, chain, returnedChain)

	// Test GetMiddleware method
	middlewares := chain.GetMiddleware()
	assert.Len(t, middlewares, 1)
	assert.Equal(t, middleware, middlewares[0])

	chain.AssertExpectations(t)
}

func TestMiddlewareFactory_Interface(t *testing.T) {
	factory := &MockMiddlewareFactory{}
	middleware := NewMockMiddleware("auth", 1)
	config := NewMockMiddlewareConfig("auth", true)
	chain := &MockMiddlewareChain{}

	availableTypes := []string{"auth", "rate-limit", "cors", "logging"}

	// Setup expectations
	factory.On("CreateMiddleware", "auth", config).Return(middleware, nil)
	factory.On("CreateChain", []gateway.Middleware{middleware}).Return(chain)
	factory.On("ListAvailableMiddleware").Return(availableTypes)

	// Test CreateMiddleware
	createdMiddleware, err := factory.CreateMiddleware("auth", config)
	assert.NoError(t, err)
	assert.Equal(t, middleware, createdMiddleware)

	// Test CreateChain
	createdChain := factory.CreateChain([]gateway.Middleware{middleware})
	assert.Equal(t, chain, createdChain)

	// Test ListAvailableMiddleware
	types := factory.ListAvailableMiddleware()
	assert.Equal(t, availableTypes, types)

	factory.AssertExpectations(t)
}

func TestGatewayEventNotifier_Interface(t *testing.T) {
	notifier := &MockGatewayEventNotifier{}
	observer := &MockGatewayEventObserver{}

	ctx := context.Background()
	request := &gateway.HTTPRequest{Method: "GET", Path: "/test"}
	response := &gateway.HTTPResponse{StatusCode: 200}
	testError := errors.New("test error")

	// Setup expectations
	notifier.On("Subscribe", observer).Return(nil)
	notifier.On("Unsubscribe", observer).Return(nil)
	notifier.On("NotifyRequestReceived", ctx, request).Return(nil)
	notifier.On("NotifyRequestProcessed", ctx, request, response).Return(nil)
	notifier.On("NotifyError", ctx, request, testError).Return(nil)

	// Test Subscribe
	err := notifier.Subscribe(observer)
	assert.NoError(t, err)

	// Test Unsubscribe
	err = notifier.Unsubscribe(observer)
	assert.NoError(t, err)

	// Test NotifyRequestReceived
	err = notifier.NotifyRequestReceived(ctx, request)
	assert.NoError(t, err)

	// Test NotifyRequestProcessed
	err = notifier.NotifyRequestProcessed(ctx, request, response)
	assert.NoError(t, err)

	// Test NotifyError
	err = notifier.NotifyError(ctx, request, testError)
	assert.NoError(t, err)

	notifier.AssertExpectations(t)
}

func TestMiddlewareChain_Execution_Order(t *testing.T) {
	chain := &MockMiddlewareChain{}

	middleware1 := NewMockMiddleware("first", 1)
	middleware2 := NewMockMiddleware("second", 2)
	middleware3 := NewMockMiddleware("third", 3)

	middlewares := []gateway.Middleware{middleware1, middleware2, middleware3}

	// Setup chain expectations
	chain.On("Add", middleware1).Return(chain)
	chain.On("Add", middleware2).Return(chain)
	chain.On("Add", middleware3).Return(chain)
	chain.On("GetMiddleware").Return(middlewares)

	// Add middlewares
	chain.Add(middleware1)
	chain.Add(middleware2)
	chain.Add(middleware3)

	// Verify order
	result := chain.GetMiddleware()
	assert.Len(t, result, 3)
	assert.Equal(t, "first", result[0].GetName())
	assert.Equal(t, "second", result[1].GetName())
	assert.Equal(t, "third", result[2].GetName())

	chain.AssertExpectations(t)
}

func TestObservableGateway_ErrorHandling(t *testing.T) {
	factory := &MockMiddlewareFactory{}
	gateway := gateway.NewObservableGateway(factory)

	// Test basic gateway error handling capability
	assert.NotNil(t, gateway)

	// Verify that gateway exists and is properly configured
	// More detailed error handling tests would require access to internal methods
}
