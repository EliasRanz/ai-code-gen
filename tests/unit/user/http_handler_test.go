// Package user_test provides unit tests for user HTTP handler
package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/EliasRanz/ai-code-gen/internal/user"
)

// mockLogger is a simple mock logger for testing
type mockLogger struct {
	lastMessage string
	lastFields  map[string]interface{}
}

func (m *mockLogger) Debug(message string, fields ...map[string]interface{}) {
	m.lastMessage = message
	if len(fields) > 0 {
		m.lastFields = fields[0]
	}
}

func (m *mockLogger) Info(message string, fields ...map[string]interface{}) {
	m.lastMessage = message
	if len(fields) > 0 {
		m.lastFields = fields[0]
	}
}

func (m *mockLogger) Warn(message string, fields ...map[string]interface{}) {
	m.lastMessage = message
	if len(fields) > 0 {
		m.lastFields = fields[0]
	}
}

func (m *mockLogger) Error(message string, err error, fields ...map[string]interface{}) {
	m.lastMessage = message
	if len(fields) > 0 {
		m.lastFields = fields[0]
	}
}

func (m *mockLogger) Fatal(message string, err error, fields ...map[string]interface{}) {
	m.lastMessage = message
	if len(fields) > 0 {
		m.lastFields = fields[0]
	}
}

func (m *mockLogger) With(fields map[string]interface{}) observability.Logger {
	m.lastFields = fields
	return m
}

// TestUserHandlerHealthCheck tests handler health check functionality
func TestUserHandlerHealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func() *user.HTTPHandler
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "HealthCheck_MissingDependencies",
			setupHandler: func() *user.HTTPHandler {
				mockLogger := &mockLogger{}
				return user.NewHTTPHandler(nil, nil, nil, nil, nil, mockLogger)
			},
			expectedError:  true,
			expectedErrMsg: "user handler dependencies not properly initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler()
			err := handler.HealthCheck()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUserHandlerValidateRoutes tests route validation functionality
func TestUserHandlerValidateRoutes(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func() *user.HTTPHandler
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "ValidateRoutes_MissingUseCase",
			setupHandler: func() *user.HTTPHandler {
				mockLogger := &mockLogger{}
				return user.NewHTTPHandler(nil, nil, nil, nil, nil, mockLogger)
			},
			expectedError:  true,
			expectedErrMsg: "user creator use case is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler()
			err := handler.ValidateRoutes()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
