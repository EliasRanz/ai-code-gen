#!/bin/bash
# Coverage script that excludes generated files for accurate reporting

set -e

echo "Running tests with coverage (excluding generated files)..."

# Run tests with coverage, excluding generated files
gotestsum --format testdox -- -coverprofile=coverage_filtered.out -covermode=atomic $(go list ./... | grep -v -E "(proto|generated|vendor)")

echo ""
echo "=== COVERAGE SUMMARY (Excluding Generated Files) ==="
echo ""

# Show overall coverage percentage
go tool cover -func=coverage_filtered.out | tail -1

echo ""
echo "=== LOW COVERAGE FILES (< 50%) ==="
echo ""

# Show files with low coverage
go tool cover -func=coverage_filtered.out | awk '$3 != "total:" && $3 != "(statements)" && substr($3, 1, length($3)-1) < 50.0 { print $1 " - " $3 }' | head -20

echo ""
echo "=== HIGH COVERAGE FILES (>= 80%) ==="
echo ""

# Show files with high coverage  
go tool cover -func=coverage_filtered.out | awk '$3 != "total:" && $3 != "(statements)" && substr($3, 1, length($3)-1) >= 80.0 { print $1 " - " $3 }' | wc -l | xargs echo "Number of files with >=80% coverage:"

echo ""
echo "Coverage report saved to: coverage_filtered.out"
echo "To generate HTML report: go tool cover -html=coverage_filtered.out -o coverage.html"
