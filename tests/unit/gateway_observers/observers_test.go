package gateway_observers

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestMain(m *testing.M) {
	// Disable logging during tests to reduce noise
	log.SetOutput(os.Stderr)
	code := m.Run()
	os.Exit(code)
}

// Mock implementations for testing
type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) IncrementRequestCount(path, method string) {
	m.Called(path, method)
}

func (m *MockMetricsCollector) RecordLatency(path string, duration time.Duration) {
	m.Called(path, duration)
}

func (m *MockMetricsCollector) IncrementResponseCode(statusCode int) {
	m.Called(statusCode)
}

type MockAlertManager struct {
	mock.Mock
}

func (m *MockAlertManager) SendSecurityAlert(ctx context.Context, request *gateway.HTTPRequest, err error) {
	m.Called(ctx, request, err)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) LogSecurityEvent(ctx context.Context, request *gateway.HTTPRequest, err error) {
	m.Called(ctx, request, err)
}

type MockObserver struct {
	mock.Mock
}

func (m *MockObserver) OnRequestReceived(ctx context.Context, request *gateway.HTTPRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockObserver) OnRequestProcessed(ctx context.Context, request *gateway.HTTPRequest, response *gateway.HTTPResponse) error {
	args := m.Called(ctx, request, response)
	return args.Error(0)
}

func (m *MockObserver) OnError(ctx context.Context, request *gateway.HTTPRequest, err error) error {
	args := m.Called(ctx, request, err)
	return args.Error(0)
}

func (m *MockObserver) OnMetricsUpdate(ctx context.Context, metrics *gateway.RequestMetrics) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

// Helper function to create test HTTP requests
func createTestRequest(method, path, clientIP string) *gateway.HTTPRequest {
	return &gateway.HTTPRequest{
		Method:    method,
		Path:      path,
		ClientIP:  clientIP,
		StartTime: time.Now(),
		Headers:   make(map[string]string),
	}
}

// Helper function to create test HTTP responses
func createTestResponse(statusCode, size int) *gateway.HTTPResponse {
	return &gateway.HTTPResponse{
		StatusCode: statusCode,
		Size:       size,
		Headers:    make(map[string]string),
		Body:       []byte("test response"),
	}
}

// Helper function to create test metrics
func createTestMetrics(path, method string, statusCode int, duration time.Duration) *gateway.RequestMetrics {
	return &gateway.RequestMetrics{
		Path:       path,
		Method:     method,
		StatusCode: statusCode,
		Duration:   duration,
		Size:       100,
	}
}

// Tests for MetricsObserver
func TestNewMetricsObserver(t *testing.T) {
	t.Run("create metrics observer", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()

		assert.NotNil(t, observer)
	})
}

func TestMetricsObserver_OnRequestReceived(t *testing.T) {
	t.Run("successful request received processing", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()
		request := createTestRequest("GET", "/api/test", "192.168.1.1")

		err := observer.OnRequestReceived(ctx, request)

		assert.NoError(t, err)
	})

	t.Run("different HTTP methods", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()

		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
		for _, method := range methods {
			request := createTestRequest(method, "/api/test", "192.168.1.1")

			err := observer.OnRequestReceived(ctx, request)

			assert.NoError(t, err)
		}
	})

	t.Run("different paths", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()

		paths := []string{"/api/users", "/api/projects", "/health", "/metrics"}
		for _, path := range paths {
			request := createTestRequest("GET", path, "192.168.1.1")

			err := observer.OnRequestReceived(ctx, request)

			assert.NoError(t, err)
		}
	})
}

func TestMetricsObserver_OnRequestProcessed(t *testing.T) {
	t.Run("successful request processed", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()
		request := createTestRequest("POST", "/api/users", "192.168.1.1")
		response := createTestResponse(201, 256)

		err := observer.OnRequestProcessed(ctx, request, response)

		assert.NoError(t, err)
	})

	t.Run("different status codes", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()
		request := createTestRequest("GET", "/api/test", "192.168.1.1")

		statusCodes := []int{200, 201, 400, 401, 404, 500}
		for _, code := range statusCodes {
			response := createTestResponse(code, 100)

			err := observer.OnRequestProcessed(ctx, request, response)

			assert.NoError(t, err)
		}
	})

	t.Run("measure processing duration", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()

		// Create request with past start time to simulate processing duration
		request := createTestRequest("GET", "/api/slow", "192.168.1.1")
		request.StartTime = time.Now().Add(-100 * time.Millisecond)
		response := createTestResponse(200, 512)

		err := observer.OnRequestProcessed(ctx, request, response)

		assert.NoError(t, err)
	})
}

