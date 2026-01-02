package scela

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPublishWithPriority(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received []string
	var mu sync.Mutex

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg.Payload().(string))
		return nil
	})

	_, err := bus.Subscribe("test", handler)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	ctx := context.Background()

	// Publish with different priorities
	bus.PublishWithPriority(ctx, "test", "low", PriorityLow)
	bus.PublishWithPriority(ctx, "test", "urgent", PriorityUrgent)
	bus.PublishWithPriority(ctx, "test", "normal", PriorityNormal)
	bus.PublishWithPriority(ctx, "test", "high", PriorityHigh)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 4 {
		t.Errorf("Expected 4 messages, got %d", len(received))
	}
}

func TestBusPublishWithPriority_Closed(t *testing.T) {
	bus := New()
	bus.Close()

	ctx := context.Background()
	err := bus.PublishWithPriority(ctx, "test", "data", PriorityHigh)
	if err == nil {
		t.Error("Expected error when publishing to closed bus")
	}
}

func TestBusPublishWithPriority_ContextCanceled(t *testing.T) {
	bus := New(WithWorkers(1))
	defer bus.Close()

	// Fill the queue
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		bus.PublishWithPriority(ctx, "test", i, PriorityNormal)
	}

	// Try to publish with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err := bus.PublishWithPriority(canceledCtx, "test", "data", PriorityHigh)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}
