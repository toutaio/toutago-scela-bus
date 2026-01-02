# Migration Guide: From Internal Message Bus to Scéla

This guide helps you migrate from Toutā's internal message bus to the new **Scéla** (toutago-scela-bus) component library.

## Why Migrate?

The new Scéla message bus provides:

- **Better performance** - Optimized worker pool and pattern matching
- **More features** - Middleware, retry logic, DLQ, observers
- **Zero dependencies** - Only standard library
- **Production-ready** - 92.8% test coverage, thoroughly documented
- **Flexible** - Can be used standalone or with Toutā

## Quick Migration

### Before (Internal Bus)

```go
// Old internal message bus (if you were using one)
type MessageBus struct {
    handlers map[string][]func(interface{})
}

func (b *MessageBus) Publish(topic string, data interface{}) {
    for _, handler := range b.handlers[topic] {
        go handler(data)
    }
}

func (b *MessageBus) Subscribe(topic string, handler func(interface{})) {
    b.handlers[topic] = append(b.handlers[topic], handler)
}
```

### After (Scéla)

```go
import (
    "context"
    "github.com/toutaio/toutago-scela-bus/pkg/scela"
    "github.com/toutaio/toutago/pkg/touta/integration"
)

// Create bus
bus := integration.NewScelaBus()
defer bus.Close()

// Subscribe with typed handler
_, err := bus.Subscribe("user.created", scela.HandlerFunc(
    func(ctx context.Context, msg scela.Message) error {
        user := msg.Payload().(UserData)
        // Handle user creation
        return nil
    },
))

// Publish
ctx := context.Background()
err = bus.Publish(ctx, "user.created", UserData{
    ID: "123",
    Email: "user@example.com",
})
```

## Key Differences

### 1. Context Support

**Before**: No context support
```go
bus.Publish("event", data)
```

**After**: Context-aware for timeouts and cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
bus.Publish(ctx, "event", data)
```

### 2. Error Handling

**Before**: No error handling
```go
bus.Publish("event", data) // Fire and forget
```

**After**: Proper error handling and retry
```go
if err := bus.Publish(ctx, "event", data); err != nil {
    log.Printf("Publish failed: %v", err)
}
```

### 3. Pattern Matching

**Before**: Exact topic match only
```go
bus.Subscribe("user.created", handler)
```

**After**: Wildcard patterns supported
```go
// Match all user events
bus.Subscribe("user.*", handler)

// Match all creation events
bus.Subscribe("*.created", handler)

// Match everything
bus.Subscribe("*", handler)
```

### 4. Synchronous vs Asynchronous

**Before**: Always asynchronous
```go
bus.Publish("event", data) // Returns immediately
```

**After**: Choose your mode
```go
// Async (fire and forget)
bus.Publish(ctx, "event", data)

// Sync (wait for handlers)
bus.PublishSync(ctx, "event", data)
```

## Migration Steps

### Step 1: Add Dependency

```bash
go get github.com/toutaio/toutago-scela-bus
```

### Step 2: Replace Bus Creation

```go
// Old
bus := &MessageBus{
    handlers: make(map[string][]func(interface{})),
}

// New
bus := integration.NewScelaBus()
defer bus.Close() // Important: graceful shutdown
```

### Step 3: Update Subscriptions

```go
// Old
bus.Subscribe("user.created", func(data interface{}) {
    user := data.(UserData)
    // Handle event
})

// New
bus.Subscribe("user.created", scela.HandlerFunc(
    func(ctx context.Context, msg scela.Message) error {
        user := msg.Payload().(UserData)
        // Handle event
        return nil // or return error for retry
    },
))
```

### Step 4: Update Publishing

```go
// Old
bus.Publish("user.created", userData)

// New
ctx := context.Background()
bus.Publish(ctx, "user.created", userData)
```

### Step 5: Add Graceful Shutdown

```go
func main() {
    bus := integration.NewScelaBus()
    defer bus.Close() // Ensure clean shutdown

    // ... rest of your app
}
```

## Advanced Features

### Middleware

Add cross-cutting concerns like logging:

```go
loggingMiddleware := func(next scela.Handler) scela.Handler {
    return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        log.Printf("Processing: %s", msg.Topic())
        return next.Handle(ctx, msg)
    })
}

