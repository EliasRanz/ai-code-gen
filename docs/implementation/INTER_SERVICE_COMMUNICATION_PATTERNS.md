# Inter-Service Communication Patterns

*Comprehensive guidance for guaranteed request/response, high availability, high performance, and replayability with transaction monitoring*

## **Overview**

This document provides clear guidance for implementing robust inter-service communication patterns that ensure **guaranteed delivery**, **high availability**, **high performance**, and **replayability** of failed requests with comprehensive transaction monitoring.

## **Current System Assessment**

### **Existing Patterns**
- ✅ **gRPC**: Synchronous request-response between services
- ✅ **HTTP/REST**: API Gateway external communication
- ✅ **Redis Pub/Sub**: Real-time messaging for SSE scaling
- ✅ **Server-Sent Events (SSE)**: Real-time client updates
- ⚠️ **Limited**: Basic health checks and circuit breaker concepts
- ❌ **Missing**: Guaranteed delivery, replay mechanisms, comprehensive monitoring

### **Current Gaps**
1. **No guaranteed delivery** - Messages can be lost during failures
2. **No replay mechanisms** - Failed requests cannot be retried systematically
3. **Limited transaction monitoring** - Basic health checks only
4. **No message persistence** - In-memory channels and Redis pub/sub are volatile
5. **No dead letter queues** - Failed messages are lost
6. **No circuit breaker implementation** - Only conceptual in plans

---

## **Recommended Communication Patterns**

### **1. Synchronous Communication (Request-Response)**

#### **Pattern: Resilient gRPC with Circuit Breaker**

**Use Cases:**
- User service operations (CRUD)
- Auth validation requests  
- AI generation requests
- Real-time queries requiring immediate response

**Technologies:**
- **gRPC** with **grpc-go-middleware**
- **Circuit Breaker**: `sony/gobreaker` or `hystrix-go`
- **Retry Logic**: `avast/retry-go`
- **Load Balancing**: gRPC built-in or `grpc-proxy`

**Implementation with Interface Pattern:**

```go
// 🏗️ SERVICE CLIENT INTERFACE PATTERN - Technology agnostic service communication
type ServiceClient interface {
    Call(ctx context.Context, operation string, request interface{}) (interface{}, error)
    CallWithRetry(ctx context.Context, operation string, request interface{}, options RetryOptions) (interface{}, error)
    Health(ctx context.Context) error
    Metrics() ClientMetrics
    Close() error
}

// Circuit Breaker Interface - Swappable resilience implementations
type CircuitBreaker interface {
    Execute(fn func() (interface{}, error)) (interface{}, error)
    State() CircuitBreakerState
    Metrics() CircuitBreakerMetrics
}

// Retry Strategy Interface - Configurable retry behaviors
type RetryStrategy interface {
    ShouldRetry(attempt int, err error) bool
    NextDelay(attempt int) time.Duration
    MaxAttempts() int
}

// User Service Domain Interface - Business operations abstracted from transport
type UserServiceInterface interface {
    CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error)
    GetUser(ctx context.Context, userID string) (*GetUserResponse, error)
    UpdateUser(ctx context.Context, req UpdateUserRequest) (*UpdateUserResponse, error)
    DeleteUser(ctx context.Context, userID string) error
    ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error)
}

// 🏗️ FACTORY PATTERN - Dynamic client creation based on configuration
type ServiceClientFactory interface {
    CreateUserServiceClient(config ClientConfig) (UserServiceInterface, error)
    CreateAIServiceClient(config ClientConfig) (AIServiceInterface, error)
    CreateAuthServiceClient(config ClientConfig) (AuthServiceInterface, error)
}

// Resilient Service Client - Technology agnostic implementation
type ResilientServiceClient struct {
    transport     TransportClient  // gRPC, HTTP, or other
    breaker       CircuitBreaker   // Sony, Hystrix, or custom
    retryStrategy RetryStrategy    // Exponential, linear, or custom
    metrics       MetricsCollector // Prometheus, DataDog, or custom
    config        ClientConfig
}

func (c *ResilientServiceClient) Call(ctx context.Context, operation string, request interface{}) (interface{}, error) {
    // Circuit breaker wrapper - implementation agnostic
    return c.breaker.Execute(func() (interface{}, error) {
        // Add timeout from configuration
        ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
        defer cancel()
        
        // Execute with retry strategy
        return c.executeWithRetry(ctx, operation, request)
    })
}

func (c *ResilientServiceClient) executeWithRetry(ctx context.Context, operation string, request interface{}) (interface{}, error) {
    var lastErr error
    
    for attempt := 1; attempt <= c.retryStrategy.MaxAttempts(); attempt++ {
        response, err := c.transport.Send(ctx, operation, request)
        if err == nil {
            c.metrics.RecordSuccess(operation, attempt)
            return response, nil
        }
        
        lastErr = err
        c.metrics.RecordError(operation, attempt, err)
        
        // Check if we should retry
        if !c.retryStrategy.ShouldRetry(attempt, err) {
            break
        }
        
        // Wait before retry
        if attempt < c.retryStrategy.MaxAttempts() {
            delay := c.retryStrategy.NextDelay(attempt)
            select {
            case <-time.After(delay):
                continue
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }
    }
    
    return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// 🏗️ TRANSPORT INTERFACE PATTERN - Swappable communication protocols
type TransportClient interface {
    Send(ctx context.Context, operation string, request interface{}) (interface{}, error)
    Health(ctx context.Context) error
    Close() error
}

// gRPC Transport Implementation
type GRPCTransportClient struct {
    conn   *grpc.ClientConn
    client interface{} // Specific gRPC client (pb.UserServiceClient, etc.)
}

func (g *GRPCTransportClient) Send(ctx context.Context, operation string, request interface{}) (interface{}, error) {
    // Use reflection or type switches to call appropriate gRPC method
    switch operation {
    case "CreateUser":
        req := request.(*pb.CreateUserRequest)
        return g.client.(pb.UserServiceClient).CreateUser(ctx, req)
    case "GetUser":
        req := request.(*pb.GetUserRequest)
        return g.client.(pb.UserServiceClient).GetUser(ctx, req)
    // ... other operations
    default:
        return nil, fmt.Errorf("unknown operation: %s", operation)
    }
}

// HTTP Transport Implementation (alternative to gRPC)
type HTTPTransportClient struct {
    baseURL    string
    httpClient *http.Client
    serializer Serializer // JSON, XML, etc.
}

func (h *HTTPTransportClient) Send(ctx context.Context, operation string, request interface{}) (interface{}, error) {
    endpoint := h.getEndpointForOperation(operation)
    
    body, err := h.serializer.Serialize(request)
    if err != nil {
        return nil, fmt.Errorf("serialization failed: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+endpoint, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("request creation failed: %w", err)
    }
    
    resp, err := h.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
    }
    
    return h.serializer.Deserialize(resp.Body, h.getResponseTypeForOperation(operation))
}
```