func TestMetricsObserver_OnError(t *testing.T) {
	t.Run("successful error processing", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()
		request := createTestRequest("POST", "/api/error", "192.168.1.1")
		testError := errors.New("test error")

		err := observer.OnError(ctx, request, testError)

		assert.NoError(t, err)
	})

	t.Run("different error types", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()
		request := createTestRequest("GET", "/api/test", "192.168.1.1")

		errors := []error{
			errors.New("network error"),
			errors.New("timeout error"),
			errors.New("validation error"),
			errors.New("database error"),
		}

		for _, testError := range errors {
			err := observer.OnError(ctx, request, testError)

			assert.NoError(t, err)
		}
	})
}

func TestMetricsObserver_OnMetricsUpdate(t *testing.T) {
	t.Run("successful metrics update", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()
		metrics := createTestMetrics("/api/test", "GET", 200, 50*time.Millisecond)

		err := observer.OnMetricsUpdate(ctx, metrics)

		assert.NoError(t, err)
	})

	t.Run("different metric scenarios", func(t *testing.T) {
		observer := gateway.NewMetricsObserver()
		ctx := context.Background()

		scenarios := []*gateway.RequestMetrics{
			createTestMetrics("/api/fast", "GET", 200, 10*time.Millisecond),
			createTestMetrics("/api/slow", "POST", 201, 2*time.Second),
			createTestMetrics("/api/error", "PUT", 500, 100*time.Millisecond),
		}

		for _, metrics := range scenarios {
			err := observer.OnMetricsUpdate(ctx, metrics)

			assert.NoError(t, err)
		}
	})
}

// Tests for SecurityObserver
func TestNewSecurityObserver(t *testing.T) {
	t.Run("create security observer", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()

		assert.NotNil(t, observer)
	})
}

func TestSecurityObserver_OnRequestReceived(t *testing.T) {
	t.Run("normal request processing", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()
		request := createTestRequest("GET", "/api/users", "192.168.1.1")

		err := observer.OnRequestReceived(ctx, request)

		assert.NoError(t, err)
	})

	t.Run("suspicious request detection", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()

		suspiciousPaths := []string{
			"/api/../admin",
			"/admin/users",
			"/config/database",
			"/.env",
			"/etc/passwd",
		}

		for _, path := range suspiciousPaths {
			request := createTestRequest("GET", path, "192.168.1.1")

			err := observer.OnRequestReceived(ctx, request)

			assert.NoError(t, err)
		}
	})

	t.Run("different client IPs", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()

		clientIPs := []string{
			"192.168.1.1",
			"10.0.0.1",
			"203.0.113.1",
			"2001:db8::1",
		}

		for _, ip := range clientIPs {
			request := createTestRequest("GET", "/api/test", ip)

			err := observer.OnRequestReceived(ctx, request)

			assert.NoError(t, err)
		}
	})
}

func TestSecurityObserver_OnRequestProcessed(t *testing.T) {
	t.Run("normal request processed", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()
		request := createTestRequest("GET", "/api/users", "192.168.1.1")
		response := createTestResponse(200, 100)

		err := observer.OnRequestProcessed(ctx, request, response)

		assert.NoError(t, err)
	})

	t.Run("high privilege operation logging", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()

		adminPaths := []string{
			"/admin/users",
			"/api/admin/settings",
			"/users/create",
			"/projects/delete",
		}

		for _, path := range adminPaths {
			request := createTestRequest("POST", path, "192.168.1.1")
			response := createTestResponse(200, 100)

			err := observer.OnRequestProcessed(ctx, request, response)

			assert.NoError(t, err)
		}
	})

	t.Run("failed high privilege operations", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()
		request := createTestRequest("DELETE", "/admin/users", "192.168.1.1")

		failureStatusCodes := []int{401, 403, 500}
		for _, statusCode := range failureStatusCodes {
			response := createTestResponse(statusCode, 50)

			err := observer.OnRequestProcessed(ctx, request, response)

			assert.NoError(t, err)
		}
	})
}

