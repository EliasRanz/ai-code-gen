#!/bin/bash
# Minimal testing script that only uses built-in Go tools
set -e

echo "🧪 Running Go Tests with Built-in Tools"
echo "======================================"

# Run tests with coverage
echo "Running tests..."
go test -v ./tests/... \
    -coverpkg=./internal/... \
    -coverprofile=coverage.out \
    -timeout=30s

# Generate basic coverage report
echo "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Show coverage summary
echo "Coverage Summary:"
go tool cover -func=coverage.out | tail -1

# Package-level analysis (using built-in tools)
echo ""
echo "Package Coverage Breakdown:"
echo "=========================="
go tool cover -func=coverage.out | grep -E '^github.com/EliasRanz/ai-code-gen/internal/' | \
    awk -F'/' '{print $NF}' | awk '{print $1 "\t" $3}' | sort -k2 -nr | \
    head -20

echo ""
echo "📊 Coverage report generated: coverage.html"
echo "Open in browser: file://$(pwd)/coverage.html"
