# 🚀 Stub to Production LLM Migration Guide

## 📋 **CRITICAL TODOs: Replace All Stubbed LLM Implementations**

This document provides a comprehensive roadmap for transitioning from stubbed/mock LLM implementations to production-ready, legitimate LLM provider integrations.

## 🎯 **PHASE 1: Immediate LLM Provider Integration (Priority: HIGH)**

### **Step 1.1: OpenAI Integration (RECOMMENDED FIRST - 1-2 days)**

**Current State**: Stubbed LLM calls in generation service  
**Target State**: Real OpenAI API integration with cost tracking  

**Implementation Tasks:**

1. **📦 Add OpenAI SDK Dependency**
   ```bash
   cd /mnt/c/Users/josef/Documents/ai-code-gen/ai-ui-generator
   go get github.com/sashabaranov/go-openai
   ```

2. **📝 Implement OpenAI Client** 
   - **File**: `internal/llm/openai_client.go`
   - **Template**: Use `internal/llm/openai_client_todo.go` as guide
   - **Interface**: Must implement `LLMClient` from `internal/llm/types.go`
   
   **Key Methods to Implement:**
   ```go
   func (c *OpenAIClient) Generate(ctx context.Context, req *GenerationRequest) (*GenerationResponse, error)
   func (c *OpenAIClient) GenerateStream(ctx context.Context, req *GenerationRequest) (<-chan *GenerationResponse, error)
   func (c *OpenAIClient) GetModels(ctx context.Context) ([]Model, error)
   func (c *OpenAIClient) Health(ctx context.Context) error
   ```

3. **⚙️ Update Configuration**
   - **File**: `internal/config/config.go`
   - **Add**: OpenAI-specific configuration structure
   - **Environment Variables**: 
     ```bash
     export OPENAI_API_KEY="sk-..." 
     export OPENAI_ORG_ID="org-..."  # Optional
     ```

4. **🔧 Update Generation Service**
   - **File**: `internal/generation/service.go`
   - **Replace**: VLLM-only client with multi-provider support
   - **Remove**: All stubbed LLM fallback logic

### **Step 1.2: Claude Integration (1-2 days)**

**Current State**: No Claude implementation  
**Target State**: Anthropic Claude API integration  

**Implementation Tasks:**

1. **📦 Add Anthropic SDK**
   ```bash
   go get github.com/anthropics/anthropic-sdk-go
   ```

2. **📝 Create Claude Client**
   - **File**: `internal/llm/claude_client.go`
   - **Interface**: Implement `LLMClient` interface
   - **Models**: Support Claude-3, Claude-3.5-Sonnet, etc.

3. **⚙️ Configuration**
   ```yaml
   llm:
     providers:
       claude:
         api_key: "${ANTHROPIC_API_KEY}"
         base_url: "https://api.anthropic.com"
         default_model: "claude-3-sonnet-20240229"
   ```

### **Step 1.3: Ollama Integration (Local Development - 1 day)**

**Current State**: No local LLM support  
**Target State**: Local model inference for development  

**Implementation Tasks:**

1. **📦 Add Ollama SDK**
   ```bash
   go get github.com/ollama/ollama/api
   ```

2. **📝 Create Ollama Client**
   - **File**: `internal/llm/ollama_client.go`
   - **Purpose**: Local development and testing
   - **Models**: llama2, codellama, mistral, etc.

3. **🐳 Docker Setup**
   - **File**: `docker-compose.local.yml`
   - **Include**: Ollama service for local development

---

## 🎯 **PHASE 2: Multi-Provider Architecture (Priority: HIGH)**

### **Step 2.1: Provider Factory Implementation**

**Current State**: Single VLLM client with stubs  
**Target State**: Dynamic provider selection and failover  

**Implementation Tasks:**

1. **📝 Implement Provider Factory**
   - **File**: `internal/llm/factory.go`
   - **Template**: Use `internal/llm/factory_todo.go` as guide
   
   **Key Features:**
   ```go
   type ProviderFactory interface {
       CreateClient(providerType ProviderType, config interface{}) (LLMClient, error)
       GetAvailableProviders() []ProviderType
       SelectBestProvider(requirements ProviderRequirements) (ProviderType, error)
   }
   ```

