# Best Practices

## General Guidelines

### 1. Error Handling

Always check errors from bus operations:

```go
// Good
if err := bus.Publish(ctx, "topic", data); err != nil {
    log.Printf("failed to publish: %v", err)
    return err
}

// Bad
bus.Publish(ctx, "topic", data) // Ignoring error
```

### 2. Context Usage

Always pass a proper context:

```go
// Good: Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
bus.Publish(ctx, "topic", data)

// Good: Use request context in HTTP handlers
func handler(w http.ResponseWriter, r *http.Request) {
    bus.Publish(r.Context(), "topic", data)
}

// Bad: Using context.Background() everywhere
bus.Publish(context.Background(), "topic", data)
```

### 3. Graceful Shutdown

Always close the bus on application shutdown:

```go
func main() {
    bus := scela.New()
    defer bus.Close()
    
    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Shutting down...")
        bus.Close()
        os.Exit(0)
    }()
    
    // Your application code
}
```

## Topic Naming Conventions

### Use Hierarchical Naming

```go
// Good: Clear hierarchy
"user.created"
"user.updated"
"user.deleted"
"order.placed"
"order.shipped"
"payment.succeeded"

// Bad: Flat naming
"userCreated"
"updateUser"
"newOrder"
```

### Use Consistent Verb Tense

```go
// Good: Past tense (events that happened)
"user.created"
"order.placed"
"email.sent"

// Also Good: Present tense (commands to execute)
"user.create"
"order.place"
"email.send"

// Bad: Mixed tenses
"user.created"
"order.place"
"email.sending"
```

### Namespace by Domain

```go
// Good: Domain-based namespacing
"auth.user.logged_in"
"billing.invoice.generated"
"inventory.stock.updated"

// Better with patterns
bus.Subscribe("auth.**", authHandler)
bus.Subscribe("billing.**", billingHandler)
```

## Handler Design

### Keep Handlers Focused

```go
// Good: Single responsibility
func HandleUserCreated(ctx context.Context, msg scela.Message) error {
    user := msg.Payload().(*User)
    return sendWelcomeEmail(user)
}

// Bad: Multiple responsibilities
func HandleUserCreated(ctx context.Context, msg scela.Message) error {
    user := msg.Payload().(*User)
    sendWelcomeEmail(user)
    updateAnalytics(user)
    notifyAdmin(user)
    createUserDirectory(user)
    // Too many things!
}
```

### Make Handlers Idempotent

```go
// Good: Idempotent handler
func HandlePayment(ctx context.Context, msg scela.Message) error {
    payment := msg.Payload().(*Payment)
    
    // Check if already processed
    if isProcessed(payment.ID) {
        return nil // Already done
    }
    
    if err := processPayment(payment); err != nil {
        return err
    }
    
    markProcessed(payment.ID)
    return nil
}
```

### Handle Errors Gracefully

```go
// Good: Proper error handling
func HandleOrder(ctx context.Context, msg scela.Message) error {
    order := msg.Payload().(*Order)
    
    if err := validateOrder(order); err != nil {
        // Log and return error for retry
        log.Printf("invalid order %s: %v", order.ID, err)
        return fmt.Errorf("validation failed: %w", err)
    }
    
    if err := processOrder(order); err != nil {
        // Determine if retriable
        if isRetriable(err) {
            return err // Will retry
        }
        // Non-retriable, log and skip
        log.Printf("permanent failure for order %s: %v", order.ID, err)
        return nil // Don't retry
    }
    
    return nil
}
```

## Performance Optimization

### 1. Choose the Right Worker Count

```go
// CPU-bound work
bus := scela.New(scela.WithWorkers(runtime.NumCPU()))

// I/O-bound work (database, API calls)
bus := scela.New(scela.WithWorkers(runtime.NumCPU() * 4))

// Mixed workload
bus := scela.New(scela.WithWorkers(runtime.NumCPU() * 2))
```

### 2. Use Batching for High Volume

```go
// Good: Batch processing
type BatchHandler struct {
    batch []Message
    mu    sync.Mutex
}

func (h *BatchHandler) Handle(ctx context.Context, msg scela.Message) error {
    h.mu.Lock()
    h.batch = append(h.batch, msg)
    
    if len(h.batch) >= 100 {
        batch := h.batch
        h.batch = nil
        h.mu.Unlock()
        return processBatch(batch)
    }
    h.mu.Unlock()
    return nil
}
```

### 3. Avoid Heavy Work in Handlers

```go
// Good: Offload heavy work
func HandleImageUpload(ctx context.Context, msg scela.Message) error {
    image := msg.Payload().(*Image)
    
    // Quick validation only
    if !isValidImage(image) {
        return errors.New("invalid image")
    }
    
    // Offload heavy processing to worker pool
    go processImageAsync(image)
    return nil
}

// Bad: Heavy work blocks handler
func HandleImageUpload(ctx context.Context, msg scela.Message) error {
    image := msg.Payload().(*Image)
    
    // Blocks for seconds!
    resized := resizeImage(image)
    watermarked := addWatermark(resized)
    optimized := optimize(watermarked)
    
    return saveImage(optimized)
}
```

