package utilities_service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/utilities"
	"github.com/stretchr/testify/assert"
)

func TestDomainErrorTypes(t *testing.T) {
	tests := []struct {
		name         string
		errorFunc    func() error
		expectedType string
		checkFunc    func(error) bool
	}{
		{
			name: "validation_error_creation",
			errorFunc: func() error {
				return utilities.NewValidationError("invalid input", errors.New("base error"))
			},
			expectedType: "validation",
			checkFunc:    utilities.IsValidationError,
		},
		{
			name: "not_found_error_creation",
			errorFunc: func() error {
				return utilities.NewNotFoundError("resource not found")
			},
			expectedType: "not_found",
			checkFunc:    utilities.IsNotFoundError,
		},
		{
			name: "conflict_error_creation",
			errorFunc: func() error {
				return utilities.NewConflictError("resource conflict")
			},
			expectedType: "conflict",
			checkFunc:    utilities.IsConflictError,
		},
		{
			name: "unauthorized_error_creation",
			errorFunc: func() error {
				return utilities.NewUnauthorizedError("unauthorized access")
			},
			expectedType: "unauthorized",
			checkFunc:    nil, // No checker function for unauthorized
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedType)

			if tt.checkFunc != nil {
				assert.True(t, tt.checkFunc(err))
			}
		})
	}
}

func TestDomainErrorWrapping(t *testing.T) {
	t.Run("error_with_cause", func(t *testing.T) {
		baseErr := errors.New("base error")
		domainErr := utilities.NewValidationError("validation failed", baseErr)

		assert.Error(t, domainErr)
		assert.Contains(t, domainErr.Error(), "validation")
		assert.Contains(t, domainErr.Error(), "validation failed")
		assert.Contains(t, domainErr.Error(), "base error")

		// Test unwrapping
		unwrapped := errors.Unwrap(domainErr)
		assert.Equal(t, baseErr, unwrapped)
	})

	t.Run("error_without_cause", func(t *testing.T) {
		domainErr := utilities.NewNotFoundError("resource not found")

		assert.Error(t, domainErr)
		assert.Contains(t, domainErr.Error(), "not_found")
		assert.Contains(t, domainErr.Error(), "resource not found")
		assert.NotContains(t, domainErr.Error(), "(")

		// Test unwrapping returns nil
		unwrapped := errors.Unwrap(domainErr)
		assert.Nil(t, unwrapped)
	})
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name         string
		error        error
		isValidation bool
		isNotFound   bool
		isConflict   bool
	}{
		{
			name:         "validation_error",
			error:        utilities.NewValidationError("invalid", nil),
			isValidation: true,
			isNotFound:   false,
			isConflict:   false,
		},
		{
			name:         "not_found_error",
			error:        utilities.NewNotFoundError("not found"),
			isValidation: false,
			isNotFound:   true,
			isConflict:   false,
		},
		{
			name:         "conflict_error",
			error:        utilities.NewConflictError("conflict"),
			isValidation: false,
			isNotFound:   false,
			isConflict:   true,
		},
		{
			name:         "regular_error",
			error:        errors.New("regular error"),
			isValidation: false,
			isNotFound:   false,
			isConflict:   false,
		},
		{
			name:         "nil_error",
			error:        nil,
			isValidation: false,
			isNotFound:   false,
			isConflict:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValidation, utilities.IsValidationError(tt.error))
			assert.Equal(t, tt.isNotFound, utilities.IsNotFoundError(tt.error))
			assert.Equal(t, tt.isConflict, utilities.IsConflictError(tt.error))
		})
	}
}

