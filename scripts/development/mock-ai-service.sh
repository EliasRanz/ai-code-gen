#!/bin/bash

# Mock AI Service for Development
# This provides a reliable alternative to vLLM for local development

set -e

# Configuration
PORT=${PORT:-8000}
HOST=${HOST:-0.0.0.0}

echo "🤖 Starting Mock AI Service on port $PORT"
echo "This service provides mock responses for development testing"
echo "=================================================="

# Function to generate mock AI responses
generate_mock_response() {
    local prompt="$1"

    # Simple mock responses based on prompt content
    if [[ "$prompt" =~ "button" ]]; then
        cat << 'EOF'
{
  "id": "mock-chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "mock-phi-2",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "```jsx\nfunction Button({ children, onClick, variant = 'primary' }) {\n  return (\n    <button \n      onClick={onClick}\n      className={`px-4 py-2 rounded-md font-medium transition-colors \n        ${variant === 'primary' \n          ? 'bg-blue-500 text-white hover:bg-blue-600' \n          : 'bg-gray-200 text-gray-800 hover:bg-gray-300'}`}\n    >\n      {children}\n    </button>\n  );\n}\n\nexport default Button;\n```"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 50,
    "completion_tokens": 150,
    "total_tokens": 200
  }
}
EOF
    elif [[ "$prompt" =~ "input" ]] || [[ "$prompt" =~ "form" ]]; then
        cat << 'EOF'
{
  "id": "mock-chatcmpl-124",
  "object": "chat.completion",
  "created": 1677652289,
  "model": "mock-phi-2",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "```jsx\nfunction TextInput({ label, placeholder, value, onChange, error }) {\n  return (\n    <div className=\"mb-4\">\n      {label && (\n        <label className=\"block text-sm font-medium text-gray-700 mb-1\">\n          {label}\n        </label>\n      )}\n      <input\n        type=\"text\"\n        value={value}\n        onChange={(e) => onChange(e.target.value)}\n        placeholder={placeholder}\n        className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 \n          ${error ? 'border-red-500 focus:ring-red-500' : 'border-gray-300'}`}\n      />\n      {error && <p className=\"mt-1 text-sm text-red-600\">{error}</p>}\n    </div>\n  );\n}\n\nexport default TextInput;\n```"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 45,
    "completion_tokens": 180,
    "total_tokens": 225
  }
}
EOF
    else
        cat << 'EOF'
{
  "id": "mock-chatcmpl-125",
  "object": "chat.completion",
  "created": 1677652290,
  "model": "mock-phi-2",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "```jsx\nimport React from 'react';\n\nfunction Card({ title, children, className = '' }) {\n  return (\n    <div className={`bg-white rounded-lg shadow-md p-6 ${className}`}>\n      {title && (\n        <h3 className=\"text-lg font-semibold text-gray-900 mb-4\">\n          {title}\n        </h3>\n      )}\n      {children}\n    </div>\n  );\n}\n\nexport default Card;\n```\n\nThis is a basic card component with optional title and custom styling. You can use it to wrap content in a clean, elevated container."
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 30,
    "completion_tokens": 120,
    "total_tokens": 150
  }
}
EOF
    fi
}

# Health check endpoint
handle_health() {
    cat << 'EOF'
{
  "status": "healthy",
  "model": "mock-phi-2",
  "version": "1.0.0"
}
EOF
}

# Main request handler
handle_request() {
    local method="$1"
    local path="$2"
    local body="$3"

    case "$path" in
        "/health")
            handle_health
            ;;
        "/v1/chat/completions")
            if [ "$method" = "POST" ]; then
                # Extract prompt from request body (simple parsing)
                local prompt=$(echo "$body" | grep -o '"content":"[^"]*"' | head -1 | sed 's/"content":"//' | sed 's/"$//')
                if [ -z "$prompt" ]; then
                    prompt="Create a simple UI component"
                fi
                generate_mock_response "$prompt"
            else
                echo '{"error": "Method not allowed"}'
            fi
            ;;
        "/v1/models")
            cat << 'EOF'
{
  "object": "list",
  "data": [{
    "id": "mock-phi-2",
    "object": "model",
    "created": 1677652288,
    "owned_by": "mock-provider"
  }]
}
EOF
            ;;
        *)
            echo '{"error": "Not found"}'
            ;;
    esac
}

# Start the mock server
echo "🚀 Mock AI Service is running!"
echo "📡 Available endpoints:"
echo "  GET  /health"
echo "  POST /v1/chat/completions"
echo "  GET  /v1/models"
echo ""
echo "💡 This service provides realistic mock responses for development"
echo "🔄 Switch to real vLLM for production by updating docker-compose.yml"
echo ""

# Simple HTTP server using netcat (if available) or a basic implementation
if command -v nc >/dev/null 2>&1; then
    echo "Using netcat for HTTP server..."
    # This is a very basic implementation - in production you'd use a proper HTTP server
    while true; do
        echo -e "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n$(handle_health)" | nc -l -p $PORT -q 1
    done
else
    echo "Netcat not available. Installing simple HTTP server..."
    # Fallback: use Python if available
    if command -v python3 >/dev/null 2>&1; then
        python3 -c "
import http.server
import socketserver
import json
from urllib.parse import urlparse, parse_qs

class MockAIHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            response = {'status': 'healthy', 'model': 'mock-phi-2', 'version': '1.0.0'}
            self.wfile.write(json.dumps(response).encode())
        else:
            self.send_response(404)
            self.end_headers()

    def do_POST(self):
        if self.path == '/v1/chat/completions':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            # Mock response for chat completions
            response = {
                'id': 'mock-chatcmpl-123',
                'object': 'chat.completion',
                'created': 1677652288,
                'model': 'mock-phi-2',
                'choices': [{
                    'index': 0,
                    'message': {
                        'role': 'assistant',
                        'content': '''\`\`\`jsx
function Button({ children, onClick }) {
  return (
    <button 
      onClick={onClick}
      className=\"px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600\"
    >
      {children}
    </button>
  );
}

export default Button;
\`\`\`

This creates a simple, reusable button component with hover effects.'''
                    },
                    'finish_reason': 'stop'
                }],
                'usage': {
                    'prompt_tokens': 50,
                    'completion_tokens': 150,
                    'total_tokens': 200
                }
            }
            self.wfile.write(json.dumps(response).encode())
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        pass  # Suppress default logging

with socketserver.TCPServer(('', $PORT), MockAIHandler) as httpd:
    print(f'Server running on port $PORT')
    httpd.serve_forever()
"
    else
        echo "Python3 not available. Please install Python3 or netcat for the mock server."
        exit 1
    fi
fi