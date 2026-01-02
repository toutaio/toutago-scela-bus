# Scéla Usage Guide

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Publishing Messages](#publishing-messages)
- [Subscribing to Topics](#subscribing-to-topics)
- [Pattern Matching](#pattern-matching)
- [Middleware](#middleware)
- [Error Handling](#error-handling)
- [Observability](#observability)
- [Configuration](#configuration)
- [Best Practices](#best-practices)

## Installation

```bash
go get github.com/toutaio/toutago-scela-bus
```

Requirements: Go 1.22 or higher

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/toutaio/toutago-scela-bus/pkg/scela"
)

func main() {
    // Create a bus
    bus := scela.New()
    defer bus.Close()

    // Subscribe
    bus.Subscribe("user.created", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        fmt.Printf("User created: %v\n", msg.Payload())
        return nil
    }))

    // Publish
    ctx := context.Background()
    bus.Publish(ctx, "user.created", map[string]interface{}{
        "id": "123",
        "email": "user@example.com",
    })
}
```

## Publishing Messages

### Asynchronous Publishing

Publishes immediately and returns without waiting for handlers:

```go
err := bus.Publish(ctx, "order.created", orderData)
```

Use when:
- You don't need immediate confirmation
- High throughput is important
- Handlers may be slow

### Synchronous Publishing

Waits for all handlers to complete:

```go
err := bus.PublishSync(ctx, "payment.processed", paymentData)
```

Use when:
- You need to know if handlers succeeded
- Order of operations matters
- Running in a transaction

### Context Usage

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := bus.PublishSync(ctx, "task.execute", taskData)
if err == context.DeadlineExceeded {
    // Handler took too long
}

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
// Cancel from another goroutine
go func() {
    time.Sleep(1 * time.Second)
    cancel()
}()

bus.Publish(ctx, "long.task", data)
```

## Subscribing to Topics

### Basic Subscription

```go
sub, err := bus.Subscribe("user.created", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
    // Handle message
    return nil
}))
```

### Handler Interface

Implement the `Handler` interface for stateful handlers:

```go
type EmailHandler struct {
    smtp SMTPClient
}

func (h *EmailHandler) Handle(ctx context.Context, msg scela.Message) error {
    data := msg.Payload().(map[string]interface{})
    return h.smtp.Send(data["email"].(string), data["subject"].(string))
}

bus.Subscribe("email.send", &EmailHandler{smtp: smtpClient})
```

### Unsubscribing

```go
sub, _ := bus.Subscribe("temp.topic", handler)

// Later...
sub.Unsubscribe()
```

## Pattern Matching

### Exact Match

```go
bus.Subscribe("user.created", handler)  // Only "user.created"
```

### Wildcard Patterns

```go
// Match all user events
bus.Subscribe("user.*", handler)  // user.created, user.updated, user.deleted

// Match all creation events
bus.Subscribe("*.created", handler)  // user.created, order.created, etc.

// Match everything
bus.Subscribe("*", handler)  // All topics
```

### Multiple Subscriptions

A single topic can have multiple handlers:

```go
// All three will receive the message
bus.Subscribe("user.created", emailHandler)
bus.Subscribe("user.created", analyticsHandler)
bus.Subscribe("user.*", auditHandler)
```

## Middleware

Middleware wraps handlers to add cross-cutting concerns.

### Logging Middleware

```go
loggingMiddleware := func(next scela.Handler) scela.Handler {
    return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        log.Printf("Processing: %s", msg.Topic())
        start := time.Now()
        err := next.Handle(ctx, msg)
        log.Printf("Completed in %v", time.Since(start))
        return err
    })
}

bus.Use(loggingMiddleware)
```

### Authentication Middleware

```go
authMiddleware := func(next scela.Handler) scela.Handler {
    return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        user := msg.Metadata()["user"]
        if user == nil {
            return fmt.Errorf("unauthorized")
        }
        return next.Handle(ctx, msg)
    })
}
```

### Chaining Middleware

```go
bus.Use(
    loggingMiddleware,
    authMiddleware,
    metricsMiddleware,
)
```

Execution order: logging → auth → metrics → handler → metrics → auth → logging

## Error Handling

### Retry Configuration

```go
bus := scela.New(
    scela.WithMaxRetries(5),  // Retry failed messages up to 5 times
)
```

### Dead Letter Queue

```go
bus := scela.New(
    scela.WithMaxRetries(3),
    scela.WithDeadLetterHandler(scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        log.Printf("Failed message: %s - %v", msg.Topic(), msg.Payload())
        // Store in database, send alert, etc.
        return nil
    })),
)
```

### Error Handling in Handlers

```go
handler := scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
    data := msg.Payload().(map[string]interface{})
    
    if err := validate(data); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    if err := process(data); err != nil {
        return fmt.Errorf("processing failed: %w", err)
    }
    
    return nil  // Success
})
```

## Observability

### Metrics Observer

```go
type MetricsObserver struct {
    published int64
    processed int64
}

func (m *MetricsObserver) OnPublish(ctx context.Context, topic string, msg scela.Message) {
    atomic.AddInt64(&m.published, 1)
}

func (m *MetricsObserver) OnMessageProcessed(ctx context.Context, msg scela.Message, err error) {
    atomic.AddInt64(&m.processed, 1)
}

// Implement other Observer methods...

bus := scela.New(scela.WithObserver(&MetricsObserver{}))
```

### Distributed Tracing

```go
tracingMiddleware := func(next scela.Handler) scela.Handler {
    return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        span := trace.StartSpan(ctx, msg.Topic())
        defer span.End()
        
        return next.Handle(trace.WithSpan(ctx, span), msg)
    })
}
```

## Configuration

### Worker Pool Size

```go
bus := scela.New(
    scela.WithWorkers(20),  // 20 worker goroutines
)
```

More workers = higher concurrency, but more memory usage.

### Queue Size

Not directly configurable, but buffered at 1000 messages by default.

## Best Practices

### Topic Naming

Use hierarchical names with dots:

```go
// Good
"user.created"
"order.payment.processed"
"analytics.page.viewed"

// Avoid
"userCreated"  // No hierarchy
"user_created"  // Use dots, not underscores
"USER.CREATED"  // Use lowercase
```

### Message Payloads

Use structured data:

```go
// Good
bus.Publish(ctx, "user.created", map[string]interface{}{
    "id": "123",
    "email": "user@example.com",
    "timestamp": time.Now(),
})

// Better - use structs
type UserCreatedEvent struct {
    ID        string
    Email     string
    Timestamp time.Time
}

bus.Publish(ctx, "user.created", UserCreatedEvent{
    ID: "123",
    Email: "user@example.com",
    Timestamp: time.Now(),
})
```

### Error Handling

Always handle errors from handlers:

```go
handler := scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
    if err := doSomething(); err != nil {
        log.Printf("Error handling %s: %v", msg.Topic(), err)
        return err  // Will trigger retry
    }
    return nil
})
```

### Resource Cleanup

Always close the bus when done:

```go
bus := scela.New()
defer bus.Close()  // Ensures graceful shutdown
```

### Testing

Use synchronous publishing in tests:

```go
func TestUserHandler(t *testing.T) {
    bus := scela.New()
    defer bus.Close()
    
    var received bool
    bus.Subscribe("test.topic", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        received = true
        return nil
    }))
    
    // Use PublishSync to wait for handler
    bus.PublishSync(context.Background(), "test.topic", nil)
    
    if !received {
        t.Error("Handler not called")
    }
}
```

### Avoid Blocking

Don't block in handlers for long operations:

```go
// Bad - blocks worker
handler := scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
    time.Sleep(10 * time.Second)  // Blocks worker
    return nil
})

// Good - spawn goroutine for long work
handler := scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
    go func() {
        time.Sleep(10 * time.Second)
        // Do work
    }()
    return nil
})
```

### Memory Management

For high-throughput scenarios, consider pooling:

```go
var payloadPool = sync.Pool{
    New: func() interface{} {
        return make(map[string]interface{})
    },
}

handler := scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
    payload := payloadPool.Get().(map[string]interface{})
    defer func() {
        // Clear and return to pool
        for k := range payload {
            delete(payload, k)
        }
        payloadPool.Put(payload)
    }()
    
    // Use payload
    return nil
})
```

## Examples

See the [examples directory](../examples) for complete working examples:

- [basic](../examples/basic) - Simple pub/sub
- [async](../examples/async) - Asynchronous processing
- [middleware](../examples/middleware) - Middleware usage
- [dlq](../examples/dlq) - Dead letter queue and retry
- [metrics](../examples/metrics) - Metrics and observability
