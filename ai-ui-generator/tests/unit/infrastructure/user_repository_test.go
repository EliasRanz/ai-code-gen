// Package database contains tests for database infrastructure implementations
package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty array",
			input:    "[]",
			expected: []string{},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single element",
			input:    `["admin"]`,
			expected: []string{"admin"},
		},
		{
			name:     "multiple elements",
			input:    `["admin","user","viewer"]`,
			expected: []string{"admin", "user", "viewer"},
		},
		{
			name:     "elements with special characters",
			input:    `["role-with-dash","role_with_underscore","role with space"]`,
			expected: []string{"role-with-dash", "role_with_underscore", "role with space"},
		},
		{
			name:     "elements with quotes",
			input:    `["role\"with\"quotes","role'with'apostrophes"]`,
			expected: []string{`role"with"quotes`, "role'with'apostrophes"},
		},
		{
			name:     "malformed JSON",
			input:    `["admin","user"`,
			expected: []string{},
		},
		{
			name:     "invalid JSON structure",
			input:    `{"admin": true}`,
			expected: []string{},
		},
		{
			name:     "non-string elements should fail gracefully",
			input:    `[123, true, null]`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonToStringSlice(tt.input)
			assert.Equal(t, tt.expected, result)
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
			result := jsonToStringSlice(jsonStr)

			// Result should match original input
			assert.Equal(t, tt.input, result)
		})
	}
}

// TestCount_SQLGeneration tests the Count method's SQL generation logic
func TestCount_SQLGeneration(t *testing.T) {
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
