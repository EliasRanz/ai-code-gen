package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestLoggingMiddleware_Process(t *testing.T) {
	middleware := gateway.NewLoggingMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test", nil)
	c.Request = req

	ctx := gateway.WrapGinContext(c)

	// Test process with successful next
	nextCalled := false
	next := func() error {
		nextCalled = true
		c.Status(http.StatusOK)
		return nil
	}

	err := middleware.Process(ctx, next)

	assert.NoError(t, err)
	assert.True(t, nextCalled)

	// Check that request ID was set
	requestID, exists := c.Get("request_id")
	assert.True(t, exists)
	assert.NotEmpty(t, requestID)

	// Check that response header was set
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestLoggingMiddleware_ProcessWithError(t *testing.T) {
	middleware := gateway.NewLoggingMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/api/test", nil)
	c.Request = req

	ctx := gateway.WrapGinContext(c)

	// Test process with error in next
	expectedError := assert.AnError
	next := func() error {
		c.Status(http.StatusInternalServerError)
		return expectedError
	}

	err := middleware.Process(ctx, next)

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestLoggingMiddleware_Configuration(t *testing.T) {
	middleware := gateway.NewLoggingMiddleware()

	// Test configuration
	config := middleware.GetConfig()
	assert.Equal(t, "logging", config.GetName())
	assert.True(t, config.IsEnabled())

	params := config.GetParameters()
	assert.Equal(t, "api-gateway", params["service_name"])
	assert.Equal(t, true, params["metrics_enabled"])
	assert.Equal(t, true, params["tracing_enabled"])

	// Test health check
	assert.NoError(t, middleware.HealthCheck())

	// Test validation
	assert.NoError(t, middleware.ValidateConfig())

	// Test metadata
	assert.Equal(t, "logging", middleware.GetName())
	assert.Equal(t, 10, middleware.GetOrder())
}

func TestLoggingMiddleware_LogRequest(t *testing.T) {
	middleware := gateway.NewLoggingMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("User-Agent", "test-agent")
	c.Request = req

	ctx := gateway.WrapGinContext(c)

	err := middleware.LogRequest(ctx)
	assert.NoError(t, err)
}

func TestLoggingMiddleware_LogResponse(t *testing.T) {
	middleware := gateway.NewLoggingMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("POST", "/api/generate", nil)
	c.Request = req
	c.Status(http.StatusCreated)

	ctx := gateway.WrapGinContext(c)

	err := middleware.LogResponse(ctx)
	assert.NoError(t, err)
}

func TestLoggingMiddleware_RequestIDGeneration(t *testing.T) {
	middleware := gateway.NewLoggingMiddleware()

	gin.SetMode(gin.TestMode)
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	req1, _ := http.NewRequest("GET", "/test1", nil)
	c1.Request = req1
	ctx1 := gateway.WrapGinContext(c1)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	req2, _ := http.NewRequest("GET", "/test2", nil)
	c2.Request = req2
	ctx2 := gateway.WrapGinContext(c2)

	next := func() error { return nil }

	// Process two requests
	err1 := middleware.Process(ctx1, next)
	time.Sleep(time.Millisecond) // Ensure different timestamps
	err2 := middleware.Process(ctx2, next)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Get request IDs
	id1, _ := c1.Get("request_id")
	id2, _ := c2.Get("request_id")

	// Should be different
	assert.NotEqual(t, id1, id2)
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
}

func TestLoggingMiddleware_ExistingRequestID(t *testing.T) {
	middleware := gateway.NewLoggingMiddleware()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	existingID := "existing-request-id-123"
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	c.Request = req

	ctx := gateway.WrapGinContext(c)

	next := func() error { return nil }
	err := middleware.Process(ctx, next)

	assert.NoError(t, err)

	// Should use existing request ID
	requestID, exists := c.Get("request_id")
	assert.True(t, exists)
	assert.Equal(t, existingID, requestID)

	// Should set in response header
	assert.Equal(t, existingID, w.Header().Get("X-Request-ID"))
}

func TestLoggingMiddleware_Integration(t *testing.T) {
	// Test integration with gin router and metrics
	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := gateway.NewLoggingMiddleware()

	router.Use(func(c *gin.Context) {
		ctx := gateway.WrapGinContext(c)

		next := func() error {
			c.Next()
			return nil
		}

		err := middleware.Process(ctx, next)
		if err != nil {
			t.Errorf("Middleware processing failed: %v", err)
		}
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}
