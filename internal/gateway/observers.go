package gateway

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// MetricsObserver observes requests and updates metrics
type MetricsObserver struct {
	metricsCollector *MetricsCollector
	name             string
}

// NewMetricsObserver creates a new metrics observer
func NewMetricsObserver() *MetricsObserver {
	return &MetricsObserver{
		metricsCollector: NewMetricsCollector(),
		name:             "metrics-observer",
	}
}

// OnRequestReceived handles incoming request events
func (m *MetricsObserver) OnRequestReceived(ctx context.Context, request *HTTPRequest) error {
	m.metricsCollector.IncrementRequestCount(request.Path, request.Method)

	log.Debug().
		Str("observer", m.name).
		Str("path", request.Path).
		Str("method", request.Method).
		Msg("Observer: Request received")

	return nil
}

// OnRequestProcessed handles processed request events
func (m *MetricsObserver) OnRequestProcessed(ctx context.Context, request *HTTPRequest, response *HTTPResponse) error {
	duration := time.Since(request.StartTime)
	m.metricsCollector.RecordLatency(request.Path, duration)
	m.metricsCollector.IncrementResponseCode(response.StatusCode)

	log.Debug().
		Str("observer", m.name).
		Str("path", request.Path).
		Int("status_code", response.StatusCode).
		Dur("duration", duration).
		Msg("Observer: Request processed")

	return nil
}

// OnError handles error events
func (m *MetricsObserver) OnError(ctx context.Context, request *HTTPRequest, err error) error {
	log.Debug().
		Str("observer", m.name).
		Str("path", request.Path).
		Err(err).
		Msg("Observer: Request error")

	return nil
}

// OnMetricsUpdate handles metrics update events
func (m *MetricsObserver) OnMetricsUpdate(ctx context.Context, metrics *RequestMetrics) error {
	log.Debug().
		Str("observer", m.name).
		Str("path", metrics.Path).
		Dur("duration", metrics.Duration).
		Msg("Observer: Metrics update")

	return nil
}

// SecurityObserver observes security-related events
type SecurityObserver struct {
	alertManager *AlertManager
	logger       Logger
	name         string
}

// NewSecurityObserver creates a new security observer
func NewSecurityObserver() *SecurityObserver {
	return &SecurityObserver{
		alertManager: NewAlertManager(10), // Alert after 10 security events
		logger:       NewSecurityLogger(),
		name:         "security-observer",
	}
}

// OnRequestReceived handles incoming request events
func (s *SecurityObserver) OnRequestReceived(ctx context.Context, request *HTTPRequest) error {
	// Check for suspicious patterns
	if s.isSuspiciousRequest(request) {
		log.Warn().
			Str("observer", s.name).
			Str("path", request.Path).
			Str("client_ip", request.ClientIP).
			Msg("Observer: Suspicious request detected")
	}

	return nil
}

// OnRequestProcessed handles processed request events
func (s *SecurityObserver) OnRequestProcessed(ctx context.Context, request *HTTPRequest, response *HTTPResponse) error {
	// Log successful high-privilege operations
	if response.StatusCode == 200 && s.isHighPrivilegeOperation(request) {
		log.Info().
			Str("observer", s.name).
			Str("path", request.Path).
			Str("client_ip", request.ClientIP).
			Msg("Observer: High-privilege operation completed")
	}

	return nil
}

// OnError handles error events
func (s *SecurityObserver) OnError(ctx context.Context, request *HTTPRequest, err error) error {
	if s.IsSecurityError(err) {
		s.alertManager.SendSecurityAlert(ctx, request, err)
		s.logger.LogSecurityEvent(ctx, request, err)

		log.Error().
			Str("observer", s.name).
			Str("path", request.Path).
			Err(err).
			Msg("Observer: Security error processed")
	}

	return nil
}

// OnMetricsUpdate handles metrics update events
func (s *SecurityObserver) OnMetricsUpdate(ctx context.Context, metrics *RequestMetrics) error {
	// Monitor for unusual patterns
	if metrics.Duration > 30*time.Second {
		log.Warn().
			Str("observer", s.name).
			Str("path", metrics.Path).
			Dur("duration", metrics.Duration).
			Msg("Observer: Unusually long request duration")
	}

	return nil
}

