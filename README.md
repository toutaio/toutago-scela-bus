# Scéla - Toutā Message Bus

[![CI](https://github.com/toutaio/toutago-scela-bus/workflows/CI/badge.svg)](https://github.com/toutaio/toutago-scela-bus/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/toutaio/toutago-scela-bus)](https://goreportcard.com/report/github.com/toutaio/toutago-scela-bus)
[![Go Reference](https://pkg.go.dev/badge/github.com/toutaio/toutago-scela-bus.svg)](https://pkg.go.dev/github.com/toutaio/toutago-scela-bus)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Scéla** (Old Irish: "news, tidings, messages") is a production-ready, in-process message bus for Go applications. Part of the [Toutā Framework](https://github.com/toutaio/toutago), Scéla provides pub/sub messaging with both synchronous and asynchronous delivery, pattern matching, middleware support, and more.

## Features

- 🔌 **Pub/Sub Pattern** - Topic-based message routing with multiple subscribers
- ⚡ **Sync & Async** - Choose between synchronous or asynchronous message delivery
- 🎯 **Pattern Matching** - Subscribe with wildcards (`user.*`, `*.created`)
- 🔗 **Middleware Pipeline** - Intercept messages for logging, metrics, validation
- 🔄 **Retry Logic** - Configurable retry logic for failed messages
- 💀 **Dead Letter Queue** - Handle failed messages gracefully
- 🧵 **Thread-Safe** - Fully concurrent, safe for goroutines
- 🎛️ **Context Support** - Cancellation, timeouts, request-scoped data
- 📊 **Observable** - Hooks for metrics and monitoring
- 💾 **Persistence** - File-based and SQL database message persistence
- 📜 **Audit Trail** - Complete message history and event tracking
- 🚀 **Zero Dependencies** - Only standard library (persistence features optional)
- ✅ **Production Ready** - 87.5% test coverage, comprehensive CI/CD

## Installation

```bash
go get github.com/toutaio/toutago-scela-bus
```

**Requirements:** Go 1.22 or higher

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/toutaio/toutago-scela-bus/pkg/scela"
)

func main() {
    // Create a new message bus
    bus := scela.New()
    defer bus.Close()
    
    // Subscribe to messages
    sub, err := bus.Subscribe("user.created", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        fmt.Printf("Received: %s - %v\n", msg.Topic(), msg.Payload())
        return nil
    }))
    if err != nil {
        log.Fatal(err)
    }
    defer sub.Unsubscribe()
    
    // Publish a message (async)
    err = bus.Publish(context.Background(), "user.created", map[string]interface{}{
        "id":    "123",
        "email": "user@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Publish synchronously (wait for handlers)
    err = bus.PublishSync(context.Background(), "user.created", map[string]interface{}{
        "id":    "456",
        "email": "another@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Usage Examples

### Pattern Matching

```go
// Subscribe to all user events
bus.Subscribe("user.*", handler)

// Subscribe to all creation events
bus.Subscribe("*.created", handler)

// Subscribe to all events
bus.Subscribe("*", handler)
```

### Middleware

```go
// Logging middleware
loggingMiddleware := func(next scela.Handler) scela.Handler {
    return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
        log.Printf("Processing message: %s", msg.Topic())
        return next.Handle(ctx, msg)
    })
}

bus.Use(loggingMiddleware)
```

### Dead Letter Queue

```go
bus := scela.New(
    scela.WithMaxRetries(3),
    scela.WithDeadLetterHandler(scela.HandlerFunc(
        func(ctx context.Context, msg scela.Message) error {
            log.Printf("Message failed after retries: %s", msg.Topic())
            return nil
        },
    )),
)
```

### Message Persistence

```go
// Use file-based persistence
fileStore := scela.NewFileStore("messages.json")
persistentBus := scela.NewPersistentBus(bus, fileStore)
defer persistentBus.Close()

// Or use database persistence (SQLite, PostgreSQL, MySQL, etc.)
db, _ := sql.Open("sqlite3", "messages.db")
sqlStore, _ := scela.NewSQLStore(scela.SQLStoreConfig{
    DB:        db,
    TableName: "messages",
})
persistentBus = scela.NewPersistentBus(bus, sqlStore)

// Messages are automatically persisted
persistentBus.Publish(ctx, "orders.created", order)

// Replay persisted messages (e.g., after restart)
persistentBus.Replay(ctx)

// Query specific messages
messages, _ := sqlStore.LoadByTopic(ctx, "orders.created")
recent, _ := sqlStore.LoadAfter(ctx, time.Now().Add(-1*time.Hour))
```

### Audit Trail

```go
// Create audit history
history := scela.NewMessageHistory(1000)
auditBus := scela.NewAuditableBus(bus, history)

// Use history middleware for detailed tracking
auditBus.Subscribe("*", scela.HistoryMiddleware(history)(handler))

// Query audit trail
published := history.GetByEvent("published")
failed := history.GetByEvent("failed")
orderEvents := history.GetByTopic("orders.created")
recent := history.GetInTimeRange(yesterday, now)

fmt.Printf("Total events tracked: %d\n", history.Count())
```

### Observability

```go
type MetricsObserver struct {
    published int64
}

func (m *MetricsObserver) OnPublish(ctx context.Context, topic string, msg scela.Message) {
    atomic.AddInt64(&m.published, 1)
}

// Implement other Observer methods...

bus := scela.New(scela.WithObserver(&MetricsObserver{}))
```

## Celtic Name

**Scéla** (pronounced "SHKAY-la") comes from Old Irish, meaning "news, tidings, or messages." It's the perfect name for a message bus that carries information between parts of your application, just as ancient Irish messengers carried scéla between tribes.

## Part of Toutā Framework

Scéla is part of the **Toutā Framework** ecosystem:

- **[toutago](https://github.com/toutaio/toutago)** - Core framework
- **[toutago-cosan-router](https://github.com/toutaio/toutago-cosan-router)** - HTTP router
- **[toutago-fith-renderer](https://github.com/toutaio/toutago-fith-renderer)** - Template engine
- **[toutago-nasc-dependency-injector](https://github.com/toutaio/toutago-nasc-dependency-injector)** - DI container
- **[toutago-sil-migrator](https://github.com/toutaio/toutago-sil-migrator)** - Database migrations
- **[toutago-datamapper](https://github.com/toutaio/toutago-datamapper)** - Data mapping

Each component works standalone or together as a cohesive ecosystem.

## Documentation

- [Usage Guide](./docs/usage.md) - Complete guide with examples and best practices
- [Architecture](./docs/architecture.md) - Design decisions and internals
- [Migration Guide](./docs/migration.md) - Migrate from internal message bus
- [API Reference](https://pkg.go.dev/github.com/toutaio/toutago-scela-bus) - Full API documentation
- [Examples](./examples) - Working code examples

## Performance

Scéla is designed for high-throughput, low-latency messaging:

```
BenchmarkPublishSync-8      500000    2500 ns/op    400 B/op    5 allocs/op
BenchmarkPublishAsync-8    2000000     800 ns/op    200 B/op    3 allocs/op
BenchmarkPatternMatch-8   10000000     150 ns/op      0 B/op    0 allocs/op
```

- **Async publish**: ~800 nanoseconds per operation
- **Sync publish**: ~2.5 microseconds per operation
- **Pattern matching**: ~150 nanoseconds per match

Run benchmarks: `go test -bench=. ./pkg/scela`

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Support

- 📖 [Documentation](https://pkg.go.dev/github.com/toutaio/toutago-scela-bus)
- 💬 [Discussions](https://github.com/toutaio/toutago-scela-bus/discussions)
- 🐛 [Issues](https://github.com/toutaio/toutago-scela-bus/issues)

---

Made with ❤️ by the Toutā Framework team
