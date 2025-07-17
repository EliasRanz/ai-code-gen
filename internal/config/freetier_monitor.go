package config

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FreeTierMonitor provides automated checks to prevent accidental paid API usage
type FreeTierMonitor struct {
	validator    ConfigValidator
	alertChannel chan FreeTierAlert
	rules        []FreeTierRule
}

// FreeTierAlert represents a free tier violation alert
type FreeTierAlert struct {
	Timestamp   time.Time
	Service     string
	Provider    string
	Violation   string
	ConfigValue interface{}
	Limit       interface{}
	Severity    AlertSeverity
}

// AlertSeverity defines the severity level of free tier alerts
type AlertSeverity string

const (
	SeverityWarning  AlertSeverity = "warning"
	SeverityError    AlertSeverity = "error"
	SeverityCritical AlertSeverity = "critical"
)

// FreeTierRule defines a rule for free tier monitoring
type FreeTierRule struct {
	Service     string
	Provider    string
	ConfigKey   string
	MaxValue    interface{}
	CheckType   string // "max", "exact", "contains"
	Severity    AlertSeverity
	Description string
}

// NewFreeTierMonitor creates a new free tier monitor
func NewFreeTierMonitor() *FreeTierMonitor {
	monitor := &FreeTierMonitor{
		validator:    NewConfigValidator(),
		alertChannel: make(chan FreeTierAlert, 100),
		rules:        make([]FreeTierRule, 0),
	}

	// Add default free tier rules
	monitor.addDefaultRules()
	return monitor
}

// addDefaultRules adds standard free tier monitoring rules
func (m *FreeTierMonitor) addDefaultRules() {
	// OpenAI free tier rules
	m.rules = append(m.rules, []FreeTierRule{
		{
			Service:     "ai",
			Provider:    "openai",
			ConfigKey:   "max_tokens",
			MaxValue:    2000,
			CheckType:   "max",
			Severity:    SeverityError,
			Description: "OpenAI free tier max tokens limit",
		},
		{
			Service:     "ai",
			Provider:    "openai",
			ConfigKey:   "api_key",
			MaxValue:    "sk-",
			CheckType:   "contains",
			Severity:    SeverityCritical,
			Description: "OpenAI API key should not be production key",
		},
		{
			Service:     "ai",
			Provider:    "openai",
			ConfigKey:   "free_tier_only",
			MaxValue:    true,
			CheckType:   "exact",
			Severity:    SeverityCritical,
			Description: "Free tier enforcement must be enabled",
		},
	}...)

	// vLLM free tier rules
	m.rules = append(m.rules, []FreeTierRule{
		{
			Service:     "ai",
			Provider:    "vllm",
			ConfigKey:   "max_tokens",
			MaxValue:    1500,
			CheckType:   "max",
			Severity:    SeverityError,
			Description: "vLLM free tier max tokens limit",
		},
		{
			Service:     "ai",
			Provider:    "vllm",
			ConfigKey:   "concurrent_requests",
			MaxValue:    5,
			CheckType:   "max",
			Severity:    SeverityWarning,
			Description: "vLLM concurrent request limit",
		},
	}...)
}

// ValidateFreeTierCompliance validates configuration against free tier rules
func (m *FreeTierMonitor) ValidateFreeTierCompliance(ctx context.Context, data ConfigData) error {
	var violations []FreeTierAlert

	for _, rule := range m.rules {
		if alert := m.checkRule(rule, data); alert != nil {
			violations = append(violations, *alert)

			// Send alert to channel
			select {
			case m.alertChannel <- *alert:
			default:
				// Channel full, log error
			}
		}
	}

	if len(violations) > 0 {
		return m.formatViolationError(violations)
	}

	return nil
}

// checkRule checks a single free tier rule against configuration data
func (m *FreeTierMonitor) checkRule(rule FreeTierRule, data ConfigData) *FreeTierAlert {
	// Build the key path for nested configuration
	keyPath := fmt.Sprintf("%s.%s.%s", rule.Service, rule.Provider, rule.ConfigKey)
	value, exists := m.getNestedValue(data, keyPath)

	if !exists {
		// Missing required free tier configuration
		if rule.Severity == SeverityCritical {
			return &FreeTierAlert{
				Timestamp:   time.Now(),
				Service:     rule.Service,
				Provider:    rule.Provider,
				Violation:   fmt.Sprintf("Missing required free tier config: %s", rule.ConfigKey),
				ConfigValue: nil,
				Limit:       rule.MaxValue,
				Severity:    rule.Severity,
			}
		}
		return nil
	}

	// Check the rule based on type
	violated := false
	switch rule.CheckType {
	case "max":
		violated = m.checkMaxValue(value, rule.MaxValue)
	case "exact":
		violated = m.checkExactValue(value, rule.MaxValue)
	case "contains":
		violated = m.checkContainsValue(value, rule.MaxValue)
	}

	if violated {
		return &FreeTierAlert{
			Timestamp:   time.Now(),
			Service:     rule.Service,
			Provider:    rule.Provider,
			Violation:   rule.Description,
			ConfigValue: value,
			Limit:       rule.MaxValue,
			Severity:    rule.Severity,
		}
	}

	return nil
}

// Helper methods for rule checking
func (m *FreeTierMonitor) checkMaxValue(value, maxValue interface{}) bool {
	valFloat, valOk := convertToFloat64(value)
	maxFloat, maxOk := convertToFloat64(maxValue)
	return valOk && maxOk && valFloat > maxFloat
}

func (m *FreeTierMonitor) checkExactValue(value, expectedValue interface{}) bool {
	return value != expectedValue
}

func (m *FreeTierMonitor) checkContainsValue(value, expectedValue interface{}) bool {
	valueStr, ok1 := value.(string)
	expectedStr, ok2 := expectedValue.(string)
	if !ok1 || !ok2 {
		return false
	}
	return !contains([]string{valueStr}, expectedStr)
}

// getNestedValue retrieves a value from nested configuration data using dot notation
func (m *FreeTierMonitor) getNestedValue(data ConfigData, keyPath string) (interface{}, bool) {
	keys := strings.Split(keyPath, ".")
	current := data

	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key, return the value
			value, exists := current[key]
			return value, exists
		}

		// Navigate deeper into nested structure
		if nested, ok := current[key].(map[string]interface{}); ok {
			current = nested
		} else {
			return nil, false
		}
	}

	return nil, false
}

// formatViolationError formats multiple violations into a single error
func (m *FreeTierMonitor) formatViolationError(violations []FreeTierAlert) error {
	var messages []string
	for _, violation := range violations {
		msg := fmt.Sprintf("[%s] %s.%s: %s (value: %v, limit: %v)",
			violation.Severity,
			violation.Service,
			violation.Provider,
			violation.Violation,
			violation.ConfigValue,
			violation.Limit)
		messages = append(messages, msg)
	}

	return fmt.Errorf("free tier violations detected: %s", strings.Join(messages, "; "))
}

// GetAlertChannel returns the alert channel for monitoring
func (m *FreeTierMonitor) GetAlertChannel() <-chan FreeTierAlert {
	return m.alertChannel
}

// AddCustomRule adds a custom free tier rule
func (m *FreeTierMonitor) AddCustomRule(rule FreeTierRule) error {
	if rule.Service == "" || rule.Provider == "" || rule.ConfigKey == "" {
		return fmt.Errorf("service, provider, and config_key are required for free tier rules")
	}

	m.rules = append(m.rules, rule)
	return nil
}
