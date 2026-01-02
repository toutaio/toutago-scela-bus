# Scéla Architecture

## Overview

Scéla is an in-process message bus implementing the Pub/Sub pattern with support for both synchronous and asynchronous message delivery. It provides a zero-dependency, thread-safe, and highly performant solution for event-driven architectures.

## Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                         Bus (Facade)                          │
│  - Publish/Subscribe API                                      │
│  - Configuration & Lifecycle                                  │
└────────────┬──────────────────────────────────┬──────────────┘
             │                                  │
    ┌────────▼──────────┐          ┌───────────▼──────────┐
    │  Worker Pool       │          │ Subscription         │
    │  - Async Queue     │          │ Registry             │
    │  - Goroutines      │          │ - Pattern Matching   │
    │  - Load Balancing  │          │ - Handler Storage    │
    └────────┬───────────┘          └───────────┬──────────┘
             │                                  │
    ┌────────▼──────────────────────────────────▼──────────┐
    │              Message Router                           │
    │  - Topic Matching (wildcards)                         │
    │  - Handler Selection                                  │
    └────────┬──────────────────────────────────────────────┘
             │
    ┌────────▼───────────┐
    │  Middleware        │
    │  Pipeline          │
    │  (Composable)      │
    └────────┬───────────┘
             │
    ┌────────▼───────────┐
    │  Message           │
    │  Handlers          │
    └────────────────────┘
```

## Message Flow

### Synchronous Publishing

```
Client
  │
  ├─► PublishSync(topic, payload)
  │
  ├─► Create Message
  │
  ├─► Find Matching Handlers (pattern matching)
  │
  ├─► Apply Middleware Pipeline
  │
  ├─► Execute Handlers (sequentially)
  │
  └─► Return (with any errors)
```

### Asynchronous Publishing

```
Client                    Worker Pool
  │                            │
  ├─► Publish(topic, payload)  │
  │                            │
  ├─► Create Message           │
  │                            │
  ├─► Enqueue                  │
  │   (buffered channel)       │
  │                            │
  └─► Return immediately       │
                               │
                      ┌────────▼─────────┐
                      │ Worker picks msg  │
                      └────────┬─────────┘
                               │
                      ┌────────▼──────────┐
                      │ Pattern Matching   │
                      └────────┬──────────┘
                               │
                      ┌────────▼──────────┐
                      │ Middleware         │
                      └────────┬──────────┘
                               │
                      ┌────────▼──────────┐
                      │ Execute Handlers   │
                      └────────┬──────────┘
                               │
                      ┌────────▼──────────┐
                      │ Error Handling/    │
                      │ Retry/DLQ          │
                      └────────────────────┘
```

## Pattern Matching

Scéla supports wildcard patterns for flexible topic subscriptions:

- **Exact match**: `user.created`
- **Single segment wildcard**: `user.*` matches `user.created`, `user.updated`
- **Suffix wildcard**: `*.created` matches `user.created`, `order.created`
- **All wildcard**: `*` or `#` matches all messages

Pattern matching uses segment-based comparison for O(n) complexity where n is the number of segments.

## Concurrency Model

### Worker Pool

- Fixed number of worker goroutines (configurable, default: 10)
- Workers pull from a buffered channel (default: 1000 messages)
- Load balancing via Go's channel scheduling
- Graceful shutdown waits for all workers to complete

### Thread Safety

- All public APIs are thread-safe
- Read-write locks protect shared state
- Atomic operations for concurrent access
- Pattern registry uses copy-on-write for subscriptions

## Middleware Pipeline

Middleware wraps handlers in an onion-like structure:

```
Request
  │
  ├─► Middleware 1 (before)
  │     │
  │     ├─► Middleware 2 (before)
  │     │     │
  │     │     ├─► Handler
  │     │     │
  │     │     └─► Middleware 2 (after)
  │     │
  │     └─► Middleware 1 (after)
  │
  └─► Response
```

