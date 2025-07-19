package gateway_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EliasRanz/ai-code-gen/internal/gateway"
)

func TestMetricsObserver_OnRequestReceived(t *testing.T) {
	observer := gateway.NewMetricsObserver()

	request := &gateway.HTTPRequest{
		Method:   "GET",
		Path:     "/api/test",
		ClientIP: "192.168.1.1",
	}

	err := observer.OnRequestReceived(context.Background(), request)
	assert.NoError(t, err)
}

func TestMetricsObserver_OnRequestProcessed(t *testing.T) {
	observer := gateway.NewMetricsObserver()

	request := &gateway.HTTPRequest{
		Method:    "POST",
		Path:      "/api/generate",
		StartTime: time.Now().Add(-100 * time.Millisecond),
		ClientIP:  "192.168.1.1",
	}

	response := &gateway.HTTPResponse{
		StatusCode: 200,
		Size:       1024,
	}

	err := observer.OnRequestProcessed(context.Background(), request, response)
	assert.NoError(t, err)
}

func TestMetricsObserver_OnError(t *testing.T) {
	observer := gateway.NewMetricsObserver()

	request := &gateway.HTTPRequest{
		Method: "GET",
		Path:   "/api/users",
	}

	testError := assert.AnError

	err := observer.OnError(context.Background(), request, testError)
	assert.NoError(t, err)
}

func TestMetricsObserver_OnMetricsUpdate(t *testing.T) {
	observer := gateway.NewMetricsObserver()

	metrics := &gateway.RequestMetrics{
		Path:       "/api/health",
		Method:     "GET",
		StatusCode: 200,
		Duration:   50 * time.Millisecond,
		Size:       256,
	}

	err := observer.OnMetricsUpdate(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestSecurityObserver_OnRequestReceived(t *testing.T) {
	observer := gateway.NewSecurityObserver()

	tests := []struct {
		name        string
		path        string
		expectAlert bool
	}{
		{
			name:        "normal request",
			path:        "/api/users",
			expectAlert: false,
		},
		{
			name:        "suspicious admin path",
			path:        "/admin/config",
			expectAlert: true,
		},
		{
			name:        "path traversal attempt",
			path:        "/api/../etc/passwd",
			expectAlert: true,
		},
		{
			name:        "env file access",
			path:        "/.env",
			expectAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &gateway.HTTPRequest{
				Method: "GET",
				Path:   tt.path,
			}

			err := observer.OnRequestReceived(context.Background(), request)
			assert.NoError(t, err)
		})
	}
}

func TestSecurityObserver_OnError(t *testing.T) {
	observer := gateway.NewSecurityObserver()

	tests := []struct {
		name           string
		error          error
		expectSecAlert bool
	}{
		{
			name:           "generic error",
			error:          assert.AnError,
			expectSecAlert: false,
		},
		{
			name:           "unauthorized error",
			error:          &testError{"unauthorized access"},
			expectSecAlert: true,
		},
		{
			name:           "forbidden error",
			error:          &testError{"forbidden operation"},
			expectSecAlert: true,
		},
		{
			name:           "authentication error",
			error:          &testError{"authentication failed"},
			expectSecAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &gateway.HTTPRequest{
				Method: "GET",
				Path:   "/api/test",
			}

			err := observer.OnError(context.Background(), request, tt.error)
			assert.NoError(t, err)
		})
	}
}

func TestSecurityObserver_OnRequestProcessed(t *testing.T) {
	observer := gateway.NewSecurityObserver()

	tests := []struct {
		name       string
		path       string
		statusCode int
		expectLog  bool
	}{
		{
			name:       "normal user request",
			path:       "/api/profile",
			statusCode: 200,
			expectLog:  false,
		},
		{
			name:       "admin operation",
			path:       "/admin/users",
			statusCode: 200,
			expectLog:  true,
		},
		{
			name:       "user management",
			path:       "/users/create",
			statusCode: 201,
			expectLog:  true,
		},
		{
			name:       "failed admin operation",
			path:       "/admin/config",
			statusCode: 500,
			expectLog:  false, // Only log successful high-privilege ops
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &gateway.HTTPRequest{
				Method: "POST",
				Path:   tt.path,
			}

			response := &gateway.HTTPResponse{
				StatusCode: tt.statusCode,
			}

			err := observer.OnRequestProcessed(context.Background(), request, response)
			assert.NoError(t, err)
		})
	}
}