2. **📊 Smart Routing Logic**
   - Cost-based routing (cheapest for simple tasks)
   - Quality-based routing (best models for complex tasks)
   - Latency-based routing (fastest for real-time needs)
   - Failover support (automatic provider switching)

3. **⚙️ Multi-Provider Configuration**
   - **File**: `internal/config/llm_config.go`
   - **Template**: Use `internal/config/llm_config_todo.go`
   
   **Example Configuration:**
   ```yaml
   llm:
     default_provider: "openai"
     providers:
       openai:
         api_key: "${OPENAI_API_KEY}"
         models: ["gpt-4", "gpt-3.5-turbo"]
       claude:
         api_key: "${ANTHROPIC_API_KEY}"
         models: ["claude-3-sonnet-20240229"]
       vllm:
         base_url: "${VLLM_BASE_URL}"
         models: ["llama-2-7b-chat"]
     routing_rules:
       - condition: "cost_sensitive"
         provider: "vllm"
       - condition: "high_quality"
         provider: "claude"
       - condition: "default"
         provider: "openai"
   ```

### **Step 2.2: Load Balancing and Failover**

**Implementation Tasks:**

1. **🔄 Provider Health Monitoring**
   ```go
   type HealthChecker interface {
       CheckHealth(ctx context.Context, provider ProviderType) error
       GetProviderStatus(provider ProviderType) ProviderStatus
   }
   ```

2. **⚖️ Load Balancing**
   - Round-robin across healthy providers
   - Weighted routing based on capacity
   - Rate limit management per provider

3. **🛡️ Circuit Breaker Pattern**
   - Automatic provider disabling on failures
   - Gradual re-enabling after recovery
   - Fallback provider chains

---

## 🎯 **PHASE 3: Remove All Stubs (Priority: HIGH)**

### **Step 3.1: Identify and Remove Stubbed Code**

**Files with Stubs to Remove:**

1. **Generation Service Stubs**
   - **File**: `internal/generation/service.go`
   - **Lines**: ~572-640 (VLLM fallback stubs)
   - **Action**: Replace with real provider factory calls

2. **VLLM Client Stubs**
   - **File**: `internal/llm/vllm_client.go`
   - **Lines**: All stubbed fallback methods
   - **Action**: Remove stub methods, keep only real HTTP implementation

3. **Auth Middleware Stubs**
   - **File**: `internal/middleware/auth.go`
   - **Lines**: 37, 43, 64, 78 (TODO comments)
   - **Action**: Implement real JWT validation with auth service

### **Step 3.2: Replace Stubbed Auth Implementation**

**Current TODOs in Auth Middleware:**

```go
// TODO: Validate JWT token with auth service
// TODO: Set user context from validated token  
// TODO: Validate token and set user context if valid
// TODO: Check user role from context
```

**Implementation Tasks:**

1. **🔐 Real JWT Validation**
   ```go
   func validateJWTToken(token string) (*UserClaims, error) {
       // Use jwt-go library to validate token
       // Verify signature with secret key
       // Check expiration
       // Return user claims
   }
   ```

2. **👤 User Context Management**
   ```go
   func setUserContext(c *gin.Context, claims *UserClaims) {
       c.Set("user_id", claims.UserID)
       c.Set("user_email", claims.Email)
       c.Set("user_role", claims.Role)
   }
   ```

3. **🛡️ Role-Based Authorization**
   ```go
   func RequireRole(role string) gin.HandlerFunc {
       return func(c *gin.Context) {
           userRole := c.GetString("user_role")
           if userRole != role {
               c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient privileges"})
               c.Abort()
               return
           }
           c.Next()
       }
   }
   ```

---

## 🎯 **PHASE 4: Production Features (Priority: MEDIUM)**

### **Step 4.1: Cost Tracking and Analytics**

**Implementation Tasks:**

1. **💰 Token Usage Tracking**
   ```go
   type CostTracker interface {
       TrackRequest(userID string, provider ProviderType, model string, tokens int)
       GetUserCosts(userID string, period time.Duration) (float64, error)
       GetProviderCosts(provider ProviderType, period time.Duration) (float64, error)
   }
   ```