**Configuration-Driven Factory Pattern:**

```go
// 🏗️ CONFIGURATION INTERFACE PATTERN - Environment-agnostic configuration
type ClientConfig struct {
    Transport    TransportType     `yaml:"transport"`     // "grpc", "http", "websocket"
    Address      string           `yaml:"address"`
    Timeout      time.Duration    `yaml:"timeout"`
    CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
    Retry        RetryConfig      `yaml:"retry"`
    Metrics      MetricsConfig    `yaml:"metrics"`
}

type TransportType string

const (
    TransportGRPC      TransportType = "grpc"
    TransportHTTP      TransportType = "http"
    TransportWebSocket TransportType = "websocket"
)

// Service Client Factory Implementation
type DefaultServiceClientFactory struct {
    transportFactory TransportFactory
    breakerFactory   CircuitBreakerFactory
    metricsFactory   MetricsFactory
}

func (f *DefaultServiceClientFactory) CreateUserServiceClient(config ClientConfig) (UserServiceInterface, error) {
    // Create transport based on configuration
    transport, err := f.transportFactory.CreateTransport(config.Transport, config.Address)
    if err != nil {
        return nil, fmt.Errorf("failed to create transport: %w", err)
    }
    
    // Create circuit breaker based on configuration
    breaker, err := f.breakerFactory.CreateCircuitBreaker(config.CircuitBreaker)
    if err != nil {
        return nil, fmt.Errorf("failed to create circuit breaker: %w", err)
    }
    
    // Create metrics collector based on configuration
    metrics, err := f.metricsFactory.CreateMetricsCollector(config.Metrics)
    if err != nil {
        return nil, fmt.Errorf("failed to create metrics collector: %w", err)
    }
    
    // Create retry strategy based on configuration
    retryStrategy := CreateRetryStrategy(config.Retry)
    
    // Assemble the resilient client
    client := &ResilientServiceClient{
        transport:     transport,
        breaker:       breaker,
        retryStrategy: retryStrategy,
        metrics:       metrics,
        config:        config,
    }
    
    // Wrap with user service specific interface
    return &UserServiceClientAdapter{client: client}, nil
}

// 🎯 ADAPTER PATTERN - Domain-specific interface implementation
type UserServiceClientAdapter struct {
    client *ResilientServiceClient
}

func (u *UserServiceClientAdapter) CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
    // Convert domain request to transport request
    transportReq := u.convertToTransportRequest("CreateUser", req)
    
    response, err := u.client.Call(ctx, "CreateUser", transportReq)
    if err != nil {
        return nil, err
    }
    
    // Convert transport response to domain response
    return u.convertFromTransportResponse("CreateUser", response)
}

func (u *UserServiceClientAdapter) GetUser(ctx context.Context, userID string) (*GetUserResponse, error) {
    transportReq := map[string]interface{}{"user_id": userID}
    
    response, err := u.client.Call(ctx, "GetUser", transportReq)
    if err != nil {
        return nil, err
    }
    
    return u.convertFromTransportResponse("GetUser", response)
}

// Technology switching example configuration
func ExampleConfigurationBasedUsage() {
    // Development: Use HTTP for easier debugging
    devConfig := ClientConfig{
        Transport: TransportHTTP,
        Address:   "http://localhost:8080",
        Timeout:   5 * time.Second,
        CircuitBreaker: CircuitBreakerConfig{
            FailureThreshold: 3,
            RecoveryTimeout:  30 * time.Second,
        },
        Retry: RetryConfig{
            MaxAttempts: 3,
            Strategy:    "exponential",
            BaseDelay:   100 * time.Millisecond,
        },
    }
    
    // Production: Use gRPC for performance
    prodConfig := ClientConfig{
        Transport: TransportGRPC,
        Address:   "user-service.production.svc.cluster.local:9090",
        Timeout:   2 * time.Second,
        CircuitBreaker: CircuitBreakerConfig{
            FailureThreshold: 5,
            RecoveryTimeout:  60 * time.Second,
        },
        Retry: RetryConfig{
            MaxAttempts: 5,
            Strategy:    "exponential_jitter",
            BaseDelay:   50 * time.Millisecond,
        },
    }
    
    // Same interface, different implementations - zero code changes!
    factory := NewDefaultServiceClientFactory()
    
    var userService UserServiceInterface
    var err error
    
    if os.Getenv("ENV") == "production" {
        userService, err = factory.CreateUserServiceClient(prodConfig)
    } else {
        userService, err = factory.CreateUserServiceClient(devConfig)
    }
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Business logic remains unchanged regardless of transport
    user, err := userService.CreateUser(context.Background(), CreateUserRequest{
        Name:  "John Doe",
        Email: "john@example.com",
    })
}
```

