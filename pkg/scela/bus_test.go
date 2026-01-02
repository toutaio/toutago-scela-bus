package scela

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBus_PublishSubscribe(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received int32
	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&received, 1)
		return nil
	})

	_, err := bus.Subscribe("user.created", handler)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	ctx := context.Background()
	if err := bus.PublishSync(ctx, "user.created", "test"); err != nil {
		t.Fatalf("PublishSync() error = %v", err)
	}

	if got := atomic.LoadInt32(&received); got != 1 {
		t.Errorf("Expected 1 message received, got %d", got)
	}
}

func TestBus_PatternSubscription(t *testing.T) {
	bus := New()
	defer bus.Close()

	var mu sync.Mutex
	messages := make([]string, 0)

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		mu.Lock()
		messages = append(messages, msg.Topic())
		mu.Unlock()
		return nil
	})

	_, err := bus.Subscribe("user.*", handler)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	ctx := context.Background()
	topics := []string{"user.created", "user.updated", "user.deleted"}

	for _, topic := range topics {
		if err := bus.PublishSync(ctx, topic, nil); err != nil {
			t.Fatalf("PublishSync(%s) error = %v", topic, err)
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}
}

func TestBus_MultipleSubscribers(t *testing.T) {
	bus := New()
	defer bus.Close()

	var count int32

	handler1 := HandlerFunc(func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	handler2 := HandlerFunc(func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&count, 10)
		return nil
	})

	_, err := bus.Subscribe("test.topic", handler1)
	if err != nil {
		t.Fatalf("Subscribe(handler1) error = %v", err)
	}

	_, err = bus.Subscribe("test.topic", handler2)
	if err != nil {
		t.Fatalf("Subscribe(handler2) error = %v", err)
	}

	ctx := context.Background()
	if err := bus.PublishSync(ctx, "test.topic", nil); err != nil {
		t.Fatalf("PublishSync() error = %v", err)
	}

	if got := atomic.LoadInt32(&count); got != 11 {
		t.Errorf("Expected count = 11, got %d", got)
	}
}

func TestBus_Unsubscribe(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received int32
	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&received, 1)
		return nil
	})

	sub, err := bus.Subscribe("test.topic", handler)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	ctx := context.Background()

	// Publish before unsubscribe
	if err := bus.PublishSync(ctx, "test.topic", nil); err != nil {
		t.Fatalf("PublishSync() error = %v", err)
	}

	if got := atomic.LoadInt32(&received); got != 1 {
		t.Errorf("Expected 1 message before unsubscribe, got %d", got)
	}

	// Unsubscribe
	if err := sub.Unsubscribe(); err != nil {
		t.Fatalf("Unsubscribe() error = %v", err)
	}

	// Publish after unsubscribe
	if err := bus.PublishSync(ctx, "test.topic", nil); err != nil {
		t.Fatalf("PublishSync() error = %v", err)
	}

	// Should still be 1 (not 2)
	if got := atomic.LoadInt32(&received); got != 1 {
		t.Errorf("Expected 1 message after unsubscribe, got %d", got)
	}
}

func TestBus_Async(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received int32
	done := make(chan struct{})

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&received, 1)
		close(done)
		return nil
	})

	_, err := bus.Subscribe("test.async", handler)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	ctx := context.Background()
	if err := bus.Publish(ctx, "test.async", "payload"); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	// Wait for async processing
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for async message")
	}

	if got := atomic.LoadInt32(&received); got != 1 {
		t.Errorf("Expected 1 message received, got %d", got)
	}
}

func TestBus_Middleware(t *testing.T) {
	bus := New()
	defer bus.Close()

	var mu sync.Mutex
	order := make([]string, 0)

	middleware1 := func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg Message) error {
			mu.Lock()
			order = append(order, "middleware1-before")
			mu.Unlock()
			err := next.Handle(ctx, msg)
			mu.Lock()
			order = append(order, "middleware1-after")
			mu.Unlock()
			return err
		})
	}

	middleware2 := func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg Message) error {
			mu.Lock()
			order = append(order, "middleware2-before")
			mu.Unlock()
			err := next.Handle(ctx, msg)
			mu.Lock()
			order = append(order, "middleware2-after")
			mu.Unlock()
			return err
		})
	}

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		mu.Lock()
		order = append(order, "handler")
		mu.Unlock()
		return nil
	})

	bus.Use(middleware1, middleware2)

	_, err := bus.Subscribe("test.middleware", handler)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	ctx := context.Background()
	if err := bus.PublishSync(ctx, "test.middleware", nil); err != nil {
		t.Fatalf("PublishSync() error = %v", err)
	}

	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}

	mu.Lock()
	defer mu.Unlock()

	if len(order) != len(expected) {
		t.Fatalf("Expected %d execution steps, got %d", len(expected), len(order))
	}

	for i, exp := range expected {
		if order[i] != exp {
			t.Errorf("Step %d: expected %q, got %q", i, exp, order[i])
		}
	}
}

func TestBus_Close(t *testing.T) {
	bus := New()

	ctx := context.Background()
	_, err := bus.Subscribe("test.topic", HandlerFunc(func(ctx context.Context, msg Message) error {
		return nil
	}))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	if err := bus.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Operations after close should fail
	if err := bus.Publish(ctx, "test.topic", nil); err == nil {
		t.Error("Expected error when publishing to closed bus, got nil")
	}

	if _, err := bus.Subscribe("test.topic", nil); err == nil {
		t.Error("Expected error when subscribing to closed bus, got nil")
	}
}

func TestBus_ConcurrentPublish(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received int32
	var wg sync.WaitGroup

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&received, 1)
		return nil
	})

	_, err := bus.Subscribe("test.concurrent", handler)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	// Publish 100 messages concurrently
	numMessages := 100
	wg.Add(numMessages)

	for i := 0; i < numMessages; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			if err := bus.PublishSync(ctx, "test.concurrent", nil); err != nil {
				t.Errorf("PublishSync() error = %v", err)
			}
		}()
	}

	wg.Wait()

	if got := atomic.LoadInt32(&received); got != int32(numMessages) {
		t.Errorf("Expected %d messages received, got %d", numMessages, got)
	}
}

func BenchmarkBus_PublishSync(b *testing.B) {
	bus := New()
	defer bus.Close()

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		return nil
	})

	bus.Subscribe("bench.topic", handler)

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.PublishSync(ctx, "bench.topic", "payload")
	}
}

func BenchmarkBus_PublishAsync(b *testing.B) {
	bus := New()
	defer bus.Close()

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		return nil
	})

	bus.Subscribe("bench.topic", handler)

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.Publish(ctx, "bench.topic", "payload")
	}
}
