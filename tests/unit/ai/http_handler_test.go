// Package ai_test provides unit tests for AI HTTP handler
package ai_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EliasRanz/ai-code-gen/internal/ai"
	"github.com/EliasRanz/ai-code-gen/internal/infrastructure/observability"
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

// TestAIHandlerHealthCheck tests handler health check functionality
func TestAIHandlerHealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func() *ai.HTTPHandler
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "HealthCheck_MissingDependencies",
			setupHandler: func() *ai.HTTPHandler {
				mockLogger := &mockLogger{}
				return ai.NewHTTPHandler(nil, nil, mockLogger)
			},
			expectedError:  true,
			expectedErrMsg: "AI handler dependencies not properly initialized",
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

// TestAIHandlerValidateRoutes tests route validation functionality
func TestAIHandlerValidateRoutes(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func() *ai.HTTPHandler
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "ValidateRoutes_MissingService",
			setupHandler: func() *ai.HTTPHandler {
				mockLogger := &mockLogger{}
				return ai.NewHTTPHandler(nil, nil, mockLogger)
			},
			expectedError:  true,
			expectedErrMsg: "generate code service is required",
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