**Benefits:**
- **Guaranteed Response**: Circuit breaker prevents cascade failures
- **High Performance**: Connection pooling and keep-alive
- **Automatic Recovery**: Circuit breaker auto-reopens when service recovers
- **Comprehensive Metrics**: Request latency, success/failure rates, circuit state

---

### **2. Asynchronous Communication (Fire-and-Forget)**

#### **Pattern: Message Queue with Guaranteed Delivery**

**Use Cases:**
- Background AI processing
- Notification dispatch
- Audit logging
- Analytics events
- Non-critical updates

**Technologies:**
- **Primary**: **Redis Streams** (persistent, ordered, replayable)
- **Alternative**: **Apache Kafka** (high throughput, distributed)
- **Fallback**: **RabbitMQ** (reliable, feature-rich)

**Implementation with Interface Pattern:**

```go
// 🏗️ MESSAGE QUEUE INTERFACE PATTERN - Technology agnostic messaging
type MessageQueue interface {
    Publish(ctx context.Context, topic string, message Message) error
    Subscribe(ctx context.Context, topic string, consumer string, handler MessageHandler) error
    CreateConsumerGroup(ctx context.Context, topic, group string) error
    AckMessage(ctx context.Context, topic, group, messageID string) error
    GetPendingMessages(ctx context.Context, topic, group string) ([]PendingMessage, error)
    Health(ctx context.Context) error
    Close() error
}

// Message Interface - Standardized message structure
type Message interface {
    GetID() string
    GetType() string
    GetPayload() interface{}
    GetMetadata() MessageMetadata
    GetTimestamp() time.Time
    GetRetryCount() int
    SetRetryCount(count int)
}

// Message Handler Interface - Pluggable message processing
type MessageHandler interface {
    Handle(ctx context.Context, message Message) error
    ShouldRetry(err error) bool
    GetMaxRetries() int
}

// Dead Letter Queue Interface - Failed message handling
type DeadLetterQueue interface {
    Send(ctx context.Context, message Message, reason string) error
    List(ctx context.Context, filter DLQFilter) ([]DeadLetterMessage, error)
    Replay(ctx context.Context, messageID string) error
    BatchReplay(ctx context.Context, filter DLQFilter) error
}

// 🏗️ FACTORY PATTERN - Dynamic queue creation based on configuration
type MessageQueueFactory interface {
    CreateMessageQueue(config QueueConfig) (MessageQueue, error)
    CreateDeadLetterQueue(config DLQConfig) (DeadLetterQueue, error)
}

// Universal Message Implementation
type StandardMessage struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    Payload     interface{}           `json:"payload"`
    Metadata    MessageMetadata       `json:"metadata"`
    Timestamp   time.Time             `json:"timestamp"`
    RetryCount  int                   `json:"retry_count"`
    MaxRetries  int                   `json:"max_retries"`
}

type MessageMetadata struct {
    UserID        string            `json:"user_id"`
    CorrelationID string            `json:"correlation_id"`
    Source        string            `json:"source"`
    Headers       map[string]string `json:"headers"`
}

// Guaranteed Delivery Publisher - Technology agnostic
type GuaranteedPublisher struct {
    queue         MessageQueue
    dlq           DeadLetterQueue
    retryStrategy RetryStrategy
    metrics       MessageMetrics
}

func (p *GuaranteedPublisher) PublishUserEvent(ctx context.Context, event UserEvent) error {
    message := &StandardMessage{
        ID:        uuid.New().String(),
        Type:      "user.event." + event.EventType,
        Payload:   event,
        Timestamp: time.Now(),
        RetryCount: 0,
        MaxRetries: p.retryStrategy.MaxAttempts(),
        Metadata: MessageMetadata{
            UserID:        event.UserID,
            CorrelationID: uuid.New().String(),
            Source:        "user-service",
            Headers:       event.Headers,
        },
    }
    
    return p.publishWithRetry(ctx, "users:events", message)
}

func (p *GuaranteedPublisher) publishWithRetry(ctx context.Context, topic string, message Message) error {
    var lastErr error
    
    for attempt := 1; attempt <= message.GetRetryCount()+1; attempt++ {
        err := p.queue.Publish(ctx, topic, message)
        if err == nil {
            p.metrics.RecordPublishSuccess(topic)
            return nil
        }
        
        lastErr = err
        p.metrics.RecordPublishError(topic, err)
        
        // Check if we should retry
        if !p.retryStrategy.ShouldRetry(attempt, err) {
            break
        }
        
        // Wait before retry
        if attempt <= p.retryStrategy.MaxAttempts() {
            delay := p.retryStrategy.NextDelay(attempt)
            select {
            case <-time.After(delay):
                continue
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }
    
    // Send to dead letter queue if all retries failed
    dlqErr := p.dlq.Send(ctx, message, fmt.Sprintf("publish failed after %d attempts: %v", p.retryStrategy.MaxAttempts(), lastErr))
    if dlqErr != nil {
        p.metrics.RecordDLQError(topic, dlqErr)
        return fmt.Errorf("publish failed and DLQ send failed: publish_error=%w, dlq_error=%v", lastErr, dlqErr)
    }
    
    return fmt.Errorf("publish failed after %d attempts, sent to DLQ: %w", p.retryStrategy.MaxAttempts(), lastErr)
}

// Technology-specific implementations
// Redis Streams Implementation
type RedisStreamQueue struct {
    client   RedisClient
    config   RedisConfig
    serializer MessageSerializer
}

func (r *RedisStreamQueue) Publish(ctx context.Context, topic string, message Message) error {
    data, err := r.serializer.Serialize(message)
    if err != nil {
        return fmt.Errorf("serialization failed: %w", err)
    }
    
    args := &redis.XAddArgs{
        Stream: topic,
        ID:     "*",
        Values: map[string]interface{}{
            "data": data,
            "type": message.GetType(),
            "id":   message.GetID(),
        },
    }
    
    _, err = r.client.XAdd(ctx, args).Result()
    return err
}

// Apache Kafka Implementation (alternative)
type KafkaQueue struct {
    producer   kafka.Producer
    consumer   kafka.Consumer
    config     KafkaConfig
    serializer MessageSerializer
}

func (k *KafkaQueue) Publish(ctx context.Context, topic string, message Message) error {
    data, err := k.serializer.Serialize(message)
    if err != nil {
        return fmt.Errorf("serialization failed: %w", err)
    }
    
    msg := &kafka.Message{
        Topic: topic,
        Key:   []byte(message.GetID()),
        Value: data,
        Headers: map[string][]byte{
            "message-type": []byte(message.GetType()),
            "correlation-id": []byte(message.GetMetadata().CorrelationID),
        },
    }
    
    return k.producer.Send(ctx, msg)
}

// RabbitMQ Implementation (alternative)
type RabbitMQQueue struct {
    connection *amqp.Connection
    channel    *amqp.Channel
    config     RabbitMQConfig
    serializer MessageSerializer
}

func (r *RabbitMQQueue) Publish(ctx context.Context, topic string, message Message) error {
    data, err := r.serializer.Serialize(message)
    if err != nil {
        return fmt.Errorf("serialization failed: %w", err)
    }
    
    return r.channel.PublishWithContext(
        ctx,
        "",    // exchange
        topic, // routing key
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType:   "application/json",
            Body:          data,
            MessageId:     message.GetID(),
            Type:          message.GetType(),
            Timestamp:     message.GetTimestamp(),
            CorrelationId: message.GetMetadata().CorrelationID,
        },
    )
}
```

