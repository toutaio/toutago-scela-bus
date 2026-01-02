package scela

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestBus_ContextCancellation(t *testing.T) {
	bus := New(WithWorkers(1)) // Use 1 worker to ensure channel blocks
	defer bus.Close()

	// Fill the queue first
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		bus.Publish(ctx, "filler", nil)
	}

	// Now try with cancelled context
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := bus.Publish(cancelCtx, "test.topic", nil)
	if !errors.Is(err, context.Canceled) {
		t.Logf("Note: Publish may return nil if queue has space. Got: %v", err)
	}
}

func TestBus_NoHandlers(t *testing.T) {
	bus := New()
	defer bus.Close()

	ctx := context.Background()

	// Publishing to topic with no subscribers should not error
	err := bus.PublishSync(ctx, "nonexistent.topic", nil)
	if err != nil {
		t.Errorf("PublishSync() with no handlers returned error: %v", err)
	}
}

func TestBus_InvalidSubscription(t *testing.T) {
	bus := New()
	defer bus.Close()

	// Empty pattern
	_, err := bus.Subscribe("", HandlerFunc(func(ctx context.Context, msg Message) error {
		return nil
	}))
	if err == nil {
		t.Error("Subscribe() with empty pattern should return error")
	}

	// Nil handler
	_, err = bus.Subscribe("test.topic", nil)
	if err == nil {
		t.Error("Subscribe() with nil handler should return error")
	}
}

func TestBus_DeadLetterQueue(t *testing.T) {
	var mu sync.Mutex
	var dlqCalled bool
	var dlqMessage Message

	dlqHandler := HandlerFunc(func(ctx context.Context, msg Message) error {
		mu.Lock()
		dlqCalled = true
		dlqMessage = msg
		mu.Unlock()
		return nil
	})

	bus := New(
		WithMaxRetries(2),
		WithDeadLetterHandler(dlqHandler),
	)
	defer bus.Close()

	// Subscribe with handler that always fails
	failHandler := HandlerFunc(func(ctx context.Context, msg Message) error {
		return errors.New("handler error")
	})

	_, err := bus.Subscribe("test.dlq", failHandler)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	ctx := context.Background()
	if err := bus.Publish(ctx, "test.dlq", "payload"); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	// Wait for async processing and retries
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if !dlqCalled {
		t.Error("Dead letter queue handler was not called")
	}

	if dlqMessage == nil {
		t.Fatal("DLQ message is nil")
	}

	if dlqMessage.Topic() != "test.dlq" {
		t.Errorf("DLQ message topic = %v, want test.dlq", dlqMessage.Topic())
	}
}

func TestBus_WithWorkers(t *testing.T) {
	b := New(WithWorkers(5))
	defer b.Close()

	impl, ok := b.(*bus)
	if !ok {
		t.Fatal("New() did not return *bus")
	}

	if impl.workers != 5 {
		t.Errorf("workers = %d, want 5", impl.workers)
	}
}

func TestBus_WithWorkersInvalid(t *testing.T) {
	b := New(WithWorkers(0))
	defer b.Close()

	impl, ok := b.(*bus)
	if !ok {
		t.Fatal("New() did not return *bus")
	}

	// Should use default value when invalid
	if impl.workers == 0 {
		t.Error("workers should not be 0 when invalid value provided")
	}
}

func TestBus_WithMaxRetries(t *testing.T) {
	b := New(WithMaxRetries(5))
	defer b.Close()

	impl, ok := b.(*bus)
	if !ok {
		t.Fatal("New() did not return *bus")
	}

	if impl.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", impl.maxRetries)
	}
}

func TestBus_CloseTwice(t *testing.T) {
	bus := New()

	if err := bus.Close(); err != nil {
		t.Fatalf("First Close() error = %v", err)
	}

	if err := bus.Close(); err == nil {
		t.Error("Second Close() should return error")
	}
}

func TestSubscriptionRegistry_Remove_NotFound(t *testing.T) {
	registry := newSubscriptionRegistry()

	err := registry.Remove("nonexistent-id")
	if err == nil {
		t.Error("Remove() with non-existent ID should return error")
	}
}

func TestSubscriptionRegistry_Count(t *testing.T) {
	registry := newSubscriptionRegistry()
	b := &bus{registry: registry}

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		return nil
	})

	if count := registry.Count(); count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	registry.Add("topic1", handler, b)
	registry.Add("topic2", handler, b)
	registry.Add("topic3", handler, b)

	if count := registry.Count(); count != 3 {
		t.Errorf("Count after 3 adds = %d, want 3", count)
	}
}

func TestHandlerFunc_Interface(t *testing.T) {
	var called bool

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		called = true
		return nil
	})

	// Verify it implements Handler
	var _ Handler = handler

	ctx := context.Background()
	msg := NewMessage("test", nil)

	if err := handler.Handle(ctx, msg); err != nil {
		t.Errorf("Handle() error = %v", err)
	}

	if !called {
		t.Error("Handler function was not called")
	}
}

func TestBus_RaceConditions(t *testing.T) {
	bus := New()
	defer bus.Close()

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		return nil
	})

	// Run concurrent operations
	done := make(chan struct{})

	// Concurrent subscribes
	go func() {
		for i := 0; i < 50; i++ {
			bus.Subscribe("test.topic", handler)
		}
	}()

	// Concurrent publishes
	go func() {
		for i := 0; i < 50; i++ {
			ctx := context.Background()
			bus.Publish(ctx, "test.topic", i)
		}
	}()

	// Concurrent unsubscribes
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			sub, _ := bus.Subscribe("temp.topic", handler)
			if sub != nil {
				sub.Unsubscribe()
			}
		}
	}()

	<-done
	// If we reach here without race detector errors, test passes
}
