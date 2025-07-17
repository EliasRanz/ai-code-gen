#!/bin/bash

# Mock generation script for interface patterns
# This script generates mocks for all interface patterns in the codebase

set -e

echo "🔧 Generating mocks for interface patterns..."

# Create mocks directory if it doesn't exist
mkdir -p tests/mocks

# Generate cache interface mocks
echo "📦 Generating cache interface mocks..."
mockgen -source=internal/cache/interfaces.go -destination=tests/mocks/cache_mocks.go -package=mocks -mock_names="HealthOperations=MockCacheHealthOperations"

# Generate LLM interface mocks
echo "🤖 Generating LLM interface mocks..."
mockgen -source=internal/ai/llm/types.go -destination=tests/mocks/llm_mocks.go -package=mocks -mock_names="HealthOperations=MockLLMHealthOperations"

# Generate AI service LLMClient interface mocks
echo "🤖 Generating AI service interface mocks..."
mockgen -source=internal/ai/llm_abstractor.go -destination=tests/mocks/ai_service_mocks.go -package=mocks

# Generate auth service interface mocks
echo "🔐 Generating auth interface mocks..."
mockgen -source=internal/auth/types.go -destination=tests/mocks/auth_mocks.go -package=mocks -mock_names="UserRepository=MockAuthUserRepository"

# Generate utilities interface mocks  
echo "🛠️ Generating utilities interface mocks..."
mockgen -source=internal/utilities/repository_template.go -destination=tests/mocks/utilities_template_mocks.go -package=mocks

# Generate builder interface mocks
echo "🏗️ Generating builder interface mocks..."
mockgen -source=internal/ai/llm/builder.go -destination=tests/mocks/builder_mocks.go -package=mocks

# Generate config interface mocks
echo "⚙️ Generating config interface mocks..."
mockgen -source=internal/config/interfaces.go -destination=tests/mocks/config_mocks.go -package=mocks

echo "✅ Mock generation completed successfully!"
echo "📁 Mocks available in: tests/mocks/"

# Verify that generated mocks compile
echo "🔍 Verifying generated mocks compile..."
if go build ./tests/mocks/... > /dev/null 2>&1; then
    echo "✅ All generated mocks compile successfully!"
else
    echo "❌ Mock compilation failed. Please check the generated code."
    exit 1
fi