**Configuration-Driven Queue Selection:**

```go
// 🏗️ QUEUE CONFIGURATION PATTERN - Technology selection via configuration
type QueueConfig struct {
    Provider     QueueProvider    `yaml:"provider"`      // "redis", "kafka", "rabbitmq"
    Address      string          `yaml:"address"`
    MaxRetries   int             `yaml:"max_retries"`
    BatchSize    int             `yaml:"batch_size"`
    Serializer   SerializerType  `yaml:"serializer"`    // "json", "protobuf", "avro"
    DLQConfig    DLQConfig       `yaml:"dlq"`
    Monitoring   MonitoringConfig `yaml:"monitoring"`
}

type QueueProvider string

const (
    QueueProviderRedis    QueueProvider = "redis"
    QueueProviderKafka    QueueProvider = "kafka"
    QueueProviderRabbitMQ QueueProvider = "rabbitmq"
    QueueProviderInMemory QueueProvider = "inmemory" // For testing
)

// Queue Factory Implementation
type DefaultMessageQueueFactory struct {
    serializerFactory SerializerFactory
    metricsFactory    MetricsFactory
}

func (f *DefaultMessageQueueFactory) CreateMessageQueue(config QueueConfig) (MessageQueue, error) {
    serializer, err := f.serializerFactory.CreateSerializer(config.Serializer)
    if err != nil {
        return nil, fmt.Errorf("failed to create serializer: %w", err)
    }
    
    metrics, err := f.metricsFactory.CreateMessageMetrics(config.Monitoring)
    if err != nil {
        return nil, fmt.Errorf("failed to create metrics: %w", err)
    }
    
    switch config.Provider {
    case QueueProviderRedis:
        return NewRedisStreamQueue(config, serializer, metrics)
    case QueueProviderKafka:
        return NewKafkaQueue(config, serializer, metrics)
    case QueueProviderRabbitMQ:
        return NewRabbitMQQueue(config, serializer, metrics)
    case QueueProviderInMemory:
        return NewInMemoryQueue(config, serializer, metrics)
    default:
        return nil, fmt.Errorf("unsupported queue provider: %s", config.Provider)
    }
}

// Environment-specific configurations
func ExampleQueueConfiguration() {
    // Development: Use in-memory for fast tests
    devConfig := QueueConfig{
        Provider:   QueueProviderInMemory,
        MaxRetries: 3,
        BatchSize:  10,
        Serializer: SerializerJSON,
    }
    
    // Staging: Use Redis for similarity to production
    stagingConfig := QueueConfig{
        Provider:   QueueProviderRedis,
        Address:    "redis.staging.internal:6379",
        MaxRetries: 5,
        BatchSize:  50,
        Serializer: SerializerJSON,
        DLQConfig: DLQConfig{
            Enabled:    true,
            TTL:        24 * time.Hour,
            MaxRetries: 3,
        },
    }
    
    // Production: Use Kafka for high throughput
    prodConfig := QueueConfig{
        Provider:   QueueProviderKafka,
        Address:    "kafka.production.internal:9092",
        MaxRetries: 10,
        BatchSize:  1000,
        Serializer: SerializerProtobuf, // More efficient
        DLQConfig: DLQConfig{
            Enabled:    true,
            TTL:        7 * 24 * time.Hour,
            MaxRetries: 5,
        },
        Monitoring: MonitoringConfig{
            MetricsEnabled: true,
            TracingEnabled: true,
            AlertingEnabled: true,
        },
    }
    
    // Same code, different technology - zero business logic changes!
    factory := NewDefaultMessageQueueFactory()
    
    var config QueueConfig
    switch os.Getenv("ENV") {
    case "production":
        config = prodConfig
    case "staging":
        config = stagingConfig
    default:
        config = devConfig
    }
    
    queue, err := factory.CreateMessageQueue(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Business logic remains identical across all environments
    publisher := &GuaranteedPublisher{
        queue:         queue,
        retryStrategy: CreateRetryStrategy(config.MaxRetries),
    }
    
    // Publish user events regardless of underlying technology
    err = publisher.PublishUserEvent(context.Background(), UserEvent{
        EventType: "user_created",
        UserID:    "user-123",
        Data:      map[string]interface{}{"name": "John Doe"},
    })
}
```

