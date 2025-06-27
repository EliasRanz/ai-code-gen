package llm

import (
	"net/http"
	"time"
)

// HTTPClientInterface interface for dependency injection and testing
type HTTPClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient implements HTTPClientInterface using standard library
type DefaultHTTPClient struct {
	client *http.Client
}

// NewDefaultHTTPClient creates a new default HTTP client
func NewDefaultHTTPClient(timeout time.Duration) *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Do executes an HTTP request
func (c *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
