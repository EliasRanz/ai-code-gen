package utilities_http_interfaces

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
)

// Setup
func init() {
	gin.SetMode(gin.TestMode)
}

// Mock implementations for testing

type MockHTTPHandler struct {
	mock.Mock
}

func (m *MockHTTPHandler) RegisterRoutes(router utilities.Router) error {
	args := m.Called(router)
	return args.Error(0)
}

func (m *MockHTTPHandler) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHTTPHandler) ValidateRoutes() error {
	args := m.Called()
	return args.Error(0)
}

type MockRouter struct {
	mock.Mock
}

func (m *MockRouter) GET(relativePath string, handlers ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) POST(relativePath string, handlers ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) PUT(relativePath string, handlers ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) DELETE(relativePath string, handlers ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) Group(relativePath string) utilities.RouterGroup {
	args := m.Called(relativePath)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouter) Use(middleware ...utilities.HandlerFunc) utilities.Router {
	args := m.Called(middleware)
	return args.Get(0).(utilities.Router)
}

func (m *MockRouter) Engine() interface{} {
	args := m.Called()
	return args.Get(0)
}

type MockRouterGroup struct {
	mock.Mock
}

func (m *MockRouterGroup) GET(relativePath string, handlers ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) POST(relativePath string, handlers ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) PUT(relativePath string, handlers ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) DELETE(relativePath string, handlers ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(relativePath, handlers)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) Group(relativePath string) utilities.RouterGroup {
	args := m.Called(relativePath)
	return args.Get(0).(utilities.RouterGroup)
}

func (m *MockRouterGroup) Use(middleware ...utilities.HandlerFunc) utilities.RouterGroup {
	args := m.Called(middleware)
	return args.Get(0).(utilities.RouterGroup)
}

type MockHandlerFactory struct {
	mock.Mock
}

