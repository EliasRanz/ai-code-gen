package observability

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

// UserInterface represents a minimal user interface for validation
type UserInterface interface {
	GetEmail() string
	GetName() string
}

// Validator provides validation functionality
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateStruct validates a struct using tags - compatible with user.Validator interface
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateUser validates a user struct specifically - compatible with user.Validator interface signature
func (v *Validator) ValidateUser(user interface{}) error {
	return v.validate.Struct(user)
}

// GetValidationErrors returns structured validation errors (additional functionality)
func (v *Validator) GetValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if err == nil {
		return validationErrors
	}

	// Handle validation errors
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validationErr {
			validationErrors = append(validationErrors, ValidationError{
				Field: err.Field(),
				Tag:   err.Tag(),
				Value: getFieldValue(err.Value()),
			})
		}
	}

	return validationErrors
}

// IsValidStruct returns true if struct passes validation
func (v *Validator) IsValidStruct(s interface{}) bool {
	return v.validate.Struct(s) == nil
}

// RegisterCustomValidation allows registering custom validation functions
func (v *Validator) RegisterCustomValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// GetValidationErrors returns human-readable validation errors (global function)
func GetValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if err == nil {
		return validationErrors
	}

	// Handle validation errors
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validationErr {
			validationErrors = append(validationErrors, ValidationError{
				Field: err.Field(),
				Tag:   err.Tag(),
				Value: getFieldValue(err.Value()),
			})
		}
	}

	return validationErrors
}

// getFieldValue converts field value to string
func getFieldValue(v interface{}) string {
	if v == nil {
		return ""
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.String:
		return v.(string)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v)
	case reflect.Bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
