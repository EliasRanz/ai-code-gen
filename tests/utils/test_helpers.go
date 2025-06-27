package utils

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelper provides common test utilities
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// AssertNoError checks that error is nil
func (h *TestHelper) AssertNoError(err error) {
	require.NoError(h.t, err)
}

// AssertEqual checks that values are equal
func (h *TestHelper) AssertEqual(expected, actual interface{}) {
	assert.Equal(h.t, expected, actual)
}

// AssertNotNil checks that value is not nil
func (h *TestHelper) AssertNotNil(value interface{}) {
	assert.NotNil(h.t, value)
}