func TestSecurityObserver_OnMetricsUpdate(t *testing.T) {
	observer := gateway.NewSecurityObserver()

	tests := []struct {
		name        string
		duration    time.Duration
		expectAlert bool
	}{
		{
			name:        "normal duration",
			duration:    100 * time.Millisecond,
			expectAlert: false,
		},
		{
			name:        "long duration",
			duration:    35 * time.Second,
			expectAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &gateway.RequestMetrics{
				Path:     "/api/test",
				Duration: tt.duration,
			}

			err := observer.OnMetricsUpdate(context.Background(), metrics)
			assert.NoError(t, err)
		})
	}
}

func TestGatewayEventNotifier_Subscribe(t *testing.T) {
	notifier := gateway.NewGatewayEventNotifier()
	observer := gateway.NewMetricsObserver()

	err := notifier.Subscribe(observer)
	assert.NoError(t, err)
}

func TestGatewayEventNotifier_Unsubscribe(t *testing.T) {
	notifier := gateway.NewGatewayEventNotifier()
	observer := gateway.NewMetricsObserver()

	// Subscribe first
	err := notifier.Subscribe(observer)
	require.NoError(t, err)

	// Then unsubscribe
	err = notifier.Unsubscribe(observer)
	assert.NoError(t, err)
}

func TestGatewayEventNotifier_NotifyRequestReceived(t *testing.T) {
	notifier := gateway.NewGatewayEventNotifier()
	observer := gateway.NewMetricsObserver()

	err := notifier.Subscribe(observer)
	require.NoError(t, err)

	request := &gateway.HTTPRequest{
		Method: "GET",
		Path:   "/api/test",
	}

	err = notifier.NotifyRequestReceived(context.Background(), request)
	assert.NoError(t, err)
}

func TestGatewayEventNotifier_NotifyRequestProcessed(t *testing.T) {
	notifier := gateway.NewGatewayEventNotifier()
	observer := gateway.NewMetricsObserver()

	err := notifier.Subscribe(observer)
	require.NoError(t, err)

	request := &gateway.HTTPRequest{
		Method:    "POST",
		Path:      "/api/generate",
		StartTime: time.Now().Add(-50 * time.Millisecond),
	}

	response := &gateway.HTTPResponse{
		StatusCode: 201,
		Size:       512,
	}

	err = notifier.NotifyRequestProcessed(context.Background(), request, response)
	assert.NoError(t, err)
}

func TestGatewayEventNotifier_NotifyError(t *testing.T) {
	notifier := gateway.NewGatewayEventNotifier()
	observer := gateway.NewSecurityObserver()

	err := notifier.Subscribe(observer)
	require.NoError(t, err)

	request := &gateway.HTTPRequest{
		Method: "GET",
		Path:   "/api/admin",
	}

	testError := &testError{"unauthorized access"}

	err = notifier.NotifyError(context.Background(), request, testError)
	assert.NoError(t, err)
}

func TestGatewayEventNotifier_MultipleObservers(t *testing.T) {
	notifier := gateway.NewGatewayEventNotifier()

	metricsObserver := gateway.NewMetricsObserver()
	securityObserver := gateway.NewSecurityObserver()

	err := notifier.Subscribe(metricsObserver)
	require.NoError(t, err)

	err = notifier.Subscribe(securityObserver)
	require.NoError(t, err)

	request := &gateway.HTTPRequest{
		Method:    "GET",
		Path:      "/api/test",
		StartTime: time.Now(),
	}

	// Should notify both observers without error
	err = notifier.NotifyRequestReceived(context.Background(), request)
	assert.NoError(t, err)

	response := &gateway.HTTPResponse{
		StatusCode: 200,
		Size:       256,
	}

	err = notifier.NotifyRequestProcessed(context.Background(), request, response)
	assert.NoError(t, err)
}

// Helper type for testing error conditions
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
