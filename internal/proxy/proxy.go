package proxy

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

// ReverseProxy creates a reverse proxy for a service
func ReverseProxy(serviceConfig ServiceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse target URL
		target, err := url.Parse(serviceConfig.BaseURL)
		if err != nil {
			log.Error().Err(err).Str("service", serviceConfig.Name).Msg("Invalid service URL")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Service configuration error"})
			return
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
			req.Header.Set("X-Forwarded-Service", serviceConfig.Name)
			req.Header.Set("X-Request-ID", c.GetString("request_id"))

			log.Debug().
				Str("service", serviceConfig.Name).
				Str("original_path", c.Request.URL.Path).
				Str("target_path", req.URL.Path).
				Str("target_host", req.URL.Host).
				Msg("Proxying request")
		}

		// Handle proxy errors
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error().
				Err(err).
				Str("service", serviceConfig.Name).
				Str("path", r.URL.Path).
				Msg("Proxy error")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, `{"error": "Service %s unavailable", "service": "%s"}`,
				serviceConfig.Name, serviceConfig.Name)
		}

		// Execute the proxy
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthCheck creates a health check handler for a service
func HealthCheck(serviceConfig ServiceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Make a health check request to the service
		healthURL := serviceConfig.BaseURL + serviceConfig.HealthPath

		resp, err := http.Get(healthURL)
		if err != nil {
			log.Error().
				Err(err).
				Str("service", serviceConfig.Name).
				Str("health_url", healthURL).
				Msg("Health check failed")

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"service": serviceConfig.Name,
				"status":  "unhealthy",
				"error":   err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			c.JSON(http.StatusOK, gin.H{
				"service": serviceConfig.Name,
				"status":  "healthy",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"service": serviceConfig.Name,
				"status":  "unhealthy",
				"code":    resp.StatusCode,
			})
		}
	}
}