func (m *MockHandlerFactory) CreateHandler(serviceType string, config utilities.ServiceConfig) (utilities.HTTPHandler, error) {
	args := m.Called(serviceType, config)
	if handler := args.Get(0); handler != nil {
		return handler.(utilities.HTTPHandler), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockHandlerFactory) RegisterHandler(router utilities.Router, handler utilities.HTTPHandler) error {
	args := m.Called(router, handler)
	return args.Error(0)
}

func (m *MockHandlerFactory) ListAvailableHandlers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

// Test GinContextAdapter creation
func TestNewGinContextAdapter(t *testing.T) {
	// Create a test gin context
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	adapter := utilities.NewGinContextAdapter(ginCtx)

	assert.NotNil(t, adapter)
	assert.IsType(t, &utilities.GinContextAdapter{}, adapter)
}

// Test GinContextAdapter JSON method
func TestGinContextAdapter_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)

	adapter := utilities.NewGinContextAdapter(ginCtx)

	testData := map[string]string{"message": "test"}
	adapter.JSON(http.StatusOK, testData)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"test"`)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

// Test GinContextAdapter Param method
func TestGinContextAdapter_Param(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	// Set up a route with parameters
	engine.GET("/users/:id", func(c *gin.Context) {
		adapter := utilities.NewGinContextAdapter(c)
		id := adapter.Param("id")
		c.String(http.StatusOK, id)
	})

	req, _ := http.NewRequest("GET", "/users/123", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "123", w.Body.String())
}

// Test GinContextAdapter Query method
func TestGinContextAdapter_Query(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.GET("/search", func(c *gin.Context) {
		adapter := utilities.NewGinContextAdapter(c)
		query := adapter.Query("q")
		c.String(http.StatusOK, query)
	})

	req, _ := http.NewRequest("GET", "/search?q=test", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test", w.Body.String())
}

// Test GinContextAdapter DefaultQuery method
func TestGinContextAdapter_DefaultQuery(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.GET("/items", func(c *gin.Context) {
		adapter := utilities.NewGinContextAdapter(c)

		// Test with provided query parameter
		limit := adapter.DefaultQuery("limit", "10")
		// Test with missing query parameter (should use default)
		page := adapter.DefaultQuery("page", "1")

		c.String(http.StatusOK, "limit=%s,page=%s", limit, page)
	})

	req, _ := http.NewRequest("GET", "/items?limit=20", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "limit=20,page=1", w.Body.String())
}

// Test GinContextAdapter GetHeader method
func TestGinContextAdapter_GetHeader(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.GET("/headers", func(c *gin.Context) {
		adapter := utilities.NewGinContextAdapter(c)
		auth := adapter.GetHeader("Authorization")
		contentType := adapter.GetHeader("Content-Type")
		c.String(http.StatusOK, "auth=%s,type=%s", auth, contentType)
	})

	req, _ := http.NewRequest("GET", "/headers", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "auth=Bearer token123,type=application/json", w.Body.String())
}

// Test GinContextAdapter Set method
func TestGinContextAdapter_Set(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.GET("/set", func(c *gin.Context) {
		adapter := utilities.NewGinContextAdapter(c)
		adapter.Set("user_id", "123")
		adapter.Set("role", "admin")

		// Verify values were set
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "123", userID)

		role, exists := c.Get("role")
		assert.True(t, exists)
		assert.Equal(t, "admin", role)

		c.String(http.StatusOK, "values set")
	})

	req, _ := http.NewRequest("GET", "/set", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "values set", w.Body.String())
}

// Test GinContextAdapter ShouldBindJSON method
func TestGinContextAdapter_ShouldBindJSON(t *testing.T) {
	t.Run("successful binding", func(t *testing.T) {
		jsonData := `{"name":"John","age":30}`
		req, _ := http.NewRequest("POST", "/bind", bytes.NewBuffer([]byte(jsonData)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		_, engine := gin.CreateTestContext(w)

		engine.POST("/bind", func(c *gin.Context) {
			adapter := utilities.NewGinContextAdapter(c)

			var user struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}

			err := adapter.ShouldBindJSON(&user)

			assert.NoError(t, err)
			assert.Equal(t, "John", user.Name)
			assert.Equal(t, 30, user.Age)

			c.String(http.StatusOK, "binding successful")
		})

		engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		jsonData := `{"name":"John","age":}` // Invalid JSON
		req, _ := http.NewRequest("POST", "/bind", bytes.NewBuffer([]byte(jsonData)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		_, engine := gin.CreateTestContext(w)

		engine.POST("/bind", func(c *gin.Context) {
			adapter := utilities.NewGinContextAdapter(c)

			var user struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}

			err := adapter.ShouldBindJSON(&user)

			assert.Error(t, err)
			c.String(http.StatusBadRequest, "binding failed")
		})

		engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Test GinContextAdapter Request method
func TestGinContextAdapter_Request(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.GET("/request", func(c *gin.Context) {
		adapter := utilities.NewGinContextAdapter(c)
		req := adapter.Request()

		assert.NotNil(t, req)
		assert.IsType(t, &http.Request{}, req)

		httpReq := req.(*http.Request)
		assert.Equal(t, "GET", httpReq.Method)
		assert.Equal(t, "/request", httpReq.URL.Path)

		c.String(http.StatusOK, "request accessed")
	})

	req, _ := http.NewRequest("GET", "/request", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test GinContextAdapter Writer method
func TestGinContextAdapter_Writer(t *testing.T) {
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.GET("/writer", func(c *gin.Context) {
		adapter := utilities.NewGinContextAdapter(c)
		writer := adapter.Writer()

		assert.NotNil(t, writer)
		// Note: Gin's writer is a gin.ResponseWriter interface, not http.ResponseWriter

		c.String(http.StatusOK, "writer accessed")
	})

	req, _ := http.NewRequest("GET", "/writer", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test ServiceConfig structure
func TestServiceConfig(t *testing.T) {
	config := utilities.ServiceConfig{
		ServiceName: "test-service",
		Version:     "1.0.0",
		BaseURL:     "/api/v1",
		Timeout:     30,
	}

	assert.Equal(t, "test-service", config.ServiceName)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, "/api/v1", config.BaseURL)
	assert.Equal(t, 30, config.Timeout)
}

// Test HTTPHandler interface through mock
func TestHTTPHandler_Interface(t *testing.T) {
	handler := &MockHTTPHandler{}
	router := &MockRouter{}

	// Test RegisterRoutes
	handler.On("RegisterRoutes", router).Return(nil)
	err := handler.RegisterRoutes(router)
	assert.NoError(t, err)
	handler.AssertExpectations(t)

	// Test HealthCheck
	handler.On("HealthCheck").Return(nil)
	err = handler.HealthCheck()
	assert.NoError(t, err)
	handler.AssertExpectations(t)

	// Test ValidateRoutes
	handler.On("ValidateRoutes").Return(nil)
	err = handler.ValidateRoutes()
	assert.NoError(t, err)
	handler.AssertExpectations(t)
}

// Test HTTPHandler interface error cases
func TestHTTPHandler_ErrorCases(t *testing.T) {
	handler := &MockHTTPHandler{}
	router := &MockRouter{}

	// Test RegisterRoutes error
	registerErr := errors.New("failed to register routes")
	handler.On("RegisterRoutes", router).Return(registerErr)
	err := handler.RegisterRoutes(router)
	assert.Error(t, err)
	assert.Equal(t, registerErr, err)

	// Test HealthCheck error
	healthErr := errors.New("service unhealthy")
	handler.On("HealthCheck").Return(healthErr)
	err = handler.HealthCheck()
	assert.Error(t, err)
	assert.Equal(t, healthErr, err)

	// Test ValidateRoutes error
	validateErr := errors.New("invalid route configuration")
	handler.On("ValidateRoutes").Return(validateErr)
	err = handler.ValidateRoutes()
	assert.Error(t, err)
	assert.Equal(t, validateErr, err)

	handler.AssertExpectations(t)
}

// Test Router interface through mock
func TestRouter_Interface(t *testing.T) {
	router := &MockRouter{}
	group := &MockRouterGroup{}

	// Test HTTP methods
	router.On("GET", "/test", mock.Anything).Return(router)
	router.On("POST", "/test", mock.Anything).Return(router)
	router.On("PUT", "/test", mock.Anything).Return(router)
	router.On("DELETE", "/test", mock.Anything).Return(router)

	handler := func(c utilities.Context) {}

	result := router.GET("/test", handler)
	assert.Equal(t, router, result)

	result = router.POST("/test", handler)
	assert.Equal(t, router, result)

	result = router.PUT("/test", handler)
	assert.Equal(t, router, result)

	result = router.DELETE("/test", handler)
	assert.Equal(t, router, result)

	// Test Group
	router.On("Group", "/api").Return(group)
	groupResult := router.Group("/api")
	assert.Equal(t, group, groupResult)

	// Test Use middleware
	middleware := func(c utilities.Context) {}
	router.On("Use", mock.Anything).Return(router)
	result = router.Use(middleware)
	assert.Equal(t, router, result)

	// Test Engine
	engine := gin.New()
	router.On("Engine").Return(engine)
	engineResult := router.Engine()
	assert.Equal(t, engine, engineResult)

	router.AssertExpectations(t)
}

// Test RouterGroup interface through mock
func TestRouterGroup_Interface(t *testing.T) {
	group := &MockRouterGroup{}

	// Test HTTP methods
	group.On("GET", "/test", mock.Anything).Return(group)
	group.On("POST", "/test", mock.Anything).Return(group)
	group.On("PUT", "/test", mock.Anything).Return(group)
	group.On("DELETE", "/test", mock.Anything).Return(group)

	handler := func(c utilities.Context) {}

	result := group.GET("/test", handler)
	assert.Equal(t, group, result)

	result = group.POST("/test", handler)
	assert.Equal(t, group, result)

	result = group.PUT("/test", handler)
	assert.Equal(t, group, result)

	result = group.DELETE("/test", handler)
	assert.Equal(t, group, result)

	// Test nested Group
	nestedGroup := &MockRouterGroup{}
	group.On("Group", "/v1").Return(nestedGroup)
	groupResult := group.Group("/v1")
	assert.Equal(t, nestedGroup, groupResult)

	// Test Use middleware
	middleware := func(c utilities.Context) {}
	group.On("Use", mock.Anything).Return(group)
	result = group.Use(middleware)
	assert.Equal(t, group, result)

	group.AssertExpectations(t)
}

// Test HandlerFactory interface through mock
func TestHandlerFactory_Interface(t *testing.T) {
	factory := &MockHandlerFactory{}
	router := &MockRouter{}
	handler := &MockHTTPHandler{}

	config := utilities.ServiceConfig{
		ServiceName: "test-service",
		Version:     "1.0.0",
		BaseURL:     "/api",
		Timeout:     30,
	}

	// Test CreateHandler
	factory.On("CreateHandler", "api", config).Return(handler, nil)
	createdHandler, err := factory.CreateHandler("api", config)
	assert.NoError(t, err)
	assert.Equal(t, handler, createdHandler)

	// Test RegisterHandler
	factory.On("RegisterHandler", router, handler).Return(nil)
	err = factory.RegisterHandler(router, handler)
	assert.NoError(t, err)

	// Test ListAvailableHandlers
	expectedHandlers := []string{"api", "auth", "user"}
	factory.On("ListAvailableHandlers").Return(expectedHandlers)
	handlers := factory.ListAvailableHandlers()
	assert.Equal(t, expectedHandlers, handlers)

	factory.AssertExpectations(t)
}

// Test HandlerFactory error cases
func TestHandlerFactory_ErrorCases(t *testing.T) {
	factory := &MockHandlerFactory{}
	router := &MockRouter{}
	handler := &MockHTTPHandler{}

	config := utilities.ServiceConfig{
		ServiceName: "unknown-service",
		Version:     "1.0.0",
	}

	// Test CreateHandler error
	createErr := errors.New("unknown service type")
	factory.On("CreateHandler", "unknown", config).Return(nil, createErr)
	createdHandler, err := factory.CreateHandler("unknown", config)
	assert.Error(t, err)
	assert.Nil(t, createdHandler)
	assert.Equal(t, createErr, err)

	// Test RegisterHandler error
	registerErr := errors.New("failed to register handler")
	factory.On("RegisterHandler", router, handler).Return(registerErr)
	err = factory.RegisterHandler(router, handler)
	assert.Error(t, err)
	assert.Equal(t, registerErr, err)

	factory.AssertExpectations(t)
}

// Test HandlerFunc type
func TestHandlerFunc_Type(t *testing.T) {
	// Test that HandlerFunc can be assigned and called
	var handler utilities.HandlerFunc = func(c utilities.Context) {
		c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	assert.NotNil(t, handler)

	// Test calling the handler with a mock context
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	adapter := utilities.NewGinContextAdapter(ginCtx)

	handler(adapter)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"success"`)
}

