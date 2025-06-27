#!/bin/bash

# Pre-commit hook to prevent committing protobuf files in wrong locations
# This helps catch the nested directory issue before it gets committed

echo "🔍 Checking for protobuf files in wrong locations..."

# Check for nested github.com directories
NESTED_DIRS=$(find . -type d -name "github.com" 2>/dev/null)
if [ -n "$NESTED_DIRS" ]; then
    echo "❌ ERROR: Found nested github.com directories:"
    echo "$NESTED_DIRS"
    echo ""
    echo "🔧 Fix this by running: make generate-protos"
    echo "📖 See docs/PROTOBUF_GENERATION.md for more info"
    exit 1
fi

# Check for .pb.go files outside of api/proto/
MISPLACED_FILES=$(find . -name "*.pb.go" -not -path "./api/proto/*" 2>/dev/null)
if [ -n "$MISPLACED_FILES" ]; then
    echo "❌ ERROR: Found .pb.go files outside of api/proto/:"
    echo "$MISPLACED_FILES"
    echo ""
    echo "🔧 Fix this by running: make generate-protos"
    echo "📖 See docs/PROTOBUF_GENERATION.md for more info"
    exit 1
fi

echo "✅ Protobuf file locations look good!"
exit 0
