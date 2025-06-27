#!/bin/bash

# Script to generate protobuf files and ensure they're in the correct location
# This prevents the issue where protoc creates nested directories based on go_package paths

set -e

PROJECT_ROOT="/mnt/c/Users/josef/Documents/ai-code-gen"
PROTO_DIR="$PROJECT_ROOT/api/proto"

echo "🔧 Generating protobuf files..."

# Clean up any existing generated files
echo "🧹 Cleaning up existing generated files..."
find "$PROJECT_ROOT" -name "*.pb.go" -delete 2>/dev/null || true
find "$PROJECT_ROOT" -type d -name "github.com" -exec rm -rf {} + 2>/dev/null || true

# Generate protobuf files
echo "📦 Generating user service protobuf files..."
cd "$PROJECT_ROOT"
protoc --go_out=. --go-grpc_out=. api/proto/user.proto

echo "🔐 Generating auth service protobuf files..."
cd "$PROJECT_ROOT"
protoc --go_out=. --go-grpc_out=. api/proto/auth.proto

# Function to move generated files to correct location
move_proto_files() {
    local service_name=$1
    local source_path="$PROJECT_ROOT/github.com/EliasRanz/ai-code-gen/api/proto/$service_name"
    local target_path="$PROJECT_ROOT/api/proto/$service_name"
    
    if [ -d "$source_path" ]; then
        echo "📁 Moving $service_name protobuf files to correct location..."
        
        # Create target directory if it doesn't exist
        mkdir -p "$target_path"
        
        # Move all .pb.go files
        if ls "$source_path"/*.pb.go 1> /dev/null 2>&1; then
            mv "$source_path"/*.pb.go "$target_path/"
            echo "✅ Moved $service_name protobuf files successfully"
        else
            echo "⚠️  No .pb.go files found in $source_path"
        fi
    else
        echo "ℹ️  No nested directory found for $service_name (files may already be in correct location)"
    fi
}

# Move generated files to correct locations
move_proto_files "user"
move_proto_files "auth"

# Clean up the nested github.com directory structure
echo "🧹 Cleaning up nested directory structure..."
if [ -d "$PROJECT_ROOT/github.com" ]; then
    rm -rf "$PROJECT_ROOT/github.com"
    echo "✅ Removed nested github.com directory structure"
fi

# Verify files are in correct locations
echo "🔍 Verifying generated files..."
echo "User service files:"
ls -la "$PROJECT_ROOT/api/proto/user/" 2>/dev/null || echo "  ❌ No user protobuf files found"

echo "Auth service files:"
ls -la "$PROJECT_ROOT/api/proto/auth/" 2>/dev/null || echo "  ❌ No auth protobuf files found"

echo "🎉 Protobuf generation complete!"

# Check for any remaining nested directories
if find "$PROJECT_ROOT" -type d -name "github.com" 2>/dev/null | grep -q .; then
    echo "⚠️  Warning: Found remaining nested github.com directories:"
    find "$PROJECT_ROOT" -type d -name "github.com" 2>/dev/null
else
    echo "✅ No nested github.com directories found"
fi
