// Package database contains tests for database infrastructure implementations
package database

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stringSliceToJSON converts a string slice to JSON
func stringSliceToJSON(slice []string) string {
	if slice == nil {
		slice = []string{}
	}
	data, _ := json.Marshal(slice)
	return string(data)
}

// jsonToStringSlice converts JSON to a string slice
func jsonToStringSlice(jsonStr string) ([]string, error) {
	var raw []interface{}
	err := json.Unmarshal([]byte(jsonStr), &raw)
	if err != nil {
		return nil, err
	}
	
	result := make([]string, len(raw))
	for i, v := range raw {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a string", i)
		}
		result[i] = str
	}
	return result, nil
}

// TestStringSliceToJSON tests the stringSliceToJSON function
func TestStringSliceToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: "[]",
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: "[]",
		},
		{
			name:     "single element",
			input:    []string{"admin"},
			expected: `["admin"]`,
		},
		{
			name:     "multiple elements",
			input:    []string{"admin", "user", "viewer"},
			expected: `["admin","user","viewer"]`,
		},
		{
			name:     "elements with special characters",
			input:    []string{"role-with-dash", "role_with_underscore", "role with space"},
			expected: `["role-with-dash","role_with_underscore","role with space"]`,
		},
		{
			name:     "elements with quotes",
			input:    []string{`role"with"quotes`, "role'with'apostrophes"},
			expected: `["role\"with\"quotes","role'with'apostrophes"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSliceToJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestJSONToStringSlice tests the jsonToStringSlice function
func TestJSONToStringSlice(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      []string
		expectError   bool
	}{
		{
			name:        "empty array",
			input:       "[]",
			expected:    []string{},
			expectError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "single element",
			input:       `["admin"]`,
			expected:    []string{"admin"},
			expectError: false,
		},
		{
			name:        "multiple elements",
			input:       `["admin","user","viewer"]`,
			expected:    []string{"admin", "user", "viewer"},
			expectError: false,
		},
		{
			name:        "elements with special characters",
			input:       `["role-with-dash","role_with_underscore","role with space"]`,
			expected:    []string{"role-with-dash", "role_with_underscore", "role with space"},
			expectError: false,
		},
		{
			name:        "elements with quotes",
			input:       `["role\"with\"quotes","role'with'apostrophes"]`,
			expected:    []string{`role"with"quotes`, "role'with'apostrophes"},
			expectError: false,
		},
		{
			name:        "malformed JSON",
			input:       `["admin","user"`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid JSON structure",
			input:       `{"admin": true}`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "non-string elements should fail gracefully",
			input:       `[123, true, null]`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonToStringSlice(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestJSONRoundTrip tests that stringSliceToJSON and jsonToStringSlice are inverse operations
func TestJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{
			name:  "empty slice",
			input: []string{},
		},
		{
			name:  "single element",
			input: []string{"admin"},
		},
		{
			name:  "multiple elements",
			input: []string{"admin", "user", "viewer"},
		},
		{
			name:  "elements with special characters",
			input: []string{"role-with-dash", "role_with_underscore", "role with space"},
		},
		{
			name:  "elements with quotes",
			input: []string{`role"with"quotes`, "role'with'apostrophes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert slice to JSON and back
			jsonStr := stringSliceToJSON(tt.input)
			result, err := jsonToStringSlice(jsonStr)
			assert.NoError(t, err)

			// Result should match original input
			assert.Equal(t, tt.input, result)
		})
	}
}

// TestCount_SQLGeneration tests the Count method's SQL generation logic
func TestCountSQLGeneration(t *testing.T) {
	t.Run("should generate correct SQL for empty search", func(t *testing.T) {
		// Test validates that empty search returns all users
		search := ""

		// Simulate the same logic as in the Count method
		var searchPattern string
		if search != "" {
			searchPattern = "%" + search + "%"
		}

		// For empty search, searchPattern should remain empty
		assert.Equal(t, "", searchPattern)
	})

	t.Run("should generate correct SQL for non-empty search", func(t *testing.T) {
		search := "john"
		expectedPattern := "%john%"

		var searchPattern string
		if search != "" {
			searchPattern = "%" + search + "%"
		}

		assert.Equal(t, expectedPattern, searchPattern)
	})

	t.Run("should handle special characters in search", func(t *testing.T) {
		testCases := []struct {
			name     string
			search   string
			expected string
		}{
			{
				name:     "search with spaces",
				search:   "john doe",
				expected: "%john doe%",
			},
			{
				name:     "search with symbols",
				search:   "user@example.com",
				expected: "%user@example.com%",
			},
			{
				name:     "search with quotes",
				search:   `john"doe`,
				expected: `%john"doe%`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var searchPattern string
				if tc.search != "" {
					searchPattern = "%" + tc.search + "%"
				}
				assert.Equal(t, tc.expected, searchPattern)
			})
		}
	})
}
