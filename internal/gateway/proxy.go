package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ServiceConfig holds configuration for a service proxy
type ServiceConfig struct {
	Name       string
	BaseURL    string
	HealthPath string
}

// ProxyService provides reverse proxy functionality
type ProxyService struct {
	config ServiceConfig
}

// NewProxyService creates a new proxy service
func NewProxyService(config ServiceConfig) *ProxyService {
	return &ProxyService{
		config: config,
	}
}

// ReverseProxy creates a reverse proxy handler for a service
func (p *ProxyService) ReverseProxy() func(Context) error {
	return func(ctx Context) error {
		// Parse target URL
		target, err := url.Parse(p.config.BaseURL)
		if err != nil {
			log.Error().Err(err).Str("service", p.config.Name).Msg("Invalid service URL")
			return p.handleProxyError(ctx, "Service configuration error", http.StatusInternalServerError)
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Customize the director to modify the request
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Remove the API prefix from the path
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api")

			// Add service-specific headers
			req.Header.Set("X-Forwarded-Service", p.config.Name)

			if requestID, ok := ctx.Get("request_id"); ok {
				if id, ok := requestID.(string); ok {
					req.Header.Set("X-Request-ID", id)
				}
			}

			log.Debug().
				Str("service", p.config.Name).
				Str("original_path", ctx.Request().URL.Path).
				Str("target_path", req.URL.Path).
				Str("target_host", req.URL.Host).
				Msg("Proxying request")
		}

		// Handle proxy errors
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error().
				Err(err).
				Str("service", p.config.Name).
				Str("path", r.URL.Path).
				Msg("Proxy error")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, `{"error": "Service %s unavailable", "service": "%s"}`,
				p.config.Name, p.config.Name)
		}

		// Execute the proxy
		proxy.ServeHTTP(ctx.Writer(), ctx.Request())
		return nil
	}
}

// HealthCheck creates a health check handler for a service
func (p *ProxyService) HealthCheck() func(Context) error {
	return func(ctx Context) error {
		// Make a health check request to the service
		healthURL := p.config.BaseURL + p.config.HealthPath

		resp, err := http.Get(healthURL)
		if err != nil {
			log.Error().
				Err(err).
				Str("service", p.config.Name).
				Str("health_url", healthURL).
				Msg("Health check failed")

			return p.handleHealthCheckError(ctx, err.Error())
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return p.handleHealthCheckSuccess(ctx)
		} else {
			return p.handleHealthCheckError(ctx, fmt.Sprintf("status code %d", resp.StatusCode))
		}
	}
}

// handleProxyError handles proxy errors
func (p *ProxyService) handleProxyError(ctx Context, message string, statusCode int) error {
	if ginCtx, ok := ctx.(*GinContextWrapper); ok {
		ginCtx.JSON(statusCode, map[string]interface{}{
			"error":   message,
			"service": p.config.Name,
		})
	}
	return fmt.Errorf("proxy error for service %s: %s", p.config.Name, message)
}

// handleHealthCheckSuccess handles successful health checks
func (p *ProxyService) handleHealthCheckSuccess(ctx Context) error {
	if ginCtx, ok := ctx.(*GinContextWrapper); ok {
		ginCtx.JSON(http.StatusOK, map[string]interface{}{
			"service": p.config.Name,
			"status":  "healthy",
		})
	}
	return nil
}

// handleHealthCheckError handles health check errors
func (p *ProxyService) handleHealthCheckError(ctx Context, errorMsg string) error {
	if ginCtx, ok := ctx.(*GinContextWrapper); ok {
		ginCtx.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"service": p.config.Name,
			"status":  "unhealthy",
			"error":   errorMsg,
		})
	}
	return fmt.Errorf("health check failed for service %s: %s", p.config.Name, errorMsg)
}

// Legacy proxy functions for backward compatibility
func CreateReverseProxy(serviceConfig ServiceConfig) func(Context) error {
	proxy := NewProxyService(serviceConfig)
	return proxy.ReverseProxy()
}

func CreateHealthCheck(serviceConfig ServiceConfig) func(Context) error {
	proxy := NewProxyService(serviceConfig)
	return proxy.HealthCheck()
}

// Gin-compatible proxy functions for main.go backward compatibility
func ReverseProxy(serviceConfig ServiceConfig) gin.HandlerFunc {
	proxy := NewProxyService(serviceConfig)
	proxyFunc := proxy.ReverseProxy()

	return func(c *gin.Context) {
		ctx := WrapGinContext(c)
		if err := proxyFunc(ctx); err != nil {
			// Error already handled by proxy function
			return
		}
	}
}

func HealthCheckHandler(serviceConfig ServiceConfig) gin.HandlerFunc {
	proxy := NewProxyService(serviceConfig)
	healthFunc := proxy.HealthCheck()

	return func(c *gin.Context) {
		ctx := WrapGinContext(c)
		if err := healthFunc(ctx); err != nil {
			// Error already handled by health check function
			return
		}
	}
}