// IsSecurityError determines if an error is security-related
func (s *SecurityObserver) IsSecurityError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := err.Error()
	securityKeywords := []string{
		"unauthorized", "forbidden", "authentication", "token",
		"permission", "access denied", "invalid credentials",
	}

	for _, keyword := range securityKeywords {
		if len(errorMsg) >= len(keyword) && errorMsg[:len(keyword)] == keyword {
			return true
		}
	}

	return false
}

// isSuspiciousRequest checks if a request is suspicious
func (s *SecurityObserver) isSuspiciousRequest(request *HTTPRequest) bool {
	// Simple heuristics - in practice, this would be more sophisticated
	suspiciousPaths := []string{
		"../", "admin", "config", ".env", "passwd",
	}

	for _, suspicious := range suspiciousPaths {
		if len(request.Path) >= len(suspicious) {
			for i := 0; i <= len(request.Path)-len(suspicious); i++ {
				if request.Path[i:i+len(suspicious)] == suspicious {
					return true
				}
			}
		}
	}

	return false
}

// isHighPrivilegeOperation checks if an operation requires high privileges
func (s *SecurityObserver) isHighPrivilegeOperation(request *HTTPRequest) bool {
	adminPaths := []string{"/admin", "/api/admin", "/users", "/projects"}

	for _, adminPath := range adminPaths {
		if len(request.Path) >= len(adminPath) && request.Path[:len(adminPath)] == adminPath {
			return true
		}
	}

	return false
}

// GatewayEventNotifierImpl manages event notifications to observers
type GatewayEventNotifierImpl struct {
	observers []GatewayEventObserver
	mu        sync.RWMutex
}

// NewGatewayEventNotifier creates a new event notifier
func NewGatewayEventNotifier() *GatewayEventNotifierImpl {
	return &GatewayEventNotifierImpl{
		observers: make([]GatewayEventObserver, 0),
	}
}

// Subscribe adds an observer to the notification list
func (n *GatewayEventNotifierImpl) Subscribe(observer GatewayEventObserver) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.observers = append(n.observers, observer)

	log.Debug().
		Int("observer_count", len(n.observers)).
		Msg("Observer subscribed to gateway events")

	return nil
}

// Unsubscribe removes an observer from the notification list
func (n *GatewayEventNotifierImpl) Unsubscribe(observer GatewayEventObserver) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for i, obs := range n.observers {
		if obs == observer {
			n.observers = append(n.observers[:i], n.observers[i+1:]...)
			break
		}
	}

	log.Debug().
		Int("observer_count", len(n.observers)).
		Msg("Observer unsubscribed from gateway events")

	return nil
}

// NotifyRequestReceived notifies all observers of an incoming request
func (n *GatewayEventNotifierImpl) NotifyRequestReceived(ctx context.Context, request *HTTPRequest) error {
	n.mu.RLock()
	observers := make([]GatewayEventObserver, len(n.observers))
	copy(observers, n.observers)
	n.mu.RUnlock()

	for _, observer := range observers {
		if err := observer.OnRequestReceived(ctx, request); err != nil {
			log.Error().Err(err).Msg("Observer failed to process request received event")
		}
	}

	return nil
}

// NotifyRequestProcessed notifies all observers of a processed request
func (n *GatewayEventNotifierImpl) NotifyRequestProcessed(ctx context.Context, request *HTTPRequest, response *HTTPResponse) error {
	n.mu.RLock()
	observers := make([]GatewayEventObserver, len(n.observers))
	copy(observers, n.observers)
	n.mu.RUnlock()

	for _, observer := range observers {
		if err := observer.OnRequestProcessed(ctx, request, response); err != nil {
			log.Error().Err(err).Msg("Observer failed to process request processed event")
		}
	}

	return nil
}

// NotifyError notifies all observers of an error
func (n *GatewayEventNotifierImpl) NotifyError(ctx context.Context, request *HTTPRequest, err error) error {
	n.mu.RLock()
	observers := make([]GatewayEventObserver, len(n.observers))
	copy(observers, n.observers)
	n.mu.RUnlock()

	for _, observer := range observers {
		if obsErr := observer.OnError(ctx, request, err); obsErr != nil {
			log.Error().Err(obsErr).Msg("Observer failed to process error event")
		}
	}

	return nil
}
