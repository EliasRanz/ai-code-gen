#!/bin/bash
# Enhanced test discovery and coverage script
set -e

echo "🚀 Enhanced Go Test Discovery and Coverage Report"
echo "=================================================="

# Install recommended tools (run once)
install_tools() {
    echo "📦 Installing testing tools..."
    
    # Check Go version and warn if needed
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    echo "Current Go version: $go_version"
    
    # Install core testing tools (these work with older Go versions)
    echo "Installing gotestsum..."
    if ! go install gotest.tools/gotestsum@latest; then
        echo "⚠️  gotestsum requires Go 1.23+, trying older version..."
        go install gotest.tools/gotestsum@v1.10.0 || echo "❌ Failed to install gotestsum"
    fi
    
    echo "Installing gocov..."
    go install github.com/axw/gocov/gocov@latest || echo "❌ Failed to install gocov"
    
    echo "Installing gocov-html (trying alternative)..."
    # Try different versions/approaches for gocov-html
    if ! go install github.com/matm/gocov-html@v1.4.0; then
        echo "⚠️  gocov-html not available, will use alternative approach"
    fi
    
    echo "Installing go-cover-treemap..."
    go install github.com/nikolaydubina/go-cover-treemap@latest || echo "❌ Failed to install go-cover-treemap"
    
    echo "Installing go-test-coverage..."
    go install github.com/vladopajic/go-test-coverage/v2@latest || echo "❌ Failed to install go-test-coverage"
    
    # Optional: GoConvey (only if requested)
    if [ "$INSTALL_GOCONVEY" = "true" ]; then
        echo "Installing GoConvey..."
        go install github.com/smartystreets/goconvey@latest || echo "❌ Failed to install GoConvey"
    fi
    
    echo "✅ Tool installation completed (some tools may have failed due to Go version requirements)"
}

# Run comprehensive tests with enhanced discovery
run_tests() {
    echo "🧪 Running comprehensive tests..."
    
    # Clean previous coverage files
    rm -f coverage*.out coverage*.html
    
    # Check if gotestsum is available, fallback to go test
    if command -v gotestsum >/dev/null 2>&1; then
        echo "Using gotestsum for enhanced output..."
        gotestsum --format testname --junitfile tests-report.xml -- \
            -v ./tests/... \
            -coverpkg=./internal/... \
            -coverprofile=coverage.out \
            -timeout=30s \
            -race
    else
        echo "gotestsum not available, using standard go test..."
        go test -v ./tests/... \
            -coverpkg=./internal/... \
            -coverprofile=coverage.out \
            -timeout=30s \
            -race
    fi
    
    echo "✅ Tests completed successfully!"
}

# Generate enhanced coverage reports
generate_coverage_reports() {
    echo "📊 Generating coverage reports..."
    
    if [ ! -f coverage.out ]; then
        echo "❌ No coverage.out file found. Run tests first."
        return 1
    fi
    
    # Standard Go coverage report (always available)
    echo "Generating standard HTML coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    
    # Enhanced HTML coverage with gocov (alternative approach)
    if command -v gocov >/dev/null 2>&1; then
        echo "Generating JSON coverage report with gocov..."
        gocov convert coverage.out > coverage.json
        echo "✅ JSON coverage report generated: coverage.json"
        
        # Try to generate HTML if gocov-html is available, otherwise skip
        if command -v gocov-html >/dev/null 2>&1; then
            echo "Generating enhanced HTML coverage report..."
            gocov convert coverage.out | gocov-html > coverage-enhanced.html
        else
            echo "⚠️  gocov-html not available, using alternative HTML generation..."
            # Alternative: Use go tool cover with better formatting
            go tool cover -html=coverage.out -o coverage-detailed.html
        fi
    else
        echo "⚠️  gocov not available, skipping enhanced coverage reports"
    fi
    
    # Coverage treemap visualization (if available)
    if command -v go-cover-treemap >/dev/null 2>&1; then
        echo "Generating coverage treemap..."
        go-cover-treemap -coverprofile coverage.out > coverage-treemap.svg
    else
        echo "⚠️  go-cover-treemap not available, skipping treemap generation"
    fi
    
    # Coverage summary
    go tool cover -func=coverage.out | tail -1
    
    echo "📈 Coverage reports generated:"
    echo "  - coverage.html (standard)"
    echo "  - coverage-enhanced.html (enhanced)"
    echo "  - coverage-treemap.svg (visual treemap)"
}

# Package-level coverage analysis
analyze_packages() {
    echo "🔍 Package-level coverage analysis..."
    
    # Get coverage by package
    go tool cover -func=coverage.out | grep -E '^github.com/EliasRanz/ai-code-gen/internal/' | \
    awk '{print $1 "\t" $3}' | sort -k2 -nr | \
    while IFS=$'\t' read -r package coverage; do
        printf "%-50s %s\n" "$package" "$coverage"
    done
}

# Test discovery analysis
discover_tests() {
    echo "🔍 Test Discovery Analysis..."
    echo "=========================================="
    
    echo "📁 Test directories found:"
    find tests -type d -name "*test*" | sort
    
    echo ""
    echo "🧪 Test files by package:"
    find tests -name "*_test.go" -exec dirname {} \; | sort | uniq -c | sort -nr
    
    echo ""
    echo "📊 Test function count by directory:"
    find tests -name "*_test.go" -exec grep -l "^func Test" {} \; | \
    xargs -I {} sh -c 'echo "$(dirname {}): $(grep -c "^func Test" {})"' | \
    sort
}

# Main execution
main() {
    case "${1:-all}" in
        "install")
            install_tools
            ;;
        "test")
            run_tests
            ;;
        "coverage")
            generate_coverage_reports
            ;;
        "analyze")
            analyze_packages
            ;;
        "discover")
            discover_tests
            ;;
        "all")
            run_tests
            generate_coverage_reports
            analyze_packages
            discover_tests
            ;;
        *)
            echo "Usage: $0 [install|test|coverage|analyze|discover|all]"
            echo ""
            echo "Commands:"
            echo "  install   - Install recommended testing tools"
            echo "  test      - Run tests with enhanced discovery"
            echo "  coverage  - Generate coverage reports"
            echo "  analyze   - Analyze package-level coverage"
            echo "  discover  - Analyze test discovery"
            echo "  all       - Run everything (default)"
            exit 1
            ;;
    esac
}

main "$@"
