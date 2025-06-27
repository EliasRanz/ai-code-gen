#!/bin/bash
# scripts/refactor-large-files.sh
# Automated refactoring helper for large files

set -e

echo "🔧 Refactoring large files to comply with coding standards..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Find files that are too large (>300 lines, excluding generated code)
large_files=$(find . -name "*.go" -not -path "./api/proto/*" -not -path "./github.com/*" -not -path "./web/node_modules/*" -exec wc -l {} + | \
    awk '$1 > 300 && $2 !~ /total$/ { print $2 }')

if [ -z "$large_files" ]; then
    echo -e "${GREEN}✅ No large files found!${NC}"
    exit 0
fi

echo -e "${YELLOW}⚠️ Large files found that need refactoring:${NC}"

for file in $large_files; do
    lines=$(wc -l < "$file")
    echo -e "${RED}❌ $file ($lines lines)${NC}"
    
    # Provide refactoring suggestions based on file type
    case "$file" in
        *"/handler.go"|*"handlers.go")
            echo "  💡 Suggestion: Split into separate files by domain (user_handlers.go, project_handlers.go, admin_handlers.go)"
            ;;
        *"/service.go")
            echo "  💡 Suggestion: Split into focused service files (user_service.go, project_service.go) with single responsibility"
            ;;
        *"/repository.go")
            echo "  💡 Suggestion: Split into separate repository files by entity (user_repository.go, project_repository.go)"
            ;;
        *"/client.go")
            echo "  💡 Suggestion: Split into separate client files by provider or functionality"
            ;;
        *)
            echo "  💡 Suggestion: Identify distinct responsibilities and split into separate files"
            ;;
    esac
done

echo ""
echo -e "${YELLOW}📋 Refactoring Guidelines:${NC}"
echo "1. Each file should have <300 lines (excluding generated code)"
echo "2. Each function should have <40 lines"
echo "3. Follow Single Responsibility Principle"
echo "4. Use clear, descriptive filenames"
echo "5. Keep related functionality together in packages"

echo ""
echo -e "${YELLOW}🔧 Next Steps:${NC}"
echo "1. Review the large files listed above"
echo "2. Identify logical boundaries for splitting"
echo "3. Create new files with focused responsibilities"
echo "4. Move functions to appropriate new files"
echo "5. Update imports and tests"
echo "6. Run 'make check-standards' to verify compliance"

echo ""
echo -e "${GREEN}✅ Refactoring analysis complete.${NC}"
echo "Manual refactoring is required for the files listed above."
