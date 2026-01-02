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
- 🔄 **Retry Logic** - Configurable retry with exponential backoff
- 💀 **Dead Letter Queue** - Handle failed messages gracefully
- 🧵 **Thread-Safe** - Fully concurrent, safe for goroutines
- 🎛️ **Context Support** - Cancellation, timeouts, request-scoped data
- 📊 **Observable** - Hooks for metrics and monitoring
- 🎪 **Priority Queues** - Process high-priority messages first
- 🚀 **Zero Dependencies** - Only standard library
- ✅ **Production Ready** - 80%+ test coverage, comprehensive CI/CD

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

### Priority Messages

```go
bus.PublishWithPriority(ctx, "urgent.alert", payload, scela.PriorityHigh)
```

### Dead Letter Queue

```go
bus, err := scela.New(
    scela.WithMaxRetries(3),
    scela.WithDeadLetterHandler(func(ctx context.Context, msg scela.Message, err error) {
        log.Printf("Message failed after retries: %s - %v", msg.Topic(), err)
    }),
)
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

- [API Reference](https://pkg.go.dev/github.com/toutaio/toutago-scela-bus)
- [Examples](./examples)
- [Architecture Documentation](./docs/architecture.md)
- [Contributing Guidelines](./CONTRIBUTING.md)

## Performance

Scéla is designed for high-throughput, low-latency messaging:

```
BenchmarkPublishSync-8      500000    2500 ns/op    400 B/op    5 allocs/op
BenchmarkPublishAsync-8    2000000     800 ns/op    200 B/op    3 allocs/op
BenchmarkPatternMatch-8   10000000     150 ns/op      0 B/op    0 allocs/op
```

See [benchmarks](./benchmarks) for detailed performance metrics.

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