2. **📊 Usage Analytics**
   - Request volume per user/provider
   - Token consumption patterns
   - Cost optimization recommendations
   - Error rate monitoring

### **Step 4.2: Rate Limiting and Quotas**

**Implementation Tasks:**

1. **🚦 Per-User Rate Limiting**
   ```go
   type RateLimiter interface {
       CheckLimit(userID string) error
       ConsumeToken(userID string) error
       GetRemainingQuota(userID string) (int, error)
   }
   ```

2. **💎 Tier-Based Quotas**
   - Free tier: 1000 tokens/day
   - Pro tier: 100K tokens/day  
   - Enterprise: Unlimited

### **Step 4.3: Monitoring and Observability**

**Implementation Tasks:**

1. **📈 Metrics Collection**
   - Provider response times
   - Error rates per provider
   - Token consumption rates
   - User activity patterns

2. **🚨 Alerting**
   - Provider downtime alerts
   - High error rate alerts
   - Cost threshold alerts
   - Quota exhaustion alerts

---

## 🎯 **PHASE 5: Advanced Features (Priority: LOW)**

### **Step 5.1: A/B Testing Framework**

**Implementation Tasks:**

1. **🧪 Provider A/B Testing**
   ```go
   type ABTestManager interface {
       GetProviderForUser(userID string, experimentID string) ProviderType
       RecordResult(userID string, experimentID string, quality float64)
       GetExperimentResults(experimentID string) ExperimentResults
   }
   ```

### **Step 5.2: Model Fine-tuning Support**

**Implementation Tasks:**

1. **🎯 Custom Model Management**
   - Fine-tuned model deployment
   - Model versioning
   - Performance comparison

---

## 📋 **MIGRATION CHECKLIST**

### **Pre-Migration Validation**
- [ ] All tests pass with current stub implementation
- [ ] Database migrations are applied
- [ ] Redis connection is working
- [ ] Auth service is functional

### **Phase 1: Basic Provider Integration**
- [ ] OpenAI client implemented and tested
- [ ] Claude client implemented and tested  
- [ ] Ollama client for local development
- [ ] Environment variables configured
- [ ] Basic provider selection working

### **Phase 2: Multi-Provider Architecture**
- [ ] Provider factory implemented
- [ ] Smart routing logic working
- [ ] Failover mechanisms tested
- [ ] Configuration validation added
- [ ] Health monitoring active

### **Phase 3: Stub Removal**
- [ ] All stubbed LLM calls removed
- [ ] Auth middleware fully implemented
- [ ] JWT validation working
- [ ] User context properly set
- [ ] Role-based authorization working

### **Phase 4: Production Features**
- [ ] Cost tracking implemented
- [ ] Rate limiting active
- [ ] Usage analytics working
- [ ] Monitoring and alerting set up

### **Post-Migration Validation**
- [ ] All tests pass with real providers
- [ ] Load testing completed
- [ ] Error handling verified
- [ ] Performance benchmarks met
- [ ] Security audit passed

---

## 🚨 **CRITICAL SUCCESS METRICS**

### **Functionality**
- ✅ Real LLM responses (no stubs)
- ✅ Multi-provider support working
- ✅ Failover mechanisms active
- ✅ Authentication fully implemented

### **Performance**
- ✅ Response time < 2 seconds (95th percentile)
- ✅ Error rate < 1%
- ✅ Provider uptime > 99.9%

### **Cost Management**
- ✅ Cost tracking accuracy > 99%
- ✅ Rate limiting working
- ✅ Quota enforcement active

### **Security**
- ✅ JWT validation implemented
- ✅ User context properly managed
- ✅ Role-based authorization working
- ✅ API keys securely managed

---

## 📞 **NEXT STEPS**

1. **Immediate**: Start with OpenAI integration (Step 1.1)
2. **Short-term**: Implement multi-provider factory (Phase 2)
3. **Medium-term**: Remove all stubs (Phase 3)
4. **Long-term**: Add production features (Phase 4-5)

**Estimated Timeline**: 1-2 weeks for complete migration from stubs to production-ready LLM integration.
