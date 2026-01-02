package scela

import (
	"context"
	"time"
)

// Message represents a message in the bus.
type Message interface {
	// Topic returns the message topic.
	Topic() string

	// Payload returns the message payload.
	Payload() interface{}

	// Metadata returns message metadata.
	Metadata() map[string]interface{}

	// ID returns the unique message ID.
	ID() string

	// Timestamp returns when the message was created.
	Timestamp() time.Time
}

// Handler handles messages.
type Handler interface {
	// Handle processes a message.
	Handle(ctx context.Context, msg Message) error
}

// HandlerFunc is a function adapter for Handler interface.
type HandlerFunc func(ctx context.Context, msg Message) error

// Handle implements the Handler interface.
func (f HandlerFunc) Handle(ctx context.Context, msg Message) error {
	return f(ctx, msg)
}

// Bus is the message bus interface.
type Bus interface {
	// Publish publishes a message asynchronously.
	Publish(ctx context.Context, topic string, payload interface{}) error

	// PublishSync publishes a message synchronously, waiting for all handlers.
	PublishSync(ctx context.Context, topic string, payload interface{}) error
	
	// PublishWithPriority publishes a message asynchronously with the specified priority.
	PublishWithPriority(ctx context.Context, topic string, payload interface{}, priority Priority) error

	// Subscribe subscribes a handler to a topic pattern.
	Subscribe(pattern string, handler Handler) (Subscription, error)

	// Use adds middleware to the bus.
	Use(middleware ...Middleware)

	// Close gracefully shuts down the bus.
	Close() error
}

// Subscription represents a subscription to messages.
type Subscription interface {
	// Topic returns the subscription pattern.
	Topic() string

	// Unsubscribe removes the subscription.
	Unsubscribe() error
}

// Middleware wraps handlers for cross-cutting concerns.
type Middleware func(Handler) Handler

// Priority defines message priority levels.
type Priority int

const (
	// PriorityLow for low-priority messages.
	PriorityLow Priority = iota
	// PriorityNormal for normal-priority messages (default).
	PriorityNormal
	// PriorityHigh for high-priority messages.
	PriorityHigh
	// PriorityUrgent for urgent messages.
	PriorityUrgent
)