**Benefits with Interface Pattern:**
- ✅ **Technology Independence**: Switch between Redis Streams, Kafka, RabbitMQ with configuration only
- ✅ **Zero Business Logic Changes**: Domain code unaffected by infrastructure changes
- ✅ **Environment Flexibility**: Different technologies per environment (dev/staging/prod)
- ✅ **Easy Testing**: In-memory implementations for fast unit tests
- ✅ **Vendor Lock-in Prevention**: No dependency on specific message queue features
- ✅ **Gradual Migration**: Migrate one service at a time without system-wide changes

**Technology Migration Example:**

```go
// Before: Tightly coupled to Redis
type TightlyCoupledService struct {
    redisClient *redis.Client
}

func (s *TightlyCoupledService) ProcessUserUpdate(userID string) error {
    // Hardcoded Redis commands - difficult to change
    return s.redisClient.Set("user:"+userID, userData, time.Hour).Err()
}

// After: Interface-based, technology agnostic
type FlexibleService struct {
    messageQueue MessageQueue
    cache        CacheInterface
    userService  UserServiceInterface
}

func (s *FlexibleService) ProcessUserUpdate(ctx context.Context, userID string) error {
    // Technology-agnostic operations
    event := UserEvent{EventType: "user_updated", UserID: userID}
    return s.messageQueue.Publish(ctx, "user:events", event)
}

// Migration path: Redis → Kafka (zero business logic changes)
// Step 1: Change configuration only
oldConfig := QueueConfig{Provider: QueueProviderRedis}
newConfig := QueueConfig{Provider: QueueProviderKafka}

// Step 2: Factory creates new implementation
queue, _ := factory.CreateMessageQueue(newConfig)

// Step 3: Same service code, different technology
service := &FlexibleService{messageQueue: queue}
```

---

### **3. Event Streaming (Real-Time)**

#### **Pattern: Event Sourcing with Redis Streams**

**Use Cases:**
- Real-time AI generation updates
- Live collaboration features
- System state synchronization
- Real-time dashboards

**Implementation:**

