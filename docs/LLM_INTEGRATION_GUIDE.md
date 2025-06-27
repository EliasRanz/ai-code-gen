# LLM Integration Implementation Guide

## 🎯 Quick Start: Replace Stubs with Real LLMs

### Option 1: OpenAI Integration (Recommended for Production)

**Why OpenAI First?**
- Most reliable and well-documented API
- Excellent Go SDK available
- Predictable costs and billing
- High-quality models (GPT-4, GPT-3.5)

**Implementation Steps:**

1. **Add OpenAI Dependency**
   ```bash
   go get github.com/sashabaranov/go-openai
   ```

2. **Create OpenAI Client Implementation**
   ```bash
   # Create the file
   touch internal/llm/openai_client.go
   
   # Copy the template from openai_client_todo.go
   # Implement the LLMClient interface
   ```

3. **Configuration Setup**
   ```yaml
   # Add to your config file
   llm:
     provider: "openai"
     openai:
       api_key: "${OPENAI_API_KEY}"
       organization: "${OPENAI_ORG_ID}"  # Optional
       base_url: "https://api.openai.com/v1"
       default_model: "gpt-3.5-turbo"
       max_retries: 3
       timeout: "30s"
   ```

4. **Environment Variables**
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   export OPENAI_ORG_ID="your-org-id"  # Optional
   ```

5. **Update Generation Service**
   ```go
   // In internal/generation/service.go
   // Replace NewVLLMClient with:
   if config.LLMConfig.Provider == "openai" {
       llmClient = llm.NewOpenAIClient(config.LLMConfig.OpenAI)
   } else {
       llmClient = llm.NewVLLMClient(config.LLMConfig)
   }
   ```

### Option 2: Local Development with Ollama

**Why Ollama for Development?**
- Free local inference
- No API keys required
- Good for testing and development
- Supports many open-source models

**Setup Steps:**

1. **Install Ollama**
   ```bash
   # macOS/Linux
   curl -fsSL https://ollama.ai/install.sh | sh
   
   # Start Ollama
   ollama serve
   ```

2. **Pull Models**
   ```bash
   ollama pull llama2
   ollama pull codellama
   ```

3. **Add Ollama Client**
   ```bash
   go get github.com/ollama/ollama/api
   ```

4. **Configuration**
   ```yaml
   llm:
     provider: "ollama"
     ollama:
       base_url: "http://localhost:11434"
       default_model: "llama2"
       timeout: "60s"
   ```

### Option 3: Production VLLM Setup

**Why VLLM for Production?**
- Self-hosted (cost control)
- High performance inference
- Custom model support
- No external API dependencies

**Deployment Steps:**

1. **Docker Setup**
   ```dockerfile
   # Use official VLLM image
   FROM vllm/vllm-openai:latest
   
   # Copy your models or configure model download
   ENV MODEL_NAME="microsoft/DialoGPT-medium"
   
   EXPOSE 8000
   CMD ["python", "-m", "vllm.entrypoints.openai.api_server", "--model", "$MODEL_NAME"]
   ```

2. **Kubernetes Deployment**
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: vllm-server
   spec:
     replicas: 2
     template:
       spec:
         containers:
         - name: vllm
           image: vllm/vllm-openai:latest
           ports:
           - containerPort: 8000
           env:
           - name: MODEL_NAME
             value: "microsoft/DialoGPT-medium"
   ```

3. **Configuration Update**
   ```yaml
   llm:
     provider: "vllm"
     vllm:
       base_url: "http://vllm-server:8000"
       api_key: "your-auth-token"  # If authentication enabled
       default_model: "microsoft/DialoGPT-medium"
   ```

## 🔄 Migration Strategy

### Phase 1: Single Provider (Week 1)
- Choose OpenAI or Ollama
- Implement basic client
- Update configuration
- Test with existing endpoints

### Phase 2: Multi-Provider Support (Week 2-3)
- Implement provider factory
- Add intelligent routing
- Configure failover logic
- Add cost tracking

### Phase 3: Production Optimization (Week 4+)
- Add monitoring and metrics
- Implement caching
- Add rate limiting
- Security hardening

## 🧪 Testing Your Implementation

1. **Unit Tests**
   ```bash
   go test ./internal/llm/... -v
   ```

2. **Integration Test**
   ```bash
   # Test with real API (optional)
   INTEGRATION_TEST=true go test ./internal/generation/... -v
   ```

3. **Manual Testing**
   ```bash
   curl -X POST http://localhost:8080/api/v1/generate \
     -H "Authorization: Bearer your-jwt-token" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "gpt-3.5-turbo",
       "prompt": "Write a hello world function in Go",
       "max_tokens": 150
     }'
   ```

## 📊 Monitoring Production Usage

1. **Add Metrics Collection**
   ```go
   // Track provider usage
   providerRequestsTotal.WithLabelValues(providerName, model).Inc()
   providerLatency.WithLabelValues(providerName).Observe(duration.Seconds())
   ```

2. **Cost Tracking**
   ```go
   // Calculate costs based on token usage
   cost := tokens * modelPricePerToken
   totalCost.Add(cost)
   ```

3. **Health Checks**
   ```go
   // Regular provider health checks
   if err := provider.Health(ctx); err != nil {
       providerHealthGauge.WithLabelValues(providerName).Set(0)
   }
   ```

## 🚀 Quick Commands to Get Started

```bash
# 1. Add OpenAI support
go get github.com/sashabaranov/go-openai

# 2. Set up environment
export OPENAI_API_KEY="your-key"

# 3. Update config and implement client
# See openai_client_todo.go for template

# 4. Test the integration
go test ./internal/generation/... -v

# 5. Run the service
go run cmd/server/main.go
```

This roadmap will transform your AI generation service from stubbed implementations to production-ready LLM integration! 🎉