bus.Use(loggingMiddleware)
```

### Retry Logic

Configure automatic retries:

```go
bus := integration.NewScelaBus(
    scela.WithMaxRetries(3),
)
```

### Dead Letter Queue

Handle failed messages:

```go
bus := integration.NewScelaBus(
    scela.WithMaxRetries(3),
    scela.WithDeadLetterHandler(scela.HandlerFunc(
        func(ctx context.Context, msg scela.Message) error {
            log.Printf("Failed message: %s", msg.Topic())
            // Store in database, send alert, etc.
            return nil
        },
    )),
)
```

### Metrics

Track message processing:

```go
type MetricsObserver struct {
    published int64
}

func (m *MetricsObserver) OnPublish(ctx context.Context, topic string, msg scela.Message) {
    atomic.AddInt64(&m.published, 1)
}

// Implement other Observer methods...

bus := integration.NewScelaBus(
    scela.WithObserver(&MetricsObserver{}),
)
```

## Common Patterns

### Request-Reply Pattern

```go
// Create correlation ID
correlationID := uuid.New().String()

// Subscribe to reply
replyChan := make(chan interface{})
bus.Subscribe(fmt.Sprintf("reply.%s", correlationID), 
    scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        replyChan <- msg.Payload()
        return nil
    }),
)

// Send request
bus.Publish(ctx, "request.process", RequestData{
    CorrelationID: correlationID,
    // ... request data
})

// Wait for reply
select {
case reply := <-replyChan:
    // Handle reply
case <-time.After(5 * time.Second):
    // Timeout
}
```

### Event Sourcing

```go
// Store all events
bus.Subscribe("*", scela.HandlerFunc(
    func(ctx context.Context, msg scela.Message) error {
        event := Event{
            Topic:     msg.Topic(),
            Payload:   msg.Payload(),
            Timestamp: msg.Timestamp(),
        }
        return eventStore.Save(event)
    },
))
```

## Testing

### Before

```go
// Hard to test
func TestUserService(t *testing.T) {
    bus := &MessageBus{}
    service := NewUserService(bus)
    service.CreateUser(userData)
    // No way to verify message was sent
}
```

### After

```go
func TestUserService(t *testing.T) {
    bus := integration.NewScelaBus()
    defer bus.Close()

    var received bool
    bus.Subscribe("user.created", scela.HandlerFunc(
        func(ctx context.Context, msg scela.Message) error {
            received = true
            return nil
        },
    ))

    service := NewUserService(bus)
    service.CreateUser(userData)

    // Use PublishSync in tests
    if !received {
        t.Error("Message not published")
    }
}
```

## Performance Considerations

### Worker Pool Size

```go
// Default: 10 workers
bus := integration.NewScelaBus()

// High throughput: more workers
bus := integration.NewScelaBus(scela.WithWorkers(50))

// Low throughput: fewer workers
bus := integration.NewScelaBus(scela.WithWorkers(3))
```

### Buffering

Queue is buffered at 1000 messages by default. If you're dropping messages:

1. Increase workers
2. Make handlers faster
3. Use async publishing where possible

## Troubleshooting

### Messages Not Being Delivered

1. Check pattern matching: `user.created` vs `user.*`
2. Verify bus not closed
3. Check for handler errors (use DLQ to see failures)

### Performance Issues

1. Use async publishing for non-critical messages
2. Increase worker count
3. Profile handlers for bottlenecks
4. Consider middleware overhead

### Memory Leaks

1. Always call `defer bus.Close()`
2. Unsubscribe unused subscriptions
3. Don't capture large values in handler closures

## Getting Help

- Documentation: [GitHub](https://github.com/toutaio/toutago-scela-bus)
- Examples: [examples/](https://github.com/toutaio/toutago-scela-bus/tree/main/examples)
- Issues: [GitHub Issues](https://github.com/toutaio/toutago-scela-bus/issues)

## Migration Checklist

- [ ] Add scela dependency to go.mod
- [ ] Replace bus creation code
- [ ] Update all Subscribe calls with new signature
- [ ] Update all Publish calls with context
- [ ] Add defer bus.Close() for graceful shutdown
- [ ] Update tests to use scela types
- [ ] Consider adding middleware (logging, metrics)
- [ ] Configure retry logic if needed
- [ ] Set up dead letter queue for failures
- [ ] Run tests to verify migration
- [ ] Profile performance and adjust workers if needed