## Message Design

### Use Strong Types

```go
// Good: Structured payload
type UserCreatedEvent struct {
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

bus.Publish(ctx, "user.created", &UserCreatedEvent{...})

// Bad: Unstructured map
bus.Publish(ctx, "user.created", map[string]interface{}{
    "id": "123",
    "email": "user@example.com",
})
```

### Include Metadata

```go
// Good: Rich message with metadata
msg := &scela.BaseMessage{
    TopicVal:   "order.placed",
    PayloadVal: order,
    MetadataVal: map[string]interface{}{
        "user_id": user.ID,
        "source": "web",
        "trace_id": traceID,
    },
}
```

### Version Your Events

```go
// Good: Versioned events
type UserCreatedEventV1 struct {
    Version string `json:"version"` // "1.0"
    UserID  string `json:"user_id"`
}

type UserCreatedEventV2 struct {
    Version   string `json:"version"` // "2.0"
    UserID    string `json:"user_id"`
    AccountID string `json:"account_id"` // New field
}

// Handler supports both versions
func HandleUserCreated(ctx context.Context, msg scela.Message) error {
    switch v := msg.Payload().(type) {
    case *UserCreatedEventV1:
        // Handle V1
    case *UserCreatedEventV2:
        // Handle V2
    }
}
```

## Testing

### Test Handlers in Isolation

```go
func TestUserCreatedHandler(t *testing.T) {
    handler := &UserCreatedHandler{}
    
    msg := &scela.BaseMessage{
        TopicVal: "user.created",
        PayloadVal: &User{ID: "123"},
    }
    
    err := handler.Handle(context.Background(), msg)
    assert.NoError(t, err)
}
```

### Use Mock Bus for Integration Tests

```go
type MockBus struct {
    published []Message
}

func (m *MockBus) Publish(ctx context.Context, topic string, payload interface{}) error {
    m.published = append(m.published, ...)
    return nil
}

func TestService(t *testing.T) {
    mockBus := &MockBus{}
    service := NewService(mockBus)
    
    service.CreateUser(&User{...})
    
    assert.Len(t, mockBus.published, 1)
    assert.Equal(t, "user.created", mockBus.published[0].Topic())
}
```

## Monitoring and Observability

### Add Logging Middleware

```go
func LoggingMiddleware() scela.Middleware {
    return func(next scela.Handler) scela.Handler {
        return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
            start := time.Now()
            err := next.Handle(ctx, msg)
            duration := time.Since(start)
            
            log.Printf(
                "topic=%s duration=%s error=%v",
                msg.Topic(),
                duration,
                err,
            )
            return err
        })
    }
}
```

### Collect Metrics

```go
type MetricsCollector struct {
    published prometheus.Counter
    processed prometheus.Histogram
    failed    prometheus.Counter
}

bus := scela.New(scela.WithMetrics(&MetricsCollector{...}))
```

### Add Tracing

```go
func TracingMiddleware(tracer trace.Tracer) scela.Middleware {
    return func(next scela.Handler) scela.Handler {
        return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
            ctx, span := tracer.Start(ctx, "message.process")
            span.SetAttributes(
                attribute.String("topic", msg.Topic()),
                attribute.String("message_id", msg.ID()),
            )
            defer span.End()
            
            err := next.Handle(ctx, msg)
            if err != nil {
                span.RecordError(err)
            }
            return err
        })
    }
}
```

## Common Pitfalls

### ❌ Don't Block in Handlers

```go
// Bad
func HandleEmail(ctx context.Context, msg scela.Message) error {
    time.Sleep(10 * time.Second) // Blocks worker!
    return sendEmail(msg.Payload())
}

// Good
func HandleEmail(ctx context.Context, msg scela.Message) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return sendEmail(ctx, msg.Payload())
}
```

### ❌ Don't Modify Shared State Without Locking

```go
// Bad
var counter int
func HandleIncrement(ctx context.Context, msg scela.Message) error {
    counter++ // Race condition!
    return nil
}

// Good
var (
    counter int
    mu      sync.Mutex
)
func HandleIncrement(ctx context.Context, msg scela.Message) error {
    mu.Lock()
    counter++
    mu.Unlock()
    return nil
}
```

### ❌ Don't Forget to Unsubscribe

```go
// Bad
func setupSubscription() {
    bus.Subscribe("temp.topic", handler) // Leaks!
}

// Good
func setupSubscription() scela.SubscriptionID {
    id, _ := bus.Subscribe("temp.topic", handler)
    return id
}

func cleanup(id scela.SubscriptionID) {
    bus.Unsubscribe(id)
}
```

## Summary

✅ **DO:**
- Use contexts properly
- Handle errors gracefully
- Keep handlers focused and idempotent
- Use strong types for messages
- Monitor and measure
- Test thoroughly

❌ **DON'T:**
- Block in handlers
- Ignore errors
- Use global state without locking
- Forget to clean up subscriptions
- Over-complicate message payloads
