#!/bin/bash
# Unified code generation script with security improvements
# Combines functionality from generate-mocks.sh and generate-protos.sh

set -euo pipefail

# Security: Use PROJECT_ROOT variable instead of absolute paths
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Security: Validate required tools
check_tools() {
    local missing_tools=()

    if ! command -v mockgen &> /dev/null; then
        missing_tools+=("mockgen")
    fi

    if ! command -v protoc &> /dev/null; then
        missing_tools+=("protoc")
    fi

    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Install missing tools:"
        for tool in "${missing_tools[@]}"; do
            case "$tool" in
                "mockgen")
                    echo "  go install go.uber.org/mock/mockgen@latest"
                    ;;
                "protoc")
                    echo "  Visit: https://grpc.io/docs/protoc-installation/"
                    ;;
                "go")
                    echo "  Visit: https://golang.org/dl/"
                    ;;
            esac
        done
        exit 1
    fi
}

# Generate mock files
generate_mocks() {
    log_info "Generating mock files..."

    # Create mocks directory
    mkdir -p tests/mocks

    # Generate mocks for each service
    local mock_files=(
        "internal/cache/interfaces.go:tests/mocks/cache_mocks.go:HealthOperations=MockCacheHealthOperations"
        "internal/ai/llm/types.go:tests/mocks/llm_mocks.go:HealthOperations=MockLLMHealthOperations"
        "internal/ai/llm_abstractor.go:tests/mocks/ai_service_mocks.go"
        "internal/auth/types.go:tests/mocks/auth_mocks.go:UserRepository=MockAuthUserRepository"
        "internal/utilities/repository_template.go:tests/mocks/utilities_template_mocks.go"
        "internal/ai/llm/builder.go:tests/mocks/builder_mocks.go"
        "internal/config/interfaces.go:tests/mocks/config_mocks.go"
    )

    for mock_spec in "${mock_files[@]}"; do
        IFS=':' read -r source dest mock_names <<< "$mock_spec"

        if [ -f "$source" ]; then
            log_info "Generating mock for $source..."

            local cmd="mockgen -source=$source -destination=$dest -package=mocks"
            if [ -n "$mock_names" ]; then
                cmd="$cmd -mock_names=\"$mock_names\""
            fi

            if eval "$cmd"; then
                log_success "Generated $dest"
            else
                log_warning "Failed to generate $dest"
            fi
        else
            log_warning "Source file not found: $source"
        fi
    done

    # Verify generated mocks compile
    log_info "Verifying generated mocks..."
    if go build ./tests/mocks/... &> /dev/null; then
        log_success "All mocks compile successfully"
    else
        log_error "Some mocks failed to compile"
        return 1
    fi
}

# Generate protobuf files
generate_protos() {
    log_info "Generating protobuf files..."

    # Clean up existing generated files
    find . -name "*.pb.go" -delete 2>/dev/null || true
    find . -type d -name "github.com" -exec rm -rf {} + 2>/dev/null || true

    # Generate protobuf files for each service
    local proto_files=(
        "api/proto/user.proto"
        "api/proto/auth.proto"
    )

    for proto_file in "${proto_files[@]}"; do
        if [ -f "$proto_file" ]; then
            log_info "Generating protobuf for $proto_file..."

            if protoc --go_out=. --go-grpc_out=. "$proto_file"; then
                log_success "Generated protobuf for $(basename "$proto_file" .proto)"
            else
                log_error "Failed to generate protobuf for $proto_file"
                return 1
            fi
        else
            log_warning "Proto file not found: $proto_file"
        fi
    done

    # Move generated files to correct location if needed
    move_proto_files

    log_success "Protobuf generation completed"
}

# Move protobuf files to correct location
move_proto_files() {
    local services=("user" "auth")

    for service in "${services[@]}"; do
        local source_path="github.com/EliasRanz/ai-code-gen/api/proto/$service"
        local target_path="api/proto/$service"

        if [ -d "$source_path" ]; then
            log_info "Moving $service protobuf files to correct location..."

            mkdir -p "$target_path"

            if ls "$source_path"/*.pb.go &> /dev/null; then
                mv "$source_path"/*.pb.go "$target_path/"
                log_success "Moved $service protobuf files"
            else
                log_warning "No .pb.go files found in $source_path"
            fi

            # Clean up empty directory
            rmdir "$source_path" 2>/dev/null || true
        fi
    done
}

# Main execution
main() {
    local command="${1:-all}"

    check_tools

    case "$command" in
        "mocks")
            generate_mocks
            ;;
        "protos")
            generate_protos
            ;;
        "all")
            generate_mocks
            generate_protos
            ;;
        *)
            echo "Usage: $0 [mocks|protos|all]"
            echo ""
            echo "Commands:"
            echo "  mocks  - Generate mock files only"
            echo "  protos - Generate protobuf files only"
            echo "  all    - Generate both mocks and protobufs"
            exit 1
            ;;
    esac
}

main "$@"