package validation

import (
	"github.com/EliasRanz/ai-code-gen/internal/domain/user"
	"github.com/go-playground/validator/v10"
)

// Validator wraps the validator library
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{validate: validator.New()}
}

// ValidateStruct validates a struct
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateUser validates a user struct
func (v *Validator) ValidateUser(user *user.User) error {
	return v.validate.Struct(user)
}
