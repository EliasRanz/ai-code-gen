package gateway_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

// MockMiddleware implements gateway.Middleware interface
type MockMiddleware struct {
	mock.Mock
	name  string
	order int
}

func NewMockMiddleware(name string, order int) *MockMiddleware {
	return &MockMiddleware{
		name:  name,
		order: order,
	}
}

func (m *MockMiddleware) Process(ctx gateway.Context, next gateway.Next) error {
	args := m.Called(ctx, next)
	return args.Error(0)
}

func (m *MockMiddleware) GetConfig() gateway.MiddlewareConfig {
	args := m.Called()
	return args.Get(0).(gateway.MiddlewareConfig)
}

func (m *MockMiddleware) GetName() string {
	return m.name
}

func (m *MockMiddleware) GetOrder() int {
	return m.order
}

func (m *MockMiddleware) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMiddleware) ValidateConfig() error {
	args := m.Called()
	return args.Error(0)
}

// MockMiddlewareConfig implements gateway.MiddlewareConfig interface
type MockMiddlewareConfig struct {
	mock.Mock
	name       string
	enabled    bool
	parameters map[string]interface{}
}

func NewMockMiddlewareConfig(name string, enabled bool) *MockMiddlewareConfig {
	return &MockMiddlewareConfig{
		name:       name,
		enabled:    enabled,
		parameters: make(map[string]interface{}),
	}
}

func (m *MockMiddlewareConfig) GetName() string {
	return m.name
}

func (m *MockMiddlewareConfig) IsEnabled() bool {
	return m.enabled
}

func (m *MockMiddlewareConfig) GetParameters() map[string]interface{} {
	return m.parameters
}

// MockGatewayEventObserver implements gateway.GatewayEventObserver interface
type MockGatewayEventObserver struct {
	mock.Mock
}

func (m *MockGatewayEventObserver) OnRequestReceived(ctx context.Context, request *gateway.HTTPRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockGatewayEventObserver) OnRequestProcessed(ctx context.Context, request *gateway.HTTPRequest, response *gateway.HTTPResponse) error {
	args := m.Called(ctx, request, response)
	return args.Error(0)
}

func (m *MockGatewayEventObserver) OnError(ctx context.Context, request *gateway.HTTPRequest, err error) error {
	args := m.Called(ctx, request, err)
	return args.Error(0)
}

func (m *MockGatewayEventObserver) OnMetricsUpdate(ctx context.Context, metrics *gateway.RequestMetrics) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func TestGinContextWrapper_Request(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Create gin context
	c, _ := gin.CreateTestContext(nil)
	c.Request = req

	// Wrap in gateway context
	wrapper := gateway.WrapGinContext(c)

	// Test Request() method
	assert.Equal(t, req, wrapper.Request())
}

func TestGinContextWrapper_ClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create gin context
	c, _ := gin.CreateTestContext(nil)
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:8080"
	c.Request = req

	// Wrap in gateway context
	wrapper := gateway.WrapGinContext(c)

	// Test ClientIP() method
	ip := wrapper.ClientIP()
	assert.NotEmpty(t, ip)
}

func TestGinContextWrapper_GetSetHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create gin context
	c, _ := gin.CreateTestContext(nil)
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	c.Request = req

	// Wrap in gateway context
	wrapper := gateway.WrapGinContext(c)

	// Test GetHeader() method
	auth := wrapper.GetHeader("Authorization")
	assert.Equal(t, "Bearer token123", auth)

	// Test non-existent header
	nonExistent := wrapper.GetHeader("X-Non-Existent")
	assert.Empty(t, nonExistent)
}

func TestGinContextWrapper_GetSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create gin context
	c, _ := gin.CreateTestContext(nil)

	// Wrap in gateway context
	wrapper := gateway.WrapGinContext(c)

	// Test Set() and Get() methods
	wrapper.Set("user_id", "123")
	wrapper.Set("session_data", map[string]string{"role": "admin"})

	userID, exists := wrapper.Get("user_id")
	assert.True(t, exists)
	assert.Equal(t, "123", userID)

	sessionData, exists := wrapper.Get("session_data")
	assert.True(t, exists)
	assert.IsType(t, map[string]string{}, sessionData)

	// Test non-existent key
	nonExistent, exists := wrapper.Get("non_existent")
	assert.False(t, exists)
	assert.Nil(t, nonExistent)
}

func TestHTTPRequest_Structure(t *testing.T) {
	startTime := time.Now()
	request := &gateway.HTTPRequest{
		Method:    "POST",
		Path:      "/api/v1/users",
		Headers:   map[string]string{"Content-Type": "application/json"},
		Body:      []byte(`{"name": "test"}`),
		StartTime: startTime,
		ClientIP:  "192.168.1.100",
	}

	assert.Equal(t, "POST", request.Method)
	assert.Equal(t, "/api/v1/users", request.Path)
	assert.Equal(t, "application/json", request.Headers["Content-Type"])
	assert.Equal(t, `{"name": "test"}`, string(request.Body))
	assert.Equal(t, startTime, request.StartTime)
	assert.Equal(t, "192.168.1.100", request.ClientIP)
}