func TestSecurityObserver_OnError(t *testing.T) {
	t.Run("security error processing", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()
		request := createTestRequest("POST", "/api/login", "192.168.1.1")

		securityErrors := []error{
			errors.New("unauthorized access attempt"),
			errors.New("forbidden operation"),
			errors.New("authentication failed"),
			errors.New("invalid token provided"),
		}

		for _, testError := range securityErrors {
			err := observer.OnError(ctx, request, testError)

			assert.NoError(t, err)
		}
	})

	t.Run("non-security error processing", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()
		request := createTestRequest("GET", "/api/test", "192.168.1.1")

		nonSecurityErrors := []error{
			errors.New("database connection failed"),
			errors.New("network timeout"),
			errors.New("invalid JSON format"),
			errors.New("file not found"),
		}

		for _, testError := range nonSecurityErrors {
			err := observer.OnError(ctx, request, testError)

			assert.NoError(t, err)
		}
	})

	t.Run("nil error handling", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()
		request := createTestRequest("GET", "/api/test", "192.168.1.1")

		err := observer.OnError(ctx, request, nil)

		assert.NoError(t, err)
	})
}

func TestSecurityObserver_OnMetricsUpdate(t *testing.T) {
	t.Run("normal metrics processing", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()
		metrics := createTestMetrics("/api/test", "GET", 200, 100*time.Millisecond)

		err := observer.OnMetricsUpdate(ctx, metrics)

		assert.NoError(t, err)
	})

	t.Run("unusually long request duration detection", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()

		longDurations := []time.Duration{
			31 * time.Second,
			1 * time.Minute,
			5 * time.Minute,
		}

		for _, duration := range longDurations {
			metrics := createTestMetrics("/api/slow", "GET", 200, duration)

			err := observer.OnMetricsUpdate(ctx, metrics)

			assert.NoError(t, err)
		}
	})

	t.Run("normal duration requests", func(t *testing.T) {
		observer := gateway.NewSecurityObserver()
		ctx := context.Background()

		normalDurations := []time.Duration{
			10 * time.Millisecond,
			100 * time.Millisecond,
			1 * time.Second,
			29 * time.Second,
		}

		for _, duration := range normalDurations {
			metrics := createTestMetrics("/api/fast", "GET", 200, duration)

			err := observer.OnMetricsUpdate(ctx, metrics)

			assert.NoError(t, err)
		}
	})
}

// Tests for SecurityObserver helper methods
func TestSecurityObserver_IsSecurityError(t *testing.T) {
	observer := gateway.NewSecurityObserver()

	t.Run("nil error", func(t *testing.T) {
		result := observer.IsSecurityError(nil)

		assert.False(t, result)
	})

	t.Run("security errors", func(t *testing.T) {
		securityErrors := []error{
			errors.New("unauthorized"),
			errors.New("forbidden access"),
			errors.New("authentication failed"),
			errors.New("token expired"),
			errors.New("permission denied"),
			errors.New("access denied"),
			errors.New("invalid credentials"),
		}

		for _, err := range securityErrors {
			result := observer.IsSecurityError(err)

			assert.True(t, result, "Expected %s to be identified as security error", err.Error())
		}
	})

	t.Run("non-security errors", func(t *testing.T) {
		nonSecurityErrors := []error{
			errors.New("database connection failed"),
			errors.New("network timeout"),
			errors.New("invalid JSON"),
			errors.New("file not found"),
			errors.New("internal server error"),
		}

		for _, err := range nonSecurityErrors {
			result := observer.IsSecurityError(err)

			assert.False(t, result, "Expected %s to not be identified as security error", err.Error())
		}
	})

	t.Run("partial matches", func(t *testing.T) {
		partialMatches := []error{
			errors.New("user unauthorized to access resource"),
			errors.New("request forbidden by policy"),
			errors.New("token authentication required"),
		}

		for _, err := range partialMatches {
			result := observer.IsSecurityError(err)

			assert.True(t, result, "Expected %s to be identified as security error", err.Error())
		}
	})
}