```go
// Event Streaming Interface
type EventStreamer interface {
    PublishEvent(ctx context.Context, event DomainEvent) error
    SubscribeToEvents(ctx context.Context, filter EventFilter, handler EventHandler) error
    ReplayEvents(ctx context.Context, from time.Time, filter EventFilter) ([]DomainEvent, error)
}

// Domain Event Structure
type DomainEvent struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    AggregateID string                 `json:"aggregate_id"`
    Version     int64                  `json:"version"`
    Data        map[string]interface{} `json:"data"`
    Metadata    EventMetadata          `json:"metadata"`
    Timestamp   time.Time              `json:"timestamp"`
}

type EventMetadata struct {
    UserID      string `json:"user_id"`
    CorrelationID string `json:"correlation_id"`
    CausationID   string `json:"causation_id"`
    Source        string `json:"source"`
}

// Event Sourcing Implementation
type RedisEventStreamer struct {
    client      *redis.Client
    publisher   MessageQueue
    subscribers map[string][]EventHandler
    metrics     *EventMetrics
}

func (e *RedisEventStreamer) PublishGenerationEvent(ctx context.Context, generationID string, event GenerationEvent) error {
    domainEvent := DomainEvent{
        ID:          uuid.New().String(),
        Type:        "generation." + event.Type,
        AggregateID: generationID,
        Version:     event.Version,
        Data: map[string]interface{}{
            "generation_id": generationID,
            "status":       event.Status,
            "progress":     event.Progress,
            "content":      event.Content,
        },
        Metadata: EventMetadata{
            UserID:        event.UserID,
            CorrelationID: event.CorrelationID,
            Source:        "ai-service",
        },
        Timestamp: time.Now(),
    }
    
    // Publish to multiple streams for different consumers
    streams := []string{
        "generations:all",
        fmt.Sprintf("generations:user:%s", event.UserID),
        fmt.Sprintf("generations:project:%s", event.ProjectID),
    }
    
    for _, stream := range streams {
        if err := e.publisher.Publish(ctx, stream, domainEvent); err != nil {
            e.metrics.IncrementPublishErrors(stream)
            return fmt.Errorf("failed to publish to stream %s: %w", stream, err)
        }
    }
    
    e.metrics.IncrementEventPublished(domainEvent.Type)
    return nil
}
```

---

### **4. Request-Response with Persistence**

#### **Pattern: Request-Reply with Message Queues**

**Use Cases:**
- Long-running AI operations
- Batch processing requests
- Operations requiring audit trails
- Cross-service transactions

**Implementation:**

```go
// Request-Reply Pattern Interface
type RequestReplyManager interface {
    SendRequest(ctx context.Context, request Request) (<-chan Response, error)
    ProcessRequests(ctx context.Context, handler RequestHandler) error
    ReplayFailedRequests(ctx context.Context, since time.Time) error
}

// Persistent Request-Reply Implementation
type RedisRequestReply struct {
    client       *redis.Client
    requestQueue string
    replyQueue   string
    timeout      time.Duration
    dlqQueue     string
}

func (r *RedisRequestReply) SendAIGenerationRequest(ctx context.Context, req AIGenerationRequest) (*AIGenerationResponse, error) {
    // Create request with correlation ID
    request := Request{
        ID:            uuid.New().String(),
        CorrelationID: uuid.New().String(),
        Type:          "ai.generation",
        Payload:       req,
        ReplyTo:       fmt.Sprintf("ai:replies:%s", req.UserID),
        Timestamp:     time.Now(),
        TTL:           time.Hour, // Request expires in 1 hour
    }
    
    // Publish request
    if err := r.publishRequest(ctx, request); err != nil {
        return nil, fmt.Errorf("failed to publish request: %w", err)
    }
    
    // Wait for reply with timeout
    response, err := r.waitForReply(ctx, request.CorrelationID, 30*time.Second)
    if err != nil {
        // Move to failed requests for replay
        r.moveToFailedQueue(ctx, request)
        return nil, fmt.Errorf("request failed: %w", err)
    }
    
    return response.(*AIGenerationResponse), nil
}

func (r *RedisRequestReply) ProcessAIRequests(ctx context.Context) error {
    return r.consumeRequests(ctx, "ai:requests", func(req Request) (*Response, error) {
        // Process AI generation
        result, err := r.aiService.Generate(ctx, req.Payload.(AIGenerationRequest))
        if err != nil {
            return &Response{
                ID:            uuid.New().String(),
                CorrelationID: req.CorrelationID,
                Type:          "ai.generation.error",
                Error:         err.Error(),
                Timestamp:     time.Now(),
            }, nil
        }
        
        return &Response{
            ID:            uuid.New().String(),
            CorrelationID: req.CorrelationID,
            Type:          "ai.generation.success",
            Payload:       result,
            Timestamp:     time.Now(),
        }, nil
    })
}
```

---

## **Transaction Monitoring & Observability**

### **Comprehensive Monitoring Stack**

**Technologies:**
- **Metrics**: Prometheus + Grafana
- **Tracing**: Jaeger or Zipkin
- **Logging**: ELK Stack (Elasticsearch, Logstash, Kibana)
- **APM**: New Relic or DataDog (optional)

