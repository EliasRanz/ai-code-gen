package observability

// UserValidatorAdapter adapts the observability.Validator to match user.Validator interface exactly
type UserValidatorAdapter struct {
	*Validator
}

// NewUserValidatorAdapter creates a new adapter that matches user.Validator interface exactly
func NewUserValidatorAdapter() *UserValidatorAdapter {
	return &UserValidatorAdapter{
		Validator: NewValidator(),
	}
}

// ValidateUser validates a user struct - matches user.Validator interface signature
func (v *UserValidatorAdapter) ValidateUser(user interface{}) error {
	return v.ValidateStruct(user)
}
