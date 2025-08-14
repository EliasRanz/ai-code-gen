package observability_test

import (
	"errors"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for validation
type TestUser struct {
	Name  string `validate:"required,min=2,max=50"`
	Email string `validate:"required,email"`
	Age   int    `validate:"min=0,max=120"`
}

type InvalidStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestValidatorCreation(t *testing.T) {
	t.Run("NewValidator", func(t *testing.T) {
		validator := observability.NewValidator()
		require.NotNil(t, validator)
	})

	t.Run("NewUserValidatorAdapter", func(t *testing.T) {
		adapter := observability.NewUserValidatorAdapter()
		require.NotNil(t, adapter)
	})
}

func TestStructValidation(t *testing.T) {
	validator := observability.NewValidator()
	require.NotNil(t, validator)

	t.Run("Valid Struct", func(t *testing.T) {
		validUser := TestUser{
			Name:  "John Doe",
			Email: "john.doe@example.com",
			Age:   30,
		}

		err := validator.ValidateStruct(validUser)
		assert.NoError(t, err)

		isValid := validator.IsValidStruct(validUser)
		assert.True(t, isValid)
	})

	t.Run("Invalid Struct - Missing Required Fields", func(t *testing.T) {
		invalidUser := TestUser{
			Name:  "", // Required but empty
			Email: "", // Required but empty
			Age:   -5, // Below minimum
		}

		err := validator.ValidateStruct(invalidUser)
		assert.Error(t, err)

		isValid := validator.IsValidStruct(invalidUser)
		assert.False(t, isValid)

		// Test validation error processing
		validationErrors := validator.GetValidationErrors(err)
		assert.NotEmpty(t, validationErrors)
		assert.True(t, len(validationErrors) >= 3) // At least 3 validation errors
	})

	t.Run("Invalid Struct - Invalid Email", func(t *testing.T) {
		invalidUser := TestUser{
			Name:  "John Doe",
			Email: "invalid-email", // Invalid email format
			Age:   30,
		}

		err := validator.ValidateStruct(invalidUser)
		assert.Error(t, err)

		validationErrors := validator.GetValidationErrors(err)
		assert.NotEmpty(t, validationErrors)

		// Check that email validation error is present
		hasEmailError := false
		for _, ve := range validationErrors {
			if ve.Field == "Email" && ve.Tag == "email" {
				hasEmailError = true
				break
			}
		}
		assert.True(t, hasEmailError)
	})

	t.Run("Invalid Struct - String Length Validation", func(t *testing.T) {
		invalidUser := TestUser{
			Name:  "A", // Below minimum length
			Email: "valid@example.com",
			Age:   30,
		}

		err := validator.ValidateStruct(invalidUser)
		assert.Error(t, err)

		validationErrors := validator.GetValidationErrors(err)
		assert.NotEmpty(t, validationErrors)

		// Check that min length validation error is present
		hasMinError := false
		for _, ve := range validationErrors {
			if ve.Field == "Name" && ve.Tag == "min" {
				hasMinError = true
				break
			}
		}
		assert.True(t, hasMinError)
	})

	t.Run("Invalid Struct - Range Validation", func(t *testing.T) {
		invalidUser := TestUser{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   150, // Above maximum
		}

		err := validator.ValidateStruct(invalidUser)
		assert.Error(t, err)

		validationErrors := validator.GetValidationErrors(err)
		assert.NotEmpty(t, validationErrors)

		// Check that max validation error is present
		hasMaxError := false
		for _, ve := range validationErrors {
			if ve.Field == "Age" && ve.Tag == "max" {
				hasMaxError = true
				break
			}
		}
		assert.True(t, hasMaxError)
	})
}

func TestUserValidation(t *testing.T) {
	validator := observability.NewValidator()
	require.NotNil(t, validator)

	t.Run("Valid User - Generic Interface", func(t *testing.T) {
		validUser := TestUser{
			Name:  "Jane Doe",
			Email: "jane.doe@example.com",
			Age:   25,
		}

		err := validator.ValidateUser(validUser)
		assert.NoError(t, err)
	})

	t.Run("Invalid User - Generic Interface", func(t *testing.T) {
		invalidUser := TestUser{
			Name:  "",
			Email: "invalid-email",
			Age:   -1,
		}

		err := validator.ValidateUser(invalidUser)
		assert.Error(t, err)
	})

	t.Run("User Adapter Validation", func(t *testing.T) {
		adapter := observability.NewUserValidatorAdapter()
		require.NotNil(t, adapter)

		validUser := TestUser{
			Name:  "Bob Smith",
			Email: "bob.smith@example.com",
			Age:   40,
		}

		err := adapter.ValidateUser(validUser)
		assert.NoError(t, err)

		err = adapter.ValidateStruct(validUser)
		assert.NoError(t, err)
	})
}

func TestValidationErrorProcessing(t *testing.T) {
	validator := observability.NewValidator()
	require.NotNil(t, validator)

	t.Run("Nil Error", func(t *testing.T) {
		validationErrors := validator.GetValidationErrors(nil)
		assert.Empty(t, validationErrors)

		// Test global function as well
		globalErrors := observability.GetValidationErrors(nil)
		assert.Empty(t, globalErrors)
	})

	t.Run("Non-Validation Error", func(t *testing.T) {
		nonValidationError := errors.New("generic error")
		validationErrors := validator.GetValidationErrors(nonValidationError)
		assert.Empty(t, validationErrors)

		globalErrors := observability.GetValidationErrors(nonValidationError)
		assert.Empty(t, globalErrors)
	})

	t.Run("Validation Errors with Different Field Types", func(t *testing.T) {
		complexStruct := struct {
			StringField  string  `validate:"required"`
			IntField     int     `validate:"min=1"`
			UintField    uint    `validate:"min=1"`
			FloatField   float64 `validate:"min=0.1"`
			BoolField    bool    `validate:"required"`
			PointerField *string `validate:"required"`
		}{
			StringField:  "",    // Invalid - required but empty
			IntField:     0,     // Invalid - below minimum
			UintField:    0,     // Invalid - below minimum
			FloatField:   0.0,   // Invalid - below minimum
			BoolField:    false, // This should be valid (false is a valid bool)
			PointerField: nil,   // Invalid - required but nil
		}

		err := validator.ValidateStruct(complexStruct)
		assert.Error(t, err)

		validationErrors := validator.GetValidationErrors(err)
		assert.NotEmpty(t, validationErrors)

		// Check that we have validation errors for different field types
		fieldTypes := make(map[string]bool)
		for _, ve := range validationErrors {
			fieldTypes[ve.Field] = true
			assert.NotEmpty(t, ve.Tag)
			// Value can be empty or non-empty, both are valid
		}

		// Should have errors for multiple fields
		assert.True(t, len(fieldTypes) >= 3)
	})
}

func TestCustomValidation(t *testing.T) {
	val := observability.NewValidator()
	require.NotNil(t, val)

	t.Run("Register Custom Validation", func(t *testing.T) {
		// Define a custom validation function
		customValidation := func(fl validator.FieldLevel) bool {
			return fl.Field().String() == "custom_value"
		}

		err := val.RegisterCustomValidation("custom", customValidation)
		assert.NoError(t, err)

		// Test struct with custom validation
		testStruct := struct {
			CustomField string `validate:"custom"`
		}{
			CustomField: "custom_value", // Valid
		}

		err = val.ValidateStruct(testStruct)
		assert.NoError(t, err)

		// Test with invalid value
		testStruct.CustomField = "invalid_value"
		err = val.ValidateStruct(testStruct)
		assert.Error(t, err)
	})

	t.Run("Register Invalid Custom Validation", func(t *testing.T) {
		// Try to register with existing tag name - this should panic
		invalidValidation := func(fl validator.FieldLevel) bool {
			return true
		}

		assert.Panics(t, func() {
			val.RegisterCustomValidation("required", invalidValidation)
		}, "Should panic when trying to register validation with reserved tag name")
	})
}

func TestValidationErrorFieldValues(t *testing.T) {
	validator := observability.NewValidator()
	require.NotNil(t, validator)

	// Test different field types to ensure getFieldValue works correctly
	testStruct := struct {
		StringField string  `validate:"required"`
		IntField    int     `validate:"min=10"`
		UintField   uint    `validate:"min=10"`
		FloatField  float64 `validate:"min=10.5"`
		BoolField   bool    `validate:"eq=true"`
	}{
		StringField: "",      // Will trigger required error
		IntField:    5,       // Will trigger min error
		UintField:   uint(5), // Will trigger min error
		FloatField:  5.5,     // Will trigger min error
		BoolField:   false,   // Will trigger eq error
	}

	err := validator.ValidateStruct(testStruct)
	assert.Error(t, err)

	validationErrors := validator.GetValidationErrors(err)
	assert.NotEmpty(t, validationErrors)

	// Verify that field values are converted to strings properly
	for _, ve := range validationErrors {
		assert.NotNil(t, ve.Field)
		assert.NotNil(t, ve.Tag)
		// Value should be a string representation
		assert.IsType(t, "", ve.Value)
	}
}

func TestValidationIntegration(t *testing.T) {
	t.Run("End-to-End Validation Flow", func(t *testing.T) {
		validator := observability.NewValidator()
		require.NotNil(t, validator)

		// Create a complex validation scenario
		user := TestUser{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   25,
		}

		// 1. Validate successfully
		err := validator.ValidateStruct(user)
		assert.NoError(t, err)

		// 2. Check that it's valid
		isValid := validator.IsValidStruct(user)
		assert.True(t, isValid)

		// 3. Validate with the user-specific method
		err = validator.ValidateUser(user)
		assert.NoError(t, err)

		// 4. Test adapter functionality
		adapter := observability.NewUserValidatorAdapter()
		err = adapter.ValidateUser(user)
		assert.NoError(t, err)
		err = adapter.ValidateStruct(user)
		assert.NoError(t, err)

		// 5. Break the user and test error handling
		user.Email = "invalid-email"
		user.Age = -1

		err = validator.ValidateStruct(user)
		assert.Error(t, err)

		validationErrors := validator.GetValidationErrors(err)
		assert.NotEmpty(t, validationErrors)
		assert.True(t, len(validationErrors) >= 2)
	})
}