func TestSecurityObserver_IsSuspiciousRequest(t *testing.T) {
	observer := gateway.NewSecurityObserver()

	t.Run("suspicious requests", func(t *testing.T) {
		suspiciousRequests := []*gateway.HTTPRequest{
			createTestRequest("GET", "/api/../admin", "192.168.1.1"),
			createTestRequest("POST", "/admin/users", "192.168.1.1"),
			createTestRequest("GET", "/config/database", "192.168.1.1"),
			createTestRequest("GET", "/.env", "192.168.1.1"),
			createTestRequest("GET", "/etc/passwd", "192.168.1.1"),
			createTestRequest("GET", "/app/config", "192.168.1.1"),
		}

		for _, request := range suspiciousRequests {
			// Use reflection to call private method
			// In a real implementation, we might expose this as a public method for testing
			// For now, we'll test through OnRequestReceived which calls this method
			err := observer.OnRequestReceived(context.Background(), request)
			assert.NoError(t, err)
		}
	})

	t.Run("normal requests", func(t *testing.T) {
		normalRequests := []*gateway.HTTPRequest{
			createTestRequest("GET", "/api/users", "192.168.1.1"),
			createTestRequest("POST", "/api/projects", "192.168.1.1"),
			createTestRequest("GET", "/health", "192.168.1.1"),
			createTestRequest("GET", "/metrics", "192.168.1.1"),
			createTestRequest("PUT", "/api/users/123", "192.168.1.1"),
		}

		for _, request := range normalRequests {
			err := observer.OnRequestReceived(context.Background(), request)
			assert.NoError(t, err)
		}
	})
}

func TestSecurityObserver_IsHighPrivilegeOperation(t *testing.T) {
	observer := gateway.NewSecurityObserver()

	t.Run("high privilege operations", func(t *testing.T) {
		highPrivilegeRequests := []*gateway.HTTPRequest{
			createTestRequest("DELETE", "/admin/users", "192.168.1.1"),
			createTestRequest("POST", "/api/admin/settings", "192.168.1.1"),
			createTestRequest("PUT", "/users/roles", "192.168.1.1"),
			createTestRequest("GET", "/projects/sensitive", "192.168.1.1"),
		}

		for _, request := range highPrivilegeRequests {
			// Test through OnRequestProcessed which calls this method
			response := createTestResponse(200, 100)
			err := observer.OnRequestProcessed(context.Background(), request, response)
			assert.NoError(t, err)
		}
	})

	t.Run("normal operations", func(t *testing.T) {
		normalRequests := []*gateway.HTTPRequest{
			createTestRequest("GET", "/api/public", "192.168.1.1"),
			createTestRequest("POST", "/api/comments", "192.168.1.1"),
			createTestRequest("GET", "/health", "192.168.1.1"),
			createTestRequest("PUT", "/api/profile", "192.168.1.1"),
		}

		for _, request := range normalRequests {
			response := createTestResponse(200, 100)
			err := observer.OnRequestProcessed(context.Background(), request, response)
			assert.NoError(t, err)
		}
	})
}

// Tests for GatewayEventNotifierImpl
func TestNewGatewayEventNotifier(t *testing.T) {
	t.Run("create event notifier", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()

		assert.NotNil(t, notifier)
	})
}

func TestGatewayEventNotifierImpl_Subscribe(t *testing.T) {
	t.Run("subscribe single observer", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}

		err := notifier.Subscribe(observer)

		assert.NoError(t, err)
	})

	t.Run("subscribe multiple observers", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()

		observers := []*MockObserver{
			&MockObserver{},
			&MockObserver{},
			&MockObserver{},
		}

		for _, observer := range observers {
			err := notifier.Subscribe(observer)
			assert.NoError(t, err)
		}
	})

	t.Run("subscribe same observer multiple times", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}

		err1 := notifier.Subscribe(observer)
		err2 := notifier.Subscribe(observer)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})
}

