package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// StandardConfigValidator implements ConfigValidator interface
type StandardConfigValidator struct {
	rules []ValidationRule
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() ConfigValidator {
	return &StandardConfigValidator{
		rules: make([]ValidationRule, 0),
	}
}

// AddRule adds a validation rule
func (v *StandardConfigValidator) AddRule(rule ValidationRule) error {
	if rule.Key == "" {
		return fmt.Errorf("validation rule key cannot be empty")
	}

	// Validate the rule itself
	if rule.Type != "" {
		validTypes := []string{"string", "int", "float", "bool", "duration", "slice"}
		if !contains(validTypes, rule.Type) {
			return fmt.Errorf("invalid validation type: %s", rule.Type)
		}
	}

	if rule.Pattern != "" {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern for key %s: %w", rule.Key, err)
		}
	}

	v.rules = append(v.rules, rule)
	return nil
}

// Validate validates configuration data against all rules
func (v *StandardConfigValidator) Validate(data ConfigData) error {
	var errors []string

	for _, rule := range v.rules {
		if err := v.validateRule(rule, data); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetRules returns all validation rules
func (v *StandardConfigValidator) GetRules() []ValidationRule {
	return v.rules
}

// validateRule validates a single rule against configuration data
func (v *StandardConfigValidator) validateRule(rule ValidationRule, data ConfigData) error {
	value, exists := data[rule.Key]

	// Check if required field is missing
	if rule.Required && !exists {
		return fmt.Errorf("required field '%s' is missing", rule.Key)
	}

	// If field is not required and doesn't exist, skip validation
	if !exists {
		return nil
	}

	// Type validation
	if rule.Type != "" {
		if err := v.validateType(rule.Key, value, rule.Type); err != nil {
			return err
		}
	}

	// Range validation
	if rule.MinValue != nil || rule.MaxValue != nil {
		if err := v.validateRange(rule.Key, value, rule.MinValue, rule.MaxValue); err != nil {
			return err
		}
	}

	// Pattern validation
	if rule.Pattern != "" {
		if err := v.validatePattern(rule.Key, value, rule.Pattern); err != nil {
			return err
		}
	}

	// Custom validator
	if rule.Validator != nil {
		if err := rule.Validator(value); err != nil {
			return fmt.Errorf("custom validation failed for '%s': %w", rule.Key, err)
		}
	}

	return nil
}

// validateType validates the type of a configuration value
func (v *StandardConfigValidator) validateType(key string, value interface{}, expectedType string) error {
	// Handle nil values
	if value == nil {
		return fmt.Errorf("field '%s' is nil, expected %s", key, expectedType)
	}

	actualType := reflect.TypeOf(value).Kind()

	switch expectedType {
	case "string":
		if actualType != reflect.String {
			return fmt.Errorf("field '%s' must be a string, got %s", key, actualType)
		}
	case "int":
		if actualType != reflect.Int && actualType != reflect.Int64 && actualType != reflect.Float64 {
			return fmt.Errorf("field '%s' must be an integer, got %s", key, actualType)
		}
	case "float":
		if actualType != reflect.Float64 && actualType != reflect.Int && actualType != reflect.Int64 {
			return fmt.Errorf("field '%s' must be a number, got %s", key, actualType)
		}
	case "bool":
		if actualType != reflect.Bool {
			return fmt.Errorf("field '%s' must be a boolean, got %s", key, actualType)
		}
	case "duration":
		if actualType != reflect.String {
			return fmt.Errorf("field '%s' must be a duration string, got %s", key, actualType)
		}
	case "slice":
		if actualType != reflect.Slice {
			return fmt.Errorf("field '%s' must be a slice, got %s", key, actualType)
		}
	}

	return nil
}

// validateRange validates numeric ranges
func (v *StandardConfigValidator) validateRange(key string, value interface{}, minValue, maxValue interface{}) error {
	var numValue float64
	var ok bool

	// Convert value to float64 for comparison
	switch val := value.(type) {
	case int:
		numValue = float64(val)
		ok = true
	case int64:
		numValue = float64(val)
		ok = true
	case float64:
		numValue = val
		ok = true
	default:
		return fmt.Errorf("field '%s' is not numeric, cannot validate range", key)
	}

	if !ok {
		return fmt.Errorf("field '%s' is not numeric", key)
	}

	if minValue != nil {
		if min, ok := convertToFloat64(minValue); ok && numValue < min {
			return fmt.Errorf("field '%s' value %v is below minimum %v", key, numValue, min)
		}
	}

	if maxValue != nil {
		if max, ok := convertToFloat64(maxValue); ok && numValue > max {
			return fmt.Errorf("field '%s' value %v is above maximum %v", key, numValue, max)
		}
	}

	return nil
}

// validatePattern validates string patterns using regex
func (v *StandardConfigValidator) validatePattern(key string, value interface{}, pattern string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field '%s' is not a string, cannot validate pattern", key)
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern for field '%s': %w", key, err)
	}

	if !regex.MatchString(str) {
		return fmt.Errorf("field '%s' value '%s' does not match pattern '%s'", key, str, pattern)
	}

	return nil
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func convertToFloat64(value interface{}) (float64, bool) {
	switch val := value.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}