func TestUserIDType(t *testing.T) {
	tests := []struct {
		name     string
		userID   utilities.UserID
		expected string
		isEmpty  bool
	}{
		{
			name:     "valid_user_id",
			userID:   utilities.UserID("user-123"),
			expected: "user-123",
			isEmpty:  false,
		},
		{
			name:     "empty_user_id",
			userID:   utilities.UserID(""),
			expected: "",
			isEmpty:  true,
		},
		{
			name:     "uuid_user_id",
			userID:   utilities.UserID("550e8400-e29b-41d4-a716-446655440000"),
			expected: "550e8400-e29b-41d4-a716-446655440000",
			isEmpty:  false,
		},
		{
			name:     "numeric_user_id",
			userID:   utilities.UserID("12345"),
			expected: "12345",
			isEmpty:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.userID.String())
			assert.Equal(t, tt.isEmpty, tt.userID.IsEmpty())
		})
	}
}

func TestProjectIDType(t *testing.T) {
	tests := []struct {
		name      string
		projectID utilities.ProjectID
		expected  string
	}{
		{
			name:      "valid_project_id",
			projectID: utilities.ProjectID("project-456"),
			expected:  "project-456",
		},
		{
			name:      "empty_project_id",
			projectID: utilities.ProjectID(""),
			expected:  "",
		},
		{
			name:      "uuid_project_id",
			projectID: utilities.ProjectID("550e8400-e29b-41d4-a716-446655440001"),
			expected:  "550e8400-e29b-41d4-a716-446655440001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.projectID.String())
		})
	}
}

func TestSessionIDType(t *testing.T) {
	tests := []struct {
		name      string
		sessionID utilities.SessionID
		expected  string
	}{
		{
			name:      "valid_session_id",
			sessionID: utilities.SessionID("session-789"),
			expected:  "session-789",
		},
		{
			name:      "empty_session_id",
			sessionID: utilities.SessionID(""),
			expected:  "",
		},
		{
			name:      "token_session_id",
			sessionID: utilities.SessionID("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"),
			expected:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.sessionID.String())
		})
	}
}