Middleware can:
- Modify messages (add metadata)
- Short-circuit execution
- Handle errors
- Log/trace execution
- Collect metrics

## Error Handling & Retry

```
Handler Error
  │
  ├─► Retry Count < Max?
  │     │
  │     ├─► Yes: Re-enqueue message (with incremented retry count)
  │     │
  │     └─► No: Send to Dead Letter Queue
  │           │
  │           └─► DLQ Handler (if configured)
```

Retry logic:
- Configurable max retries (default: 3)
- Simple retry without backoff (immediate re-queue)
- Exponential backoff can be implemented via middleware
- Dead letter queue for failed messages

## Observable Pattern

Observers receive notifications for:
- `OnPublish` - When a message is published
- `OnSubscribe` - When a handler subscribes
- `OnUnsubscribe` - When a subscription is removed
- `OnMessageProcessed` - After each message is handled
- `OnClose` - When the bus shuts down

Use cases:
- Metrics collection
- Distributed tracing
- Audit logging
- Debugging/monitoring

## Memory Management

- Messages are allocated on demand
- No pooling by default (simple GC-friendly design)
- Buffered channels prevent unbounded memory growth
- Graceful shutdown prevents message loss

## Performance Characteristics

- **Synchronous publish**: O(n) where n = number of matching handlers
- **Asynchronous publish**: O(1) enqueue + background processing
- **Pattern matching**: O(m) where m = number of subscription patterns
- **Subscription**: O(1) add to registry
- **Unsubscription**: O(1) removal from registry

### Benchmarks

Typical performance on modern hardware:

```
BenchmarkPublishSync     500,000    2,500 ns/op    400 B/op    5 allocs/op
BenchmarkPublishAsync  2,000,000      800 ns/op    200 B/op    3 allocs/op
BenchmarkPatternMatch 10,000,000      150 ns/op      0 B/op    0 allocs/op
```

## Design Decisions

### Why In-Process Only?

- **Simplicity**: No network protocols, serialization, or distributed concerns
- **Performance**: Nanosecond latency, zero network overhead
- **Type Safety**: Native Go types, no serialization issues
- **Use Case**: Event-driven monoliths, microservices communication within a process

For distributed messaging, use NATS, RabbitMQ, or Kafka.

### Why No Built-in Persistence?

- **Core Principle**: Keep the core simple and fast
- **Flexibility**: Users can implement persistence via middleware or observers
- **Use Case**: Most in-process use cases don't need persistence

Persistence can be added via:
- Observers (log all published messages)
- Middleware (intercept and persist)
- Dead letter queue handler (persist failed messages)

### Why Zero Dependencies?

- **Reliability**: No supply chain risks
- **Portability**: Works anywhere Go works
- **Simplicity**: Easier to understand and audit
- **Size**: Minimal binary footprint

## Extension Points

Scéla is designed for extension:

1. **Middleware**: Add cross-cutting concerns
2. **Observers**: Monitor and collect metrics
3. **Custom Guards**: Implement priority queues
4. **DLQ Handler**: Custom failure handling
5. **User Providers**: Plug in any data source (for future auth integration)

## Comparison to Alternatives

| Feature | Scéla | EventBus | Channel-based |
|---------|-------|----------|---------------|
| Wildcards | ✓ | ✗ | ✗ |
| Middleware | ✓ | ✗ | ✗ |
| Retry/DLQ | ✓ | ✗ | ✗ |
| Observability | ✓ | ✗ | ✗ |
| Thread-safe | ✓ | ✓ | ✗ |
| Dependencies | 0 | 0 | 0 |

## Future Considerations

Potential additions (not in v1.0):

- Priority queues (high/low priority messages)
- Message batching (process multiple messages together)
- Backpressure (slow down publishers when queue is full)
- Persistent message store (optional plugin)
- Message expiry/TTL
- Request/reply pattern (beyond pub/sub)