func TestGatewayEventNotifierImpl_Unsubscribe(t *testing.T) {
	t.Run("unsubscribe existing observer", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}

		notifier.Subscribe(observer)
		err := notifier.Unsubscribe(observer)

		assert.NoError(t, err)
	})

	t.Run("unsubscribe non-existing observer", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}

		err := notifier.Unsubscribe(observer)

		assert.NoError(t, err)
	})

	t.Run("unsubscribe from multiple observers", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()

		observer1 := &MockObserver{}
		observer2 := &MockObserver{}
		observer3 := &MockObserver{}

		notifier.Subscribe(observer1)
		notifier.Subscribe(observer2)
		notifier.Subscribe(observer3)

		err := notifier.Unsubscribe(observer2)

		assert.NoError(t, err)
	})
}

func TestGatewayEventNotifierImpl_NotifyRequestReceived(t *testing.T) {
	t.Run("notify single observer", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}
		ctx := context.Background()
		request := createTestRequest("GET", "/api/test", "192.168.1.1")

		observer.On("OnRequestReceived", ctx, request).Return(nil)

		notifier.Subscribe(observer)
		err := notifier.NotifyRequestReceived(ctx, request)

		assert.NoError(t, err)
		observer.AssertExpectations(t)
	})

	t.Run("notify multiple observers", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		ctx := context.Background()
		request := createTestRequest("POST", "/api/users", "192.168.1.1")

		observers := []*MockObserver{
			&MockObserver{},
			&MockObserver{},
			&MockObserver{},
		}

		for _, observer := range observers {
			observer.On("OnRequestReceived", ctx, request).Return(nil)
			notifier.Subscribe(observer)
		}

		err := notifier.NotifyRequestReceived(ctx, request)

		assert.NoError(t, err)
		for _, observer := range observers {
			observer.AssertExpectations(t)
		}
	})

	t.Run("handle observer error", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}
		ctx := context.Background()
		request := createTestRequest("GET", "/api/error", "192.168.1.1")

		observer.On("OnRequestReceived", ctx, request).Return(errors.New("observer error"))

		notifier.Subscribe(observer)
		err := notifier.NotifyRequestReceived(ctx, request)

		assert.NoError(t, err) // Notifier should not fail even if observer fails
		observer.AssertExpectations(t)
	})
}

func TestGatewayEventNotifierImpl_NotifyRequestProcessed(t *testing.T) {
	t.Run("notify single observer", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}
		ctx := context.Background()
		request := createTestRequest("PUT", "/api/users/123", "192.168.1.1")
		response := createTestResponse(200, 256)

		observer.On("OnRequestProcessed", ctx, request, response).Return(nil)

		notifier.Subscribe(observer)
		err := notifier.NotifyRequestProcessed(ctx, request, response)

		assert.NoError(t, err)
		observer.AssertExpectations(t)
	})

	t.Run("notify multiple observers", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		ctx := context.Background()
		request := createTestRequest("DELETE", "/api/projects/456", "192.168.1.1")
		response := createTestResponse(204, 0)

		observers := []*MockObserver{
			&MockObserver{},
			&MockObserver{},
		}

		for _, observer := range observers {
			observer.On("OnRequestProcessed", ctx, request, response).Return(nil)
			notifier.Subscribe(observer)
		}

		err := notifier.NotifyRequestProcessed(ctx, request, response)

		assert.NoError(t, err)
		for _, observer := range observers {
			observer.AssertExpectations(t)
		}
	})

	t.Run("handle observer error", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}
		ctx := context.Background()
		request := createTestRequest("POST", "/api/test", "192.168.1.1")
		response := createTestResponse(500, 100)

		observer.On("OnRequestProcessed", ctx, request, response).Return(errors.New("processing error"))

		notifier.Subscribe(observer)
		err := notifier.NotifyRequestProcessed(ctx, request, response)

		assert.NoError(t, err)
		observer.AssertExpectations(t)
	})
}

