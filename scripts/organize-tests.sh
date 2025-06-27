#!/bin/bash
# scripts/organize-tests.sh
# Move all test files to proper test directory structure

set -e

echo "📁 Organizing test files according to coding standards..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create test directory structure
echo "Creating test directory structure..."
mkdir -p tests/{unit,integration,fixtures,utils}

# Create subdirectories mirroring source structure
mkdir -p tests/unit/{middleware,user,auth,generation,llm,application,infrastructure}

echo -e "${YELLOW}Moving test files to proper locations...${NC}"

# Find all test files not already in tests directory
test_files=$(find . -name "*_test.go" -not -path "./tests/*")

moved_count=0

for test_file in $test_files; do
    # Get the directory structure relative to the source
    rel_path=$(echo "$test_file" | sed 's|^\./||' | sed 's|internal/||')
    
    # Determine target directory based on source location
    if [[ "$test_file" == *"/middleware/"* ]]; then
        target_dir="tests/unit/middleware"
    elif [[ "$test_file" == *"/user/"* ]]; then
        target_dir="tests/unit/user"
    elif [[ "$test_file" == *"/auth/"* ]]; then
        target_dir="tests/unit/auth"
    elif [[ "$test_file" == *"/generation/"* ]]; then
        target_dir="tests/unit/generation"
    elif [[ "$test_file" == *"/llm/"* ]]; then
        target_dir="tests/unit/llm"
    elif [[ "$test_file" == *"/application/"* ]]; then
        target_dir="tests/unit/application"
    elif [[ "$test_file" == *"/infrastructure/"* ]]; then
        target_dir="tests/unit/infrastructure"
    else
        # Default to unit tests
        target_dir="tests/unit"
    fi
    
    # Create target directory if it doesn't exist
    mkdir -p "$target_dir"
    
    # Move the file
    filename=$(basename "$test_file")
    target_file="$target_dir/$filename"
    
    echo "Moving $test_file -> $target_file"
    mv "$test_file" "$target_file"
    moved_count=$((moved_count + 1))
done

echo ""
echo -e "${GREEN}✅ Moved $moved_count test files to proper locations${NC}"

# Create a test utilities file
echo "Creating test utilities..."
cat > tests/utils/test_helpers.go << 'EOF'
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
EOF

# Create test fixtures
echo "Creating test fixtures..."
cat > tests/fixtures/users.go << 'EOF'
package fixtures

import (
	pb "github.com/ai-code-gen/ai-ui-generator/api/proto/user"
)

// MockUser returns a mock user for testing
func MockUser() *pb.User {
	return &pb.User{
		Id:        "test-user-id",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarUrl: "https://example.com/avatar.jpg",
		Roles:     []string{"user"},
		CreatedAt: "2023-01-01T00:00:00Z",
		UpdatedAt: "2023-01-01T00:00:00Z",
	}
}

// MockProject returns a mock project for testing
func MockProject() *pb.Project {
	return &pb.Project{
		Id:          "test-project-id",
		UserId:      "test-user-id",
		Name:        "Test Project",
		Description: "A test project",
		Status:      pb.ProjectStatus_ACTIVE,
		CreatedAt:   "2023-01-01T00:00:00Z",
		UpdatedAt:   "2023-01-01T00:00:00Z",
	}
}
EOF

echo ""
echo -e "${GREEN}✅ Test organization complete!${NC}"
echo ""
echo "📋 Next steps:"
echo "1. Update import paths in moved test files"
echo "2. Run 'go mod tidy' to clean up module dependencies"
echo "3. Run 'make test' to verify all tests still work"
echo "4. Update any build scripts that reference old test locations"
