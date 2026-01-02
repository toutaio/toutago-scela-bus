package scela

import (
	"context"
	"fmt"
	"sync"
)

// bus is the default implementation of the Bus interface.
type bus struct {
	registry   *subscriptionRegistry
	middleware []Middleware
	workers    int
	queue      chan *envelope
	wg         sync.WaitGroup
	mu         sync.RWMutex
	closed     bool
	maxRetries int
	dlqHandler Handler
	observers  *observerRegistry
}

// envelope wraps a message for internal processing.
type envelope struct {
	msg      Message
	retries  int
	priority Priority
}

// Option is a functional option for configuring the bus.
type Option func(*bus)

// WithWorkers sets the number of worker goroutines for async processing.
func WithWorkers(n int) Option {
	return func(b *bus) {
		if n > 0 {
			b.workers = n
		}
	}
}

// WithMaxRetries sets the maximum number of retries for failed messages.
func WithMaxRetries(n int) Option {
	return func(b *bus) {
		if n >= 0 {
			b.maxRetries = n
		}
	}
}

// WithDeadLetterHandler sets a handler for messages that exceed max retries.
func WithDeadLetterHandler(handler Handler) Option {
	return func(b *bus) {
		b.dlqHandler = handler
	}
}

// New creates a new message bus with the given options.
func New(opts ...Option) Bus {
	b := &bus{
		registry:   newSubscriptionRegistry(),
		middleware: make([]Middleware, 0),
		workers:    10,                         // Default number of workers
		queue:      make(chan *envelope, 1000), // Buffered channel
		maxRetries: 3,
		observers:  newObserverRegistry(),
	}

	// Apply options
	for _, opt := range opts {
		opt(b)
	}

	// Start worker pool
	for i := 0; i < b.workers; i++ {
		b.wg.Add(1)
		go b.worker()
	}

	return b
}

// worker processes messages from the queue.
func (b *bus) worker() {
	defer b.wg.Done()

	for env := range b.queue {
		b.processMessage(env)
	}
}

// processMessage processes a single message envelope.
func (b *bus) processMessage(env *envelope) {
	ctx := context.Background()

	handlers := b.registry.GetHandlers(env.msg.Topic())
	if len(handlers) == 0 {
		return
	}

	// Apply middleware
	finalHandler := b.wrapWithMiddleware(HandlerFunc(func(ctx context.Context, msg Message) error {
		// Execute all matching handlers
		var lastErr error
		for _, h := range handlers {
			if err := h.Handle(ctx, msg); err != nil {
				lastErr = err
			}
		}
		return lastErr
	}))

	// Handle the message
	err := finalHandler.Handle(ctx, env.msg)

	// Notify observers
	b.observers.NotifyMessageProcessed(ctx, env.msg, err)

	if err != nil {
		b.handleError(env)
	}
}

// handleError handles a message processing error with retry logic.
func (b *bus) handleError(env *envelope) {
	env.retries++

	if env.retries < b.maxRetries {
		// Retry the message
		b.queue <- env
		return
	}

	// Max retries exceeded, send to DLQ
	if b.dlqHandler != nil {
		ctx := context.Background()
		_ = b.dlqHandler.Handle(ctx, env.msg)
	}
}

// Publish publishes a message asynchronously.
func (b *bus) Publish(ctx context.Context, topic string, payload interface{}) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("bus is closed")
	}

	msg := NewMessage(topic, payload)

	// Notify observers
	b.observers.NotifyPublish(ctx, topic, msg)

	env := &envelope{
		msg:      msg,
		priority: PriorityNormal,
	}

	select {
	case b.queue <- env:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PublishSync publishes a message synchronously, waiting for all handlers to complete.
func (b *bus) PublishSync(ctx context.Context, topic string, payload interface{}) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("bus is closed")
	}

	msg := NewMessage(topic, payload)

	// Notify observers
	b.observers.NotifyPublish(ctx, topic, msg)

	handlers := b.registry.GetHandlers(topic)

	if len(handlers) == 0 {
		return nil
	}

	// Apply middleware
	finalHandler := b.wrapWithMiddleware(HandlerFunc(func(ctx context.Context, msg Message) error {
		// Execute all matching handlers synchronously
		var lastErr error
		for _, h := range handlers {
			if err := h.Handle(ctx, msg); err != nil {
				lastErr = err
			}
		}
		return lastErr
	}))

	err := finalHandler.Handle(ctx, msg)

	// Notify observers
	b.observers.NotifyMessageProcessed(ctx, msg, err)

	return err
}

// PublishWithPriority publishes a message asynchronously with the specified priority.
func (b *bus) PublishWithPriority(ctx context.Context, topic string, payload interface{}, priority Priority) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("bus is closed")
	}

	// Check context before proceeding
	if err := ctx.Err(); err != nil {
		return err
	}

	msg := NewMessage(topic, payload)

	// Notify observers
	b.observers.NotifyPublish(ctx, topic, msg)

	env := &envelope{
		msg:      msg,
		priority: priority,
	}

	select {
	case b.queue <- env:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Subscribe subscribes a handler to a topic pattern.
func (b *bus) Subscribe(pattern string, handler Handler) (Subscription, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return nil, fmt.Errorf("bus is closed")
	}

	sub, err := b.registry.Add(pattern, handler, b)
	if err == nil {
		b.observers.NotifySubscribe(pattern)
	}
	return sub, err
}

// unsubscribe removes a subscription by ID.
func (b *bus) unsubscribe(id string) error {
	// Get pattern before removing
	b.registry.mu.RLock()
	sub, exists := b.registry.subscriptions[id]
	b.registry.mu.RUnlock()

	err := b.registry.Remove(id)
	if err == nil && exists {
		b.observers.NotifyUnsubscribe(sub.pattern)
	}
	return err
}

// Use adds middleware to the bus.
func (b *bus) Use(middleware ...Middleware) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.middleware = append(b.middleware, middleware...)
}

// wrapWithMiddleware wraps a handler with all registered middleware.
func (b *bus) wrapWithMiddleware(handler Handler) Handler {
	// Apply middleware in reverse order so they execute in registration order
	for i := len(b.middleware) - 1; i >= 0; i-- {
		handler = b.middleware[i](handler)
	}
	return handler
}

// Close gracefully shuts down the bus.
func (b *bus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return fmt.Errorf("bus already closed")
	}
	b.closed = true
	b.mu.Unlock()

	// Close the queue to signal workers to stop
	close(b.queue)

	// Wait for all workers to finish
	b.wg.Wait()

	// Clear all subscriptions
	b.registry.Clear()

	// Notify observers
	b.observers.NotifyClose()

	return nil
}
