package gateway

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// AlertManager provides alerting functionality
type AlertManager struct {
	alertThreshold int
	alertCount     int
	mu             sync.RWMutex
}

// NewAlertManager creates a new alert manager
func NewAlertManager(threshold int) *AlertManager {
	return &AlertManager{
		alertThreshold: threshold,
	}
}

// SendSecurityAlert sends a security alert
func (a *AlertManager) SendSecurityAlert(ctx context.Context, request *HTTPRequest, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.alertCount++

	log.Warn().
		Str("path", request.Path).
		Str("method", request.Method).
		Str("client_ip", request.ClientIP).
		Err(err).
		Int("alert_count", a.alertCount).
		Msg("Security alert triggered")
}

// Logger provides structured logging
type Logger interface {
	LogSecurityEvent(ctx context.Context, request *HTTPRequest, err error)
}

// SecurityLogger implements security-specific logging
type SecurityLogger struct{}

// NewSecurityLogger creates a new security logger
func NewSecurityLogger() *SecurityLogger {
	return &SecurityLogger{}
}

// LogSecurityEvent logs security events
func (s *SecurityLogger) LogSecurityEvent(ctx context.Context, request *HTTPRequest, err error) {
	log.Error().
		Str("event_type", "security").
		Str("path", request.Path).
		Str("method", request.Method).
		Str("client_ip", request.ClientIP).
		Err(err).
		Msg("Security event logged")
}
