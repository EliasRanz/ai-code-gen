# Protobuf Generation Guidelines

## Problem
When using `protoc` to generate Go code from `.proto` files, the generated files are placed in directories based on the `go_package` option in the proto file. This creates nested directory structures like:
```
github.com/EliasRanz/ai-code-gen/api/proto/user/
```

## Solution
Always use the provided script and Makefile target to generate protobuf files.

## Usage

### Method 1: Using Makefile (Recommended)
```bash
make generate-protos
```

### Method 2: Using the script directly
```bash
./scripts/generate-protos.sh
```

### Method 3: Manual generation (NOT RECOMMENDED)
If you must generate protobuf files manually, always run the cleanup afterward:
```bash
# Generate files
protoc --go_out=. --go-grpc_out=. api/proto/user.proto
protoc --go_out=. --go-grpc_out=. api/proto/auth.proto

# Move files to correct location and cleanup
./scripts/generate-protos.sh
```

## What the script does:
1. Cleans up any existing generated files
2. Generates protobuf files using `protoc`
3. Automatically moves generated files from nested directories to correct locations:
   - From: `github.com/EliasRanz/ai-code-gen/api/proto/user/` 
   - To: `api/proto/user/`
4. Removes the nested `github.com` directory structure
5. Verifies all files are in correct locations

## File Locations
- **Proto definitions**: `api/proto/*.proto`
- **Generated Go files**: `api/proto/{service}/*.pb.go`

## Prevention
- Always use `make generate-protos` instead of running `protoc` directly
- The script includes verification steps to catch issues early
- Generated files are automatically moved to correct locations

## Troubleshooting
If you find files in the wrong location:
```bash
# Run the script to fix the layout
./scripts/generate-protos.sh

# Or check for nested directories
find . -type d -name "github.com"
```