func TestHTTPResponse_Structure(t *testing.T) {
	response := &gateway.HTTPResponse{
		StatusCode: 201,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"id": 1, "name": "test"}`),
		Size:       25,
	}

	assert.Equal(t, 201, response.StatusCode)
	assert.Equal(t, "application/json", response.Headers["Content-Type"])
	assert.Equal(t, `{"id": 1, "name": "test"}`, string(response.Body))
	assert.Equal(t, 25, response.Size)
}

func TestRequestMetrics_Structure(t *testing.T) {
	duration := 150 * time.Millisecond
	metrics := &gateway.RequestMetrics{
		Path:       "/api/v1/users",
		Method:     "GET",
		StatusCode: 200,
		Duration:   duration,
		Size:       1024,
	}

	assert.Equal(t, "/api/v1/users", metrics.Path)
	assert.Equal(t, "GET", metrics.Method)
	assert.Equal(t, 200, metrics.StatusCode)
	assert.Equal(t, duration, metrics.Duration)
	assert.Equal(t, 1024, metrics.Size)
}

func TestUserContext_Structure(t *testing.T) {
	user := &gateway.UserContext{
		UserID: "user123",
		Email:  "test@example.com",
		Role:   "admin",
		Active: true,
	}

	assert.Equal(t, "user123", user.UserID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "admin", user.Role)
	assert.True(t, user.Active)
}

func TestLimitInfo_Structure(t *testing.T) {
	resetTime := time.Now().Add(time.Hour)
	windowStart := time.Now().Add(-30 * time.Minute)

	limitInfo := &gateway.LimitInfo{
		Remaining:   45,
		ResetTime:   resetTime,
		WindowStart: windowStart,
	}

	assert.Equal(t, 45, limitInfo.Remaining)
	assert.Equal(t, resetTime, limitInfo.ResetTime)
	assert.Equal(t, windowStart, limitInfo.WindowStart)
}

func TestMockMiddleware_Interface(t *testing.T) {
	middleware := NewMockMiddleware("test-middleware", 1)
	config := NewMockMiddlewareConfig("test-config", true)

	// Test basic properties
	assert.Equal(t, "test-middleware", middleware.GetName())
	assert.Equal(t, 1, middleware.GetOrder())

	// Setup mock expectations
	middleware.On("GetConfig").Return(config)
	middleware.On("HealthCheck").Return(nil)
	middleware.On("ValidateConfig").Return(nil)

	// Test method calls
	returnedConfig := middleware.GetConfig()
	assert.Equal(t, config, returnedConfig)

	err := middleware.HealthCheck()
	assert.NoError(t, err)

	err = middleware.ValidateConfig()
	assert.NoError(t, err)

	middleware.AssertExpectations(t)
}

func TestMockMiddlewareConfig_Interface(t *testing.T) {
	config := NewMockMiddlewareConfig("auth-middleware", true)
	config.parameters["jwt_secret"] = "secret123"
	config.parameters["timeout"] = 30 * time.Second

	assert.Equal(t, "auth-middleware", config.GetName())
	assert.True(t, config.IsEnabled())

	params := config.GetParameters()
	assert.Equal(t, "secret123", params["jwt_secret"])
	assert.Equal(t, 30*time.Second, params["timeout"])
}

func TestMockGatewayEventObserver_Interface(t *testing.T) {
	observer := &MockGatewayEventObserver{}

	request := &gateway.HTTPRequest{
		Method: "GET",
		Path:   "/test",
	}

	response := &gateway.HTTPResponse{
		StatusCode: 200,
	}

	metrics := &gateway.RequestMetrics{
		Path:       "/test",
		Method:     "GET",
		StatusCode: 200,
		Duration:   100 * time.Millisecond,
	}

	ctx := context.Background()
	testErr := assert.AnError

	// Setup expectations
	observer.On("OnRequestReceived", ctx, request).Return(nil)
	observer.On("OnRequestProcessed", ctx, request, response).Return(nil)
	observer.On("OnError", ctx, request, testErr).Return(nil)
	observer.On("OnMetricsUpdate", ctx, metrics).Return(nil)

	// Test method calls
	err := observer.OnRequestReceived(ctx, request)
	assert.NoError(t, err)

	err = observer.OnRequestProcessed(ctx, request, response)
	assert.NoError(t, err)

	err = observer.OnError(ctx, request, testErr)
	assert.NoError(t, err)

	err = observer.OnMetricsUpdate(ctx, metrics)
	assert.NoError(t, err)

	observer.AssertExpectations(t)
}
