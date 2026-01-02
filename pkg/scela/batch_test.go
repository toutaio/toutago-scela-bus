package scela

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestBatch(t *testing.T) {
	batch := NewBatch()

	if batch.Size() != 0 {
		t.Errorf("Expected empty batch, got size %d", batch.Size())
	}

	msg1 := NewMessage("test1", "data1")
	msg2 := NewMessage("test2", "data2")

	batch.Add(msg1)
	batch.Add(msg2)

	if batch.Size() != 2 {
		t.Errorf("Expected batch size 2, got %d", batch.Size())
	}

	messages := batch.Clear()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if batch.Size() != 0 {
		t.Errorf("Expected empty batch after clear, got size %d", batch.Size())
	}
}

func TestBatchPublisher_Size(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received int32

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&received, 1)
		return nil
	})

	bus.Subscribe("*", handler)

	bp := NewBatchPublisher(bus, WithBatchSize(5))
	defer bp.Close()

	ctx := context.Background()

	// Publish 10 messages (should trigger 2 batches of 5)
	for i := 0; i < 10; i++ {
		bp.Publish(ctx, "test", i)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	count := atomic.LoadInt32(&received)
	if count != 10 {
		t.Errorf("Expected 10 messages processed, got %d", count)
	}
}

func TestBatchPublisher_Timeout(t *testing.T) {
	bus := New()
	defer bus.Close()

	var batchCount int32

	bp := NewBatchPublisher(bus,
		WithBatchSize(100),
		WithBatchWait(100*time.Millisecond),
		WithBatchCallback(func(messages []Message) {
			atomic.AddInt32(&batchCount, 1)
		}),
	)
	defer bp.Close()

	ctx := context.Background()

	// Publish 3 messages (not enough to trigger size-based flush)
	bp.Publish(ctx, "test", 1)
	bp.Publish(ctx, "test", 2)
	bp.Publish(ctx, "test", 3)

	// Wait for timeout
	time.Sleep(200 * time.Millisecond)

	count := atomic.LoadInt32(&batchCount)
	if count < 1 {
		t.Errorf("Expected at least 1 batch, got %d", count)
	}
}

func TestBatchPublisher_Flush(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received int32

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&received, 1)
		return nil
	})

	bus.Subscribe("*", handler)

	bp := NewBatchPublisher(bus,
		WithBatchSize(100),
		WithBatchWait(10*time.Second),
	)
	defer bp.Close()

	ctx := context.Background()

	// Publish a few messages
	bp.Publish(ctx, "test", 1)
	bp.Publish(ctx, "test", 2)
	bp.Publish(ctx, "test", 3)

	// Flush immediately
	bp.Flush(ctx)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	count := atomic.LoadInt32(&received)
	if count != 3 {
		t.Errorf("Expected 3 messages processed, got %d", count)
	}
}
