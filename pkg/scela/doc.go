// Package scela provides a production-ready, in-process message bus for Go applications.
//
// Scéla (Old Irish: "news, tidings, messages") enables pub/sub messaging patterns
// with support for synchronous and asynchronous delivery, pattern matching,
// middleware, retry logic, dead letter queues, message persistence, and audit trails.
// Part of the Toutā Framework ecosystem.
//
// # Features
//
//   - Pub/Sub Pattern - Topic-based message routing with multiple subscribers
//   - Sync & Async - Choose between synchronous or asynchronous message delivery
//   - Pattern Matching - Subscribe with wildcards (user.*, *.created)
//   - Middleware Pipeline - Intercept messages for logging, metrics, validation
//   - Retry Logic - Configurable retry logic for failed messages
//   - Dead Letter Queue - Handle failed messages gracefully
//   - Thread-Safe - Fully concurrent, safe for goroutines
//   - Context Support - Cancellation, timeouts, request-scoped data
//   - Observable - Hooks for metrics and monitoring
//   - Persistence - File-based and SQL database message persistence
//   - Audit Trail - Complete message history and event tracking
//   - Zero Dependencies - Only standard library (persistence features optional)
//   - Production Ready - 87.5% test coverage, comprehensive CI/CD
//
// # Quick Start
//
//	bus := scela.New()
//	defer bus.Close()
//
//	// Subscribe to messages
//	sub, _ := bus.Subscribe("user.created", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
//	    fmt.Printf("User created: %v\n", msg.Payload())
//	    return nil
//	}))
//	defer sub.Unsubscribe()
//
//	// Publish message asynchronously
//	bus.Publish(context.Background(), "user.created", userData)
//
//	// Publish message synchronously (wait for handlers)
//	bus.PublishSync(context.Background(), "user.created", userData)
//
// # Pattern Matching
//
// Subscribe using wildcard patterns:
//
//	bus.Subscribe("user.*", handler)       // All user events
//	bus.Subscribe("*.created", handler)    // All creation events
//	bus.Subscribe("*", handler)            // All events
//
// # Middleware
//
// Add cross-cutting concerns with middleware:
//
//	loggingMiddleware := func(next scela.Handler) scela.Handler {
//	    return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
//	        log.Printf("Processing: %s", msg.Topic())
//	        return next.Handle(ctx, msg)
//	    })
//	}
//	bus.Use(loggingMiddleware)
//
// # Retry and Dead Letter Queue
//
//	bus := scela.New(
//	    scela.WithMaxRetries(3),
//	    scela.WithDeadLetterHandler(scela.HandlerFunc(
//	        func(ctx context.Context, msg scela.Message) error {
//	            log.Printf("Message failed after retries: %s", msg.Topic())
//	            return nil
//	        },
//	    )),
//	)
//
// # Message Persistence
//
// Persist messages to disk or database:
//
//	// File-based persistence
//	fileStore := scela.NewFileStore("messages.json")
//	persistentBus := scela.NewPersistentBus(bus, fileStore)
//
//	// SQL database persistence (SQLite, PostgreSQL, MySQL, etc.)
//	db, _ := sql.Open("sqlite3", "messages.db")
//	sqlStore, _ := scela.NewSQLStore(scela.SQLStoreConfig{
//	    DB:        db,
//	    TableName: "messages",
//	})
//	persistentBus = scela.NewPersistentBus(bus, sqlStore)
//
//	// Replay persisted messages after restart
//	persistentBus.Replay(ctx)
//
// # Audit Trail
//
// Track message lifecycle events:
//
//	history := scela.NewMessageHistory(1000)
//	auditBus := scela.NewAuditableBus(bus, history)
//
//	// Query audit trail
//	published := history.GetByEvent("published")
//	failed := history.GetByEvent("failed")
//	orderEvents := history.GetByTopic("orders.created")
//
// # Observability
//
// Implement the Observer interface for metrics and monitoring:
//
//	type MetricsObserver struct{}
//
//	func (m *MetricsObserver) OnPublish(ctx context.Context, topic string, msg scela.Message) {
//	    // Track published messages
//	}
//
//	bus := scela.New(scela.WithObserver(&MetricsObserver{}))
//
// # Thread Safety
//
// All operations are thread-safe and can be used concurrently from multiple goroutines.
// The message bus uses channels and mutexes internally to ensure safe concurrent access.
//
// # Version
//
// This is version 1.5.1 - production ready with 87.5% test coverage.
// Requires Go 1.22 or higher.
package scela