// Test Context interface completeness
func TestContext_InterfaceCompleteness(t *testing.T) {
	// Create a proper HTTP request context
	req, _ := http.NewRequest("GET", "/test?q=value", nil)
	req.Header.Set("Authorization", "Bearer token")

	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.GET("/test", func(c *gin.Context) {
		// Test that GinContextAdapter implements all Context methods
		var ctx utilities.Context = utilities.NewGinContextAdapter(c)

		// Test all methods exist and can be called
		assert.NotPanics(t, func() {
			ctx.JSON(200, map[string]string{"test": "data"})
			_ = ctx.Param("id")
			_ = ctx.Query("q")
			_ = ctx.DefaultQuery("limit", "10")
			_ = ctx.GetHeader("Authorization")
			ctx.Set("key", "value")

			var testStruct struct {
				Name string `json:"name"`
			}
			_ = ctx.ShouldBindJSON(&testStruct) // May error, but should not panic

			_ = ctx.Request()
			_ = ctx.Writer()
		})

		c.String(http.StatusOK, "completeness test passed")
	})

	engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test integration scenario
func TestGinContextAdapter_Integration(t *testing.T) {
	// Test a realistic HTTP handler scenario
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)

	engine.POST("/users/:id", func(c *gin.Context) {
		adapter := utilities.NewGinContextAdapter(c)

		// Get path parameter
		id := adapter.Param("id")

		// Get query parameter with default
		format := adapter.DefaultQuery("format", "json")

		// Check authorization header
		auth := adapter.GetHeader("Authorization")
		if auth == "" {
			adapter.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing authorization",
			})
			return
		}

		// Bind JSON body
		var requestBody struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		if err := adapter.ShouldBindJSON(&requestBody); err != nil {
			adapter.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
			return
		}

		// Set context values
		adapter.Set("user_id", id)
		adapter.Set("format", format)

		// Return success response
		adapter.JSON(http.StatusOK, map[string]interface{}{
			"id":     id,
			"name":   requestBody.Name,
			"email":  requestBody.Email,
			"format": format,
		})
	})

	// Test the integration
	jsonBody := `{"name":"John Doe","email":"john@example.com"}`
	req, _ := http.NewRequest("POST", "/users/123?format=xml", bytes.NewBuffer([]byte(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"123"`)
	assert.Contains(t, w.Body.String(), `"name":"John Doe"`)
	assert.Contains(t, w.Body.String(), `"email":"john@example.com"`)
	assert.Contains(t, w.Body.String(), `"format":"xml"`)
}