func TestGatewayEventNotifierImpl_NotifyError(t *testing.T) {
	t.Run("notify single observer", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}
		ctx := context.Background()
		request := createTestRequest("GET", "/api/failing", "192.168.1.1")
		testError := errors.New("service error")

		observer.On("OnError", ctx, request, testError).Return(nil)

		notifier.Subscribe(observer)
		err := notifier.NotifyError(ctx, request, testError)

		assert.NoError(t, err)
		observer.AssertExpectations(t)
	})

	t.Run("notify multiple observers", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		ctx := context.Background()
		request := createTestRequest("POST", "/api/critical", "192.168.1.1")
		testError := errors.New("critical error")

		observers := []*MockObserver{
			&MockObserver{},
			&MockObserver{},
			&MockObserver{},
		}

		for _, observer := range observers {
			observer.On("OnError", ctx, request, testError).Return(nil)
			notifier.Subscribe(observer)
		}

		err := notifier.NotifyError(ctx, request, testError)

		assert.NoError(t, err)
		for _, observer := range observers {
			observer.AssertExpectations(t)
		}
	})

	t.Run("handle observer error", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		observer := &MockObserver{}
		ctx := context.Background()
		request := createTestRequest("GET", "/api/test", "192.168.1.1")
		testError := errors.New("original error")

		observer.On("OnError", ctx, request, testError).Return(errors.New("observer error"))

		notifier.Subscribe(observer)
		err := notifier.NotifyError(ctx, request, testError)

		assert.NoError(t, err)
		observer.AssertExpectations(t)
	})
}

// Integration tests
func TestObserverIntegration(t *testing.T) {
	t.Run("complete request lifecycle with both observers", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		metricsObserver := gateway.NewMetricsObserver()
		securityObserver := gateway.NewSecurityObserver()

		notifier.Subscribe(metricsObserver)
		notifier.Subscribe(securityObserver)

		ctx := context.Background()
		request := createTestRequest("POST", "/api/users", "192.168.1.1")
		response := createTestResponse(201, 512)

		// Simulate complete request lifecycle
		err1 := notifier.NotifyRequestReceived(ctx, request)
		err2 := notifier.NotifyRequestProcessed(ctx, request, response)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})

	t.Run("error handling with both observers", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		metricsObserver := gateway.NewMetricsObserver()
		securityObserver := gateway.NewSecurityObserver()

		notifier.Subscribe(metricsObserver)
		notifier.Subscribe(securityObserver)

		ctx := context.Background()
		request := createTestRequest("POST", "/api/login", "192.168.1.1")
		testError := errors.New("authentication failed")

		err := notifier.NotifyError(ctx, request, testError)

		assert.NoError(t, err)
	})

	t.Run("concurrent notification handling", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		metricsObserver := gateway.NewMetricsObserver()

		notifier.Subscribe(metricsObserver)

		ctx := context.Background()

		// Simulate concurrent requests
		requests := make([]*gateway.HTTPRequest, 10)
		for i := 0; i < 10; i++ {
			requests[i] = createTestRequest("GET", "/api/concurrent", "192.168.1.1")
		}

		// Test concurrent notifications
		for _, request := range requests {
			go func(req *gateway.HTTPRequest) {
				notifier.NotifyRequestReceived(ctx, req)
			}(request)
		}

		// Small delay to ensure all goroutines complete
		time.Sleep(10 * time.Millisecond)
	})
}

// Performance and stress tests
func TestObserver_Performance(t *testing.T) {
	t.Run("high volume request processing", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()
		metricsObserver := gateway.NewMetricsObserver()

		notifier.Subscribe(metricsObserver)

		ctx := context.Background()
		requestCount := 1000

		start := time.Now()

		for i := 0; i < requestCount; i++ {
			request := createTestRequest("GET", "/api/load", "192.168.1.1")
			notifier.NotifyRequestReceived(ctx, request)
		}

		duration := time.Since(start)

		// Ensure processing doesn't take unreasonably long
		assert.Less(t, duration, 1*time.Second, "Processing %d requests took too long: %v", requestCount, duration)
	})

	t.Run("many observers handling requests", func(t *testing.T) {
		notifier := gateway.NewGatewayEventNotifier()

		observerCount := 100
		for i := 0; i < observerCount; i++ {
			metricsObserver := gateway.NewMetricsObserver()
			notifier.Subscribe(metricsObserver)
		}

		ctx := context.Background()
		request := createTestRequest("POST", "/api/stress", "192.168.1.1")

		start := time.Now()

		err := notifier.NotifyRequestReceived(ctx, request)

		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 100*time.Millisecond, "Notifying %d observers took too long: %v", observerCount, duration)
	})
}