func TestTimestamps(t *testing.T) {
	t.Run("timestamps_initialization", func(t *testing.T) {
		now := time.Now()
		ts := utilities.Timestamps{
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.Equal(t, now, ts.CreatedAt)
		assert.Equal(t, now, ts.UpdatedAt)
	})

	t.Run("touch_updates_timestamp", func(t *testing.T) {
		now := time.Now()
		ts := utilities.Timestamps{
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Wait a small amount to ensure time difference
		time.Sleep(time.Millisecond)

		ts.Touch()

		// CreatedAt should remain unchanged
		assert.Equal(t, now, ts.CreatedAt)

		// UpdatedAt should be newer than original time
		assert.True(t, ts.UpdatedAt.After(now))
	})

	t.Run("multiple_touches", func(t *testing.T) {
		now := time.Now()
		ts := utilities.Timestamps{
			CreatedAt: now,
			UpdatedAt: now,
		}

		ts.Touch()
		firstUpdate := ts.UpdatedAt

		time.Sleep(time.Millisecond)
		ts.Touch()
		secondUpdate := ts.UpdatedAt

		assert.Equal(t, now, ts.CreatedAt)
		assert.True(t, firstUpdate.After(now))
		assert.True(t, secondUpdate.After(firstUpdate))
	})
}

func TestPaginationParams(t *testing.T) {
	tests := []struct {
		name           string
		pagination     utilities.PaginationParams
		expectedValid  bool
		expectedError  string
		expectedOffset int32
	}{
		{
			name: "valid_first_page",
			pagination: utilities.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			expectedValid:  true,
			expectedOffset: 0,
		},
		{
			name: "valid_second_page",
			pagination: utilities.PaginationParams{
				Page:  2,
				Limit: 10,
			},
			expectedValid:  true,
			expectedOffset: 10,
		},
		{
			name: "valid_high_page",
			pagination: utilities.PaginationParams{
				Page:  5,
				Limit: 25,
			},
			expectedValid:  true,
			expectedOffset: 100,
		},
		{
			name: "valid_max_limit",
			pagination: utilities.PaginationParams{
				Page:  1,
				Limit: 100,
			},
			expectedValid:  true,
			expectedOffset: 0,
		},
		{
			name: "invalid_zero_page",
			pagination: utilities.PaginationParams{
				Page:  0,
				Limit: 10,
			},
			expectedValid:  false,
			expectedError:  "invalid input",
			expectedOffset: -10,
		},
		{
			name: "invalid_negative_page",
			pagination: utilities.PaginationParams{
				Page:  -1,
				Limit: 10,
			},
			expectedValid:  false,
			expectedError:  "invalid input",
			expectedOffset: -20,
		},
		{
			name: "invalid_zero_limit",
			pagination: utilities.PaginationParams{
				Page:  1,
				Limit: 0,
			},
			expectedValid:  false,
			expectedError:  "invalid input",
			expectedOffset: 0,
		},
		{
			name: "invalid_limit_too_high",
			pagination: utilities.PaginationParams{
				Page:  1,
				Limit: 101,
			},
			expectedValid:  false,
			expectedError:  "invalid input",
			expectedOffset: 0,
		},
		{
			name: "invalid_negative_limit",
			pagination: utilities.PaginationParams{
				Page:  1,
				Limit: -1,
			},
			expectedValid:  false,
			expectedError:  "invalid input",
			expectedOffset: 0, // (1-1) * -1 = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pagination.Validate()

			if tt.expectedValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			// Test offset calculation regardless of validation
			offset := tt.pagination.Offset()
			assert.Equal(t, tt.expectedOffset, offset)
		})
	}
}

func TestPaginationOffsetCalculations(t *testing.T) {
	tests := []struct {
		name           string
		page           int32
		limit          int32
		expectedOffset int32
	}{
		{name: "first_page_10_items", page: 1, limit: 10, expectedOffset: 0},
		{name: "second_page_10_items", page: 2, limit: 10, expectedOffset: 10},
		{name: "third_page_10_items", page: 3, limit: 10, expectedOffset: 20},
		{name: "first_page_25_items", page: 1, limit: 25, expectedOffset: 0},
		{name: "second_page_25_items", page: 2, limit: 25, expectedOffset: 25},
		{name: "tenth_page_5_items", page: 10, limit: 5, expectedOffset: 45},
		{name: "large_page_number", page: 100, limit: 10, expectedOffset: 990},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagination := utilities.PaginationParams{
				Page:  tt.page,
				Limit: tt.limit,
			}

			offset := pagination.Offset()
			assert.Equal(t, tt.expectedOffset, offset)
		})
	}
}

func TestErrorIntegration(t *testing.T) {
	t.Run("error_chain_validation", func(t *testing.T) {
		baseErr := errors.New("database connection failed")
		validationErr := utilities.NewValidationError("invalid user data", baseErr)

		// Test error messages
		assert.Contains(t, validationErr.Error(), "validation")
		assert.Contains(t, validationErr.Error(), "invalid user data")
		assert.Contains(t, validationErr.Error(), "database connection failed")

		// Test type checking
		assert.True(t, utilities.IsValidationError(validationErr))
		assert.False(t, utilities.IsNotFoundError(validationErr))
		assert.False(t, utilities.IsConflictError(validationErr))

		// Test unwrapping
		unwrapped := errors.Unwrap(validationErr)
		assert.Equal(t, baseErr, unwrapped)
	})

	t.Run("complex_error_scenarios", func(t *testing.T) {
		// Create a chain of errors
		repoErr := utilities.NewNotFoundError("user not found")
		serviceErr := utilities.NewValidationError("service validation failed", repoErr)

		// Test the service error
		assert.True(t, utilities.IsValidationError(serviceErr))
		assert.False(t, utilities.IsNotFoundError(serviceErr))

		// Unwrap to get repository error
		unwrapped := errors.Unwrap(serviceErr)
		assert.True(t, utilities.IsNotFoundError(unwrapped))

		// The chain should contain multiple error types
		assert.Contains(t, serviceErr.Error(), "validation")
		assert.Contains(t, serviceErr.Error(), "not_found")
	})
}
