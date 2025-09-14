package utilities_service_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/config"
	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHandler for testing HTTP functionality
type mockHandler struct {
	called bool
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.called = true
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func TestService_New(t *testing.T) {
	tests := []struct {
		name      string
		svcName   string
		version   string
		config    *config.Config
		wantPanic bool
	}{
		{
			name:    "valid service creation",
			svcName: "test-service",
			version: "1.0.0",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantPanic: false,
		},
		{
			name:    "service with empty name",
			svcName: "",
			version: "1.0.0",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 8081,
				},
				Logging: config.LoggingConfig{
					Level:  "debug",
					Format: "console",
				},
			},
			wantPanic: false,
		},
		{
			name:    "service with empty version",
			svcName: "test-service",
			version: "",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "0.0.0.0",
					Port: 9000,
				},
				Logging: config.LoggingConfig{
					Level:  "warn",
					Format: "json",
				},
			},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					utilities.New(tt.svcName, tt.version, tt.config)
				})
				return
			}

			service := utilities.New(tt.svcName, tt.version, tt.config)
			require.NotNil(t, service)
			assert.Equal(t, tt.svcName, service.Name)
			assert.Equal(t, tt.version, service.Version)
			assert.Equal(t, tt.config, service.Config)
			assert.Nil(t, service.HTTPServer)
		})
	}
}

func TestService_Initialize(t *testing.T) {
	t.Run("successful initialization with all observability disabled", func(t *testing.T) {
		config := &config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
			},
			Logging: config.LoggingConfig{
				Level:  "info",
				Format: "json",
			},
			Observability: config.ObservabilityConfig{
				MetricsEnabled: false,
				TracingEnabled: false,
			},
		}

		service := utilities.New("test-service", "1.0.0", config)
		require.NotNil(t, service)

		err := service.Initialize()
		assert.NoError(t, err)
	})

	t.Run("successful initialization with tracing enabled", func(t *testing.T) {
		config := &config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8082,
			},
			Logging: config.LoggingConfig{
				Level:  "warn",
				Format: "json",
			},
			Observability: config.ObservabilityConfig{
				MetricsEnabled: false,
				TracingEnabled: true,
				JaegerEndpoint: "http://localhost:14268/api/traces",
			},
		}

		service := utilities.New("test-service-tracing", "1.0.0", config)
		require.NotNil(t, service)

		err := service.Initialize()
		assert.NoError(t, err)
	})
}

// Test metrics initialization separately to avoid conflicts
func TestService_Initialize_WithMetrics(t *testing.T) {
	config := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8081,
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "console",
		},
		Observability: config.ObservabilityConfig{
			MetricsEnabled: true,
			TracingEnabled: false,
		},
	}

	service := utilities.New("test-service-metrics", "1.0.0", config)
	require.NotNil(t, service)

	err := service.Initialize()
	assert.NoError(t, err)
}

func TestService_SetupHTTPServer(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		handler  http.Handler
		wantAddr string
	}{
		{
			name: "standard HTTP server setup",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
			},
			handler:  &mockHandler{},
			wantAddr: "localhost:8080",
		},
		{
			name: "HTTP server with different host and port",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "0.0.0.0",
					Port: 9000,
				},
			},
			handler:  http.DefaultServeMux,
			wantAddr: "0.0.0.0:9000",
		},
		{
			name: "HTTP server with nil handler",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "127.0.0.1",
					Port: 3000,
				},
			},
			handler:  nil,
			wantAddr: "127.0.0.1:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := utilities.New("test-service", "1.0.0", tt.config)
			require.NotNil(t, service)

			service.SetupHTTPServer(tt.handler)

			require.NotNil(t, service.HTTPServer)
			assert.Equal(t, tt.wantAddr, service.HTTPServer.Addr)
			assert.Equal(t, tt.handler, service.HTTPServer.Handler)
			assert.Equal(t, 15*time.Second, service.HTTPServer.ReadTimeout)
			assert.Equal(t, 15*time.Second, service.HTTPServer.WriteTimeout)
			assert.Equal(t, 60*time.Second, service.HTTPServer.IdleTimeout)
		})
	}
}