**Implementation:**

```go
// Transaction Monitoring Interface
type TransactionMonitor interface {
    StartTransaction(ctx context.Context, name string) Transaction
    RecordEvent(ctx context.Context, transaction Transaction, event string, data map[string]interface{})
    CompleteTransaction(ctx context.Context, transaction Transaction, status TransactionStatus)
    GetTransactionMetrics(ctx context.Context, filter MetricsFilter) (TransactionMetrics, error)
}

// Transaction Implementation
type Transaction struct {
    ID            string                 `json:"id"`
    Name          string                 `json:"name"`
    CorrelationID string                 `json:"correlation_id"`
    UserID        string                 `json:"user_id"`
    StartTime     time.Time              `json:"start_time"`
    EndTime       *time.Time             `json:"end_time,omitempty"`
    Status        TransactionStatus      `json:"status"`
    Events        []TransactionEvent     `json:"events"`
    Metadata      map[string]interface{} `json:"metadata"`
    Spans         []Span                 `json:"spans"`
}

// Distributed Tracing Integration
type TracingMonitor struct {
    tracer opentracing.Tracer
    logger *zerolog.Logger
    metrics *prometheus.Registry
    // ...
}

func (t *TracingMonitor) TraceAIGeneration(ctx context.Context, req AIGenerationRequest) (context.Context, Transaction) {
    // Start distributed trace
    span, ctx := opentracing.StartSpanFromContext(ctx, "ai.generation")
    span.SetTag("user_id", req.UserID)
    span.SetTag("model", req.Model)
    span.SetTag("prompt_length", len(req.Prompt))
    
    // Create transaction
    transaction := Transaction{
        ID:            uuid.New().String(),
        Name:          "ai.generation",
        CorrelationID: span.Context().(jaeger.SpanContext).TraceID().String(),
        UserID:        req.UserID,
        StartTime:     time.Now(),
        Status:        TransactionStatusRunning,
        Metadata: map[string]interface{}{
            "model":         req.Model,
            "prompt_length": len(req.Prompt),
            "max_tokens":    req.MaxTokens,
        },
    }
    
    // Record metrics
    t.metrics.WithLabelValues("ai.generation", "started").Inc()
    
    return ctx, transaction
}

// Real-time Metrics Collection
type MetricsCollector struct {
    promClient prometheus.Registerer
    counters   map[string]prometheus.Counter
    histograms map[string]prometheus.Histogram
    gauges     map[string]prometheus.Gauge
}

func (m *MetricsCollector) RecordRequestLatency(service, method string, duration time.Duration) {
    histogram := m.histograms[fmt.Sprintf("%s_%s_duration", service, method)]
    histogram.Observe(duration.Seconds())
}

func (m *MetricsCollector) RecordMessageQueueMetrics(queue string, depth int, processing int) {
    m.gauges[fmt.Sprintf("queue_%s_depth", queue)].Set(float64(depth))
    m.gauges[fmt.Sprintf("queue_%s_processing", queue)].Set(float64(processing))
}
```

### **Key Metrics to Monitor**

**Performance Metrics:**
- Request latency (p50, p95, p99)
- Throughput (requests per second)
- Queue depth and processing time
- Circuit breaker state and trip count

**Reliability Metrics:**
- Success/failure rates
- Retry attempts and success rates
- Dead letter queue accumulation
- Service availability (uptime)

**Business Metrics:**
- User conversion rates
- Feature usage patterns
- Generation completion rates
- Error categorization

---

## **Technology Recommendations**

### **Message Queue Selection Matrix**

| Use Case | Primary Choice | Reasoning |
|----------|---------------|-----------|
| **Real-time Events** | Redis Streams | Low latency, ordering, replay capability |
| **High Throughput** | Apache Kafka | Distributed, high performance, persistence |
| **Complex Routing** | RabbitMQ | Advanced routing, reliability features |
| **Simple Pub/Sub** | Redis Pub/Sub | Minimal setup, good for notifications |

### **Circuit Breaker Libraries**

| Library | Language | Features |
|---------|----------|----------|
| `sony/gobreaker` | Go | Simple, efficient, good metrics |
| `hystrix-go` | Go | Netflix proven, comprehensive |
| `rubyist/circuitbreaker` | Go | Lightweight, customizable |

### **Retry Libraries**

| Library | Features |
|---------|----------|
| `avast/retry-go` | Exponential backoff, jitter, context support |
| `cenkalti/backoff` | Configurable backoff strategies |
| `go-redsync/redsync` | Distributed locking for coordinated retries |

---

## **Implementation Roadmap**

### **Phase 1: Foundation (1-2 weeks)**
1. **Implement Circuit Breaker Pattern** for existing gRPC clients
2. **Add Retry Logic** with exponential backoff
3. **Enhance Health Checks** with dependency validation
4. **Basic Metrics Collection** with Prometheus

### **Phase 2: Message Queues (2-3 weeks)**
1. **Deploy Redis Streams** for persistent messaging
2. **Implement Consumer Groups** for load balancing
3. **Add Dead Letter Queues** for failed message handling
4. **Create Message Replay** mechanisms

