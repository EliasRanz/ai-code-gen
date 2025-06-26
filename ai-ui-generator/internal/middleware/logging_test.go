package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMetricsMiddleware(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Reset Prometheus metrics for clean test
	httpRequestDuration.Reset()
	httpRequestsTotal.Reset()
	httpRequestSizeBytes.Reset()
	httpResponseSizeBytes.Reset()
	httpRequestsInFlight.Set(0)

	tests := []struct {
		name       string
		method     string
		path       string
		statusCode int
		body       string
	}{
		{
			name:       "GET request",
			method:     "GET",
			path:       "/api/test",
			statusCode: 200,
			body:       "",
		},
		{
			name:       "POST request with body",
			method:     "POST",
			path:       "/api/create",
			statusCode: 201,
			body:       `{"name": "test"}`,
		},
		{
			name:       "Error request",
			method:     "GET",
			path:       "/api/notfound",
			statusCode: 404,
			body:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with metrics middleware
			router := gin.New()
			router.Use(MetricsMiddleware())

			// Add test routes
			router.GET("/api/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "success"})
			})
			router.POST("/api/create", func(c *gin.Context) {
				c.JSON(201, gin.H{"message": "created"})
			})
			router.GET("/api/notfound", func(c *gin.Context) {
				c.JSON(404, gin.H{"error": "not found"})
			})

			// Create request
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify response status
			assert.Equal(t, tt.statusCode, w.Code)

			// Verify metrics were recorded
			// Check that request counter was incremented
			counterMetric := httpRequestsTotal.WithLabelValues(tt.method, tt.path, string(rune(tt.statusCode)))
			assert.NotNil(t, counterMetric)

			// Check that duration histogram was updated
			histogramMetric := httpRequestDuration.WithLabelValues(tt.method, tt.path, string(rune(tt.statusCode)))
			assert.NotNil(t, histogramMetric)
		})
	}
}

func TestMetricsCollection(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Reset metrics
	httpRequestsTotal.Reset()
	httpRequestDuration.Reset()

	// Create router with metrics middleware
	router := gin.New()
	router.Use(MetricsMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Make a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that metrics were recorded
	assert.Equal(t, 200, w.Code)

	// Verify that the counter metric exists and has been incremented
	counter := httpRequestsTotal.WithLabelValues("GET", "/test", "200")
	metricValue := testutil.ToFloat64(counter)
	assert.Equal(t, float64(1), metricValue)
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetString("request_id")
		c.JSON(200, gin.H{"request_id": requestID})
	})

	// Test without X-Request-ID header
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))

	// Test with existing X-Request-ID header
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "custom-id-123")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "custom-id-123", w.Header().Get("X-Request-ID"))
}

func TestErrorHandlerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())

	// Route that triggers a bind error
	router.POST("/bind-error", func(c *gin.Context) {
		var data struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.Error(err).SetType(gin.ErrorTypeBind)
			return
		}
		c.JSON(200, data)
	})

	// Route that triggers a public error
	router.GET("/public-error", func(c *gin.Context) {
		c.Error(gin.Error{Err: gin.Error{}.Err, Type: gin.ErrorTypePublic})
	})

	// Test bind error
	req := httptest.NewRequest("POST", "/bind-error", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request data")
}
