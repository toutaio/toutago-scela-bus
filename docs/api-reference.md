# API Reference

## Package `scela`

The main package providing the message bus implementation.

### Core Interfaces

#### `Bus`

The primary interface for the message bus.

```go
type Bus interface {
    Publish(ctx context.Context, topic string, payload interface{}) error
    PublishWithPriority(ctx context.Context, topic string, payload interface{}, priority Priority) error
    Subscribe(pattern string, handler Handler) (SubscriptionID, error)
    Unsubscribe(id SubscriptionID) error
    Close() error
}
```

**Methods:**

- `Publish(ctx, topic, payload)` - Publishes a message to a topic
- `PublishWithPriority(ctx, topic, payload, priority)` - Publishes a message with a specific priority
- `Subscribe(pattern, handler)` - Subscribes a handler to topics matching the pattern
- `Unsubscribe(id)` - Unsubscribes a handler by ID
- `Close()` - Gracefully shuts down the bus

#### `Message`

Represents a message in the bus.

```go
type Message interface {
    ID() string
    Topic() string
    Payload() interface{}
    Timestamp() time.Time
    Priority() Priority
    Metadata() map[string]interface{}
}
```

#### `Handler`

Interface for message handlers.

```go
type Handler interface {
    Handle(ctx context.Context, msg Message) error
}
```

**Helper:**

```go
type HandlerFunc func(ctx context.Context, msg Message) error
func (f HandlerFunc) Handle(ctx context.Context, msg Message) error
```

### Configuration Options

Create a bus with configuration options:

```go
bus := scela.New(
    scela.WithWorkers(10),
    scela.WithBufferSize(1000),
    scela.WithMiddleware(loggingMiddleware),
    scela.WithPersistence(store),
    scela.WithDeadLetterQueue(dlqHandler),
)
```

#### `WithWorkers(n int)`

Sets the number of worker goroutines for async processing. Default: 10.

#### `WithBufferSize(size int)`

Sets the message queue buffer size. Default: 100.

#### `WithMiddleware(middleware ...Middleware)`

Adds middleware to the processing pipeline.

#### `WithPersistence(store PersistenceStore)`

Enables message persistence with the given store.

#### `WithDeadLetterQueue(handler Handler)`

Sets a handler for messages that fail processing after retries.

#### `WithMetrics(collector MetricsCollector)`

Enables metrics collection.

### Middleware

Middleware allows you to intercept and modify message processing.

```go
type Middleware func(Handler) Handler
```

**Built-in Middleware:**

- `LoggingMiddleware()` - Logs message processing
- `RetryMiddleware(attempts int, backoff time.Duration)` - Retries failed messages
- `TimeoutMiddleware(timeout time.Duration)` - Adds timeout to handlers
- `FilterMiddleware(predicate func(Message) bool)` - Filters messages

### Persistence

#### `PersistenceStore`

Interface for message persistence.

```go
type PersistenceStore interface {
    Save(msg Message) error
    Load(id string) (Message, error)
    LoadAll() ([]Message, error)
    Delete(id string) error
    Clear() error
}
```

**Built-in Stores:**

- `NewMemoryStore()` - In-memory persistence (default)
- `NewFileStore(path string)` - File-based persistence
- `NewSQLStore(db *sql.DB, tableName string)` - SQL database persistence

### Priority Levels

```go
const (
    PriorityLow    Priority = 0
    PriorityNormal Priority = 1
    PriorityHigh   Priority = 2
    PriorityUrgent Priority = 3
)
```

Higher priority messages are processed first.

### Pattern Matching

Topic subscription patterns support:

- `*` - Matches any single segment (e.g., `user.*` matches `user.created`, `user.updated`)
- `**` - Matches any segments (e.g., `user.**` matches `user.created.admin`, `user.updated`)
- Exact match (e.g., `user.created` matches only `user.created`)

### Metrics

#### `MetricsCollector`

Interface for collecting metrics.

```go
type MetricsCollector interface {
    MessagePublished(topic string)
    MessageProcessed(topic string, duration time.Duration)
    MessageFailed(topic string, err error)
    SubscriberAdded(pattern string)
    SubscriberRemoved(id SubscriptionID)
}
```

### Error Handling

All errors returned by Scéla are wrapped with context. Common error scenarios:

- `ErrBusClosed` - Operation attempted on closed bus
- `ErrInvalidPattern` - Invalid subscription pattern
- `ErrSubscriptionNotFound` - Subscription ID not found
- `ErrHandlerFailed` - Handler execution failed

### Context Propagation

Scéla fully supports Go contexts for:

- Cancellation
- Timeouts
- Deadline propagation
- Value propagation

Example:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := bus.Publish(ctx, "topic", data)
```

### Graceful Shutdown

```go
// Signal shutdown
bus.Close()

// Bus will:
// 1. Stop accepting new messages
// 2. Process all queued messages
// 3. Wait for all handlers to complete
// 4. Close all resources
```

### Thread Safety

All Scéla operations are thread-safe and can be safely called from multiple goroutines concurrently.