### **Phase 3: Advanced Monitoring (1-2 weeks)**
1. **Distributed Tracing** integration with Jaeger
2. **Comprehensive Metrics** dashboard in Grafana
3. **Alerting Rules** for critical failures
4. **Transaction Monitoring** across service boundaries

### **Phase 4: Event Sourcing (2-3 weeks)**
1. **Event Store** implementation with Redis Streams
2. **Event Replay** capabilities for system recovery
3. **CQRS Integration** for read/write separation
4. **Event-driven Architecture** for real-time features

---

## **Success Criteria**

### **Reliability Goals**
- **99.9% Uptime** for critical user-facing operations
- **Zero Message Loss** for persistent operations
- **< 5% Retry Rate** for inter-service communication
- **< 1 hour Recovery Time** from major outages

### **Performance Goals**
- **< 100ms p95** for synchronous gRPC calls
- **< 1 second** for message queue processing
- **< 50ms p95** for cache operations
- **> 1000 RPS** sustained throughput per service

### **Monitoring Goals**
- **100% Transaction Visibility** across service boundaries
- **< 1 minute** time to alert on critical issues
- **< 30 seconds** dashboard refresh rates
- **7 days** of detailed transaction history

This comprehensive communication pattern implementation will ensure your AI Code Generator system achieves enterprise-grade reliability, performance, and observability while maintaining the flexibility to scale and evolve.

## **Design Patterns Integration Summary**

This communication patterns guide integrates **12 key design patterns** from our established architecture, ensuring technology changes require minimal refactoring:

### **🏗️ Interface Patterns (Primary)**
- **Service Client Interface**: Technology-agnostic service communication
- **Message Queue Interface**: Swappable messaging implementations  
- **Transport Interface**: Protocol-independent communication (gRPC, HTTP, WebSocket)
- **Circuit Breaker Interface**: Pluggable resilience strategies
- **Retry Strategy Interface**: Configurable retry behaviors
- **Dead Letter Queue Interface**: Standardized failure handling

### **🏗️ Factory Patterns (Configuration-Driven)**
- **Service Client Factory**: Dynamic client creation based on environment
- **Message Queue Factory**: Runtime queue technology selection
- **Transport Factory**: Protocol selection via configuration
- **Metrics Factory**: Monitoring technology abstraction

### **🎯 Adapter Pattern (Protocol Translation)**
- **Domain Service Adapters**: Convert between business domain and transport protocols
- **Message Format Adapters**: Handle different serialization formats (JSON, Protobuf, Avro)

### **📋 Template Method Pattern (Standardized Workflows)**
- **Retry Template**: Consistent retry logic across all communication types
- **Circuit Breaker Template**: Standard failure detection and recovery patterns
- **Transaction Template**: Uniform transaction monitoring across services

### **🎯 Strategy Pattern (Runtime Behavior Selection)**
- **Retry Strategies**: Exponential backoff, linear, jitter-based
- **Serialization Strategies**: JSON for development, Protobuf for production
- **Routing Strategies**: Round-robin, weighted, circuit-breaker-aware

### **🔄 Observer Pattern (Event-Driven Architecture)**
- **Message Event Observers**: React to message lifecycle events
- **Circuit Breaker State Observers**: Monitor and alert on breaker state changes
- **Transaction Event Observers**: Distributed tracing and monitoring

## **Technology Change Impact Analysis**

| Change Scenario | Code Changes Required | Configuration Changes | Test Changes |
|-----------------|----------------------|----------------------|--------------|
| **Redis → Kafka** | ❌ None | ✅ Config only | ✅ New integration tests |
| **gRPC → HTTP** | ❌ None | ✅ Config + transport | ✅ Protocol-specific tests |
| **JSON → Protobuf** | ❌ None | ✅ Serializer config | ✅ Serialization tests |
| **Sony → Hystrix CB** | ❌ None | ✅ Breaker config | ✅ Breaker behavior tests |
| **Development → Production** | ❌ None | ✅ Environment config | ❌ None |

### **Key Design Pattern Benefits**

1. **🔧 Technology Abstraction**: Business logic isolated from infrastructure concerns
2. **📦 Dependency Injection**: Easy testing with mock implementations  
3. **🏗️ Factory Creation**: Runtime technology selection based on configuration
4. **🎯 Interface Segregation**: Clean boundaries between different communication concerns
5. **📋 Template Methods**: Consistent patterns for retries, circuit breaking, monitoring
6. **🔄 Open/Closed Principle**: Easy to extend with new technologies without modifying existing code

**Example: Complete Technology Stack Migration**

```yaml
# Before: Redis-based development environment
communication:
  message_queue:
    provider: "redis"
    address: "localhost:6379"
    serializer: "json"
  service_clients:
    transport: "http"
    circuit_breaker: "simple"
    
# After: Kafka-based production environment  
communication:
  message_queue:
    provider: "kafka"
    address: "kafka.prod.internal:9092"
    serializer: "protobuf"
  service_clients:
    transport: "grpc"
    circuit_breaker: "hystrix"
```

**Result**: Zero business logic code changes, complete technology stack migration through configuration alone.