func TestService_Start_NoHTTPServer(t *testing.T) {
	config := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	service := utilities.New("test-service", "1.0.0", config)
	require.NotNil(t, service)

	// Don't setup HTTP server
	err := service.Start()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP server not configured")
}

func TestService_AddHealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		serviceName  string
		version      string
		expectedBody string
	}{
		{
			name:         "standard health check",
			serviceName:  "test-service",
			version:      "1.0.0",
			expectedBody: `{"status":"ok","service":"test-service","version":"1.0.0"}`,
		},
		{
			name:         "health check with different service name",
			serviceName:  "api-gateway",
			version:      "2.1.0",
			expectedBody: `{"status":"ok","service":"api-gateway","version":"2.1.0"}`,
		},
		{
			name:         "health check with empty version",
			serviceName:  "user-service",
			version:      "",
			expectedBody: `{"status":"ok","service":"user-service","version":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
			}

			service := utilities.New(tt.serviceName, tt.version, config)
			require.NotNil(t, service)

			// Create a new ServeMux and add health check
			mux := http.NewServeMux()
			service.AddHealthCheck(mux)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			// Serve the request
			mux.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestService_AddMetricsEndpoint(t *testing.T) {
	t.Run("metrics endpoint disabled", func(t *testing.T) {
		config := &config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
			},
			Observability: config.ObservabilityConfig{
				MetricsEnabled: false,
			},
		}

		service := utilities.New("test-service", "1.0.0", config)
		require.NotNil(t, service)

		// Create a new ServeMux and add metrics endpoint
		mux := http.NewServeMux()
		service.AddMetricsEndpoint(mux)

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		// Serve the request
		mux.ServeHTTP(w, req)

		// Should return 404 since endpoint not registered
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("metrics endpoint configuration check", func(t *testing.T) {
		config := &config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
			},
			Observability: config.ObservabilityConfig{
				MetricsEnabled: true,
			},
		}

		service := utilities.New("test-service", "1.0.0", config)
		require.NotNil(t, service)

		// Should be able to call AddMetricsEndpoint without error
		mux := http.NewServeMux()
		assert.NotPanics(t, func() {
			service.AddMetricsEndpoint(mux)
		})
	})
}

func TestService_HTTPServerTimeouts(t *testing.T) {
	config := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	service := utilities.New("test-service", "1.0.0", config)
	require.NotNil(t, service)

	handler := &mockHandler{}
	service.SetupHTTPServer(handler)

	require.NotNil(t, service.HTTPServer)

	// Test timeout values are correctly set
	assert.Equal(t, 15*time.Second, service.HTTPServer.ReadTimeout)
	assert.Equal(t, 15*time.Second, service.HTTPServer.WriteTimeout)
	assert.Equal(t, 60*time.Second, service.HTTPServer.IdleTimeout)
}

func TestService_ServiceConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "complete service configuration",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				Logging: config.LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: config.ObservabilityConfig{
					MetricsEnabled: false, // Disable to avoid conflicts
					TracingEnabled: true,
					JaegerEndpoint: "http://localhost:14268/api/traces",
				},
			},
			wantErr: false,
		},
		{
			name: "minimal service configuration",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 3000,
				},
				Logging: config.LoggingConfig{
					Level:  "error",
					Format: "console",
				},
				Observability: config.ObservabilityConfig{
					MetricsEnabled: false,
					TracingEnabled: false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := utilities.New("config-test", "1.2.3", tt.config)
			require.NotNil(t, service)

			// Test service properties
			assert.Equal(t, "config-test", service.Name)
			assert.Equal(t, "1.2.3", service.Version)
			assert.Equal(t, tt.config, service.Config)

			// Test initialization
			err := service.Initialize()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestService_LifecycleIntegration tests the full service lifecycle
func TestService_LifecycleIntegration(t *testing.T) {
	config := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Observability: config.ObservabilityConfig{
			MetricsEnabled: false, // Disable to avoid conflicts
			TracingEnabled: false,
		},
	}

	service := utilities.New("integration-test", "1.0.0", config)
	require.NotNil(t, service)

	// 1. Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// 2. Setup HTTP server with handler
	mux := http.NewServeMux()
	service.AddHealthCheck(mux)
	service.AddMetricsEndpoint(mux)

	// Add a test endpoint
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test endpoint"))
	})

	service.SetupHTTPServer(mux)
	require.NotNil(t, service.HTTPServer)

	// 3. Test that the server configuration is correct
	assert.Equal(t, "localhost:8080", service.HTTPServer.Addr)
	assert.Equal(t, mux, service.HTTPServer.Handler)

	// 4. Test endpoints work by creating test server
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	// Test health endpoint
	resp, err := http.Get(testServer.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test metrics endpoint (should return 404 since metrics disabled)
	resp, err = http.Get(testServer.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Test custom endpoint
	resp, err = http.Get(testServer.URL + "/test")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
} // TestService_ErrorScenarios tests various error conditions
func TestService_ErrorScenarios(t *testing.T) {
	t.Run("nil config handling", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Service creation with nil config panicked as expected: %v", r)
			}
		}()

		service := utilities.New("test-service", "1.0.0", nil)
		// If we get here without panic, the service handled nil config gracefully
		if service != nil {
			assert.Nil(t, service.Config)
		}
	})

	t.Run("starting service without initialization", func(t *testing.T) {
		config := &config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
			},
		}

		service := utilities.New("test-service", "1.0.0", config)
		require.NotNil(t, service)

		handler := &mockHandler{}
		service.SetupHTTPServer(handler)

		// Try to use service without initialization
		// This should still work for basic HTTP functionality
		assert.NotNil(t, service.HTTPServer)
		assert.Equal(t, handler, service.HTTPServer.Handler)
	})
}

// TestService_ConcurrentRequests tests handling of concurrent requests
func TestService_ConcurrentRequests(t *testing.T) {
	config := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	service := utilities.New("concurrent-test", "1.0.0", config)
	require.NotNil(t, service)

	err := service.Initialize()
	require.NoError(t, err)

	// Setup handler that simulates some processing time
	requestCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		time.Sleep(10 * time.Millisecond) // Simulate processing
		fmt.Fprintf(w, "Request %d completed", requestCount)
	})

	service.AddHealthCheck(mux)
	service.SetupHTTPServer(mux)

	// Create test server
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	// Make concurrent requests
	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(testServer.URL + "/slow")
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
				return
			}
			results <- nil
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// Verify health check still works
	resp, err := http.Get(testServer.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestService_EdgeCases tests various edge cases and boundary conditions
func TestService_EdgeCases(t *testing.T) {
	t.Run("service with extreme port numbers", func(t *testing.T) {
		tests := []struct {
			name string
			port int
		}{
			{"minimum port", 1},
			{"maximum port", 65535},
			{"standard HTTP port", 80},
			{"standard HTTPS port", 443},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &config.Config{
					Server: config.ServerConfig{
						Host: "localhost",
						Port: tt.port,
					},
				}

				service := utilities.New("edge-test", "1.0.0", config)
				require.NotNil(t, service)

				service.SetupHTTPServer(&mockHandler{})
				require.NotNil(t, service.HTTPServer)

				expectedAddr := fmt.Sprintf("localhost:%d", tt.port)
				assert.Equal(t, expectedAddr, service.HTTPServer.Addr)
			})
		}
	})

	t.Run("service with special characters in name and version", func(t *testing.T) {
		specialNames := []struct {
			name    string
			version string
		}{
			{"test-service_v1", "1.0.0-beta"},
			{"service.name", "2.1.0+build.123"},
			{"service@domain", "1.0.0-alpha.1"},
			{"", ""}, // empty values
		}

		for _, tt := range specialNames {
			config := &config.Config{
				Server: config.ServerConfig{Host: "localhost", Port: 8080},
			}

			service := utilities.New(tt.name, tt.version, config)
			require.NotNil(t, service)

			assert.Equal(t, tt.name, service.Name)
			assert.Equal(t, tt.version, service.Version)

			// Test health check with special characters
			mux := http.NewServeMux()
			service.AddHealthCheck(mux)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			expectedBody := fmt.Sprintf(`{"status":"ok","service":"%s","version":"%s"}`, tt.name, tt.version)
			assert.Equal(t, expectedBody, w.Body.String())
		}
	})
}
